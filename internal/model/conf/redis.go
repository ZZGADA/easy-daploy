package conf

import (
	"fmt"
	"github.com/ZZGADA/easy-deploy/internal/config"
	log "github.com/sirupsen/logrus"

	"github.com/go-redis/redis/v8"
)

var RedisClient *redis.Client

func InitRedis() {
	redisConfig := config.GlobalConfig.Redis

	dsn := fmt.Sprintf("%s:%d", redisConfig.Host, redisConfig.Port)

	RedisClient = redis.NewClient(&redis.Options{
		Addr:     dsn,
		Password: redisConfig.Password,
		DB:       redisConfig.DB,
	})
	_, err := RedisClient.Ping(RedisClient.Context()).Result()
	if err != nil {
		log.Fatalf("failed to connect redis: %v", err)
	}
}
