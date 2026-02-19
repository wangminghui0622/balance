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
	"balance/backend/internal/ratelimit"
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

	r := router.SetupRouter(cfg.App.Mode, cfg)

	addr := fmt.Sprintf(":%d", cfg.App.Port)
	log.Printf("服务器启动于 %s", addr)

	srv := &http.Server{
		Addr:         addr,
		Handler:      r,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}
	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("启动HTTP服务器失败: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit,
		syscall.SIGINT,
		syscall.SIGTERM,
		syscall.SIGQUIT,
		syscall.SIGHUP,
	)
	<-quit

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatal("服务器强制关闭:", err)
	}

	log.Println("正在关闭redis连接...")
	database.CloseRedis()

	log.Println("正在关闭mysql连接...")
	database.Close()

	log.Println("服务器已退出")
}
