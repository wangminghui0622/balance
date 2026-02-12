package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"balance/backend/app/internal/router"
	"balance/backend/internal/config"
	"balance/backend/internal/database"
	"balance/backend/internal/services"
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

	// 初始化Redis
	if err := database.InitRedis(&cfg.Redis); err != nil {
		log.Fatalf("初始化Redis失败: %v", err)
	}
	defer database.CloseRedis()
	log.Println("Redis连接成功")

	// 启动定时同步调度器
	syncScheduler := services.NewSyncScheduler(database.GetDB())
	syncScheduler.Start()
	defer syncScheduler.Stop()

	// 设置路由
	r := router.SetupRouter(cfg.App.Mode)

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
