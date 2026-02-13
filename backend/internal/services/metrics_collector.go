package services

import (
	"context"
	"fmt"
	"log"
	"time"

	"balance/backend/internal/database"
	"balance/backend/internal/middleware"

	"gorm.io/gorm"
)

// MetricsCollector 指标收集器
type MetricsCollector struct {
	db       *gorm.DB
	interval time.Duration
	stopChan chan struct{}
}

// NewMetricsCollector 创建指标收集器
func NewMetricsCollector() *MetricsCollector {
	return &MetricsCollector{
		db:       database.GetDB(),
		interval: 1 * time.Minute,
		stopChan: make(chan struct{}),
	}
}

// Start 启动指标收集
func (m *MetricsCollector) Start() {
	log.Println("[Metrics] 启动指标收集器...")
	go m.collect()
}

// Stop 停止指标收集
func (m *MetricsCollector) Stop() {
	close(m.stopChan)
	log.Println("[Metrics] 指标收集器已停止")
}

func (m *MetricsCollector) collect() {
	ticker := time.NewTicker(m.interval)
	defer ticker.Stop()

	// 启动时立即收集一次
	m.collectMetrics()

	for {
		select {
		case <-ticker.C:
			m.collectMetrics()
		case <-m.stopChan:
			return
		}
	}
}

func (m *MetricsCollector) collectMetrics() {
	ctx := context.Background()

	// 收集分表数据量
	m.collectShardTableMetrics()

	// 收集数据库连接池状态
	m.collectDBPoolMetrics()

	// 收集Redis队列状态
	m.collectRedisQueueMetrics(ctx)
}

// collectShardTableMetrics 收集分表数据量指标
func (m *MetricsCollector) collectShardTableMetrics() {
	tables := []string{
		"orders", "order_items", "order_addresses",
		"order_escrows", "order_escrow_items",
		"order_settlements", "order_shipment_records",
		"shipments", "finance_incomes", "operation_logs",
		"account_transactions",
	}

	for _, table := range tables {
		for i := 0; i < database.ShardCount; i++ {
			tableName := fmt.Sprintf("%s_%d", table, i)
			var count int64
			if err := m.db.Table(tableName).Count(&count).Error; err == nil {
				middleware.ShardTableRows.WithLabelValues(tableName).Set(float64(count))
			}
		}
	}
}

// collectDBPoolMetrics 收集数据库连接池指标
func (m *MetricsCollector) collectDBPoolMetrics() {
	sqlDB, err := m.db.DB()
	if err != nil {
		return
	}

	stats := sqlDB.Stats()
	middleware.DBConnectionsActive.Set(float64(stats.InUse))
	middleware.DBConnectionsIdle.Set(float64(stats.Idle))
}

// collectRedisQueueMetrics 收集Redis队列指标
func (m *MetricsCollector) collectRedisQueueMetrics(ctx context.Context) {
	rdb := database.GetRedis()
	if rdb == nil {
		return
	}

	// 同步任务队列长度
	queueLen, err := rdb.LLen(ctx, "sync:shop:queue").Result()
	if err == nil {
		middleware.SyncTasksQueued.Set(float64(queueLen))
	}

	// 正在处理的任务数
	processingCount, err := rdb.SCard(ctx, "sync:shop:processing").Result()
	if err == nil {
		middleware.SyncTasksProcessing.Set(float64(processingCount))
	}
}
