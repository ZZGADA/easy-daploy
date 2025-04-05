package middleware

import (
	"fmt"
	"github.com/ZZGADA/easy-deploy/internal/define"
	"github.com/ZZGADA/easy-deploy/internal/model/conf"
	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
	"net/http"
	"strconv"
)

// CustomAuthMiddleware 自定义中间件函数
func CustomAuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 从请求头中获取 Authorization 字段
		token := c.GetHeader("Authorization")
		if token == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Missing Authorization header"})
			c.Abort()
			return
		}

		// 从 Redis 中根据 token 获取 user_id
		userIdS, redisErr := conf.RedisClient.Get(c, fmt.Sprintf(define.UserToken, token)).Result()
		if redisErr != nil {
			if redisErr != redis.Nil {
				c.JSON(http.StatusBadRequest, gin.H{
					"code": 400,
					"msg":  "Redis error",
				})
				return
			} else if redisErr == redis.Nil {
				c.JSON(401, gin.H{"error": "Invalid token"})
			}
		}

		userId, terr := strconv.ParseUint(userIdS, 10, 32)
		if terr != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"code": 400,
				"msg":  "Redis error",
			})
			return
		}

		// 将 user_id 写入 Gin 上下文
		c.Set("user_id", uint(userId))
		c.Next()
	}
}

// CustomAuthMiddleware 自定义中间件函数
func WsAuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 从请求头中获取 Authorization 字段
		token := c.Query("token")
		if token == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Missing Authorization header"})
			c.Abort()
			return
		}

		// 从 Redis 中根据 token 获取 user_id
		userIdS, redisErr := conf.RedisClient.Get(c, fmt.Sprintf(define.UserToken, token)).Result()
		if redisErr != nil {
			if redisErr != redis.Nil {
				c.JSON(http.StatusBadRequest, gin.H{
					"code": 400,
					"msg":  "Redis error",
				})
				return
			} else if redisErr == redis.Nil {
				c.JSON(401, gin.H{"error": "Invalid token"})
			}
		}

		userId, terr := strconv.ParseUint(userIdS, 10, 32)
		if terr != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"code": 400,
				"msg":  "Redis error",
			})
			return
		}

		// 将 user_id 写入 Gin 上下文
		c.Set("user_id", uint(userId))
		c.Next()
	}
}
