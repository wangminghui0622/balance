package models

import (
	"time"
)

// ShopSyncRecord 店铺同步记录
type ShopSyncRecord struct {
	ID                    uint64     `gorm:"primaryKey" json:"id"`
	ShopID                uint64     `gorm:"not null;uniqueIndex:uk_shop_sync_type" json:"shop_id"`
	SyncType              string     `gorm:"size:50;not null;uniqueIndex:uk_shop_sync_type" json:"sync_type"`
	LastSyncTime          int64      `gorm:"not null;default:0" json:"last_sync_time"`
	LastTransactionID     int64      `gorm:"not null;default:0" json:"last_transaction_id"`
	LastSyncAt            *time.Time `json:"last_sync_at"`
	TotalSyncedCount      int64      `gorm:"not null;default:0" json:"total_synced_count"`
	LastSyncCount         int        `gorm:"not null;default:0" json:"last_sync_count"`
	LastError             string     `gorm:"size:500;not null;default:''" json:"last_error"`
	ConsecutiveFailCount  int        `gorm:"not null;default:0" json:"consecutive_fail_count"`
	Status                int8       `gorm:"not null;default:1" json:"status"`
	CreatedAt             time.Time  `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt             time.Time  `gorm:"autoUpdateTime" json:"updated_at"`
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
