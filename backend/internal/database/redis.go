package database

import (
	"context"

	"github.com/redis/go-redis/v9"
)

// InitRedis 初始化Redis连接
func InitRedis(addr, password string) (*redis.Client, error) {
	client := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: password,
		DB:       0,
	})

	// 测试连接
	ctx := context.Background()
	_, err := client.Ping(ctx).Result()
	if err != nil {
		return nil, err
	}

	return client, nil
}
