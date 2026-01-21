package utils

import (
	"balance/internal/constants"
	"context"
	"fmt"
	"math/rand"
	"sync"
	"time"

	"github.com/redis/go-redis/v9"
)

var rng = rand.New(rand.NewSource(time.Now().UnixNano()))

// generateTimestampID 生成时间戳ID（降级方案）
// 格式：8 + 毫秒时间戳（12位） + 随机数（0-999，3位）
// 例如：81705747200000123 = 8 + 1705747200000 + 123
func generateTimestampID() int64 {
	// 获取当前毫秒时间戳
	timestamp := time.Now().UnixMilli()
	
	// 生成0-999的随机数，避免同一毫秒内的冲突
	random := rng.Int63n(constants.FallbackRandomMax + 1)
	
	// 组合：8 + timestamp + random
	// 8是前缀，timestamp是12位，random是3位
	id := constants.FallbackIDPrefix + timestamp*1000 + random
	
	return id
}

// Lua脚本：原子性地生成ID
// KEYS[1]: Redis key
// ARGV[1]: 初始值
// ARGV[2]: 增量
// 返回值：新生成的ID
const generateIDScript = `
local current = redis.call('GET', KEYS[1])
if current == false then
	-- key不存在，设置初始值并返回初始值本身
	redis.call('SET', KEYS[1], ARGV[1])
	return tonumber(ARGV[1])
else
	-- key存在，读取当前值，加上增量，更新并返回
	local newID = tonumber(current) + tonumber(ARGV[2])
	redis.call('SET', KEYS[1], newID)
	return newID
end
`

// IDGenerator ID生成器
type IDGenerator struct {
	client        *redis.Client
	generateIDSha string    // Lua脚本的SHA1哈希，用于缓存
	shaMu         sync.Mutex // 保护generateIDSha的互斥锁
}

// NewIDGenerator 创建ID生成器
func NewIDGenerator(client *redis.Client) *IDGenerator {
	gen := &IDGenerator{client: client}
	// 预加载Lua脚本并缓存SHA1
	sha, err := client.ScriptLoad(context.Background(), generateIDScript).Result()
	if err == nil {
		gen.generateIDSha = sha
	}
	return gen
}

// generateIDWithLua 使用Lua脚本原子性地生成ID
func (g *IDGenerator) generateIDWithLua(ctx context.Context, key string, initialValue int64, increment int64) (int64, error) {
	var result interface{}
	var err error

	// 优先使用缓存的SHA1执行脚本（更快）
	g.shaMu.Lock()
	cachedSha := g.generateIDSha
	g.shaMu.Unlock()

	if cachedSha != "" {
		result, err = g.client.EvalSha(ctx, cachedSha, []string{key}, initialValue, increment).Result()
		// 如果脚本不存在（可能Redis重启），清除缓存的SHA1
		if err != nil && err.Error() == "NOSCRIPT No matching script. Use EVAL." {
			g.shaMu.Lock()
			g.generateIDSha = ""
			g.shaMu.Unlock()
		}
		// 如果成功，直接返回
		if err == nil {
			id, ok := result.(int64)
			if !ok {
				return 0, fmt.Errorf("unexpected result type: %T", result)
			}
			return id, nil
		}
	}

	// 如果没有缓存的SHA1或执行失败，使用脚本内容
	result, err = g.client.Eval(ctx, generateIDScript, []string{key}, initialValue, increment).Result()
	if err != nil {
		// Redis失败时，使用时间戳ID作为降级方案
		return generateTimestampID(), nil
	}

	// 如果成功，尝试缓存SHA1（用于下次调用）
	g.shaMu.Lock()
	if g.generateIDSha == "" {
		sha, shaErr := g.client.ScriptLoad(ctx, generateIDScript).Result()
		if shaErr == nil {
			g.generateIDSha = sha
		}
	}
	g.shaMu.Unlock()

	// 转换结果为int64
	id, ok := result.(int64)
	if !ok {
		return 0, fmt.Errorf("unexpected result type: %T", result)
	}

	return id, nil
}

// GenerateShopOwnerID 生成店铺ID（10000000000开头，每个加100-500）
func (g *IDGenerator) GenerateShopOwnerID(ctx context.Context) (int64, error) {
	key := "id:generator:shopowner"
	
	// 生成100-500之间的随机增量
	// 计算范围：max - min + 1 = 500 - 100 + 1 = 401
	rangeSize := constants.IDIncrementShopOwnerMax - constants.IDIncrementShopOwnerMin + 1
	increment := rng.Int63n(rangeSize) + constants.IDIncrementShopOwnerMin

	// 使用Lua脚本原子性地生成ID
	return g.generateIDWithLua(ctx, key, constants.IDInitialShopOwner, increment)
}

// GenerateOperatorID 生成运营ID（50000000000开头，每个加30-50）
func (g *IDGenerator) GenerateOperatorID(ctx context.Context) (int64, error) {
	key := "id:generator:operator"
	
	// 生成30-50之间的随机增量
	// 计算范围：max - min + 1 = 50 - 30 + 1 = 21
	rangeSize := constants.IDIncrementOperatorMax - constants.IDIncrementOperatorMin + 1
	increment := rng.Int63n(rangeSize) + constants.IDIncrementOperatorMin

	// 使用Lua脚本原子性地生成ID
	return g.generateIDWithLua(ctx, key, constants.IDInitialOperator, increment)
}

// GeneratePlatformID 生成平台ID（90000000000开头，每个加10-20）
func (g *IDGenerator) GeneratePlatformID(ctx context.Context) (int64, error) {
	key := "id:generator:platform"
	
	// 生成10-20之间的随机增量
	// 计算范围：max - min + 1 = 20 - 10 + 1 = 11
	rangeSize := constants.IDIncrementPlatformMax - constants.IDIncrementPlatformMin + 1
	increment := rng.Int63n(rangeSize) + constants.IDIncrementPlatformMin

	// 使用Lua脚本原子性地生成ID
	return g.generateIDWithLua(ctx, key, constants.IDInitialPlatform, increment)
}
