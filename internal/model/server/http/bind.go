package http

import (
	"fmt"
	"github.com/ZZGADA/easy-deploy/internal/model/conf"
	"github.com/ZZGADA/easy-deploy/internal/model/define"
	"github.com/ZZGADA/easy-deploy/internal/model/service/user_manage"
	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
	"net/http"
	"strconv"
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
	redirectURL := c.Query("redirect_url") // 获取前端传来的重定向URL

	if code == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"code": 400,
			"msg":  "Missing GitHub code",
		})
		return
	}

	// 从 token 中获取用户 ID
	userToken := c.Query("token")
	if userToken == "" {
		c.JSON(http.StatusUnauthorized, gin.H{
			"code": 401,
			"msg":  "Unauthorized",
		})
		return
	}

	userIdS, redisErr := conf.RedisClient.Get(c, fmt.Sprintf(define.UserToken, userToken)).Result()
	if redisErr != nil && redisErr != redis.Nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code": 400,
			"msg":  "Redis error",
		})
		return
	}
	userId, terr := strconv.ParseUint(userIdS, 10, 32)
	if terr != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code": 400,
			"msg":  "Redis error",
		})
		return
	}
	_, err := h.bindService.BindGithub(c, uint32(userId), code)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code": 500,
			"msg":  err.Error(),
		})
		return
	}

	// 绑定成功后重定向到前端页面
	if redirectURL != "" {
		c.Redirect(http.StatusTemporaryRedirect, redirectURL)
		return
	}

	// 如果没有重定向URL，返回JSON响应
	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"msg":  "GitHub account bound successfully",
	})
}

// CheckGithubBinding 检查用户是否已绑定 GitHub
func (h *BindHandler) CheckGithubBinding(c *gin.Context) {
	// 从 token 中获取用户 ID
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"code": 401,
			"msg":  "未授权",
		})
		return
	}

	// 检查绑定状态
	bound, userGithub, err := h.bindService.CheckGithubBinding(c, uint(userID.(uint64)))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code": 500,
			"msg":  err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"data": gin.H{
			"bound":      bound,
			"avatar_url": userGithub.AvatarUrl,
			"id":         userGithub.UserId,
			"email":      userGithub.Email,
			"name":       userGithub.Name,
			"github_id":  userGithub.GithubId,
		},
	})
}

// UnbindGithub 解绑 GitHub 账号
func (h *BindHandler) UnbindGithub(c *gin.Context) {
	// 从 token 中获取用户 ID
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"code": 401,
			"msg":  "未授权",
		})
		return
	}

	// 执行解绑
	if err := h.bindService.UnbindGithub(c, uint(userID.(uint64))); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code": 500,
			"msg":  err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"msg":  "GitHub 账号解绑成功",
	})
}
