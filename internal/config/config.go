package config

import (
	"fmt"

	"github.com/spf13/viper"
)

type Config struct {
	// 服务器配置
	Server struct {
		Port string
	}

	// 数据库配置
	MySQL struct {
		Host     string
		Port     int
		Username string
		Password string
		Database string
	}

	// Redis配置
	Redis struct {
		Host     string
		Port     int
		Password string
		DB       int
	}

	// GitHub OAuth 配置
	Github struct {
		ClientID     string
		ClientSecret string
	}

	WebSocket struct {
		Port            int    `mapstructure:"port"`
		Path            string `mapstructure:"path"`
		PathK8s         string `mapstructure:"path_k8s"`
		ReadBufferSize  int    `mapstructure:"read_buffer_size"`
		WriteBufferSize int    `mapstructure:"write_buffer_size"`
	}

	Elastic struct {
		Host     string `mapstructure:"host"`
		Port     int    `mapstructure:"port"`
		Username string `mapstructure:"username"`
		Password string `mapstructure:"password"`
	}

	Kafka struct {
		Brokers []string `mapstructure:"brokers"`
		Topic   string   `mapstructure:"topic"`
		GroupId string   `mapstructure:"group_id"`
	}

	Smtp struct {
		From     string `mapstructure:"from"`
		Host     string `mapstructure:"host"`
		Port     int    `mapstructure:"port"`
		User     string `mapstructure:"user"`
		Password string `mapstructure:"password"`
	}
}

var GlobalConfig Config

// InitConfig 初始化配置
func InitConfig(env string) error {
	// 1. 首先读取 application-{env}.yaml
	v := viper.New()
	v.SetConfigName("application-" + env)
	v.SetConfigType("yaml")
	v.AddConfigPath("conf")

	if err := v.ReadInConfig(); err != nil {
		return fmt.Errorf("读取应用配置文件失败: %w", err)
	}

	// 2. 读取 .env 文件
	envViper := viper.New()
	envViper.SetConfigName(".env")
	envViper.SetConfigType("env")
	envViper.AddConfigPath(".")
	envViper.AutomaticEnv()

	if err := envViper.ReadInConfig(); err != nil {
		return fmt.Errorf("读取环境变量配置文件失败: %w", err)
	}

	// 3. 将配置映射到结构体
	// 3.1 从 application-{env}.yaml 映射配置
	if err := v.Unmarshal(&GlobalConfig); err != nil {
		return fmt.Errorf("解析应用配置失败: %w", err)
	}

	// 3.2 从 .env 文件映射 GitHub 配置
	GlobalConfig.Github.ClientID = envViper.GetString("GITHUB_CLIENT_ID")
	GlobalConfig.Github.ClientSecret = envViper.GetString("GITHUB_CLIENT_SECRET")

	return nil
}
