package main

import (
	"fmt"
	"haven_camp_server/internal/config"
	"haven_camp_server/internal/https_server"
	"haven_camp_server/internal/service/chat"
	"haven_camp_server/internal/service/kafka"
	myredis "haven_camp_server/internal/service/redis"
	"haven_camp_server/pkg/zlog"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	conf := config.GetConfig()
	host := conf.MainConfig.Host
	port := conf.MainConfig.Port
	kafkaConfig := conf.KafkaConfig
	if kafkaConfig.MessageMode == "kafka" {
		kafka.KafkaService.KafkaInit()
	}

	if kafkaConfig.MessageMode == "channel" {
		go chat.ChatServer.Start()
	} else {
		go chat.KafkaChatServer.Start()
	}

	go func() {
		// 检查SSL证书文件是否存在
		certFile := "/etc/ssl/certs/server.crt"
		keyFile := "/etc/ssl/private/server.key"

		if _, err := os.Stat(certFile); err == nil {
			if _, err := os.Stat(keyFile); err == nil {
				// SSL证书存在，使用HTTPS
				if err := https_server.GE.RunTLS(fmt.Sprintf("%s:%d", host, port), certFile, keyFile); err != nil {
					zlog.Fatal("HTTPS server running fault")
					return
				}
			} else {
				// 证书文件不存在，使用HTTP
				if err := https_server.GE.Run(fmt.Sprintf("%s:%d", host, port)); err != nil {
					zlog.Fatal("HTTP server running fault")
					return
				}
			}
		} else {
			// 证书文件不存在，使用HTTP
			if err := https_server.GE.Run(fmt.Sprintf("%s:%d", host, port)); err != nil {
				zlog.Fatal("HTTP server running fault")
				return
			}
		}
	}()

	// 设置信号监听
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	// 等待信号
	<-quit

	if kafkaConfig.MessageMode == "kafka" {
		kafka.KafkaService.KafkaClose()
	}

	chat.ChatServer.Close()

	zlog.Info("关闭服务器...")

	// 删除所有Redis键
	if err := myredis.DeleteAllRedisKeys(); err != nil {
		zlog.Error(err.Error())
	} else {
		zlog.Info("所有Redis键已删除")
	}

	zlog.Info("服务器已关闭")

}
