package services

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"balance/backend/internal/database"
	"balance/backend/internal/services/shopower"
	"balance/backend/internal/utils"

	"github.com/go-redsync/redsync/v4"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

const (
	// Redis Keys
	syncSchedulerLock  = "sync:scheduler:lock"      // 调度器分布式锁
	syncShopQueue      = "sync:shop:queue"          // 待同步店铺队列
	syncShopLockKey    = "sync:shop:lock:%d"        // 店铺级别锁
	syncShopProcessing = "sync:shop:processing"     // 正在处理的店铺集合

	// 配置
	schedulerLockTTL      = 2 * time.Minute   // 调度器锁过期时间（调度任务较快）
	shopLockTTL           = 10 * time.Minute  // 店铺锁过期时间（同步可能耗时较长）
	shopLockExtendInterval = shopLockTTL / 3  // 续期间隔
	workerCount           = 10                // 每台服务器的 Worker 数量
	queuePopTimeout       = 30 * time.Second  // 队列阻塞等待超时
)

// DistributedSyncScheduler 分布式同步调度器
type DistributedSyncScheduler struct {
	db           *gorm.DB
	rdb          *redis.Client
	rs           *redsync.Redsync
	orderService *shopower.OrderService
	shopService  *shopower.ShopService
	interval     time.Duration
	stopChan     chan struct{}
	wg           sync.WaitGroup
	running      bool
	mu           sync.Mutex
}

// NewDistributedSyncScheduler 创建分布式同步调度器
func NewDistributedSyncScheduler(db *gorm.DB, rdb *redis.Client) *DistributedSyncScheduler {
	return &DistributedSyncScheduler{
		db:           db,
		rdb:          rdb,
		rs:           database.GetRedsync(),
		orderService: shopower.NewOrderService(),
		shopService:  shopower.NewShopService(),
		interval:     30 * time.Minute,
		stopChan:     make(chan struct{}),
	}
}

// Start 启动分布式同步
func (s *DistributedSyncScheduler) Start() {
	s.mu.Lock()
	if s.running {
		s.mu.Unlock()
		return
	}
	s.running = true
	s.mu.Unlock()

	// 启动调度器 (尝试成为主节点)
	s.wg.Add(1)
	go s.runScheduler()

	// 启动多个 Worker
	for i := 0; i < workerCount; i++ {
		s.wg.Add(1)
		go s.runWorker(i)
	}

	log.Printf("[DistributedSync] 分布式同步已启动，Worker数量: %d", workerCount)
}

// Stop 停止分布式同步
func (s *DistributedSyncScheduler) Stop() {
	s.mu.Lock()
	if !s.running {
		s.mu.Unlock()
		return
	}
	s.running = false
	s.mu.Unlock()

	close(s.stopChan)
	s.wg.Wait()
	log.Println("[DistributedSync] 分布式同步已停止")
}

// runScheduler 运行调度器 (尝试成为主节点)
func (s *DistributedSyncScheduler) runScheduler() {
	defer s.wg.Done()

	// 启动时立即尝试调度一次
	s.trySchedule()

	ticker := time.NewTicker(s.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			s.trySchedule()
		case <-s.stopChan:
			return
		}
	}
}

// trySchedule 尝试成为主节点并执行调度
func (s *DistributedSyncScheduler) trySchedule() {
	ctx := context.Background()

	// 使用 redsync 创建调度器锁
	mutex := s.rs.NewMutex(syncSchedulerLock,
		redsync.WithExpiry(schedulerLockTTL),
		redsync.WithTries(1),
	)

	// 尝试获取锁
	if err := mutex.TryLockContext(ctx); err != nil {
		log.Println("[DistributedSync] 其他节点正在调度，跳过")
		return
	}
	defer mutex.Unlock()

	log.Println("[DistributedSync] 获取调度器锁成功，开始分配任务...")

	// 获取所有需要同步的店铺
	shops, err := s.shopService.GetAuthorizedShops()
	if err != nil {
		log.Printf("[DistributedSync] 获取店铺列表失败: %v", err)
		return
	}

	if len(shops) == 0 {
		log.Println("[DistributedSync] 没有需要同步的店铺")
		return
	}

	// 将店铺ID推入队列
	pipe := s.rdb.Pipeline()
	for _, shop := range shops {
		if shop.SyncOrders {
			pipe.LPush(ctx, syncShopQueue, shop.ShopID)
		}
	}
	_, err = pipe.Exec(ctx)
	if err != nil {
		log.Printf("[DistributedSync] 推入任务队列失败: %v", err)
		return
	}

	log.Printf("[DistributedSync] 已将 %d 个店铺推入同步队列", len(shops))
}

// runWorker 运行 Worker
func (s *DistributedSyncScheduler) runWorker(workerID int) {
	defer s.wg.Done()

	log.Printf("[Worker-%d] 启动", workerID)

	for {
		select {
		case <-s.stopChan:
			log.Printf("[Worker-%d] 停止", workerID)
			return
		default:
			s.processOneShop(workerID)
		}
	}
}

// processOneShop 处理一个店铺
// 从队列取任务，获取锁，执行同步
func (s *DistributedSyncScheduler) processOneShop(workerID int) {
	ctx := context.Background()

	// 从队列取出一个店铺ID
	result, err := s.rdb.RPop(ctx, syncShopQueue).Result()
	if err != nil {
		if err == redis.Nil {
			// 队列为空，等待一段时间后重试
			time.Sleep(queuePopTimeout)
			return
		}
		log.Printf("[Worker-%d] 从队列取任务失败: %v", workerID, err)
		time.Sleep(time.Second)
		return
	}

	var shopID uint64
	fmt.Sscanf(result, "%d", &shopID)

	// 使用 redsync 创建店铺锁
	lockKey := fmt.Sprintf(syncShopLockKey, shopID)
	mutex := s.rs.NewMutex(lockKey,
		redsync.WithExpiry(shopLockTTL),
		redsync.WithTries(1),
	)

	// 尝试获取锁并自动续期
	unlockFunc, acquired := utils.TryLockWithAutoExtend(ctx, mutex, shopLockExtendInterval)
	if !acquired {
		// 获取锁失败，将任务放回队列头部
		s.rdb.LPush(ctx, syncShopQueue, shopID)
		time.Sleep(100 * time.Millisecond)
		return
	}

	// 标记正在处理
	s.rdb.SAdd(ctx, syncShopProcessing, shopID)

	// 确保释放锁和移除处理标记
	defer func() {
		unlockFunc()
		s.rdb.SRem(ctx, syncShopProcessing, shopID)
	}()

	// 执行增量同步
	s.syncShopOrders(workerID, shopID)
}

// syncShopOrders 同步店铺订单 (增量)
func (s *DistributedSyncScheduler) syncShopOrders(workerID int, shopID uint64) {
	log.Printf("[Worker-%d] 开始同步店铺 %d", workerID, shopID)

	// context超时应小于锁TTL，确保任务在锁过期前完成
	ctx, cancel := context.WithTimeout(context.Background(), 8*time.Minute)
	defer cancel()

	// 获取店铺信息，包括上次同步时间
	shop, err := s.shopService.GetShopByID(shopID)
	if err != nil {
		log.Printf("[Worker-%d] 获取店铺 %d 信息失败: %v", workerID, shopID, err)
		return
	}

	// 增量同步：从上次同步时间开始
	var timeFrom time.Time
	if shop.LastSyncAt != nil && !shop.LastSyncAt.IsZero() {
		timeFrom = *shop.LastSyncAt
	} else {
		// 首次同步，拉取最近30天
		timeFrom = time.Now().Add(-30 * 24 * time.Hour)
	}
	timeTo := time.Now()

	// 调用同步服务
	count, err := s.orderService.SyncOrders(ctx, 0, int64(shopID), timeFrom, timeTo)
	if err != nil {
		log.Printf("[Worker-%d] 店铺 %d 同步失败: %v", workerID, shopID, err)
		return
	}

	// 更新最后同步时间
	s.shopService.UpdateLastSyncTime(shopID)

	log.Printf("[Worker-%d] 店铺 %d 同步完成，新增/更新 %d 条订单", workerID, shopID, count)
}

// GetQueueLength 获取队列长度 (用于监控)
func (s *DistributedSyncScheduler) GetQueueLength() (int64, error) {
	ctx := context.Background()
	return s.rdb.LLen(ctx, syncShopQueue).Result()
}

// GetProcessingCount 获取正在处理的店铺数量 (用于监控)
func (s *DistributedSyncScheduler) GetProcessingCount() (int64, error) {
	ctx := context.Background()
	return s.rdb.SCard(ctx, syncShopProcessing).Result()
}
