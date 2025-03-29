package http

import (
	"github.com/ZZGADA/easy-deploy/internal/middleware"
	"github.com/ZZGADA/easy-deploy/internal/model/conf"
	"github.com/ZZGADA/easy-deploy/internal/model/dao"
	"github.com/ZZGADA/easy-deploy/internal/model/service/user_manage"
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

	// 创建 BindHandler 实例
	bindHandler := NewBindHandler(user_manage.NewBindService(dao.NewUserGithubDao(conf.DB)))

	// GitHub 绑定相关路由
	bindGroup := r.Group("/api/user/github")
	{
		bindGroup.GET("/bind/callback", bindHandler.GithubCallback)
		bindGroup.GET("/status", middleware.CustomAuthMiddleware(), bindHandler.CheckGithubBinding)
		bindGroup.POST("/unbind", middleware.CustomAuthMiddleware(), bindHandler.UnbindGithub)

		// 开发者令牌相关路由
		bindGroup.POST("/developer/token/save", middleware.CustomAuthMiddleware(), bindHandler.SaveDeveloperToken)
		bindGroup.POST("/developer/token/update", middleware.CustomAuthMiddleware(), bindHandler.UpdateDeveloperToken)
		bindGroup.GET("/developer/token/query", middleware.CustomAuthMiddleware(), bindHandler.QueryDeveloperToken)
	}

	// check health
	r.GET("/api/ping", func(c *gin.Context) {
		c.JSON(200, gin.H{"message": "pong"})
	})
}
