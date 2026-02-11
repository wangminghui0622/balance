package database

import (
	"context"
	"fmt"
	"time"

	"balance/backend/internal/config"

	"github.com/redis/go-redis/v9"
)

var rdb *redis.Client

// InitRedis 初始化Redis连接
func InitRedis(cfg *config.RedisConfig) error {
	rdb = redis.NewClient(&redis.Options{
		Addr:     cfg.Addr(),
		Password: cfg.Password,
		DB:       cfg.DB,
		PoolSize: cfg.PoolSize,
	})

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := rdb.Ping(ctx).Err(); err != nil {
		return fmt.Errorf("连接Redis失败: %w", err)
	}

	return nil
}

// GetRedis 获取Redis客户端
func GetRedis() *redis.Client {
	return rdb
}

// CloseRedis 关闭Redis连接
func CloseRedis() error {
	if rdb != nil {
		return rdb.Close()
	}
	return nil
}
