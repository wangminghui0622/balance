package database

import (
	"fmt"
	"time"

	"balance/backend/internal/config"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var db *gorm.DB

// InitMySQL 初始化MySQL连接
func InitMySQL(cfg *config.MySQLConfig) error {
	var err error

	// 配置GORM日志级别
	logLevel := logger.Info
	if config.Get().App.Mode == "release" {
		logLevel = logger.Error
	}

	db, err = gorm.Open(mysql.Open(cfg.DSN()), &gorm.Config{
		Logger: logger.Default.LogMode(logLevel),
	})
	if err != nil {
		return fmt.Errorf("连接MySQL失败: %w", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		return fmt.Errorf("获取底层DB失败: %w", err)
	}

	// 设置连接池参数
	sqlDB.SetMaxIdleConns(cfg.MaxIdleConns)
	sqlDB.SetMaxOpenConns(cfg.MaxOpenConns)
	sqlDB.SetConnMaxLifetime(time.Duration(cfg.ConnMaxLifetime) * time.Second)

	// 测试连接
	if err := sqlDB.Ping(); err != nil {
		return fmt.Errorf("Ping MySQL失败: %w", err)
	}

	return nil
}

// GetDB 获取数据库实例
func GetDB() *gorm.DB {
	return db
}

// Close 关闭数据库连接
func Close() error {
	if db != nil {
		sqlDB, err := db.DB()
		if err != nil {
			return err
		}
		return sqlDB.Close()
	}
	return nil
}
