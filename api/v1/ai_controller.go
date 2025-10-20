package v1

import (
	"encoding/json"
	"fmt"
	"haven_camp_server/internal/config"
	"haven_camp_server/internal/dao"
	"haven_camp_server/internal/dto/request"
	"haven_camp_server/internal/dto/respond"
	"haven_camp_server/internal/model"
	"haven_camp_server/internal/service/ai"
	"haven_camp_server/internal/service/chat"
	"haven_camp_server/internal/service/gorm"
	myredis "haven_camp_server/internal/service/redis"
	"haven_camp_server/pkg/constants"
	"haven_camp_server/pkg/enum/message/message_status_enum"
	"haven_camp_server/pkg/enum/message/message_type_enum"
	"haven_camp_server/pkg/util/random"
	"haven_camp_server/pkg/zlog"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
)

// AiChat AI对话接口
func AiChat(c *gin.Context) {
	var req request.AiChatRequest
	if err := c.BindJSON(&req); err != nil {
		zlog.Error(err.Error())
		c.JSON(http.StatusOK, gin.H{
			"code":    500,
			"message": constants.SYSTEM_ERROR,
		})
		return
	}

	conf := config.GetConfig().DifyConfig
	aiUserId := conf.AiUserId
	aiName := conf.AiName
	aiAvatar := conf.AiAvatar

	// 如果 session_id 为空，创建或获取会话
	sessionId := req.SessionId
	if sessionId == "" {
		openSessionReq := request.OpenSessionRequest{
			SendId:    req.OwnerId,
			ReceiveId: aiUserId,
		}
		message, sid, ret := gorm.SessionService.OpenSession(openSessionReq)
		if ret != 0 {
			JsonBack(c, message, ret, nil)
			return
		}
		sessionId = sid
	}

	// 调用 Dify 服务
	answer, err := ai.DifyService.Ask(req.Question, req.OwnerId, sessionId, req.Meta)
	if err != nil {
		zlog.Error("调用 Dify 失败: " + err.Error())
		JsonBack(c, "AI服务调用失败", -1, nil)
		return
	}

	// 构造 AI 回复消息并存库
	aiMessage := model.Message{
		Uuid:       fmt.Sprintf("M%s", random.GetNowAndLenRandomString(11)),
		SessionId:  sessionId,
		Type:       message_type_enum.Text,
		Content:    answer,
		Url:        "",
		SendId:     aiUserId,
		SendName:   aiName,
		SendAvatar: aiAvatar,
		ReceiveId:  req.OwnerId,
		FileSize:   "0B",
		FileType:   "",
		FileName:   "",
		Status:     message_status_enum.Unsent,
		CreatedAt:  time.Now(),
		AVdata:     "",
	}

	if res := dao.GormDB.Create(&aiMessage); res.Error != nil {
		zlog.Error("AI消息存库失败: " + res.Error.Error())
		JsonBack(c, constants.SYSTEM_ERROR, -1, nil)
		return
	}

	// 通过 WebSocket 推送消息给用户
	messageRsp := respond.GetMessageListRespond{
		SendId:     aiMessage.SendId,
		SendName:   aiMessage.SendName,
		SendAvatar: aiMessage.SendAvatar,
		ReceiveId:  aiMessage.ReceiveId,
		Type:       aiMessage.Type,
		Content:    aiMessage.Content,
		Url:        aiMessage.Url,
		FileSize:   aiMessage.FileSize,
		FileName:   aiMessage.FileName,
		FileType:   aiMessage.FileType,
		CreatedAt:  aiMessage.CreatedAt.Format("2006-01-02 15:04:05"),
	}

	jsonMessage, err := json.Marshal(messageRsp)
	if err != nil {
		zlog.Error(err.Error())
	}

	messageBack := &chat.MessageBack{
		Message: jsonMessage,
		Uuid:    aiMessage.Uuid,
	}

	// 根据消息模式推送
	kafkaConfig := config.GetConfig().KafkaConfig
	if kafkaConfig.MessageMode == "channel" {
		// Channel 模式
		if receiveClient, ok := chat.ChatServer.Clients[aiMessage.ReceiveId]; ok {
			receiveClient.SendBack <- messageBack
		}
	} else {
		// Kafka 模式
		if receiveClient, ok := chat.KafkaChatServer.Clients[aiMessage.ReceiveId]; ok {
			receiveClient.SendBack <- messageBack
		}
	}

	// 更新消息状态为已发送
	if res := dao.GormDB.Model(&model.Message{}).Where("uuid = ?", aiMessage.Uuid).Update("status", message_status_enum.Sent); res.Error != nil {
		zlog.Error(res.Error.Error())
	}

	// 更新 Redis 缓存
	updateRedisMessageCache(aiMessage.SendId, aiMessage.ReceiveId, messageRsp)

	// 返回响应
	rsp := respond.AiChatRespond{
		SessionId: sessionId,
		Answer:    answer,
	}
	JsonBack(c, "AI对话成功", 0, rsp)
}

// updateRedisMessageCache 更新 Redis 中的消息缓存
func updateRedisMessageCache(sendId, receiveId string, messageRsp respond.GetMessageListRespond) {
	// 正向缓存
	rspString, err := myredis.GetKeyNilIsErr("message_list_" + sendId + "_" + receiveId)
	if err == nil {
		var rsp []respond.GetMessageListRespond
		if err := json.Unmarshal([]byte(rspString), &rsp); err != nil {
			zlog.Error(err.Error())
		}
		rsp = append(rsp, messageRsp)
		rspString2, err := json.Marshal(rsp)
		if err != nil {
			zlog.Error(err.Error())
		}
		if err := myredis.SetKeyEx("message_list_"+sendId+"_"+receiveId, string(rspString2), time.Minute*constants.REDIS_TIMEOUT); err != nil {
			zlog.Error(err.Error())
		}
	} else if err == redis.Nil {
		// 缓存不存在时不创建，让用户下次查询时从数据库加载
		zlog.Info("Redis 缓存不存在，跳过更新")
	}

	// 反向缓存
	rspString, err = myredis.GetKeyNilIsErr("message_list_" + receiveId + "_" + sendId)
	if err == nil {
		var rsp []respond.GetMessageListRespond
		if err := json.Unmarshal([]byte(rspString), &rsp); err != nil {
			zlog.Error(err.Error())
		}
		rsp = append(rsp, messageRsp)
		rspString2, err := json.Marshal(rsp)
		if err != nil {
			zlog.Error(err.Error())
		}
		if err := myredis.SetKeyEx("message_list_"+receiveId+"_"+sendId, string(rspString2), time.Minute*constants.REDIS_TIMEOUT); err != nil {
			zlog.Error(err.Error())
		}
	}
}
