package http

import (
	"github.com/ZZGADA/easy-deploy/internal/model/service/user_manage"
	"net/http"

	"github.com/gin-gonic/gin"
)

type BindHandler struct {
	bindService *user_manage.BindService
}

func NewBindHandler(bindService *user_manage.BindService) *BindHandler {
	return &BindHandler{
		bindService: bindService,
	}
}

// GithubCallback 处理 GitHub OAuth 回调
func (h *BindHandler) GithubCallback(c *gin.Context) {
	code := c.Query("code")
	if code == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"code": 400,
			"msg":  "Missing GitHub code",
		})
		return
	}

	// 从 token 中获取用户 ID
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"code": 401,
			"msg":  "Unauthorized",
		})
		return
	}

	err := h.bindService.BindGithub(c, userID.(uint32), code)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code": 500,
			"msg":  err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"msg":  "GitHub account bound successfully",
	})
}
