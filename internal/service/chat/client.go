package chat

import (
	"context"
	"encoding/json"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/segmentio/kafka-go"
	"haven_camp_server/internal/config"
	"haven_camp_server/internal/dao"
	"haven_camp_server/internal/dto/request"
	"haven_camp_server/internal/model"
	myKafka "haven_camp_server/internal/service/kafka"
	"haven_camp_server/pkg/constants"
	"haven_camp_server/pkg/enum/message/message_status_enum"
	"haven_camp_server/pkg/zlog"
	"log"
	"net/http"
	"strconv"
)

// MessageBack 表示需要返回给前端的消息及其唯一标识
type MessageBack struct {
	Message []byte // 消息内容
	Uuid    string // 消息唯一标识
}

// Client 表示一个连接到服务器的客户端
type Client struct {
	Conn     *websocket.Conn     // WebSocket连接对象
	Uuid     string              // 客户端唯一标识
	SendTo   chan []byte         // 发送到服务器的消息通道
	SendBack chan *MessageBack   // 发送回客户端的消息通道
}

// upgrader 用于将HTTP连接升级为WebSocket连接
var upgrader = websocket.Upgrader{
	ReadBufferSize:  2048,       // 读取缓冲区大小
	WriteBufferSize: 2048,       // 写入缓冲区大小
	// 检查连接的Origin头，允许所有来源的连接
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

// ctx 全局上下文
var ctx = context.Background()

// messageMode 消息传输模式，从配置中获取，支持"channel"和"kafka"两种模式
var messageMode = config.GetConfig().KafkaConfig.MessageMode

// Read 从WebSocket读取客户端消息并处理
// 该方法在独立的goroutine中运行，持续监听客户端发送的消息
func (c *Client) Read() {
	zlog.Info("ws read goroutine start")
	for {
		// 读取WebSocket消息（阻塞操作）
		// messageType: 消息类型（文本或二进制）
		// jsonMessage: 消息内容
		// err: 错误信息
		_, jsonMessage, err := c.Conn.ReadMessage()
		if err != nil {
			// 读取错误时记录日志并退出循环，关闭连接
			zlog.Error(err.Error())
			return
		} else {
			// 解析JSON消息为ChatMessageRequest结构
			var message = request.ChatMessageRequest{}
			if err := json.Unmarshal(jsonMessage, &message); err != nil {
				zlog.Error(err.Error())
			}
			log.Println("接受到消息为: ", jsonMessage)
			
			// 根据配置的消息模式选择不同的处理方式
			if messageMode == "channel" {
				// 通道模式：使用内存通道进行消息传输
				
				// 先处理客户端SendTo通道中积压的消息
				// 将SendTo通道中的消息发送到服务器的Transmit通道
				for len(ChatServer.Transmit) < constants.CHANNEL_SIZE && len(c.SendTo) > 0 {
					sendToMessage := <-c.SendTo
					ChatServer.SendMessageToTransmit(sendToMessage)
				}
				
				// 如果服务器的Transmit通道未满，直接将新消息发送到服务器
				if len(ChatServer.Transmit) < constants.CHANNEL_SIZE {
					ChatServer.SendMessageToTransmit(jsonMessage)
				} else if len(c.SendTo) < constants.CHANNEL_SIZE {
					// 如果服务器通道已满但客户端SendTo通道未满，将消息放入客户端SendTo通道
					c.SendTo <- jsonMessage
				} else {
					// 通道已满，通知客户端稍后重试
					if err := c.Conn.WriteMessage(websocket.TextMessage, []byte("由于目前同一时间过多用户发送消息，消息发送失败，请稍后重试")); err != nil {
						zlog.Error(err.Error())
					}
				}
			} else {
				// Kafka模式：使用Kafka进行消息传输
				// 将消息写入Kafka
				if err := myKafka.KafkaService.ChatWriter.WriteMessages(ctx, kafka.Message{
					Key:   []byte(strconv.Itoa(config.GetConfig().KafkaConfig.Partition)),
					Value: jsonMessage,
				}); err != nil {
					zlog.Error(err.Error())
				}
				zlog.Info("已发送消息：" + string(jsonMessage))
			}
		}
	}
}

// Write 从SendBack通道读取消息并发送给WebSocket客户端
// 该方法在独立的goroutine中运行，持续监听SendBack通道中的消息
func (c *Client) Write() {
	zlog.Info("ws write goroutine start")
	for messageBack := range c.SendBack { // 阻塞状态，等待消息
		// 通过WebSocket发送消息
		err := c.Conn.WriteMessage(websocket.TextMessage, messageBack.Message)
		if err != nil {
			// 发送错误时记录日志并退出循环，关闭连接
			zlog.Error(err.Error())
			return
		}
		// log.Println("已发送消息：", messageBack.Message)
		
		// 消息发送成功后，更新数据库中消息状态为"已发送"
		if res := dao.GormDB.Model(&model.Message{}).Where("uuid = ?", messageBack.Uuid).Update("status", message_status_enum.Sent); res.Error != nil {
			zlog.Error(res.Error.Error())
		}
	}
}

// NewClientInit 当接受到前端有登录消息时，会调用该函数
// 初始化新的客户端连接
func NewClientInit(c *gin.Context, clientId string) {
	// 获取Kafka配置
	kafkaConfig := config.GetConfig().KafkaConfig
	
	// 将HTTP连接升级为WebSocket连接
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		zlog.Error(err.Error())
	}
	
	// 创建新的客户端对象
	client := &Client{
		Conn:     conn,            // WebSocket连接
		Uuid:     clientId,        // 客户端唯一标识
		SendTo:   make(chan []byte, constants.CHANNEL_SIZE),       // 发送到服务器的消息通道
		SendBack: make(chan *MessageBack, constants.CHANNEL_SIZE), // 发送回客户端的消息通道
	}
	
	// 根据配置的消息模式，将客户端注册到相应的服务器
	if kafkaConfig.MessageMode == "channel" {
		ChatServer.SendClientToLogin(client)
	} else {
		KafkaChatServer.SendClientToLogin(client)
	}
	
	// 启动读取和写入协程
	go client.Read()
	go client.Write()
	zlog.Info("ws连接成功")
}

// ClientLogout 当接受到前端有登出消息时，会调用该函数
// 处理客户端登出逻辑
func ClientLogout(clientId string) (string, int) {
	// 获取Kafka配置
	kafkaConfig := config.GetConfig().KafkaConfig
	
	// 从服务器客户端列表中获取客户端对象
	client := ChatServer.Clients[clientId]
	if client != nil {
		// 根据配置的消息模式，将客户端从相应的服务器中注销
		if kafkaConfig.MessageMode == "channel" {
			ChatServer.SendClientToLogout(client)
		} else {
			KafkaChatServer.SendClientToLogout(client)
		}
		
		// 关闭WebSocket连接
		if err := client.Conn.Close(); err != nil {
			zlog.Error(err.Error())
			return constants.SYSTEM_ERROR, -1
		}
		
		// 关闭消息通道
		close(client.SendTo)
		close(client.SendBack)
	}
	return "退出成功", 0
}    