package http

import (
	"github.com/ZZGADA/easy-deploy/internal/middleware"
	"github.com/ZZGADA/easy-deploy/internal/model/conf"
	"github.com/ZZGADA/easy-deploy/internal/model/dao"
	"github.com/ZZGADA/easy-deploy/internal/model/server/websocket"
	"github.com/ZZGADA/easy-deploy/internal/model/service/docker_manage"
	"github.com/ZZGADA/easy-deploy/internal/model/service/k8s_manage"
	"github.com/ZZGADA/easy-deploy/internal/model/service/oss_manage"
	"github.com/ZZGADA/easy-deploy/internal/model/service/team_manage"
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

	websocketHandler := websocket.NewSocketDockerHandler(
		websocket2.NewSocketService(
			dao.NewUserDockerfileDao(conf.DB),
			dao.NewUserDockerDao(conf.DB),
			dao.NewUserGithubDao(conf.DB),
			dao.NewUserK8sResourceDao(conf.DB),
			dao.NewUserOssDao(conf.DB),
			dao.NewUserK8sResourceOperationLogDao(conf.DB)),
		docker_manage.NewDockerImageService(
			dao.NewUserDockerImageDao(conf.DB), dao.NewUsersDao(conf.DB)),
		user_manage.NewDockerAccountService(
			dao.NewUserDockerDao(conf.DB)))

	r.GET(conf.WSServer.Path, middleware.WsAuthMiddleware(), websocketHandler.HandleWebSocketDockerBuild)
	r.GET(conf.WSServer.PathK8s, middleware.WsAuthMiddleware(), websocketHandler.HandleWebSocketK8s)

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
	dockerfile := r.Group("/api/user/dockerfile", middleware.CustomAuthMiddleware())
	{
		dockerfile.POST("/repository/upload", dockerfileHandler.UploadDockerfile) // Dockerfile 首次上传
		dockerfile.GET("/repository/query", dockerfileHandler.QueryDockerfile)    // Dockerfile 查询
		dockerfile.POST("/repository/update", dockerfileHandler.UpdateDockerfile) // Dockerfile 更新
		dockerfile.POST("/repository/delete", dockerfileHandler.DeleteDockerfile) // Dockerfile 删除

		dockerfile.POST("/bind/shell/save", dockerfileHandler.SaveShellPath) // Mq-UtilityBillService/build&test.shell
	}

	// docker 账号管理  & docker 镜像管理
	dockerHandler := NewDockerHandler(user_manage.NewDockerAccountService(dao.NewUserDockerDao(conf.DB)))
	dockerImageHandler := NewDockerImageHandler(docker_manage.NewDockerImageService(dao.NewUserDockerImageDao(conf.DB), dao.NewUsersDao(conf.DB)))

	// 查询 docker 镜像列表
	docker := r.Group("/api/user/docker", middleware.CustomAuthMiddleware())
	{
		docker.POST("/info/save", dockerHandler.SaveDockerAccount)
		docker.POST("/info/update", dockerHandler.UpdateDockerAccount)
		docker.POST("/info/delete", dockerHandler.DeleteDockerAccount)
		docker.GET("/info/query", dockerHandler.QueryDockerAccounts)
		docker.POST("/info/setDefault", dockerHandler.SetDefaultDockerAccount)
		docker.POST("/login", dockerHandler.LoginDockerAccount)
		docker.GET("/info/login/query", dockerHandler.QueryLoginDockerAccount)

		// 镜像管理接口
		docker.GET("/images/query", dockerImageHandler.QueryDockerImages)
	}

	// k8s 资源管理
	k8sResourceHandler := NewK8sResourceHandler(k8s_manage.NewK8sResourceService(dao.NewUserK8sResourceDao(conf.DB)))
	k8sResourceOperationLogHandler := NewK8sResourceOperationLogHandler(k8s_manage.NewK8sResourceOperationLogService(dao.NewUserK8sResourceOperationLogDao(conf.DB)))
	k8s := r.Group("/api/user/k8s", middleware.CustomAuthMiddleware())
	{
		k8s.POST("/resource/save", k8sResourceHandler.SaveResource)
		k8s.POST("/resource/update", k8sResourceHandler.UpdateResource)
		k8s.GET("/resource/query", k8sResourceHandler.QueryResources)
		k8s.POST("/resource/delete", k8sResourceHandler.DeleteResource)
		k8s.GET("/resource/operation/log/query", k8sResourceOperationLogHandler.QueryOperationLogs)
	}

	// OSS 访问信息管理
	ossHandler := NewOssHandler(oss_manage.NewOssService(dao.NewUserOssDao(conf.DB)))
	oss := r.Group("/api/user/oss", middleware.CustomAuthMiddleware())
	{
		oss.POST("/access/save", ossHandler.SaveOssAccess)
		oss.POST("/access/update", ossHandler.UpdateOssAccess)
		oss.GET("/access/query", ossHandler.QueryOssAccess)
		oss.POST("/access/delete", ossHandler.DeleteOssAccess)
	}

	// 团队管理
	teamService := team_manage.NewTeamService(dao.NewTeamDao(conf.DB), dao.NewUsersDao(conf.DB))
	teamHandler := NewTeamHandler(teamService)
	team := r.Group("/api/team", middleware.CustomAuthMiddleware())
	{
		team.POST("/create", teamHandler.CreateTeam)
		team.POST("/update", teamHandler.UpdateTeam)
		team.POST("/delete", teamHandler.DeleteTeam)
		team.GET("/info/member", teamHandler.GetTeamMemberByID)
		team.GET("/list", teamHandler.QueryTeams)
		team.GET("/info/self", teamHandler.GetUserTeam) // 用户自己所属团队信息
	}

	// 团队申请管理
	teamRequestHandler := NewTeamRequestHandler(team_manage.NewTeamRequestService(
		dao.NewTeamRequestDao(conf.DB), dao.NewTeamDao(conf.DB), dao.NewUsersDao(conf.DB)), teamService)
	teamRequest := r.Group("/api/team/request", middleware.CustomAuthMiddleware())
	{
		teamRequest.POST("/create", teamRequestHandler.CreateTeamRequest)
		teamRequest.POST("/check", teamRequestHandler.CheckTeamRequest)
		teamRequest.GET("/list", teamRequestHandler.GetTeamRequestsByTeamID)
	}

	// check health
	r.GET("/api/ping", func(c *gin.Context) {
		c.JSON(200, gin.H{"message": "pong"})
	})
}
