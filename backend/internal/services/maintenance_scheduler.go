package services

import (
	"context"
	"sync"
	"time"

	"balance/backend/internal/database"
	"balance/backend/internal/utils"

	"github.com/go-redsync/redsync/v4"
	"github.com/robfig/cron/v3"
	"go.uber.org/zap"
)

const (
	maintenanceLockTTL = 5 * time.Minute
)

// MaintenanceScheduler 维护任务调度器（多机部署时通过分布式锁保证单机执行）
type MaintenanceScheduler struct {
	cron              *cron.Cron
	archiveService    *ArchiveService
	statsService      *StatsService
	settlementService *SettlementService
	rs                *redsync.Redsync
	wg                sync.WaitGroup
	stopChan          chan struct{}
	logger            *zap.SugaredLogger
}

// NewMaintenanceScheduler 创建维护任务调度器
func NewMaintenanceScheduler(logger ...*zap.SugaredLogger) *MaintenanceScheduler {
	var l *zap.SugaredLogger
	if len(logger) > 0 && logger[0] != nil {
		l = logger[0]
	} else {
		l = utils.DefaultSugaredLogger()
	}
	return &MaintenanceScheduler{
		cron:              cron.New(cron.WithSeconds()),
		archiveService:    NewArchiveService(),
		statsService:      NewStatsService(),
		settlementService: NewSettlementService(),
		rs:                database.GetRedsync(),
		stopChan:          make(chan struct{}),
		logger:            l,
	}
}

// tryRunWithLock 尝试获取分布式锁后执行，获取失败则跳过（其他节点正在执行）
func (s *MaintenanceScheduler) tryRunWithLock(lockKey string, fn func()) {
	mutex := s.rs.NewMutex(lockKey, redsync.WithExpiry(maintenanceLockTTL), redsync.WithTries(1))
	if err := mutex.LockContext(context.Background()); err != nil {
		s.logger.Infof("[Maintenance] %s: 其他节点正在执行，跳过", lockKey)
		return
	}
	defer mutex.Unlock()
	fn()
}

// Start 启动维护任务调度器
func (s *MaintenanceScheduler) Start() {
	s.logger.Info("[Maintenance] 启动维护任务调度器...")

	// 每天凌晨2点执行日志归档（分布式锁）
	_, err := s.cron.AddFunc("0 0 2 * * *", func() {
		s.tryRunWithLock("maintenance:archive", func() {
			ctx, cancel := context.WithTimeout(context.Background(), 30*time.Minute)
			defer cancel()
			count, err := s.archiveService.ArchiveOperationLogs(ctx)
			if err != nil {
				s.logger.Infof("[Maintenance] 日志归档失败: %v", err)
			} else {
				s.logger.Infof("[Maintenance] 日志归档完成，归档 %d 条记录", count)
			}
		})
	})
	if err != nil {
		s.logger.Infof("[Maintenance] 添加日志归档任务失败: %v", err)
	}

	// 每天凌晨3点生成每日统计（分布式锁）
	_, err = s.cron.AddFunc("0 0 3 * * *", func() {
		s.tryRunWithLock("maintenance:stats", func() {
			ctx, cancel := context.WithTimeout(context.Background(), 30*time.Minute)
			defer cancel()
			if err := s.statsService.GenerateDailyStats(ctx); err != nil {
				s.logger.Infof("[Maintenance] 生成每日统计失败: %v", err)
			}
		})
	})
	if err != nil {
		s.logger.Infof("[Maintenance] 添加每日统计任务失败: %v", err)
	}

	// 每月1号凌晨4点清理过期归档（分布式锁）
	_, err = s.cron.AddFunc("0 0 4 1 * *", func() {
		s.tryRunWithLock("maintenance:cleanup", func() {
			ctx, cancel := context.WithTimeout(context.Background(), 30*time.Minute)
			defer cancel()
			count, err := s.archiveService.CleanupOldArchives(ctx, 365)
			if err != nil {
				s.logger.Infof("[Maintenance] 清理过期归档失败: %v", err)
			} else {
				s.logger.Infof("[Maintenance] 清理过期归档完成，删除 %d 条记录", count)
			}
		})
	})
	if err != nil {
		s.logger.Infof("[Maintenance] 添加清理归档任务失败: %v", err)
	}

	// 每10分钟处理一次虾皮结算（分布式锁）
	_, err = s.cron.AddFunc("0 */10 * * * *", func() {
		s.tryRunWithLock("maintenance:settlement", func() {
			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
			defer cancel()
			count, err := s.settlementService.ProcessShopeeSettlement(ctx)
			if err != nil {
				s.logger.Infof("[Maintenance] 处理虾皮结算失败: %v", err)
			} else if count > 0 {
				s.logger.Infof("[Maintenance] 处理虾皮结算完成，结算 %d 笔订单", count)
			}
		})
	})
	if err != nil {
		s.logger.Infof("[Maintenance] 添加虾皮结算任务失败: %v", err)
	}

	// 每10分钟处理一次虾皮调账（分布式锁）
	_, err = s.cron.AddFunc("0 */10 * * * *", func() {
		s.tryRunWithLock("maintenance:adjustment", func() {
			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
			defer cancel()
			count, err := s.settlementService.ProcessShopeeAdjustments(ctx)
			if err != nil {
				s.logger.Infof("[Maintenance] 处理虾皮调账失败: %v", err)
			} else if count > 0 {
				s.logger.Infof("[Maintenance] 处理虾皮调账完成，处理 %d 笔调账", count)
			}
		})
	})
	if err != nil {
		s.logger.Infof("[Maintenance] 添加虾皮调账任务失败: %v", err)
	}

	s.cron.Start()
	s.logger.Info("[Maintenance] 维护任务调度器已启动")

	// 启动后检查是否需要补充历史统计
	s.wg.Add(1)
	go s.backfillStats()
}

// Stop 停止维护任务调度器（等待所有后台任务结束）
func (s *MaintenanceScheduler) Stop() {
	s.logger.Info("[Maintenance] 停止维护任务调度器...")
	close(s.stopChan) // 通知后台任务停止
	s.cron.Stop()
	s.wg.Wait() // 等待所有后台任务结束
	s.logger.Info("[Maintenance] 维护任务调度器已停止")
}

// backfillStats 补充历史统计数据（首次启动时执行，带分布式锁）
func (s *MaintenanceScheduler) backfillStats() {
	defer s.wg.Done()

	select {
	case <-time.After(10 * time.Second):
	case <-s.stopChan:
		return
	}

	mutex := s.rs.NewMutex("maintenance:backfill_stats", redsync.WithExpiry(maintenanceLockTTL), redsync.WithTries(1))
	if err := mutex.LockContext(context.Background()); err != nil {
		s.logger.Info("[Maintenance] 其他节点正在补充统计，跳过")
		return
	}
	defer mutex.Unlock()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	defer cancel()

	// 检查是否有统计数据
	var count int64
	if err := s.statsService.db.Model(&struct{}{}).Table("platform_daily_stats").Count(&count).Error; err != nil {
		s.logger.Infof("[Maintenance] 检查统计数据失败: %v", err)
		return
	}

	if count > 0 {
		s.logger.Info("[Maintenance] 已有统计数据，跳过补充")
		return
	}

	s.logger.Info("[Maintenance] 开始补充历史统计数据...")

	// 补充最近30天的统计
	for i := 30; i >= 1; i-- {
		date := time.Now().AddDate(0, 0, -i)
		statDate := time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, time.Local)

		if err := s.statsService.generateOrderDailyStats(ctx, statDate); err != nil {
			s.logger.Infof("[Maintenance] 补充 %s 订单统计失败: %v", statDate.Format("2006-01-02"), err)
		}
		if err := s.statsService.generateFinanceDailyStats(ctx, statDate); err != nil {
			s.logger.Infof("[Maintenance] 补充 %s 财务统计失败: %v", statDate.Format("2006-01-02"), err)
		}
		if err := s.statsService.generatePlatformDailyStats(ctx, statDate); err != nil {
			s.logger.Infof("[Maintenance] 补充 %s 平台统计失败: %v", statDate.Format("2006-01-02"), err)
		}
	}

	s.logger.Info("[Maintenance] 历史统计数据补充完成")
}

// TriggerArchive 手动触发归档
func (s *MaintenanceScheduler) TriggerArchive() (int64, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Minute)
	defer cancel()
	return s.archiveService.ArchiveOperationLogs(ctx)
}

// TriggerStats 手动触发统计生成
func (s *MaintenanceScheduler) TriggerStats() error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Minute)
	defer cancel()
	return s.statsService.GenerateDailyStats(ctx)
}

// GetArchiveStats 获取归档统计
func (s *MaintenanceScheduler) GetArchiveStats() map[string]interface{} {
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Minute)
	defer cancel()
	return s.archiveService.GetArchiveStats(ctx)
}

// GetStatsService 获取统计服务
func (s *MaintenanceScheduler) GetStatsService() *StatsService {
	return s.statsService
}
