package main

import (
	"balance/admin/routes"
	"balance/internal/config"
	"balance/internal/database"
	"balance/models"
	"log"
	"strings"
)

// extractHostFromDSN 从DSN中提取host
func extractHostFromDSN(dsn string) string {
	parts := strings.Split(dsn, "@tcp(")
	if len(parts) < 2 {
		return "unknown"
	}
	parts = strings.Split(parts[1], ")")
	if len(parts) < 1 {
		return "unknown"
	}
	hostPort := strings.Split(parts[0], ":")
	return hostPort[0]
}

// extractPortFromDSN 从DSN中提取port
func extractPortFromDSN(dsn string) string {
	parts := strings.Split(dsn, "@tcp(")
	if len(parts) < 2 {
		return "unknown"
	}
	parts = strings.Split(parts[1], ")")
	if len(parts) < 1 {
		return "unknown"
	}
	hostPort := strings.Split(parts[0], ":")
	if len(hostPort) < 2 {
		return "3306"
	}
	return hostPort[1]
}

// extractUserFromDSN 从DSN中提取user
func extractUserFromDSN(dsn string) string {
	parts := strings.Split(dsn, "@tcp(")
	if len(parts) < 1 {
		return "unknown"
	}
	userPass := strings.Split(parts[0], ":")
	return userPass[0]
}

// extractDBNameFromDSN 从DSN中提取dbname
func extractDBNameFromDSN(dsn string) string {
	parts := strings.Split(dsn, "/")
	if len(parts) < 2 {
		return "unknown"
	}
	dbPart := strings.Split(parts[1], "?")
	return dbPart[0]
}

func main() {
	// 加载配置
	cfg := config.LoadAdminConfig()
	
	// 打印配置信息（隐藏密码，用于调试）
	log.Printf("数据库配置: host=%s, port=%s, user=%s, dbname=%s", 
		extractHostFromDSN(cfg.DBDSN), 
		extractPortFromDSN(cfg.DBDSN),
		extractUserFromDSN(cfg.DBDSN),
		extractDBNameFromDSN(cfg.DBDSN))

	// 初始化数据库
	db, err := database.InitDB(cfg.DBDSN)
	if err != nil {
		log.Fatal("初始化数据库失败: ", err)
	}

	// 数据库迁移（admin服务需要迁移Admin模型）
	err = db.AutoMigrate(&models.Admin{})
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
