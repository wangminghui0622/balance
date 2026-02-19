package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"balance/backend/admin/internal/router"
	"balance/backend/internal/config"
	"balance/backend/internal/database"
	"balance/backend/internal/models"
	"balance/backend/internal/ratelimit"
	"balance/backend/internal/services"
	"balance/backend/internal/services/sync"
	"balance/backend/internal/utils"
)

func main() {
	// 解析命令行参数
	configPath := flag.String("config", "config/config.yaml", "配置文件路径")
	flag.Parse()

	// 加载配置
	cfg, err := config.Load(*configPath)
	if err != nil {
		log.Fatalf("加载配置失败: %v", err)
	}

	// 初始化MySQL
	if err := database.InitMySQL(&cfg.MySQL); err != nil {
		log.Fatalf("初始化MySQL失败: %v", err)
	}
	log.Println("MySQL连接成功")

	// 初始化Redis
	if err := database.InitRedis(&cfg.Redis); err != nil {
		log.Fatalf("初始化Redis失败: %v", err)
	}

	log.Println("Redis连接成功")

	// 初始化 Sentinel 限流器
	if err := ratelimit.Init(); err != nil {
		log.Printf("Sentinel 初始化警告: %v", err)
	}

	// 初始化各定时器的文件日志（不同定时器写入不同日志文件，方便调试）
	// 基于 zap + lumberjack：异步缓冲写入、日志轮转、30天自动清理、gzip压缩
	logDir := "logs/schedulers"
	distributedSyncLogger, closeDistSyncLog, err := utils.NewFileLogger(logDir, "distributed_sync.log")
	if err != nil {
		log.Printf("创建分布式同步日志失败: %v，将使用标准输出", err)
	}
	financeSyncLogger, closeFinanceSyncLog, err := utils.NewFileLogger(logDir, "finance_sync.log")
	if err != nil {
		log.Printf("创建财务同步日志失败: %v，将使用标准输出", err)
	}
	maintenanceLogger, closeMaintenanceLog, err := utils.NewFileLogger(logDir, "maintenance.log")
	if err != nil {
		log.Printf("创建维护任务日志失败: %v，将使用标准输出", err)
	}
	metricsLogger, closeMetricsLog, err := utils.NewFileLogger(logDir, "metrics.log")
	if err != nil {
		log.Printf("创建指标收集日志失败: %v，将使用标准输出", err)
	}

	// 启动分布式同步调度器（订单同步，支持多实例部署）
	distributedSyncScheduler := services.NewDistributedSyncScheduler(database.GetDB(), database.GetRedis(), distributedSyncLogger)
	distributedSyncScheduler.Start()

	// 启动财务同步调度器（增量同步，10个Worker，带分布式锁）
	financeSyncScheduler := sync.NewScheduler(10, financeSyncLogger)
	financeSyncScheduler.Start()

	// 启动维护任务调度器（日志归档、每日统计）
	maintenanceScheduler := services.NewMaintenanceScheduler(maintenanceLogger)
	maintenanceScheduler.Start()

	// 启动指标收集器（Prometheus监控）
	metricsCollector := services.NewMetricsCollector(metricsLogger)
	metricsCollector.Start()

	// 店主订单统计缓存：每 1 小时刷新一次（全部/未结算/已结算/账款调整）
	go runShopowerOrderStatsRefresh()

	// 设置路由
	r := router.SetupRouter(cfg.App.Mode, cfg)

	// 启动服务器
	addr := fmt.Sprintf(":%d", cfg.App.Port)
	log.Printf("服务器启动于 %s", addr)

	srv := &http.Server{
		Addr:         addr,
		Handler:      r,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}
	// 优雅关闭
	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("启动HTTP服务器失败: %v", err)
		}
	}()
	// 等待中断信号
	quit := make(chan os.Signal, 1)
	signal.Notify(quit,
		syscall.SIGINT,  // Ctrl+C
		syscall.SIGTERM, // kill 命令
		syscall.SIGQUIT, // Ctrl+\
		syscall.SIGHUP,  // 终端断开
	)
	<-quit

	// 设置超时上下文
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// 优雅关闭服务器
	if err := srv.Shutdown(ctx); err != nil {
		log.Fatal("服务器强制关闭:", err)
	}
	// 2. 停止所有调度器（按依赖顺序）
	log.Println("正在停止分布式同步调度器...")
	distributedSyncScheduler.Stop()

	log.Println("正在停止财务同步调度器...")
	financeSyncScheduler.Stop()

	log.Println("正在停止维护任务调度器...")
	maintenanceScheduler.Stop()

	log.Println("正在停止指标收集器...")
	metricsCollector.Stop()

	log.Println("正在关闭redis连接...")
	database.CloseRedis()

	log.Println("正在关闭mysql连接...")
	database.Close()
	// 关闭日志文件
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
	log.Println("服务器已退出")
}

// runShopowerOrderStatsRefresh 每 1 小时刷新所有店主的订单统计缓存；启动后 30 秒先跑一次预热
func runShopowerOrderStatsRefresh() {
	orderService := services.NewOrderServiceWithPrepaymentCheck()
	doRefresh := func() {
		ctx := context.Background()
		var adminIDs []int64
		if err := database.GetDB().Model(&models.Shop{}).Distinct("admin_id").Pluck("admin_id", &adminIDs).Error; err != nil {
			log.Printf("刷新店主订单统计: 获取 admin_id 失败: %v", err)
			return
		}
		for _, adminID := range adminIDs {
			stats, err := orderService.ComputeOrderStats(ctx, adminID)
			if err != nil {
				log.Printf("刷新店主订单统计 admin_id=%d: %v", adminID, err)
				continue
			}
			if err := orderService.SetOrderStatsCache(ctx, adminID, stats); err != nil {
				log.Printf("刷新店主订单统计 admin_id=%d 写缓存失败: %v", adminID, err)
			}
		}
	}
	time.Sleep(30 * time.Second)
	doRefresh()
	ticker := time.NewTicker(1 * time.Hour)
	defer ticker.Stop()
	for range ticker.C {
		doRefresh()
	}
}
