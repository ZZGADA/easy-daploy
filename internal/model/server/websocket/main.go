package websocket

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/ZZGADA/easy-deploy/internal/middleware"
	"github.com/ZZGADA/easy-deploy/internal/model/conf"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

// WSMessage WebSocket 消息结构
type WSMessage struct {
	Header struct {
		Authorization string `json:"authorization"`
		Method        string `json:"method"`
	} `json:"header"`
	Data map[string]interface{} `json:"data"`
}

// WSResponse WebSocket 响应结构
type WSResponse struct {
	Success bool        `json:"success"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

// HandleWebSocket 处理 WebSocket 连接
func HandleWebSocket(c *gin.Context) {
	userID := c.GetUint("user_id")
	if userID == 0 {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "未授权"})
		return
	}

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
			sendError(conn, "无效的消息格式")
			continue
		}

		// 处理不同的方法
		switch wsMsg.Header.Method {
		case "generate_dockerfile":
			handleGenerateDockerfile(conn, wsMsg.Data)
		case "clone_repository":
			handleCloneRepository(conn, wsMsg.Data)
		case "build_image":
			handleBuildImage(conn, wsMsg.Data)
		default:
			sendError(conn, "未知的方法")
		}
	}
}

// sendError 发送错误消息
func sendError(conn *websocket.Conn, message string) {
	response := WSResponse{
		Success: false,
		Message: message,
	}
	conn.WriteJSON(response)
}

// sendSuccess 发送成功消息
func sendSuccess(conn *websocket.Conn, message string, data interface{}) {
	response := WSResponse{
		Success: true,
		Message: message,
		Data:    data,
	}
	conn.WriteJSON(response)
}

// SetupWebSocketRoutes 设置 WebSocket 路由
func SetupWebSocketRoutes(r *gin.Engine) {
	ws := r.Group(conf.WSServer.Path)
	{
		ws.GET("", middleware.CustomAuthMiddleware(), HandleWebSocket)
	}
}
