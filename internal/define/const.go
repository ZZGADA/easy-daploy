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

const (
	K8sRunningResources = "k8s:running_resources:%d"
)

const (
	TeamRequestStatusWait     = 0 // 0: 待处理,
	TeamRequestStatusApproval = 1 // 1: 已同意,
	TeamRequestStatusReject   = 2 // 2: 已拒绝
)

const (
	TeamRequestTypeIn  = 0 // 0: 加入团队
	TeamRequestTypeOut = 1 // 1: 退出团队
)
