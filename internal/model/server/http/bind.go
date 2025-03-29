package http

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/ZZGADA/easy-deploy/internal/model/conf"
	"github.com/ZZGADA/easy-deploy/internal/model/define"
	"github.com/ZZGADA/easy-deploy/internal/model/service/user_manage"
	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
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
	bound, userGithub, err := h.bindService.CheckGithubBinding(c, userID.(uint))
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
	if err := h.bindService.UnbindGithub(c, userID.(uint)); err != nil {
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

// DeveloperTokenRequest 开发者令牌请求结构体
type DeveloperTokenRequest struct {
	DeveloperToken string `json:"developer_token" binding:"required"`
	ExpireTime     string `json:"expire_time" binding:"required"`
	Comment        string `json:"comment" binding:"required"`
	RepositoryName string `json:"repository_name" binding:"required"`
}

// SaveDeveloperToken 保存开发者令牌
func (h *BindHandler) SaveDeveloperToken(c *gin.Context) {
	// 从上下文获取用户ID
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"code":    http.StatusUnauthorized,
			"message": "未授权的访问",
		})
		return
	}

	var req DeveloperTokenRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    http.StatusBadRequest,
			"message": "无效的请求参数",
		})
		return
	}

	// 解析过期时间
	expireTime, err := time.Parse("2006-01-02 15:04:05", req.ExpireTime)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    http.StatusBadRequest,
			"message": "无效的过期时间格式",
		})
		return
	}

	// 保存开发者令牌
	err = h.bindService.SaveDeveloperToken(c.Request.Context(), userID.(uint), req.DeveloperToken, req.Comment, expireTime, req.RepositoryName)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    http.StatusInternalServerError,
			"message": "保存开发者令牌失败",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    http.StatusOK,
		"message": "保存开发者令牌成功",
	})
}

// UpdateDeveloperToken 更新开发者令牌
func (h *BindHandler) UpdateDeveloperToken(c *gin.Context) {
	// 从上下文获取用户ID
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"code":    http.StatusUnauthorized,
			"message": "未授权的访问",
		})
		return
	}

	var req DeveloperTokenRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    http.StatusBadRequest,
			"message": "无效的请求参数",
		})
		return
	}

	// 解析过期时间
	expireTime, err := time.Parse("2006-01-02 15:04:05", req.ExpireTime)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    http.StatusBadRequest,
			"message": "无效的过期时间格式",
		})
		return
	}

	// 更新开发者令牌
	err = h.bindService.UpdateDeveloperToken(c.Request.Context(), userID.(uint), req.DeveloperToken, req.Comment, expireTime, req.RepositoryName)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    http.StatusInternalServerError,
			"message": "更新开发者令牌失败",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    http.StatusOK,
		"message": "更新开发者令牌成功",
	})
}

// QueryDeveloperToken 查询开发者令牌
func (h *BindHandler) QueryDeveloperToken(c *gin.Context) {
	// 从上下文获取用户ID
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"code":    http.StatusUnauthorized,
			"message": "未授权的访问",
		})
		return
	}

	// 查询开发者令牌
	tokenInfo, err := h.bindService.GetDeveloperToken(c.Request.Context(), userID.(uint))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    http.StatusInternalServerError,
			"message": "查询开发者令牌失败",
		})
		return
	}

	// 如果未找到令牌信息
	if tokenInfo == nil {
		c.JSON(http.StatusOK, gin.H{
			"code": http.StatusOK,
			"data": nil,
		})
		return
	}

	// 处理过期时间
	var expireTimeStr string
	if tokenInfo.DeveloperTokenExpireTime != nil {
		expireTimeStr = tokenInfo.DeveloperTokenExpireTime.Format("2006-01-02 15:04:05")
	}

	c.JSON(http.StatusOK, gin.H{
		"code": http.StatusOK,
		"data": gin.H{
			"developer_token":             tokenInfo.DeveloperToken,
			"developer_token_comment":     tokenInfo.DeveloperTokenComment,
			"developer_token_expire_time": expireTimeStr,
			"developer_repository_name":   tokenInfo.DeveloperRepositoryName,
		},
	})
}
