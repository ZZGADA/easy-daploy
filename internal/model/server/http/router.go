package http

import (
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func SetupRouter(r *gin.Engine) {
	// 配置 CORS
	config := cors.DefaultConfig()
	config.AllowOrigins = []string{"*"} // 允许前端开发服务器的域名
	config.AllowMethods = []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"}
	config.AllowHeaders = []string{"Origin", "Content-Type", "Accept", "Authorization"}
	config.AllowCredentials = true

	r.Use(cors.New(config))

	// 注册路由
	// 登陆注册路由组
	auth := r.Group("/api/auth")
	{
		auth.POST("/sign_up", Register)
		auth.POST("/sign_up/verify", VerifyCode)
		auth.POST("/login", Login)
	}

	// 第三方绑定路由
	github := r.Group("/bind")
	{
		github.POST("/github/callback", func(context *gin.Context) {

		})
	}

	// check health
	r.GET("/api/ping", func(c *gin.Context) {
		c.JSON(200, gin.H{"message": "pong"})
	})
}
