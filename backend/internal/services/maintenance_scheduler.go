package services

import (
	"context"
	"log"
	"sync"
	"time"

	"github.com/robfig/cron/v3"
)

// MaintenanceScheduler 维护任务调度器
type MaintenanceScheduler struct {
	cron              *cron.Cron
	archiveService    *ArchiveService
	statsService      *StatsService
	settlementService *SettlementService
	wg                sync.WaitGroup // 用于等待后台任务结束
	stopChan          chan struct{}  // 用于通知后台任务停止
}

// NewMaintenanceScheduler 创建维护任务调度器
func NewMaintenanceScheduler() *MaintenanceScheduler {
	return &MaintenanceScheduler{
		cron:              cron.New(cron.WithSeconds()),
		archiveService:    NewArchiveService(),
		statsService:      NewStatsService(),
		settlementService: NewSettlementService(),
		stopChan:          make(chan struct{}),
	}
}

// Start 启动维护任务调度器
func (s *MaintenanceScheduler) Start() {
	log.Println("[Maintenance] 启动维护任务调度器...")

	// 每天凌晨2点执行日志归档
	_, err := s.cron.AddFunc("0 0 2 * * *", func() {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Minute)
		defer cancel()
		count, err := s.archiveService.ArchiveOperationLogs(ctx)
		if err != nil {
			log.Printf("[Maintenance] 日志归档失败: %v", err)
		} else {
			log.Printf("[Maintenance] 日志归档完成，归档 %d 条记录", count)
		}
	})
	if err != nil {
		log.Printf("[Maintenance] 添加日志归档任务失败: %v", err)
	}

	// 每天凌晨3点生成每日统计
	_, err = s.cron.AddFunc("0 0 3 * * *", func() {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Minute)
		defer cancel()
		if err := s.statsService.GenerateDailyStats(ctx); err != nil {
			log.Printf("[Maintenance] 生成每日统计失败: %v", err)
		}
	})
	if err != nil {
		log.Printf("[Maintenance] 添加每日统计任务失败: %v", err)
	}

	// 每月1号凌晨4点清理过期归档（保留365天）
	_, err = s.cron.AddFunc("0 0 4 1 * *", func() {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Minute)
		defer cancel()
		count, err := s.archiveService.CleanupOldArchives(ctx, 365)
		if err != nil {
			log.Printf("[Maintenance] 清理过期归档失败: %v", err)
		} else {
			log.Printf("[Maintenance] 清理过期归档完成，删除 %d 条记录", count)
		}
	})
	if err != nil {
		log.Printf("[Maintenance] 添加清理归档任务失败: %v", err)
	}

	// 每10分钟处理一次虾皮结算（打款）
	_, err = s.cron.AddFunc("0 */10 * * * *", func() {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
		defer cancel()
		count, err := s.settlementService.ProcessShopeeSettlement(ctx)
		if err != nil {
			log.Printf("[Maintenance] 处理虾皮结算失败: %v", err)
		} else if count > 0 {
			log.Printf("[Maintenance] 处理虾皮结算完成，结算 %d 笔订单", count)
		}
	})
	if err != nil {
		log.Printf("[Maintenance] 添加虾皮结算任务失败: %v", err)
	}

	// 每10分钟处理一次虾皮调账（退款、扣款等）
	_, err = s.cron.AddFunc("0 */10 * * * *", func() {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
		defer cancel()
		count, err := s.settlementService.ProcessShopeeAdjustments(ctx)
		if err != nil {
			log.Printf("[Maintenance] 处理虾皮调账失败: %v", err)
		} else if count > 0 {
			log.Printf("[Maintenance] 处理虾皮调账完成，处理 %d 笔调账", count)
		}
	})
	if err != nil {
		log.Printf("[Maintenance] 添加虾皮调账任务失败: %v", err)
	}

	s.cron.Start()
	log.Println("[Maintenance] 维护任务调度器已启动")

	// 启动后检查是否需要补充历史统计
	s.wg.Add(1)
	go s.backfillStats()
}

// Stop 停止维护任务调度器（等待所有后台任务结束）
func (s *MaintenanceScheduler) Stop() {
	log.Println("[Maintenance] 停止维护任务调度器...")
	close(s.stopChan) // 通知后台任务停止
	s.cron.Stop()
	s.wg.Wait() // 等待所有后台任务结束
	log.Println("[Maintenance] 维护任务调度器已停止")
}

// backfillStats 补充历史统计数据（首次启动时执行）
func (s *MaintenanceScheduler) backfillStats() {
	defer s.wg.Done()
	
	// 等待系统启动完成，但可被停止信号中断
	select {
	case <-time.After(10 * time.Second):
	case <-s.stopChan:
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	defer cancel()
	
	// 检查是否有统计数据
	var count int64
	if err := s.statsService.db.Model(&struct{}{}).Table("platform_daily_stats").Count(&count).Error; err != nil {
		log.Printf("[Maintenance] 检查统计数据失败: %v", err)
		return
	}

	if count > 0 {
		log.Println("[Maintenance] 已有统计数据，跳过补充")
		return
	}

	log.Println("[Maintenance] 开始补充历史统计数据...")

	// 补充最近30天的统计
	for i := 30; i >= 1; i-- {
		date := time.Now().AddDate(0, 0, -i)
		statDate := time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, time.Local)

		if err := s.statsService.generateOrderDailyStats(ctx, statDate); err != nil {
			log.Printf("[Maintenance] 补充 %s 订单统计失败: %v", statDate.Format("2006-01-02"), err)
		}
		if err := s.statsService.generateFinanceDailyStats(ctx, statDate); err != nil {
			log.Printf("[Maintenance] 补充 %s 财务统计失败: %v", statDate.Format("2006-01-02"), err)
		}
		if err := s.statsService.generatePlatformDailyStats(ctx, statDate); err != nil {
			log.Printf("[Maintenance] 补充 %s 平台统计失败: %v", statDate.Format("2006-01-02"), err)
		}
	}

	log.Println("[Maintenance] 历史统计数据补充完成")
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
