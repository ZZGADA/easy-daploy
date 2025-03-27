package main

import (
	"flag"
	"github.com/ZZGADA/easy-deploy/internal/model/conf"
	"github.com/ZZGADA/easy-deploy/internal/model/server/http"
	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

func init() {
	var config string
	flag.StringVar(&config, "config", "test", "Specify the configuration environment: test or prod")
	flag.Parse()

	configFileName := "application-" + config
	viper.SetConfigName(configFileName)
	viper.SetConfigType("yaml")
	viper.AddConfigPath("conf")
	if err := viper.ReadInConfig(); err != nil {
		log.Fatalf("Failed to read config file: %v", err)
	}

	// 初始化MySQL和Redis
	conf.InitMySQL()
	conf.InitRedis()
}

func main() {
	r := gin.Default()

	// 注册路由
	http.SetupRouter(r)

	port := viper.GetString("server.port")
	if port == "" {
		port = "8080"
	}
	log.Printf("Server is running on port %s", port)
	if err := r.Run(":" + port); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
