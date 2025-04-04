package websocket

import (
	"github.com/ZZGADA/easy-deploy/internal/model/dao"
	"github.com/gorilla/websocket"
)

type SocketService struct {
	userDockerfileDao              *dao.UserDockerfileDao
	userDockerDao                  dao.UserDockerDao
	userGithubDao                  *dao.UserGithubDao
	userK8sResourceDao             *dao.UserK8sResourceDao
	userOssDao                     *dao.UserOssDao
	userK8sResourceOperationLogDao *dao.UserK8sResourceOperationLogDao
}

func NewSocketService(dockerfileDao *dao.UserDockerfileDao, dockerDao dao.UserDockerDao, githubDao *dao.UserGithubDao, userK8sResourceDao *dao.UserK8sResourceDao, userOssDao *dao.UserOssDao, userK8sResourceOperationLogDao *dao.UserK8sResourceOperationLogDao) *SocketService {
	return &SocketService{
		userDockerfileDao:              dockerfileDao,
		userDockerDao:                  dockerDao,
		userGithubDao:                  githubDao,
		userK8sResourceDao:             userK8sResourceDao,
		userOssDao:                     userOssDao,
		userK8sResourceOperationLogDao: userK8sResourceOperationLogDao,
	}
}

// WSRequest WebSocket 请求结构
type WSRequest struct {
	DockerBuildStep string                 `json:"docker_build_step"`
	Data            map[string]interface{} `json:"data"`
}

// WSResponse WebSocket 响应结构
type WSResponse struct {
	Success bool        `json:"success"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

// SendError 发送错误消息
func SendError(conn *websocket.Conn, message string) {
	response := WSResponse{
		Success: false,
		Message: message,
	}
	conn.WriteJSON(response)
}

// SendSuccess 发送成功消息
func SendSuccess(conn *websocket.Conn, message string, data interface{}) {
	response := WSResponse{
		Success: true,
		Message: message,
		Data:    data,
	}
	conn.WriteJSON(response)
}
