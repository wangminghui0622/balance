package sync

import (
	"context"
	"time"

	"balance/backend/internal/database"
	"balance/backend/internal/utils"

	"github.com/go-redsync/redsync/v4"
	"github.com/robfig/cron/v3"
	"go.uber.org/zap"
)

const (
	financeSyncSchedulerLock  = "sync:finance:scheduler:lock"
	financeSyncLockTTL        = 2 * time.Minute
	financeSyncExtendInterval = financeSyncLockTTL / 3
)

// Scheduler 同步调度器
type Scheduler struct {
	cron               *cron.Cron
	financeSyncService *FinanceSyncService
	rs                 *redsync.Redsync
	logger             *zap.SugaredLogger // 文件日志
}

// NewScheduler 创建调度器
// logger: 可选的 zap SugaredLogger，传nil则使用默认标准输出
func NewScheduler(workerCount int, logger ...*zap.SugaredLogger) *Scheduler {
	var l *zap.SugaredLogger
	if len(logger) > 0 && logger[0] != nil {
		l = logger[0]
	} else {
		l = utils.DefaultSugaredLogger()
	}
	return &Scheduler{
		cron:               cron.New(cron.WithSeconds()),
		financeSyncService: NewFinanceSyncService(workerCount),
		rs:                 database.GetRedsync(),
		logger:             l,
	}
}

// Start 启动调度器
func (s *Scheduler) Start() {
	s.logger.Info("启动同步调度器...")

	// 启动财务同步服务
	s.financeSyncService.Start()

	// 每小时调度一次所有店铺的财务同步（带分布式锁）
	// 虾皮打款通常在订单完成后几天结算，无需频繁同步
	_, err := s.cron.AddFunc("0 0 * * * *", func() {
		s.tryScheduleWithLock()
	})
	if err != nil {
		s.logger.Infof("添加财务同步定时任务失败: %v", err)
	}

	// 每小时打印一次同步统计
	_, err = s.cron.AddFunc("0 0 * * * *", func() {
		stats := s.financeSyncService.GetSyncStats()
		s.logger.Infof("同步统计: 总店铺=%v, 启用=%v, 暂停=%v, 已同步=%v, 队列=%v",
			stats["total_shops"], stats["enabled_shops"], stats["paused_shops"],
			stats["total_synced"], stats["queue_size"])
	})
	if err != nil {
		s.logger.Infof("添加统计定时任务失败: %v", err)
	}

	s.cron.Start()
	s.logger.Info("同步调度器已启动")

	// 启动后延迟5秒执行一次初始同步（带分布式锁）
	go func() {
		time.Sleep(5 * time.Second)
		s.tryScheduleWithLock()
	}()
}

// tryScheduleWithLock 尝试获取分布式锁后执行调度
func (s *Scheduler) tryScheduleWithLock() {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	// 使用 redsync 创建分布式锁
	mutex := s.rs.NewMutex(financeSyncSchedulerLock,
		redsync.WithExpiry(financeSyncLockTTL),
		redsync.WithTries(1),
	)

	// 尝试获取锁并自动续期
	unlockFunc, acquired := utils.TryLockWithAutoExtend(ctx, mutex, financeSyncExtendInterval)
	if !acquired {
		s.logger.Info("[FinanceSync] 其他节点正在调度，跳过")
		return
	}
	defer unlockFunc()

	s.logger.Info("[FinanceSync] 获取分布式锁成功，开始调度财务同步任务...")
	s.financeSyncService.ScheduleAllShops()
}

// Stop 停止调度器
func (s *Scheduler) Stop() {
	s.logger.Info("停止同步调度器...")
	s.cron.Stop()
	s.financeSyncService.Stop()
	s.logger.Info("同步调度器已停止")
}

// GetFinanceSyncService 获取财务同步服务
func (s *Scheduler) GetFinanceSyncService() *FinanceSyncService {
	return s.financeSyncService
}

// TriggerSync 手动触发同步
func (s *Scheduler) TriggerSync() {
	go s.financeSyncService.ScheduleAllShops()
}
