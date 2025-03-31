package http

import (
	"net/http"

	"github.com/ZZGADA/easy-deploy/internal/model/service/user_manage"

	"github.com/gin-gonic/gin"
)

// DockerAccountRequest Docker账号请求结构体
type DockerAccountRequest struct {
	ID       uint   `json:"id,omitempty"`
	Server   string `json:"server" binding:"required"`
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
	Comment  string `json:"comment"`
}

// SetDefaultRequest 设置默认账号请求结构体
type SetDefaultRequest struct {
	DockerAccountID uint `json:"docker_account_id" binding:"required"`
}

// DockerHandler Docker账号管理处理器
type DockerHandler struct {
	dockerService *user_manage.DockerAccountService
}

// NewDockerHandler 创建 DockerHandler 实例
func NewDockerHandler(dockerService *user_manage.DockerAccountService) *DockerHandler {
	return &DockerHandler{
		dockerService: dockerService,
	}
}

// SaveDockerAccount 保存Docker账号信息
func (h *DockerHandler) SaveDockerAccount(c *gin.Context) {
	userID := c.GetUint("user_id")
	var req DockerAccountRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    http.StatusBadRequest,
			"message": "无效的请求参数",
		})
		return
	}

	h.dockerService.UserID = userID
	_, err := h.dockerService.SaveDockerAccount(req.Server, req.Username, req.Password, req.Comment)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    http.StatusInternalServerError,
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    http.StatusOK,
		"message": "save success",
	})
}

// UpdateDockerAccount 更新Docker账号信息
func (h *DockerHandler) UpdateDockerAccount(c *gin.Context) {
	userID := c.GetUint("user_id")
	var req DockerAccountRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    http.StatusBadRequest,
			"message": "无效的请求参数",
		})
		return
	}

	if req.ID == 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    http.StatusBadRequest,
			"message": "账号ID不能为空",
		})
		return
	}

	h.dockerService.UserID = userID
	_, err := h.dockerService.UpdateDockerAccount(req.ID, req.Server, req.Username, req.Password, req.Comment)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    http.StatusInternalServerError,
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    http.StatusOK,
		"message": "update success",
	})
}

// DeleteDockerAccount 删除Docker账号
func (h *DockerHandler) DeleteDockerAccount(c *gin.Context) {
	userID := c.GetUint("user_id")
	var req struct {
		ID uint `json:"id" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    http.StatusBadRequest,
			"message": "无效的请求参数",
		})
		return
	}

	h.dockerService.UserID = userID
	_, err := h.dockerService.DeleteDockerAccount(req.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    http.StatusInternalServerError,
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    http.StatusOK,
		"message": "delete success",
	})
}

// QueryDockerAccounts 查询Docker账号列表
func (h *DockerHandler) QueryDockerAccounts(c *gin.Context) {
	userID := c.GetUint("user_id")

	h.dockerService.UserID = userID
	response, err := h.dockerService.QueryDockerAccounts()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    http.StatusInternalServerError,
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code": http.StatusOK,
		"data": response,
	})
}

// SetDefaultDockerAccount 设置默认Docker账号
func (h *DockerHandler) SetDefaultDockerAccount(c *gin.Context) {
	userID := c.GetUint("user_id")
	var req SetDefaultRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    http.StatusBadRequest,
			"message": "无效的请求参数",
		})
		return
	}

	h.dockerService.UserID = userID
	_, err := h.dockerService.SetDefaultAccount(req.DockerAccountID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    http.StatusInternalServerError,
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    http.StatusOK,
		"message": "success",
	})
}
