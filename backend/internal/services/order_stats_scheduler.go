package services

import (
	"context"
	"time"

	"balance/backend/internal/database"
	"balance/backend/internal/models"
	"balance/backend/internal/services/shopower"
	"balance/backend/internal/utils"

	"github.com/go-redsync/redsync/v4"
	"github.com/robfig/cron/v3"
	"go.uber.org/zap"
)

const (
	orderStatsLockKey = "cron:order_stats_refresh"
	orderStatsLockTTL = 2 * time.Minute
)

// OrderStatsScheduler 店主订单统计缓存刷新调度器（多机部署时通过分布式锁保证单机执行）
type OrderStatsScheduler struct {
	cron         *cron.Cron
	orderService *shopower.OrderService
	rs           *redsync.Redsync
	logger       *zap.SugaredLogger
}

// NewOrderStatsScheduler 创建店主订单统计调度器
func NewOrderStatsScheduler(logger ...*zap.SugaredLogger) *OrderStatsScheduler {
	var l *zap.SugaredLogger
	if len(logger) > 0 && logger[0] != nil {
		l = logger[0]
	} else {
		l = utils.DefaultSugaredLogger()
	}
	return &OrderStatsScheduler{
		cron:         cron.New(cron.WithSeconds()),
		orderService: NewOrderServiceWithPrepaymentCheck(),
		rs:           database.GetRedsync(),
		logger:       l,
	}
}

// tryRunWithLock 尝试获取分布式锁后执行，获取失败则跳过（其他节点正在执行）
func (s *OrderStatsScheduler) tryRunWithLock(fn func()) {
	mutex := s.rs.NewMutex(orderStatsLockKey, redsync.WithExpiry(orderStatsLockTTL), redsync.WithTries(1))
	if err := mutex.LockContext(context.Background()); err != nil {
		s.logger.Info("[OrderStats] 其他节点正在刷新，跳过")
		return
	}
	defer mutex.Unlock()
	fn()
}

// refresh 执行店主订单统计缓存刷新
func (s *OrderStatsScheduler) refresh() {
	ctx := context.Background()
	var adminIDs []int64
	if err := database.GetDB().Model(&models.Shop{}).Distinct("admin_id").Pluck("admin_id", &adminIDs).Error; err != nil {
		s.logger.Errorf("[OrderStats] 获取 admin_id 失败: %v", err)
		return
	}
	for _, adminID := range adminIDs {
		stats, err := s.orderService.ComputeOrderStats(ctx, adminID)
		if err != nil {
			s.logger.Errorf("[OrderStats] admin_id=%d 计算统计失败: %v", adminID, err)
			continue
		}
		if err := s.orderService.SetOrderStatsCache(ctx, adminID, stats); err != nil {
			s.logger.Errorf("[OrderStats] admin_id=%d 写缓存失败: %v", adminID, err)
		}
	}
	s.logger.Info("[OrderStats] 店主订单统计缓存刷新完成")
}

// Start 启动店主订单统计调度器
func (s *OrderStatsScheduler) Start() {
	s.logger.Info("[OrderStats] 启动店主订单统计调度器...")

	// 每小时整点刷新一次（分布式锁）
	_, err := s.cron.AddFunc("0 0 * * * *", func() {
		s.tryRunWithLock(s.refresh)
	})
	if err != nil {
		s.logger.Errorf("[OrderStats] 添加定时任务失败: %v", err)
		return
	}

	s.cron.Start()
	s.logger.Info("[OrderStats] 店主订单统计调度器已启动")

	// 启动后延迟 30 秒执行一次初始刷新（带分布式锁）
	go func() {
		time.Sleep(30 * time.Second)
		s.tryRunWithLock(s.refresh)
	}()
}

// Stop 停止店主订单统计调度器
func (s *OrderStatsScheduler) Stop() {
	s.logger.Info("[OrderStats] 停止店主订单统计调度器...")
	s.cron.Stop()
	s.logger.Info("[OrderStats] 店主订单统计调度器已停止")
}
