package websocket

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/ZZGADA/easy-deploy/internal/model/service/docker_manage"
	websocket2 "github.com/ZZGADA/easy-deploy/internal/model/service/websocket"
	"github.com/sirupsen/logrus"

	"github.com/ZZGADA/easy-deploy/internal/model/conf"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
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

// HandleWebSocketDockerBuild 处理 WebSocket 连接
func (s *SocketDockerHandler) HandleWebSocketDockerBuild(c *gin.Context) {
	userID := c.GetUint("user_id")
	if userID == 0 {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "未授权"})
		return
	}

	// 将HTTP连接升级为WebSocket连接
	conn, err := conf.WSUpgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Printf("WebSocket upgrade error: %v\n", err)
		return
	}

	// 存储连接
	conf.WSServer.Connections[userID] = conn

	// 清理连接
	defer func() {
		conn.Close()
		delete(conf.WSServer.Connections, userID)
	}()

	// 处理消息
	for {
		_, message, err := conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("WebSocket read error: %v\n", err)
			}
			break
		}

		var wsMsg WSMessage
		if err := json.Unmarshal(message, &wsMsg); err != nil {
			websocket2.SendError(conn, "无效的消息格式")
			continue
		}

		log.Println(wsMsg)
		dockerBuildStep := wsMsg.DockerBuildStep
		if dockerBuildStep == "" {
			logrus.Warn("docker_build_step 不存在")
			break
		}
		// 处理不同的方法
		switch dockerBuildStep {
		case "init":
			websocket2.SendSuccess(conn, "connect success", "connect success")
		case "generate_dockerfile":
			s.socketService.HandleGenerateDockerfile(conn, wsMsg.Data, userID)
		case "clone_repository":
			s.socketService.HandleCloneRepository(conn, wsMsg.Data, userID)
		case "build_image":
			fullImageName := s.socketService.HandleBuildImage(conn, wsMsg.Data, userID)
			if fullImageName != "" {
				err := s.socketDockerImagesService.SaveDockerImageWS(wsMsg.Data, userID, fullImageName)
				if err != nil {
					logrus.Warn("save docker image err: ", err)
				}
			}
			break
		default:
			websocket2.SendError(conn, "未知的方法")
		}
	}
}

// HandleWebSocketK8s 处理 WebSocket 连接
func (s *SocketDockerHandler) HandleWebSocketK8s(c *gin.Context) {
	userID := c.GetUint("user_id")
	if userID == 0 {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "未授权"})
		return
	}

	// 将HTTP连接升级为WebSocket连接
	conn, err := conf.WSUpgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Printf("WebSocket upgrade error: %v\n", err)
		return
	}

	ossClient, err := s.socketService.GetOssClient(userID)
	if err != nil {
		logrus.Warnf("get ossClient error: %v\n", err)
		websocket2.SendError(conn, err.Error())
	}

	// 存储连接
	conf.WSServer.Connections[userID] = conn
	conf.WSServer.OssClient[userID] = ossClient

	// 清理连接
	defer func() {
		websocket2.SendSuccess(conn, "ws close", "ws close success")
		conn.Close()
		delete(conf.WSServer.Connections, userID)
		delete(conf.WSServer.OssClient, userID)
	}()

	// 处理消息
	// 处理消息
	for {
		_, message, err := conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("WebSocket read error: %v\n", err)
			}
			break
		}

		var wsMsg WSMessageK8s
		if err := json.Unmarshal(message, &wsMsg); err != nil {
			websocket2.SendError(conn, "无效的消息格式")
			continue
		}

		log.Println("收到消息 , userId: ", userID, "message: ", wsMsg)
		step := wsMsg.Step
		if step == "" {
			logrus.Warn("k8s command  不存在")
			break
		}
		// 处理不同的方法
		switch step {
		case "init":
			websocket2.SendSuccess(conn, "connect success", websocket2.K8sCommandResponse{
				Command: wsMsg.Command,
				Result:  "connect success",
			})
		case "connected":
			logrus.Info("")
			s.socketService.HandleKubeCommand(conn, wsMsg.Command, wsMsg.Data, userID)
		case "close":
			break
		default:
			websocket2.SendError(conn, "未知的方法")
		}
	}
}
