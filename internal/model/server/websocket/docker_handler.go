package websocket

import (
	"encoding/json"
	"github.com/ZZGADA/easy-deploy/internal/model/conf"
	websocket2 "github.com/ZZGADA/easy-deploy/internal/model/service/websocket"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/sirupsen/logrus"
	"log"
	"net/http"
)

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
			err := s.socketDockerAccountService.DockerLogin(userID)
			if err != nil {
				logrus.Error("docker login error: %v\n", err)
				websocket2.SendError(conn, err.Error())
				break
			}

			websocket2.SendSuccess(conn, "docker logins  success", "docker login  success")
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
