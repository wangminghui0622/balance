package utils

import (
	"context"
	"fmt"
	"math/rand"
	"sync"
	"time"

	"github.com/redis/go-redis/v9"
)

var rng = rand.New(rand.NewSource(time.Now().UnixNano()))

const (
	IDInitialShopOwner int64 = 19906070668
	IDInitialOperator  int64 = 58608109796
	IDInitialPlatform  int64 = 91609051906

	IDIncrementShopOwnerMin int64 = 100
	IDIncrementShopOwnerMax int64 = 500

	IDIncrementOperatorMin int64 = 30
	IDIncrementOperatorMax int64 = 50

	IDIncrementPlatformMin int64 = 10
	IDIncrementPlatformMax int64 = 20

	FallbackIDPrefix  int64 = 8
	FallbackRandomMax int64 = 999
)

const generateIDScript = `
local current = redis.call('GET', KEYS[1])
if current == false then
	redis.call('SET', KEYS[1], ARGV[1])
	return tonumber(ARGV[1])
else
	local newID = tonumber(current) + tonumber(ARGV[2])
	redis.call('SET', KEYS[1], newID)
	return newID
end
`

// IDGenerator ID生成器
type IDGenerator struct {
	client        *redis.Client
	generateIDSha string
	shaMu         sync.Mutex
}

// NewIDGenerator 创建ID生成器
func NewIDGenerator(client *redis.Client) *IDGenerator {
	gen := &IDGenerator{client: client}
	sha, err := client.ScriptLoad(context.Background(), generateIDScript).Result()
	if err == nil {
		gen.generateIDSha = sha
	}
	return gen
}

func generateTimestampID() int64 {
	timestamp := time.Now().UnixMilli()
	random := rng.Int63n(FallbackRandomMax + 1)
	return FallbackIDPrefix + timestamp*1000 + random
}

func (g *IDGenerator) generateIDWithLua(ctx context.Context, key string, initialValue int64, increment int64) (int64, error) {
	var result interface{}
	var err error

	g.shaMu.Lock()
	cachedSha := g.generateIDSha
	g.shaMu.Unlock()

	if cachedSha != "" {
		result, err = g.client.EvalSha(ctx, cachedSha, []string{key}, initialValue, increment).Result()
		if err != nil && err.Error() == "NOSCRIPT No matching script. Use EVAL." {
			g.shaMu.Lock()
			g.generateIDSha = ""
			g.shaMu.Unlock()
		}
		if err == nil {
			id, ok := result.(int64)
			if !ok {
				return 0, fmt.Errorf("unexpected result type: %T", result)
			}
			return id, nil
		}
	}

	result, err = g.client.Eval(ctx, generateIDScript, []string{key}, initialValue, increment).Result()
	if err != nil {
		return generateTimestampID(), nil
	}

	g.shaMu.Lock()
	if g.generateIDSha == "" {
		sha, shaErr := g.client.ScriptLoad(ctx, generateIDScript).Result()
		if shaErr == nil {
			g.generateIDSha = sha
		}
	}
	g.shaMu.Unlock()

	id, ok := result.(int64)
	if !ok {
		return 0, fmt.Errorf("unexpected result type: %T", result)
	}

	return id, nil
}

// GenerateShopOwnerID 生成店主ID
func (g *IDGenerator) GenerateShopOwnerID(ctx context.Context) (int64, error) {
	key := "id:generator:shopowner"
	rangeSize := IDIncrementShopOwnerMax - IDIncrementShopOwnerMin + 1
	increment := rng.Int63n(rangeSize) + IDIncrementShopOwnerMin
	return g.generateIDWithLua(ctx, key, IDInitialShopOwner, increment)
}

// GenerateOperatorID 生成运营ID
func (g *IDGenerator) GenerateOperatorID(ctx context.Context) (int64, error) {
	key := "id:generator:operator"
	rangeSize := IDIncrementOperatorMax - IDIncrementOperatorMin + 1
	increment := rng.Int63n(rangeSize) + IDIncrementOperatorMin
	return g.generateIDWithLua(ctx, key, IDInitialOperator, increment)
}

// GeneratePlatformID 生成平台ID
func (g *IDGenerator) GeneratePlatformID(ctx context.Context) (int64, error) {
	key := "id:generator:platform"
	rangeSize := IDIncrementPlatformMax - IDIncrementPlatformMin + 1
	increment := rng.Int63n(rangeSize) + IDIncrementPlatformMin
	return g.generateIDWithLua(ctx, key, IDInitialPlatform, increment)
}

// GenerateUserNo 生成用户编号
func GenerateUserNo(id int64) string {
	return fmt.Sprintf("U%011d", id)
}
