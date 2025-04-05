package websocket

import (
	"github.com/ZZGADA/easy-deploy/internal/model/service/docker_manage"
	websocket2 "github.com/ZZGADA/easy-deploy/internal/model/service/websocket"
)

// WSMessage WebSocket 消息结构
type WSMessage struct {
	Data            map[string]interface{} `json:"data"`
	DockerBuildStep string                 `json:"docker_build_step"`
}

// WSMessageK8s WSMessageK8s 消息结构
type WSMessageK8s struct {
	Data    map[string]interface{} `json:"data"`
	Command string                 `json:"command"`
	Step    string                 `json:"step"`
}

// WSResponse WebSocket 响应结构
type WSResponse struct {
	Success bool        `json:"success"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

// SocketDockerHandler docker镜像构建
type SocketDockerHandler struct {
	socketService             *websocket2.SocketService
	socketDockerImagesService *docker_manage.DockerImageService
}

// NewSocketDockerHandler 创建 SocketDockerHandler 实例
func NewSocketDockerHandler(socketService *websocket2.SocketService, socketDockerImagesService *docker_manage.DockerImageService) *SocketDockerHandler {
	return &SocketDockerHandler{
		socketService:             socketService,
		socketDockerImagesService: socketDockerImagesService,
	}
}
