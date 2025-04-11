package utils

import (
	"context"
	"crypto/rand"
	"fmt"
	"github.com/go-redis/redis/v8"
	"math/big"
	mr "math/rand"
	"time"
)

// AcquireLock 获取锁 value为时间戳+随机数
func AcquireLock(client *redis.Client, context context.Context, lockKey string, value string, expiration time.Duration) (bool, error) {
	// NX: Only set the key if it does not already exist.
	// PX: Set the specified expire time, in milliseconds.
	result, err := client.SetNX(context, lockKey, value, expiration).Result()
	return result, err
}

// ReleaseLock 释放锁 value为时间戳+随机数
func ReleaseLock(client *redis.Client, context context.Context, lockKey string, value string) (bool, error) {
	script := `
        if redis.call("get", KEYS[1]) == ARGV[1] then
            return redis.call("del", KEYS[1])
        else
            return 0
        end
    `

	result, err := client.Eval(context, script, []string{lockKey}, value).Int()
	return result != 0, err
}

// GenerateUniqueValue 生成随机value
func GenerateUniqueValue() string {
	timestamp := time.Now().UnixNano()
	randomNum, _ := rand.Int(rand.Reader, big.NewInt(1000000))
	return fmt.Sprintf("%d-%d", timestamp, randomNum)
}

// GenerateUUID 生成8位uuid
func GenerateUUID() uint32 {
	// 生成 10000000 到 99999999 之间的随机整数
	randomNum := mr.Intn(90000000) + 10000000
	// 将整数转换为字符串
	return uint32(randomNum)
}
