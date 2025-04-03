package main

import (
	"flag"

	"github.com/ZZGADA/easy-deploy/internal/config"
	"github.com/ZZGADA/easy-deploy/internal/model/conf"
	"github.com/ZZGADA/easy-deploy/internal/model/server/http"
	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
)

var configEnv string

func init() {
	flag.StringVar(&configEnv, "config", "test", "Specify the configuration environment: test or prod")
	flag.Parse()

	// 初始化配置
	if err := config.InitConfig(configEnv); err != nil {
		log.Fatalf("配置初始化失败: %v", err)
	}

	// 初始化MySQL和Redis
	// 初始化 WebSocket 服务
	conf.InitMySQL()
	conf.InitRedis()
	conf.InitWebSocketServer()
	conf.InitK8s()
}

func main() {
	r := gin.Default()

	// 设置路由
	http.SetupRouter(r)

	port := config.GlobalConfig.Server.Port
	if port == "" {
		port = "8080"
	}
	log.Printf("Server is running on port %s", port)
	if err := r.Run(":" + port); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
