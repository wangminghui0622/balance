package utils

import (
	"context"
	"log"
	"time"

	"github.com/go-redsync/redsync/v4"
)

// TryLockWithAutoExtend 尝试获取锁并自动续期（不阻塞）
// 如果获取失败，返回 nil, false
// 使用方式：
//
//	unlockFunc, acquired := TryLockWithAutoExtend(ctx, mutex, ttl/3)
//	if !acquired { return }
//	defer unlockFunc()
func TryLockWithAutoExtend(ctx context.Context, mutex *redsync.Mutex, extendInterval time.Duration) (cancel func(), acquired bool) {
	if err := mutex.TryLockContext(ctx); err != nil {
		return nil, false
	}

	stopChan := make(chan struct{})
	doneChan := make(chan struct{})

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
