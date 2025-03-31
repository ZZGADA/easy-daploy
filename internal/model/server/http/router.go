package http

import (
	"github.com/ZZGADA/easy-deploy/internal/middleware"
	"github.com/ZZGADA/easy-deploy/internal/model/conf"
	"github.com/ZZGADA/easy-deploy/internal/model/dao"
	"github.com/ZZGADA/easy-deploy/internal/model/server/websocket"
	"github.com/ZZGADA/easy-deploy/internal/model/service/docker_manage"
	"github.com/ZZGADA/easy-deploy/internal/model/service/user_manage"
	websocket2 "github.com/ZZGADA/easy-deploy/internal/model/service/websocket"
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

	// 注册 WebSocket 路由

	websocketHandler := websocket.NewSocketDockerHandler(websocket2.NewSocketDockerService(
		dao.NewUserDockerfileDao(conf.DB),
		dao.NewUserDockerDao(conf.DB),
		dao.NewUserGithubDao(conf.DB)))

	r.GET(conf.WSServer.Path, middleware.CustomAuthMiddleware(), websocketHandler.HandleWebSocket)

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
		// GitHub 绑定相关路由
		bindGroup.GET("/bind/callback", bindHandler.GithubCallback)
		bindGroup.GET("/status", middleware.CustomAuthMiddleware(), bindHandler.CheckGithubBinding)
		bindGroup.POST("/unbind", middleware.CustomAuthMiddleware(), bindHandler.UnbindGithub)

		// 开发者令牌相关路由
		bindGroup.POST("/developer/token/save", middleware.CustomAuthMiddleware(), bindHandler.SaveDeveloperToken)
		bindGroup.POST("/developer/token/update", middleware.CustomAuthMiddleware(), bindHandler.UpdateDeveloperToken)
		bindGroup.GET("/developer/token/query", middleware.CustomAuthMiddleware(), bindHandler.QueryDeveloperToken)
	}

	// 创建 DockerfileHandler 实例
	dockerfileHandler := NewDockerfileHandler(docker_manage.NewDockerfileService(dao.NewUserDockerfileDao(conf.DB)))

	// 用户仓库 Dockerfile 制作
	dockerfile := r.Group("api/user/dockerfile", middleware.CustomAuthMiddleware())
	{
		dockerfile.POST("/repository/upload", dockerfileHandler.UploadDockerfile) // Dockerfile 首次上传
		dockerfile.GET("/repository/query", dockerfileHandler.QueryDockerfile)    // Dockerfile 查询
		dockerfile.POST("/repository/update", dockerfileHandler.UpdateDockerfile) // Dockerfile 更新
		dockerfile.POST("/repository/delete", dockerfileHandler.DeleteDockerfile) // Dockerfile 删除
	}

	dockerHandler := NewDockerHandler(user_manage.NewDockerAccountService(dao.NewUserDockerDao(conf.DB)))
	docker := r.Group("/api/user/docker", middleware.CustomAuthMiddleware())
	{
		docker.POST("/info/save", dockerHandler.SaveDockerAccount)
		docker.POST("/info/update", dockerHandler.UpdateDockerAccount)
		docker.POST("/info/delete", dockerHandler.DeleteDockerAccount)
		docker.GET("/info/query", dockerHandler.QueryDockerAccounts)
		docker.POST("/info/setDefault", dockerHandler.SetDefaultDockerAccount)
		docker.POST("/login", dockerHandler.LoginDockerAccount)
		docker.GET("/info/login/query", dockerHandler.QueryLoginDockerAccount)
	}

	// check health
	r.GET("/api/ping", func(c *gin.Context) {
		c.JSON(200, gin.H{"message": "pong"})
	})
}
