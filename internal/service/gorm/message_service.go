package gorm

import (
	"encoding/json"
	"errors"
	"fmt"
	"haven_camp_server/internal/config"
	"haven_camp_server/internal/dao"
	"haven_camp_server/internal/dto/respond"
	"haven_camp_server/internal/model"
	myredis "haven_camp_server/internal/service/redis"
	"haven_camp_server/pkg/constants"
	"haven_camp_server/pkg/zlog"
	"io"
	"os"
	"path/filepath"

	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
)

type messageService struct {
}

var MessageService = new(messageService)

// GetMessageList 获取聊天记录
func (m *messageService) GetMessageList(userOneId, userTwoId string) (string, []respond.GetMessageListRespond, int) {
	rspString, err := myredis.GetKeyNilIsErr("message_list_" + userOneId + "_" + userTwoId)
	if err != nil {
		if errors.Is(err, redis.Nil) {
			zlog.Info(err.Error())
			zlog.Info(fmt.Sprintf("%s %s", userTwoId, userTwoId))
			var messageList []model.Message
			if res := dao.GormDB.Where("(send_id = ? AND receive_id = ?) OR (send_id = ? AND receive_id = ?)", userOneId, userTwoId, userTwoId, userOneId).Order("created_at ASC").Find(&messageList); res.Error != nil {
				zlog.Error(res.Error.Error())
				return constants.SYSTEM_ERROR, nil, -1
			}
			var rspList []respond.GetMessageListRespond
			for _, message := range messageList {
				rspList = append(rspList, respond.GetMessageListRespond{
					SendId:     message.SendId,
					SendName:   message.SendName,
					SendAvatar: message.SendAvatar,
					ReceiveId:  message.ReceiveId,
					Content:    message.Content,
					Url:        message.Url,
					Type:       message.Type,
					FileType:   message.FileType,
					FileName:   message.FileName,
					FileSize:   message.FileSize,
					CreatedAt:  message.CreatedAt.Format("2006-01-02 15:04:05"),
				})
			}
			//rspString, err := json.Marshal(rspList)
			//if err != nil {
			//	zlog.Error(err.Error())
			//}
			//if err := myredis.SetKeyEx("message_list_"+userOneId+"_"+userTwoId, string(rspString), time.Minute*constants.REDIS_TIMEOUT); err != nil {
			//	zlog.Error(err.Error())
			//}
			return "获取聊天记录成功", rspList, 0
		} else {
			zlog.Error(err.Error())
			return constants.SYSTEM_ERROR, nil, -1
		}
	}
	var rsp []respond.GetMessageListRespond
	if err := json.Unmarshal([]byte(rspString), &rsp); err != nil {
		zlog.Error(err.Error())
	}
	return "获取群聊记录成功", rsp, 0
}

// GetGroupMessageList 获取群聊消息记录
func (m *messageService) GetGroupMessageList(groupId string) (string, []respond.GetGroupMessageListRespond, int) {
	rspString, err := myredis.GetKeyNilIsErr("group_messagelist_" + groupId)
	if err != nil {
		if errors.Is(err, redis.Nil) {
			var messageList []model.Message
			if res := dao.GormDB.Where("receive_id = ?", groupId).Order("created_at ASC").Find(&messageList); res.Error != nil {
				zlog.Error(res.Error.Error())
				return constants.SYSTEM_ERROR, nil, -1
			}
			var rspList []respond.GetGroupMessageListRespond
			for _, message := range messageList {
				rsp := respond.GetGroupMessageListRespond{
					SendId:     message.SendId,
					SendName:   message.SendName,
					SendAvatar: message.SendAvatar,
					ReceiveId:  message.ReceiveId,
					Content:    message.Content,
					Url:        message.Url,
					Type:       message.Type,
					FileType:   message.FileType,
					FileName:   message.FileName,
					FileSize:   message.FileSize,
					CreatedAt:  message.CreatedAt.Format("2006-01-02 15:04:05"),
				}
				rspList = append(rspList, rsp)
			}
			//rspString, err := json.Marshal(rspList)
			//if err != nil {
			//	zlog.Error(err.Error())
			//}
			//if err := myredis.SetKeyEx("group_messagelist_"+groupId, string(rspString), time.Minute*constants.REDIS_TIMEOUT); err != nil {
			//	zlog.Error(err.Error())
			//}
			return "获取聊天记录成功", rspList, 0
		} else {
			zlog.Error(err.Error())
			return constants.SYSTEM_ERROR, nil, -1
		}
	}
	var rsp []respond.GetGroupMessageListRespond
	if err := json.Unmarshal([]byte(rspString), &rsp); err != nil {
		zlog.Error(err.Error())
	}
	return "获取聊天记录成功", rsp, 0
}

// UploadAvatar 上传头像
func (m *messageService) UploadAvatar(c *gin.Context) (string, int) {
	// 解析表单数据，设置最大文件大小限制
	if err := c.Request.ParseMultipartForm(constants.FILE_MAX_SIZE); err != nil {
		zlog.Error(err.Error())
		return constants.SYSTEM_ERROR, -1
	}

	// 获取表单中的所有文件
	mForm := c.Request.MultipartForm
	for key, _ := range mForm.File {
		// 获取单个文件
		file, fileHeader, err := c.Request.FormFile(key)
		if err != nil {
			zlog.Error(err.Error())
			return constants.SYSTEM_ERROR, -1
		}
		defer file.Close() // 确保文件使用完毕后关闭

		// 记录文件信息
		zlog.Info(fmt.Sprintf("文件名：%s，文件大小：%d", fileHeader.Filename, fileHeader.Size))

		// 提取文件扩展名
		ext := filepath.Ext(fileHeader.Filename)
		zlog.Info(ext)

		// 构建本地文件路径，使用原始文件名
		localFileName := config.GetConfig().StaticAvatarPath + "/" + fileHeader.Filename

		// 创建本地文件
		out, err := os.Create(localFileName)
		if err != nil {
			zlog.Error(err.Error())
			return constants.SYSTEM_ERROR, -1
		}
		defer out.Close() // 确保文件使用完毕后关闭

		// 将上传的文件内容复制到本地文件
		if _, err := io.Copy(out, file); err != nil {
			zlog.Error(err.Error())
			return constants.SYSTEM_ERROR, -1
		}

		zlog.Info("完成文件上传")
	}

	return "上传成功", 0
}

// UploadFile 上传文件
func (m *messageService) UploadFile(c *gin.Context) (string, int) {
	// 解析表单数据，设置最大文件大小限制
	if err := c.Request.ParseMultipartForm(constants.FILE_MAX_SIZE); err != nil {
		zlog.Error(err.Error())
		return constants.SYSTEM_ERROR, -1
	}

	// 获取表单中的所有文件
	mForm := c.Request.MultipartForm
	for key, _ := range mForm.File {
		// 获取单个文件
		file, fileHeader, err := c.Request.FormFile(key)
		if err != nil {
			zlog.Error(err.Error())
			return constants.SYSTEM_ERROR, -1
		}
		defer file.Close() // 确保文件使用完毕后关闭

		// 记录文件信息
		zlog.Info(fmt.Sprintf("文件名：%s，文件大小：%d", fileHeader.Filename, fileHeader.Size))

		// 提取文件扩展名（用于后续处理，如重命名或格式验证）
		ext := filepath.Ext(fileHeader.Filename)
		zlog.Info(ext)

		// 构建本地文件路径，使用原始文件名
		localFileName := config.GetConfig().StaticFilePath + "/" + fileHeader.Filename

		// 创建本地文件
		out, err := os.Create(localFileName)
		if err != nil {
			zlog.Error(err.Error())
			return constants.SYSTEM_ERROR, -1
		}
		defer out.Close() // 确保文件使用完毕后关闭

		// 将上传的文件内容复制到本地文件
		if _, err := io.Copy(out, file); err != nil {
			zlog.Error(err.Error())
			return constants.SYSTEM_ERROR, -1
		}

		zlog.Info("完成文件上传")
	}

	return "上传成功", 0
}

// UploadVoice 上传语音文件
func (m *messageService) UploadVoice(c *gin.Context) (string, int) {
	// 解析表单数据，设置最大文件大小限制
	if err := c.Request.ParseMultipartForm(constants.FILE_MAX_SIZE); err != nil {
		zlog.Error(err.Error())
		return constants.SYSTEM_ERROR, -1
	}

	// 获取表单中的所有文件
	mForm := c.Request.MultipartForm
	for key, _ := range mForm.File {
		// 获取单个文件
		file, fileHeader, err := c.Request.FormFile(key)
		if err != nil {
			zlog.Error(err.Error())
			return constants.SYSTEM_ERROR, -1
		}
		defer file.Close() // 确保文件使用完毕后关闭

		// 记录文件信息
		zlog.Info(fmt.Sprintf("语音文件名：%s，文件大小：%d", fileHeader.Filename, fileHeader.Size))

		// 提取文件扩展名
		ext := filepath.Ext(fileHeader.Filename)
		zlog.Info(ext)

		// 构建本地文件路径，使用原始文件名
		localFileName := config.GetConfig().StaticVoicePath + "/" + fileHeader.Filename

		// 创建本地文件
		out, err := os.Create(localFileName)
		if err != nil {
			zlog.Error(err.Error())
			return constants.SYSTEM_ERROR, -1
		}
		defer out.Close() // 确保文件使用完毕后关闭

		// 将上传的文件内容复制到本地文件
		if _, err := io.Copy(out, file); err != nil {
			zlog.Error(err.Error())
			return constants.SYSTEM_ERROR, -1
		}

		zlog.Info("完成语音文件上传")
	}

	return "语音上传成功", 0
}
