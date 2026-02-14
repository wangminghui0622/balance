package shopee

import (
	"context"
	"fmt"
	"sync"
	"time"

	"balance/backend/internal/consts"
	"balance/backend/internal/database"

	"github.com/redis/go-redis/v9"
)

// RateLimiter 限流器
type RateLimiter struct {
	mu       sync.Mutex
	tokens   int
	maxToken int
	interval time.Duration
	lastTime time.Time
}

// NewRateLimiter 创建限流器
func NewRateLimiter(qps int) *RateLimiter {
	return &RateLimiter{
		tokens:   qps,
		maxToken: qps,
		interval: time.Second,
		lastTime: time.Now(),
	}
}

// Allow 检查是否允许请求
func (r *RateLimiter) Allow() bool {
	r.mu.Lock()
	defer r.mu.Unlock()

	now := time.Now()
	elapsed := now.Sub(r.lastTime)

	if elapsed >= r.interval {
		r.tokens = r.maxToken
		r.lastTime = now
	}

	if r.tokens > 0 {
		r.tokens--
		return true
	}
	return false
}

// Wait 等待直到允许请求
func (r *RateLimiter) Wait(ctx context.Context) error {
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			if r.Allow() {
				return nil
			}
			time.Sleep(100 * time.Millisecond)
		}
	}
}

var (
	rateLimiters   = make(map[uint64]*RateLimiter)
	rateLimitersMu sync.RWMutex
)

// GetRateLimiter 获取店铺的限流器
func GetRateLimiter(shopID uint64) *RateLimiter {
	rateLimitersMu.RLock()
	limiter, exists := rateLimiters[shopID]
	rateLimitersMu.RUnlock()

	if exists {
		return limiter
	}

	rateLimitersMu.Lock()
	defer rateLimitersMu.Unlock()

	if limiter, exists = rateLimiters[shopID]; exists {
		return limiter
	}

	limiter = NewRateLimiter(consts.ShopeeAPIRateLimit)
	rateLimiters[shopID] = limiter
	return limiter
}

// RetryWithBackoff 带退避的重试
func RetryWithBackoff(ctx context.Context, maxRetries int, fn func() error) error {
	var lastErr error
	for i := 0; i < maxRetries; i++ {
		if err := fn(); err != nil {
			lastErr = err
			if isRateLimitError(err) {
				waitTime := time.Duration(consts.ShopeeAPIRetryInterval*(1<<i)) * time.Millisecond
				select {
				case <-ctx.Done():
					return ctx.Err()
				case <-time.After(waitTime):
					continue
				}
			}
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(time.Duration(consts.ShopeeAPIRetryInterval) * time.Millisecond):
				continue
			}
		}
		return nil
	}
	return fmt.Errorf("重试%d次后仍然失败: %w", maxRetries, lastErr)
}

func isRateLimitError(err error) bool {
	if err == nil {
		return false
	}
	errStr := err.Error()
	return contains(errStr, "error.too_many_request") ||
		contains(errStr, "rate limit") ||
		contains(errStr, "429")
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsImpl(s, substr))
}

func containsImpl(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

// rateLimitScript Lua脚本：原子性地增加计数并设置过期时间
const rateLimitScript = `
	local count = redis.call('INCR', KEYS[1])
	if count == 1 then
		redis.call('EXPIRE', KEYS[1], ARGV[1])
	end
	return count
`

// CheckRateLimit 检查并记录API调用频率（使用Lua脚本保证原子性）
func CheckRateLimit(ctx context.Context, shopID uint64, apiName string) error {
	rdb := database.GetRedis()
	key := fmt.Sprintf(consts.KeyRateLimit, shopID, apiName)

	script := redis.NewScript(rateLimitScript)
	count, err := script.Run(ctx, rdb, []string{key}, int(consts.RateLimitExpire.Seconds())).Int64()
	if err != nil {
		return fmt.Errorf("限流检查失败: %w", err)
	}

	if count > int64(consts.ShopeeAPIRateLimit*60) {
		return fmt.Errorf("API调用频率超限，请稍后再试")
	}

	return nil
}
