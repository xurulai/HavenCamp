package config

import (
	"log"
	"time"

	"github.com/BurntSushi/toml"
)

type MainConfig struct {
	AppName string `toml:"appName"`
	Host    string `toml:"host"`
	Port    int    `toml:"port"`
}

type MysqlConfig struct {
	Host         string `toml:"host"`
	Port         int    `toml:"port"`
	User         string `toml:"user"`
	Password     string `toml:"password"`
	DatabaseName string `toml:"databaseName"`
}

type RedisConfig struct {
	Host     string `toml:"host"`
	Port     int    `toml:"port"`
	Password string `toml:"password"`
	Db       int    `toml:"db"`
}

type AuthCodeConfig struct {
	AccessKeyID     string `toml:"accessKeyID"`
	AccessKeySecret string `toml:"accessKeySecret"`
	SignName        string `toml:"signName"`
	TemplateCode    string `toml:"templateCode"`
}

type LogConfig struct {
	LogPath string `toml:"logPath"`
}

type KafkaConfig struct {
	MessageMode string        `toml:"messageMode"`
	HostPort    string        `toml:"hostPort"`
	LoginTopic  string        `toml:"loginTopic"`
	LogoutTopic string        `toml:"logoutTopic"`
	ChatTopic   string        `toml:"chatTopic"`
	Partition   int           `toml:"partition"`
	Timeout     time.Duration `toml:"timeout"`
}

type StaticSrcConfig struct {
	StaticAvatarPath string `toml:"staticAvatarPath"`
	StaticFilePath   string `toml:"staticFilePath"`
}

type Config struct {
	MainConfig      `toml:"mainConfig"`
	MysqlConfig     `toml:"mysqlConfig"`
	RedisConfig     `toml:"redisConfig"`
	AuthCodeConfig  `toml:"authCodeConfig"`
	LogConfig       `toml:"logConfig"`
	KafkaConfig     `toml:"kafkaConfig"`
	StaticSrcConfig `toml:"staticSrcConfig"`
}

var config *Config

func LoadConfig() error {
	// 配置文件路径优先级：Docker环境 -> 本地环境 -> 云服务器环境
	configPaths := []string{
		"/root/configs/config.toml",                         // Docker环境
		"configs/config.toml",                               // 本地环境
		"/root/project/HavenCamp/configs/config_local.toml", // Ubuntu22.04云服务器部署
	}

	var lastErr error
	for _, path := range configPaths {
		if _, err := toml.DecodeFile(path, config); err == nil {
			log.Printf("Successfully loaded config from: %s", path)
			return nil
		} else {
			lastErr = err
			log.Printf("Failed to load config from %s: %v", path, err)
		}
	}

	log.Printf("Failed to load config from any path: %v", lastErr)
	return lastErr
}

func GetConfig() *Config {
	if config == nil {
		config = new(Config)
		_ = LoadConfig()
	}
	return config
}
