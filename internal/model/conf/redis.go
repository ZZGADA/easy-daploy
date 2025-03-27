package conf

import (
	log "github.com/sirupsen/logrus"

	"github.com/go-redis/redis/v8"
	"github.com/spf13/viper"
)

var RedisClient *redis.Client

func InitRedis() {
	RedisClient = redis.NewClient(&redis.Options{
		Addr:     viper.GetString("redis.addr"),
		Password: viper.GetString("redis.password"),
		DB:       0,
	})
	_, err := RedisClient.Ping(RedisClient.Context()).Result()
	if err != nil {
		log.Fatalf("failed to connect redis: %v", err)
	}
}
