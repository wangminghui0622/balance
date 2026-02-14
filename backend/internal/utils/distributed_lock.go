package utils

import (
	"context"
	"log"
	"time"

	"github.com/go-redsync/redsync/v4"
	"github.com/go-redsync/redsync/v4/redis/goredis/v9"
	"github.com/redis/go-redis/v9"
)

// DistributedLock 分布式锁封装
// 基于 redsync 实现，支持单机 Redis 和 Redis 集群
type DistributedLock struct {
	rs *redsync.Redsync
}

// NewDistributedLock 创建分布式锁管理器
func NewDistributedLock(rdb *redis.Client) *DistributedLock {
	pool := goredis.NewPool(rdb)
	rs := redsync.New(pool)
	return &DistributedLock{rs: rs}
}

// NewMutex 创建一个新的互斥锁
// name: 锁的名称（Redis key）
// expiry: 锁的过期时间
func (d *DistributedLock) NewMutex(name string, expiry time.Duration) *redsync.Mutex {
	return d.rs.NewMutex(name,
		redsync.WithExpiry(expiry),
		redsync.WithTries(1), // 只尝试一次，不重试
	)
}

// NewMutexWithRetry 创建一个带重试的互斥锁
// name: 锁的名称
// expiry: 锁的过期时间
// tries: 重试次数
// retryDelay: 重试间隔
func (d *DistributedLock) NewMutexWithRetry(name string, expiry time.Duration, tries int, retryDelay time.Duration) *redsync.Mutex {
	return d.rs.NewMutex(name,
		redsync.WithExpiry(expiry),
		redsync.WithTries(tries),
		redsync.WithRetryDelay(retryDelay),
	)
}

// LockWithAutoExtend 获取锁并自动续期
// 返回一个取消函数，调用它来停止续期并释放锁
func LockWithAutoExtend(ctx context.Context, mutex *redsync.Mutex, extendInterval time.Duration) (cancel func(), err error) {
	// 获取锁
	if err := mutex.LockContext(ctx); err != nil {
		return nil, err
	}

	// 创建停止通道
	stopChan := make(chan struct{})
	doneChan := make(chan struct{})

	// 启动续期 goroutine
	go func() {
		defer close(doneChan)
		ticker := time.NewTicker(extendInterval)
		defer ticker.Stop()

		for {
			select {
			case <-stopChan:
				return
			case <-ctx.Done():
				return
			case <-ticker.C:
				// 续期锁
				ok, err := mutex.ExtendContext(ctx)
				if err != nil || !ok {
					log.Printf("[DistributedLock] 锁续期失败: %v, ok=%v", err, ok)
					return
				}
			}
		}
	}()

	// 返回取消函数
	cancel = func() {
		close(stopChan)
		<-doneChan // 等待续期 goroutine 结束
		if ok, err := mutex.Unlock(); !ok || err != nil {
			log.Printf("[DistributedLock] 释放锁失败: %v, ok=%v", err, ok)
		}
	}

	return cancel, nil
}

// TryLockWithAutoExtend 尝试获取锁并自动续期（不阻塞）
// 如果获取失败，返回 nil, false
func TryLockWithAutoExtend(ctx context.Context, mutex *redsync.Mutex, extendInterval time.Duration) (cancel func(), acquired bool) {
	// 尝试获取锁
	if err := mutex.TryLockContext(ctx); err != nil {
		return nil, false
	}

	// 创建停止通道
	stopChan := make(chan struct{})
	doneChan := make(chan struct{})

	// 启动续期 goroutine
	go func() {
		defer close(doneChan)
		ticker := time.NewTicker(extendInterval)
		defer ticker.Stop()

		for {
			select {
			case <-stopChan:
				return
			case <-ctx.Done():
				return
			case <-ticker.C:
				ok, err := mutex.ExtendContext(ctx)
				if err != nil || !ok {
					log.Printf("[DistributedLock] 锁续期失败: %v, ok=%v", err, ok)
					return
				}
			}
		}
	}()

	cancel = func() {
		close(stopChan)
		<-doneChan
		if ok, err := mutex.Unlock(); !ok || err != nil {
			log.Printf("[DistributedLock] 释放锁失败: %v, ok=%v", err, ok)
		}
	}

	return cancel, true
}
