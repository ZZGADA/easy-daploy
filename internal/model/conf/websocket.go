package conf

import (
	"fmt"
	"net/http"

	"github.com/aliyun/aliyun-oss-go-sdk/oss"

	"github.com/ZZGADA/easy-deploy/internal/config"
	"github.com/gorilla/websocket"
)

var (
	WSUpgrader *websocket.Upgrader
	WSServer   *WebSocketServer
)

// WebSocketServer WebSocket 服务器
type WebSocketServer struct {
	Port            int
	Path            string
	PathK8s         string
	ReadBufferSize  int
	WriteBufferSize int
	Upgrader        *websocket.Upgrader
	Connections     map[uint]*websocket.Conn // key: userID
	OssClient       map[uint]*oss.Bucket
}

// InitWebSocketServer 初始化 WebSocket 服务器
func InitWebSocketServer() {
	WSServer = &WebSocketServer{
		Port:            config.GlobalConfig.WebSocket.Port,
		Path:            config.GlobalConfig.WebSocket.Path,
		PathK8s:         config.GlobalConfig.WebSocket.PathK8s,
		ReadBufferSize:  config.GlobalConfig.WebSocket.ReadBufferSize,
		WriteBufferSize: config.GlobalConfig.WebSocket.WriteBufferSize,
		Connections:     make(map[uint]*websocket.Conn),
		OssClient:       make(map[uint]*oss.Bucket),
	}

	WSUpgrader = &websocket.Upgrader{
		ReadBufferSize:  config.GlobalConfig.WebSocket.ReadBufferSize,
		WriteBufferSize: config.GlobalConfig.WebSocket.WriteBufferSize,
		CheckOrigin: func(r *http.Request) bool {
			return true // 允许所有来源，生产环境中应该根据实际需求设置
		},
	}

	WSServer.Upgrader = WSUpgrader
}

// GetWSServer 获取 WebSocket 服务器实例
func GetWSServer() *WebSocketServer {
	return WSServer
}

// GetWSAddress 获取 WebSocket 服务器地址
func (s *WebSocketServer) GetWSAddress() string {
	return fmt.Sprintf(":%d", s.Port)
}
