package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"balance/backend/admin/internal/router"
	"balance/backend/internal/config"
	"balance/backend/internal/database"
	"balance/backend/internal/models"
	"balance/backend/internal/ratelimit"
	"balance/backend/internal/services"
	"balance/backend/internal/services/sync"
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
	defer database.Close()
	log.Println("MySQL连接成功")

	// 自动迁移数据库表结构
	// 注意：分表的表（orders_X, shipments_X 等）需要通过 database.sql 脚本创建
	// 这里只迁移不分表的基础表
	if err := database.GetDB().AutoMigrate(
		// 用户与店铺（不分表）
		&models.Admin{},
		&models.Shop{},
		&models.ShopAuthorization{},
		&models.ShopOperatorRelation{},
		&models.ShopSyncRecord{},
		// 配置表（不分表）
		&models.ProfitShareConfig{},
		&models.LogisticsChannel{},
		// 账户表（不分表）
		&models.PrepaymentAccount{},
		&models.DepositAccount{},
		&models.OperatorAccount{},
		&models.ShopOwnerCommissionAccount{},
		&models.PlatformCommissionAccount{},
		&models.PenaltyBonusAccount{},
		&models.EscrowAccount{},
		&models.CollectionAccount{},
		// 申请表（不分表）
		&models.WithdrawApplication{},
		&models.RechargeApplication{},
		// 统计表（不分表）
		&models.OrderDailyStat{},
		&models.FinanceDailyStat{},
		&models.PlatformDailyStat{},
		// 以下表已分表，需通过 database.sql 创建：
		// orders_0~9, order_items_0~9, order_addresses_0~9
		// order_escrows_0~9, order_escrow_items_0~9, order_settlements_0~9
		// order_shipment_records_0~9, shipments_0~9, finance_incomes_0~9
		// operation_logs_0~9, account_transactions_0~9
	); err != nil {
		log.Printf("数据库迁移警告: %v", err)
	} else {
		log.Println("数据库迁移完成")
	}

	// 初始化Redis
	if err := database.InitRedis(&cfg.Redis); err != nil {
		log.Fatalf("初始化Redis失败: %v", err)
	}
	defer database.CloseRedis()
	log.Println("Redis连接成功")

	// 初始化 Sentinel 限流器
	if err := ratelimit.Init(); err != nil {
		log.Printf("Sentinel 初始化警告: %v", err)
	}

	// 启动分布式同步调度器（订单同步，支持多实例部署）
	distributedSyncScheduler := services.NewDistributedSyncScheduler(database.GetDB(), database.GetRedis())
	distributedSyncScheduler.Start()
	defer distributedSyncScheduler.Stop()

	// 启动财务同步调度器（增量同步，10个Worker，带分布式锁）
	financeSyncScheduler := sync.NewScheduler(10)
	financeSyncScheduler.Start()
	defer financeSyncScheduler.Stop()

	// 启动维护任务调度器（日志归档、每日统计）
	maintenanceScheduler := services.NewMaintenanceScheduler()
	maintenanceScheduler.Start()
	defer maintenanceScheduler.Stop()

	// 启动指标收集器（Prometheus监控）
	metricsCollector := services.NewMetricsCollector()
	metricsCollector.Start()
	defer metricsCollector.Stop()

	// 设置路由
	r := router.SetupRouter(cfg.App.Mode, cfg)

	// 启动服务器
	addr := fmt.Sprintf(":%d", cfg.App.Port)
	log.Printf("服务器启动于 %s", addr)

	// 优雅关闭
	go func() {
		if err := r.Run(addr); err != nil {
			log.Fatalf("启动服务器失败: %v", err)
		}
	}()

	// 等待中断信号
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("正在关闭服务器...")
}
