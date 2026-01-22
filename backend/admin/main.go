package main

import (
	"balance/admin/routes"
	"balance/internal/config"
	"balance/internal/database"
	"balance/internal/models"
	"log"
)

func main() {
	// 加载配置
	cfg := config.LoadAdminConfig()

	// 初始化数据库
	db, err := database.InitDB(cfg.DBDSN)
	if err != nil {
		log.Fatal("初始化数据库失败: ", err)
	}

	// 数据库迁移（admin服务需要迁移Admin模型和ShopeeToken模型）
	err = db.AutoMigrate(&models.Admin{}, &models.ShopeeToken{})
	if err != nil {
		log.Fatal("数据库迁移失败: ", err)
	}

	// 初始化Redis
	redisClient, err := database.InitRedis(cfg.RedisAddr, cfg.RedisPassword)
	if err != nil {
		log.Fatal("初始化Redis失败: ", err)
	}

	// 设置路由
	r := routes.SetupRoutes(db, redisClient, cfg)

	// 启动服务
	addr := ":" + cfg.Port
	log.Printf("服务启动在端口 %s", cfg.Port)
	if err := r.Run(addr); err != nil {
		log.Fatal("启动服务失败: ", err)
	}
}
