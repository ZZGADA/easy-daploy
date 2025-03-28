package http

import (
	"fmt"
	"github.com/ZZGADA/easy-deploy/internal/model/conf"
	"github.com/ZZGADA/easy-deploy/internal/model/define"
	"github.com/ZZGADA/easy-deploy/internal/model/service/user_manage"
	"github.com/go-redis/redis/v8"
	"net/http"
	"strconv"

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
	err := h.bindService.BindGithub(c, uint32(userId), code)
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
