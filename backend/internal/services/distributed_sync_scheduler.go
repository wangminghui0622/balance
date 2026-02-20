package services

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"balance/backend/internal/consts"
	"balance/backend/internal/database"
	"balance/backend/internal/models"
	"balance/backend/internal/services/shopower"
	"balance/backend/internal/utils"

	"github.com/go-redsync/redsync/v4"
	"github.com/panjf2000/ants/v2"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
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
	db            *gorm.DB
	rdb           *redis.Client
	rs            *redsync.Redsync
	orderService  *shopower.OrderService
	shopService   *shopower.ShopService
	returnService *ReturnService
	pool          *ants.Pool
	interval      time.Duration
	stopChan      chan struct{}
	wg            sync.WaitGroup
	running       bool
	mu            sync.Mutex
	logger        *zap.SugaredLogger
}

// NewDistributedSyncScheduler 创建分布式同步调度器
// logger: 可选的 zap SugaredLogger，传nil则使用默认标准输出
func NewDistributedSyncScheduler(db *gorm.DB, rdb *redis.Client, logger ...*zap.SugaredLogger) *DistributedSyncScheduler {
	var l *zap.SugaredLogger
	if len(logger) > 0 && logger[0] != nil {
		l = logger[0]
	} else {
		l = utils.DefaultSugaredLogger()
	}

	// 创建 ants 协程池
	pool, err := ants.NewPool(workerCount,
		ants.WithPreAlloc(true),
		ants.WithPanicHandler(func(p interface{}) {
			l.Errorf("[DistributedSync] Worker panic: %v", p)
		}),
	)
	if err != nil {
		l.Fatalf("[DistributedSync] 创建协程池失败: %v", err)
	}

	return &DistributedSyncScheduler{
		db:            db,
		rdb:           rdb,
		rs:            database.GetRedsync(),
		orderService:  NewOrderServiceWithPrepaymentCheck(),
		shopService:   shopower.NewShopService(),
		returnService: NewReturnService(),
		pool:          pool,
		interval:      30 * time.Minute,
		stopChan:      make(chan struct{}),
		logger:        l,
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

	s.logger.Infof("[DistributedSync] 分布式巡检已启动（间隔 %v），协程池大小: %d", s.interval, workerCount)
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

	s.logger.Info("[DistributedSync] 分布式巡检已停止")
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
		s.logger.Info("[DistributedSync] 其他节点正在调度，跳过")
		return
	}
	defer mutex.Unlock()

	s.logger.Info("[DistributedSync] 获取调度器锁成功，开始分配任务...")

	// 获取所有需要同步的店铺
	shops, err := s.shopService.GetAuthorizedShops()
	if err != nil {
		s.logger.Errorf("[DistributedSync] 获取店铺列表失败: %v", err)
		return
	}

	if len(shops) == 0 {
		s.logger.Info("[DistributedSync] 没有需要同步的店铺")
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
		s.logger.Errorf("[DistributedSync] 推入任务队列失败: %v", err)
		return
	}

	s.logger.Infof("[DistributedSync] 已将 %d 个店铺推入同步队列", len(shops))
}

// runDispatcher 任务分发器：从队列取任务并提交到协程池
func (s *DistributedSyncScheduler) runDispatcher() {
	defer s.wg.Done()

	// 创建可取消的 context，用于中断 BRPop 阻塞
	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		<-s.stopChan
		cancel()
	}()

	s.logger.Info("[Dispatcher] 启动")

	for {
		select {
		case <-ctx.Done():
			s.logger.Info("[Dispatcher] 停止")
			return
		default:
			s.dispatchOneTask(ctx)
		}
	}
}

// dispatchOneTask 从队列取一个任务并提交到协程池
func (s *DistributedSyncScheduler) dispatchOneTask(ctx context.Context) {
	// 从队列阻塞取出一个店铺ID（有新任务时立即返回，无需轮询等待）
	results, err := s.rdb.BRPop(ctx, queuePopTimeout, syncShopQueue).Result()
	if err != nil {
		if err == redis.Nil || ctx.Err() != nil {
			// 超时或 context 被取消，直接返回
			return
		}
		s.logger.Errorf("[Dispatcher] 从队列取任务失败: %v", err)
		time.Sleep(time.Second)
		return
	}

	// BRPop 返回 [key, value]
	var shopID uint64
	fmt.Sscanf(results[1], "%d", &shopID)

	// 提交任务到协程池
	err = s.pool.Submit(func() {
		s.processShop(shopID)
	})
	if err != nil {
		// 提交失败，将任务放回队列
		s.logger.Errorf("[Dispatcher] 提交任务到协程池失败: %v", err)
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

	// 执行巡检
	s.syncShopOrders(shopID)
}

// 订单同步任务超时时间（应小于锁TTL，确保任务在锁过期前完成）
const orderSyncTaskTimeout = 8 * time.Minute

// returnPatrolTimeout 退货巡检超时
const returnPatrolTimeout = 3 * time.Minute

// defaultPatrolLookback 无上次同步时回溯天数
const defaultPatrolLookback = 30

// syncShopOrders 巡检单个店铺：订单补录 + 退货退款
// 从 Shopee 拉取订单列表与本地 DB 比对，发现遗漏或状态不一致时补录；然后同步退货退款
func (s *DistributedSyncScheduler) syncShopOrders(shopID uint64) {
	startTime := time.Now()
	s.logger.Infof("[Patrol] 开始巡检店铺 %d", shopID)

	shop, ok := s.getShopForPatrol(shopID)
	if !ok {
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), orderSyncTaskTimeout)
	defer cancel()

	begin, end := s.determinePatrolTimeRange(shop)
	//timeFrom(LastSyncAt或者当前时刻的前30天)   timeTo(当前时刻)
	found, patched, err := s.orderService.PatrolOrders(ctx, shop, begin, end)
	elapsed := time.Since(startTime)

	if !s.logOrderPatrolResult(shopID, found, patched, elapsed, err) {
		return
	}

	s.runReturnPatrol(shopID, shop)
}

// getShopForPatrol 获取店铺信息，失败时打日志并返回 false
func (s *DistributedSyncScheduler) getShopForPatrol(shopID uint64) (*models.Shop, bool) {
	shop, err := s.shopService.GetShopByID(shopID)
	if err != nil {
		s.logger.Errorf("[Patrol] 获取店铺 %d 信息失败: %v", shopID, err)
		return nil, false
	}
	return shop, true
}

// determinePatrolTimeRange 确定巡检时间范围：优先上次同步时间，否则回溯默认天数
func (s *DistributedSyncScheduler) determinePatrolTimeRange(shop *models.Shop) (timeFrom, timeTo time.Time) {
	if shop.LastSyncAt != nil && !shop.LastSyncAt.IsZero() {
		timeFrom = *shop.LastSyncAt
	} else {
		timeFrom = time.Now().Add(-defaultPatrolLookback * 24 * time.Hour)
	}
	timeTo = time.Now()
	return timeFrom, timeTo
}

// logOrderPatrolResult 记录订单巡检结果，err 时返回 false（调用方应中止后续步骤）
func (s *DistributedSyncScheduler) logOrderPatrolResult(shopID uint64, found, patched int, elapsed time.Duration, err error) bool {
	if err != nil {
		if errors.Is(err, context.DeadlineExceeded) {
			s.logger.Errorf("[Patrol] 店铺 %d 订单巡检超时（耗时 %v）", shopID, elapsed)
		} else {
			s.logger.Errorf("[Patrol] 店铺 %d 订单巡检失败（耗时 %v）: %v", shopID, elapsed, err)
		}
		return false
	}
	if patched > 0 {
		s.logger.Infof("[Patrol] 店铺 %d 订单巡检完成，Shopee共 %d 条，补录 %d 条（耗时 %v）", shopID, found, patched, elapsed)
	} else {
		s.logger.Infof("[Patrol] 店铺 %d 订单巡检完成，Shopee共 %d 条，无遗漏（耗时 %v）", shopID, found, elapsed)
	}
	return true
}

// runReturnPatrol 执行退货退款巡检
func (s *DistributedSyncScheduler) runReturnPatrol(shopID uint64, shop *models.Shop) {
	ctx, cancel := context.WithTimeout(context.Background(), returnPatrolTimeout)
	defer cancel()

	accessToken, err := s.shopService.GetAccessToken(ctx, shopID)
	if err != nil {
		s.logger.Warnf("[Patrol] 店铺 %d 无有效授权，跳过退货巡检: %v", shopID, err)
		return
	}

	if err := s.returnService.SyncReturns(ctx, shopID, accessToken, shop.Region); err != nil {
		s.logger.Errorf("[Patrol] 店铺 %d 退货巡检失败: %v", shopID, err)
		return
	}
	s.logger.Infof("[Patrol] 店铺 %d 退货巡检完成", shopID)
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
