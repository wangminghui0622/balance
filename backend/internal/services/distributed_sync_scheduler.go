package services

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"balance/backend/internal/services/shopower"

	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

const (
	// Redis Keys
	syncSchedulerLock = "sync:scheduler:lock"       // 调度器分布式锁
	syncShopQueue     = "sync:shop:queue"           // 待同步店铺队列
	syncShopLockKey   = "sync:shop:lock:%d"         // 店铺级别锁
	syncShopProcessing = "sync:shop:processing"     // 正在处理的店铺集合

	// 配置
	schedulerLockTTL  = 60 * time.Second   // 调度器锁过期时间
	shopLockTTL       = 5 * time.Minute    // 店铺锁过期时间
	workerCount       = 10                 // 每台服务器的 Worker 数量
	queuePopTimeout   = 30 * time.Second   // 队列阻塞等待超时
)

// DistributedSyncScheduler 分布式同步调度器
type DistributedSyncScheduler struct {
	db           *gorm.DB
	rdb          *redis.Client
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

	// 尝试获取调度器锁 (只有一个节点能获取)
	ok, err := s.rdb.SetNX(ctx, syncSchedulerLock, "1", schedulerLockTTL).Result()
	if err != nil {
		log.Printf("[DistributedSync] 获取调度器锁失败: %v", err)
		return
	}
	if !ok {
		// 其他节点已经在调度，跳过
		log.Println("[DistributedSync] 其他节点正在调度，跳过")
		return
	}

	defer s.rdb.Del(ctx, syncSchedulerLock)

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
func (s *DistributedSyncScheduler) processOneShop(workerID int) {
	ctx := context.Background()

	// 从队列中阻塞获取任务
	result, err := s.rdb.BRPop(ctx, queuePopTimeout, syncShopQueue).Result()
	if err != nil {
		if err == redis.Nil {
			// 超时，没有任务，继续等待
			return
		}
		log.Printf("[Worker-%d] 获取任务失败: %v", workerID, err)
		return
	}

	// result[0] 是 key，result[1] 是 value
	shopIDStr := result[1]
	var shopID uint64
	fmt.Sscanf(shopIDStr, "%d", &shopID)

	// 尝试获取店铺锁
	lockKey := fmt.Sprintf(syncShopLockKey, shopID)
	ok, err := s.rdb.SetNX(ctx, lockKey, "1", shopLockTTL).Result()
	if err != nil {
		log.Printf("[Worker-%d] 获取店铺锁失败: %v", workerID, err)
		return
	}
	if !ok {
		// 其他 Worker 正在处理这个店铺，跳过
		log.Printf("[Worker-%d] 店铺 %d 正在被其他Worker处理，跳过", workerID, shopID)
		return
	}

	// 确保释放锁
	defer s.rdb.Del(ctx, lockKey)

	// 标记正在处理
	s.rdb.SAdd(ctx, syncShopProcessing, shopID)
	defer s.rdb.SRem(ctx, syncShopProcessing, shopID)

	// 执行增量同步
	s.syncShopOrders(workerID, shopID)
}

// syncShopOrders 同步店铺订单 (增量)
func (s *DistributedSyncScheduler) syncShopOrders(workerID int, shopID uint64) {
	log.Printf("[Worker-%d] 开始同步店铺 %d", workerID, shopID)

	ctx := context.Background()

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
