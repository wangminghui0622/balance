// cron 定时任务服务（独立部署，与 admin HTTP 服务分离）
// 多机部署时通过 Redis 分布式锁保证同一时刻只有单机执行，避免并发冲突
//
// 运行方式: cd backend && go run ./cron -config config/config.yaml
package main

import (
	"flag"
	"log"
	"os"
	"os/signal"
	"syscall"

	"balance/backend/internal/config"
	"balance/backend/internal/database"
	"balance/backend/internal/ratelimit"
	"balance/backend/internal/services"
	"balance/backend/internal/services/sync"
	"balance/backend/internal/utils"
)

func main() {
	configPath := flag.String("config", "config/config.yaml", "配置文件路径")
	flag.Parse()

	cfg, err := config.Load(*configPath)
	if err != nil {
		log.Fatalf("加载配置失败: %v", err)
	}

	if err := database.InitMySQL(&cfg.MySQL); err != nil {
		log.Fatalf("初始化MySQL失败: %v", err)
	}
	log.Println("MySQL连接成功")

	if err := database.InitRedis(&cfg.Redis); err != nil {
		log.Fatalf("初始化Redis失败: %v", err)
	}
	log.Println("Redis连接成功")

	if err := ratelimit.Init(); err != nil {
		log.Printf("Sentinel 初始化警告: %v", err)
	}

	logDir := "logs/schedulers"
	distributedSyncLogger, closeDistSyncLog, _ := utils.NewFileLogger(logDir, "distributed_sync.log")
	financeSyncLogger, closeFinanceSyncLog, _ := utils.NewFileLogger(logDir, "finance_sync.log")
	maintenanceLogger, closeMaintenanceLog, _ := utils.NewFileLogger(logDir, "maintenance.log")
	metricsLogger, closeMetricsLog, _ := utils.NewFileLogger(logDir, "metrics.log")
	orderStatsLogger, closeOrderStatsLog, _ := utils.NewFileLogger(logDir, "order_stats.log")

	// 1. 分布式同步调度器（订单+退货巡检，带分布式锁）
	distributedSyncScheduler := services.NewDistributedSyncScheduler(database.GetDB(), database.GetRedis(), distributedSyncLogger)
	distributedSyncScheduler.Start()

	// 2. 财务同步调度器（钱包流水增量同步，带分布式锁）
	financeSyncScheduler := sync.NewScheduler(10, financeSyncLogger)
	financeSyncScheduler.Start()

	// 3. 维护任务调度器（日志归档、每日统计、虾皮结算/调账，带分布式锁）
	maintenanceScheduler := services.NewMaintenanceScheduler(maintenanceLogger)
	maintenanceScheduler.Start()

	// 4. 指标收集器（Prometheus）
	metricsCollector := services.NewMetricsCollector(metricsLogger)
	metricsCollector.Start()

	// 5. 店主订单统计缓存刷新（每 1 小时，带分布式锁）
	orderStatsScheduler := services.NewOrderStatsScheduler(orderStatsLogger)
	orderStatsScheduler.Start()

	log.Println("[Cron] 定时任务服务已启动，等待退出信号...")

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT, syscall.SIGHUP)
	<-quit

	log.Println("[Cron] 收到退出信号，正在停止...")

	distributedSyncScheduler.Stop()
	financeSyncScheduler.Stop()
	maintenanceScheduler.Stop()
	metricsCollector.Stop()
	orderStatsScheduler.Stop()

	database.CloseRedis()
	database.Close()

	if closeDistSyncLog != nil {
		closeDistSyncLog()
	}
	if closeFinanceSyncLog != nil {
		closeFinanceSyncLog()
	}
	if closeMaintenanceLog != nil {
		closeMaintenanceLog()
	}
	if closeMetricsLog != nil {
		closeMetricsLog()
	}
	if closeOrderStatsLog != nil {
		closeOrderStatsLog()
	}

	log.Println("[Cron] 定时任务服务已退出")
	os.Exit(0)
}
