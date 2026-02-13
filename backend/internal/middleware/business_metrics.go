package middleware

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	// ==================== 业务指标 ====================

	// 订单相关
	OrdersSyncedTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "orders_synced_total",
			Help: "Total number of orders synced",
		},
		[]string{"shop_id", "status"},
	)

	OrdersShippedTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "orders_shipped_total",
			Help: "Total number of orders shipped",
		},
		[]string{"shop_id"},
	)

	// 财务相关
	FinanceIncomeSyncedTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "finance_income_synced_total",
			Help: "Total number of finance incomes synced",
		},
		[]string{"shop_id"},
	)

	SettlementsProcessedTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "settlements_processed_total",
			Help: "Total number of settlements processed",
		},
		[]string{"shop_id"},
	)

	// 账户相关
	AccountTransactionsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "account_transactions_total",
			Help: "Total number of account transactions",
		},
		[]string{"account_type", "transaction_type"},
	)

	WithdrawApplicationsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "withdraw_applications_total",
			Help: "Total number of withdraw applications",
		},
		[]string{"status"},
	)

	// ==================== 系统指标 ====================

	// 分布式锁
	DistributedLockAcquired = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "distributed_lock_acquired_total",
			Help: "Total number of distributed locks acquired",
		},
		[]string{"lock_name"},
	)

	DistributedLockFailed = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "distributed_lock_failed_total",
			Help: "Total number of distributed lock acquisition failures",
		},
		[]string{"lock_name"},
	)

	// 同步任务
	SyncTasksQueued = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "sync_tasks_queued",
			Help: "Number of sync tasks currently queued",
		},
	)

	SyncTasksProcessing = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "sync_tasks_processing",
			Help: "Number of sync tasks currently being processed",
		},
	)

	// 归档任务
	ArchiveRecordsTotal = promauto.NewCounter(
		prometheus.CounterOpts{
			Name: "archive_records_total",
			Help: "Total number of records archived",
		},
	)

	// 分表数据量
	ShardTableRows = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "shard_table_rows",
			Help: "Number of rows in each shard table",
		},
		[]string{"table_name"},
	)

	// Shopee API调用
	ShopeeAPICallsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "shopee_api_calls_total",
			Help: "Total number of Shopee API calls",
		},
		[]string{"api_name", "status"},
	)

	ShopeeAPILatency = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "shopee_api_latency_seconds",
			Help:    "Shopee API call latency in seconds",
			Buckets: []float64{0.1, 0.25, 0.5, 1, 2.5, 5, 10},
		},
		[]string{"api_name"},
	)

	// 数据库连接池
	DBConnectionsActive = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "db_connections_active",
			Help: "Number of active database connections",
		},
	)

	DBConnectionsIdle = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "db_connections_idle",
			Help: "Number of idle database connections",
		},
	)

	// Redis连接
	RedisCommandsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "redis_commands_total",
			Help: "Total number of Redis commands executed",
		},
		[]string{"command"},
	)
)
