package models

import (
	"time"

	"github.com/shopspring/decimal"
)

// OrderDailyStat 订单每日统计（按店铺）
type OrderDailyStat struct {
	ID            uint64          `gorm:"primaryKey" json:"id"`
	StatDate      time.Time       `gorm:"type:date;not null;index:idx_stat_shop,priority:1" json:"stat_date"`
	ShopID        uint64          `gorm:"not null;index:idx_stat_shop,priority:2" json:"shop_id"`
	OrderCount    int64           `gorm:"not null;default:0" json:"order_count"`
	TotalAmount   decimal.Decimal `gorm:"type:decimal(15,2);not null;default:0" json:"total_amount"`
	ShippedCount  int64           `gorm:"not null;default:0" json:"shipped_count"`
	SettledCount  int64           `gorm:"not null;default:0" json:"settled_count"`
	SettledAmount decimal.Decimal `gorm:"type:decimal(15,2);not null;default:0" json:"settled_amount"`
	CreatedAt     time.Time       `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt     time.Time       `gorm:"autoUpdateTime" json:"updated_at"`
}

func (OrderDailyStat) TableName() string {
	return "order_daily_stats"
}

// FinanceDailyStat 财务每日统计（按店铺）
type FinanceDailyStat struct {
	ID           uint64          `gorm:"primaryKey" json:"id"`
	StatDate     time.Time       `gorm:"type:date;not null;index:idx_stat_shop,priority:1" json:"stat_date"`
	ShopID       uint64          `gorm:"not null;index:idx_stat_shop,priority:2" json:"shop_id"`
	IncomeCount  int64           `gorm:"not null;default:0" json:"income_count"`
	IncomeAmount decimal.Decimal `gorm:"type:decimal(15,2);not null;default:0" json:"income_amount"`
	CreatedAt    time.Time       `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt    time.Time       `gorm:"autoUpdateTime" json:"updated_at"`
}

func (FinanceDailyStat) TableName() string {
	return "finance_daily_stats"
}

// PlatformDailyStat 平台每日统计（汇总）
type PlatformDailyStat struct {
	ID            uint64          `gorm:"primaryKey" json:"id"`
	StatDate      time.Time       `gorm:"type:date;not null;uniqueIndex" json:"stat_date"`
	TotalOrders   int64           `gorm:"not null;default:0" json:"total_orders"`
	TotalAmount   decimal.Decimal `gorm:"type:decimal(15,2);not null;default:0" json:"total_amount"`
	SettledAmount decimal.Decimal `gorm:"type:decimal(15,2);not null;default:0" json:"settled_amount"`
	PlatformShare decimal.Decimal `gorm:"type:decimal(15,2);not null;default:0" json:"platform_share"`
	TotalIncome   decimal.Decimal `gorm:"type:decimal(15,2);not null;default:0" json:"total_income"`
	ActiveShops   int64           `gorm:"not null;default:0" json:"active_shops"`
	CreatedAt     time.Time       `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt     time.Time       `gorm:"autoUpdateTime" json:"updated_at"`
}

func (PlatformDailyStat) TableName() string {
	return "platform_daily_stats"
}

// PlatformStatsResult 平台统计查询结果
type PlatformStatsResult struct {
	TotalOrders   int64           `json:"total_orders"`
	TotalAmount   decimal.Decimal `json:"total_amount"`
	SettledAmount decimal.Decimal `json:"settled_amount"`
	PlatformShare decimal.Decimal `json:"platform_share"`
	TotalIncome   decimal.Decimal `json:"total_income"`
}

// ShopStatsResult 店铺统计查询结果
type ShopStatsResult struct {
	TotalOrders   int64           `json:"total_orders"`
	TotalAmount   decimal.Decimal `json:"total_amount"`
	ShippedCount  int64           `json:"shipped_count"`
	SettledCount  int64           `json:"settled_count"`
	SettledAmount decimal.Decimal `json:"settled_amount"`
}
