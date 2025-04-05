package define

const (
	VerifyCodeEmail  = "verify_code_email:%s"
	RegisterPassword = "register_password:%s"
)

const (
	UserToken       = "user_token:%s"       // token->user id
	UserEmailToken  = "user_email_token:%s" // email-> token
	UserInfo        = "user_info:%s"
	UserDockerLogin = "user_docker_login:%d" // user id
)

const (
	K8sResourceStatusRun     = 1 // 1 运行正常
	K8sResourceStatusStop    = 2 // 2 运行停止
	K8sResourceStatusRestart = 3 // 3 容器重启
)
