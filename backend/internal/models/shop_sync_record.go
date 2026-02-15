package models

import (
	"time"
)

// ShopSyncRecord 店铺同步记录（记录各类数据同步进度）
type ShopSyncRecord struct {
	ID                    uint64     `gorm:"primaryKey;comment:主键ID" json:"id"`
	ShopID                uint64     `gorm:"not null;uniqueIndex:uk_shop_sync_type;comment:店铺ID" json:"shop_id"`
	SyncType              string     `gorm:"size:50;not null;uniqueIndex:uk_shop_sync_type;comment:同步类型" json:"sync_type"`
	LastSyncTime          int64      `gorm:"not null;default:0;comment:上次同步时间戳" json:"last_sync_time"`
	LastTransactionID     int64      `gorm:"not null;default:0;comment:上次同步的交易ID" json:"last_transaction_id"`
	LastSyncAt            *time.Time `gorm:"comment:上次同步时间" json:"last_sync_at"`
	TotalSyncedCount      int64      `gorm:"not null;default:0;comment:累计同步数量" json:"total_synced_count"`
	LastSyncCount         int        `gorm:"not null;default:0;comment:上次同步数量" json:"last_sync_count"`
	LastError             string     `gorm:"size:500;not null;default:'';comment:上次错误信息" json:"last_error"`
	ConsecutiveFailCount  int        `gorm:"not null;default:0;comment:连续失败次数" json:"consecutive_fail_count"`
	Status                int8       `gorm:"not null;default:1;comment:状态(0禁用/1启用/2暂停)" json:"status"`
	CreatedAt             time.Time  `gorm:"autoCreateTime;comment:创建时间" json:"created_at"`
	UpdatedAt             time.Time  `gorm:"autoUpdateTime;comment:更新时间" json:"updated_at"`
}

// TableName 指定表名
func (ShopSyncRecord) TableName() string {
	return "shop_sync_records"
}

// 同步类型
const (
	SyncTypeFinanceIncome = "finance_income" // 财务收入同步
	SyncTypeOrder         = "order"          // 订单同步
	SyncTypeEscrow        = "escrow"         // 结算明细同步
)

// 同步状态
const (
	SyncStatusDisabled = 0 // 禁用
	SyncStatusEnabled  = 1 // 启用
	SyncStatusPaused   = 2 // 暂停（连续失败过多）
)

// SyncTask 同步任务
type SyncTask struct {
	ShopID   uint64 `json:"shop_id"`
	SyncType string `json:"sync_type"`
	Priority int    `json:"priority"`
}
