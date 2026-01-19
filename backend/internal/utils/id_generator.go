package utils

import (
	"context"
	"math/rand"
	"time"

	"github.com/redis/go-redis/v9"
)

var rng = rand.New(rand.NewSource(time.Now().UnixNano()))

// IDGenerator ID生成器
type IDGenerator struct {
	client *redis.Client
}

// NewIDGenerator 创建ID生成器
func NewIDGenerator(client *redis.Client) *IDGenerator {
	return &IDGenerator{client: client}
}

// GenerateShopOwnerID 生成店铺ID（10000000000开头，每个加100-500）
func (g *IDGenerator) GenerateShopOwnerID(ctx context.Context) (int64, error) {
	key := "id:generator:shopowner"
	
	// 获取当前计数
	current, err := g.client.Get(ctx, key).Int64()
	if err == redis.Nil {
		// 如果不存在，初始化为10000000000
		current = 10000000000
	} else if err != nil {
		return 0, err
	}

	// 生成100-500之间的随机增量
	increment := rng.Int63n(401) + 100 // 100-500

	// 计算新ID
	newID := current + increment

	// 更新Redis中的值
	err = g.client.Set(ctx, key, newID, 0).Err()
	if err != nil {
		return 0, err
	}

	return newID, nil
}

// GenerateOperatorID 生成运营ID（50000000000开头，每个加30-50）
func (g *IDGenerator) GenerateOperatorID(ctx context.Context) (int64, error) {
	key := "id:generator:operator"
	
	// 获取当前计数
	current, err := g.client.Get(ctx, key).Int64()
	if err == redis.Nil {
		// 如果不存在，初始化为50000000000
		current = 50000000000
	} else if err != nil {
		return 0, err
	}

	// 生成30-50之间的随机增量
	increment := rng.Int63n(21) + 30 // 30-50

	// 计算新ID
	newID := current + increment

	// 更新Redis中的值
	err = g.client.Set(ctx, key, newID, 0).Err()
	if err != nil {
		return 0, err
	}

	return newID, nil
}

// GeneratePlatformID 生成平台ID（90000000000开头，每个加1）
func (g *IDGenerator) GeneratePlatformID(ctx context.Context) (int64, error) {
	key := "id:generator:platform"
	
	// 获取当前计数
	current, err := g.client.Get(ctx, key).Int64()
	if err == redis.Nil {
		// 如果不存在，初始化为90000000000
		current = 90000000000
	} else if err != nil {
		return 0, err
	}

	// 平台ID每次加1
	newID := current + 1

	// 更新Redis中的值
	err = g.client.Set(ctx, key, newID, 0).Err()
	if err != nil {
		return 0, err
	}

	return newID, nil
}
