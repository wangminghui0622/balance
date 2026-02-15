package services

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"balance/backend/internal/consts"
	"balance/backend/internal/database"
	"balance/backend/internal/services/shopower"
	"balance/backend/internal/utils"

	"github.com/go-redsync/redsync/v4"
	"github.com/panjf2000/ants/v2"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

const (
	// Redis Keys
	syncSchedulerLock  = "sync:scheduler:lock"  // 调度器分布式锁
	syncShopQueue      = "sync:shop:queue"      // 待同步店铺队列
	syncShopProcessing = "sync:shop:processing" // 正在处理的店铺集合
	// 注意：店铺同步锁使用 consts.KeySyncLock，与手动同步共用同一把锁

	// 配置
	schedulerLockTTL = 2 * time.Minute  // 调度器锁过期时间（调度任务较快）
	workerCount      = 16               // 每台服务器的 Worker 数量
	queuePopTimeout  = 30 * time.Second // 队列阻塞等待超时
)

// DistributedSyncScheduler 分布式同步调度器
type DistributedSyncScheduler struct {
	db           *gorm.DB
	rdb          *redis.Client
	rs           *redsync.Redsync
	orderService *shopower.OrderService
	shopService  *shopower.ShopService
	pool         *ants.Pool
	interval     time.Duration
	stopChan     chan struct{}
	wg           sync.WaitGroup
	running      bool
	mu           sync.Mutex
	logger       *log.Logger // 文件日志
}

// NewDistributedSyncScheduler 创建分布式同步调度器
// logger: 可选的文件日志，传nil则使用标准log
func NewDistributedSyncScheduler(db *gorm.DB, rdb *redis.Client, logger ...*log.Logger) *DistributedSyncScheduler {
	// 使用传入的logger或默认标准log
	var l *log.Logger
	if len(logger) > 0 && logger[0] != nil {
		l = logger[0]
	} else {
		l = log.Default()
	}

	// 创建 ants 协程池
	pool, err := ants.NewPool(workerCount,
		ants.WithPreAlloc(true),
		ants.WithPanicHandler(func(p interface{}) {
			l.Printf("[DistributedSync] Worker panic: %v", p)
		}),
	)
	if err != nil {
		log.Fatalf("[DistributedSync] 创建协程池失败: %v", err)
	}

	return &DistributedSyncScheduler{
		db:           db,
		rdb:          rdb,
		rs:           database.GetRedsync(),
		orderService: shopower.NewOrderService(),
		shopService:  shopower.NewShopService(),
		pool:         pool,
		interval:     30 * time.Minute,
		stopChan:     make(chan struct{}),
		logger:       l,
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

	// 启动任务分发器（从队列取任务并提交到协程池）
	s.wg.Add(1)
	go s.runDispatcher()

	s.logger.Printf("[DistributedSync] 分布式同步已启动，协程池大小: %d", workerCount)
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

	// 释放协程池
	s.pool.Release()

	s.logger.Println("[DistributedSync] 分布式同步已停止")
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
		s.logger.Println("[DistributedSync] 其他节点正在调度，跳过")
		return
	}
	defer mutex.Unlock()

	s.logger.Println("[DistributedSync] 获取调度器锁成功，开始分配任务...")

	// 获取所有需要同步的店铺
	shops, err := s.shopService.GetAuthorizedShops()
	if err != nil {
		s.logger.Printf("[DistributedSync] 获取店铺列表失败: %v", err)
		return
	}

	if len(shops) == 0 {
		s.logger.Println("[DistributedSync] 没有需要同步的店铺")
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
		s.logger.Printf("[DistributedSync] 推入任务队列失败: %v", err)
		return
	}

	s.logger.Printf("[DistributedSync] 已将 %d 个店铺推入同步队列", len(shops))
}

// runDispatcher 任务分发器：从队列取任务并提交到协程池
func (s *DistributedSyncScheduler) runDispatcher() {
	defer s.wg.Done()

	s.logger.Println("[Dispatcher] 启动")

	for {
		select {
		case <-s.stopChan:
			s.logger.Println("[Dispatcher] 停止")
			return
		default:
			s.dispatchOneTask()
		}
	}
}

// dispatchOneTask 从队列取一个任务并提交到协程池
func (s *DistributedSyncScheduler) dispatchOneTask() {
	ctx := context.Background()

	// 从队列取出一个店铺ID
	result, err := s.rdb.RPop(ctx, syncShopQueue).Result()
	if err != nil {
		if err == redis.Nil {
			// 队列为空，等待一段时间后重试
			time.Sleep(queuePopTimeout)
			return
		}
		s.logger.Printf("[Dispatcher] 从队列取任务失败: %v", err)
		time.Sleep(time.Second)
		return
	}

	var shopID uint64
	fmt.Sscanf(result, "%d", &shopID)

	// 提交任务到协程池
	err = s.pool.Submit(func() {
		s.processShop(shopID)
	})
	if err != nil {
		// 提交失败，将任务放回队列
		s.logger.Printf("[Dispatcher] 提交任务到协程池失败: %v", err)
		s.rdb.LPush(ctx, syncShopQueue, shopID)
		time.Sleep(100 * time.Millisecond)
	}
}

// processShop 处理一个店铺（在协程池中执行）
func (s *DistributedSyncScheduler) processShop(shopID uint64) {
	ctx := context.Background()

	// 使用 redsync 创建店铺锁（与手动同步共用同一把锁，确保互斥）
	lockKey := fmt.Sprintf(consts.KeySyncLock, shopID)
	mutex := s.rs.NewMutex(lockKey,
		redsync.WithExpiry(consts.SyncLockExpire),
		redsync.WithTries(1),
	)

	// 尝试获取锁并自动续期
	unlockFunc, acquired := utils.TryLockWithAutoExtend(ctx, mutex, consts.SyncLockExpire/3)
	if !acquired {
		// 获取锁失败（可能手动同步正在进行），将任务放回队列头部
		s.rdb.LPush(ctx, syncShopQueue, shopID)
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
	s.syncShopOrders(shopID)
}

// 订单同步任务超时时间（应小于锁TTL，确保任务在锁过期前完成）
const orderSyncTaskTimeout = 8 * time.Minute

// syncShopOrders 同步店铺订单 (增量)
func (s *DistributedSyncScheduler) syncShopOrders(shopID uint64) {
	startTime := time.Now()
	s.logger.Printf("[Sync] 开始同步店铺 %d", shopID)

	// 创建任务级超时 context
	ctx, cancel := context.WithTimeout(context.Background(), orderSyncTaskTimeout)
	defer cancel()

	// 获取店铺信息，包括上次同步时间
	shop, err := s.shopService.GetShopByID(shopID)
	if err != nil {
		s.logger.Printf("[Sync] 获取店铺 %d 信息失败: %v", shopID, err)
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

	count, err := s.orderService.SyncOrdersWithShopNoLock(ctx, shop, timeFrom, timeTo)
	elapsed := time.Since(startTime)

	if err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			s.logger.Printf("[Sync] 店铺 %d 同步任务超时（耗时 %v，超时限制 %v）", shopID, elapsed, orderSyncTaskTimeout)
		} else {
			s.logger.Printf("[Sync] 店铺 %d 同步失败（耗时 %v）: %v", shopID, elapsed, err)
		}
		return
	}
	s.logger.Printf("[Sync] 店铺 %d 同步完成，新增/更新 %d 条订单（耗时 %v）", shopID, count, elapsed)
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
