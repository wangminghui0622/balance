package utils

import (
	"context"
	"log"
	"time"

	"github.com/redis/go-redis/v9"
)

// Lua脚本：原子性地续期锁（验证所有权后再续期）
var renewLockScript = redis.NewScript(`
	local owner = redis.call('GET', KEYS[1])
	if owner == ARGV[1] then
		return redis.call('PEXPIRE', KEYS[1], ARGV[2])
	end
	return 0
`)

// LockRenewal 锁续期器（Watchdog）
// 在后台定期续期锁，防止长时间任务执行期间锁过期
type LockRenewal struct {
	rdb         *redis.Client
	lockKey     string
	lockValue   string
	ttl         time.Duration
	interval    time.Duration // 续期间隔，通常为TTL的1/3
	stopChan    chan struct{}
	stoppedChan chan struct{}
}

// NewLockRenewal 创建锁续期器
// lockKey: Redis锁的key
// lockValue: 锁的值（用于验证所有权）
// ttl: 锁的过期时间
// rdb: Redis客户端
func NewLockRenewal(rdb *redis.Client, lockKey, lockValue string, ttl time.Duration) *LockRenewal {
	return &LockRenewal{
		rdb:         rdb,
		lockKey:     lockKey,
		lockValue:   lockValue,
		ttl:         ttl,
		interval:    ttl / 3, // 续期间隔为TTL的1/3
		stopChan:    make(chan struct{}),
		stoppedChan: make(chan struct{}),
	}
}

// Start 启动锁续期（在后台goroutine中运行）
func (r *LockRenewal) Start(ctx context.Context) {
	go r.run(ctx)
}

// Stop 停止锁续期（等待goroutine结束）
func (r *LockRenewal) Stop() {
	close(r.stopChan)
	<-r.stoppedChan
}

// run 续期循环
func (r *LockRenewal) run(ctx context.Context) {
	defer close(r.stoppedChan)

	ticker := time.NewTicker(r.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-r.stopChan:
			return
		case <-ticker.C:
			if !r.renew(ctx) {
				log.Printf("[LockRenewal] 锁续期失败，锁可能已被其他节点获取: %s", r.lockKey)
				return
			}
		}
	}
}

// renew 执行一次续期
func (r *LockRenewal) renew(ctx context.Context) bool {
	ttlMs := int64(r.ttl / time.Millisecond)
	result, err := renewLockScript.Run(ctx, r.rdb, []string{r.lockKey}, r.lockValue, ttlMs).Int64()
	if err != nil {
		log.Printf("[LockRenewal] 续期脚本执行失败: %v", err)
		return false
	}
	return result == 1
}

// WithLockRenewal 执行带锁续期的任务
// 这是一个便捷函数，自动管理锁续期的生命周期
// fn: 要执行的任务函数
func WithLockRenewal(ctx context.Context, rdb *redis.Client, lockKey, lockValue string, ttl time.Duration, fn func(ctx context.Context) error) error {
	renewal := NewLockRenewal(rdb, lockKey, lockValue, ttl)
	renewal.Start(ctx)
	defer renewal.Stop()

	return fn(ctx)
}
