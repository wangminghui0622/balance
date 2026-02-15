package models

import (
	"time"

	"github.com/shopspring/decimal"
)

// OrderDailyStat 订单每日统计（按店铺统计的每日订单数据）
type OrderDailyStat struct {
	ID            uint64          `gorm:"primaryKey;comment:主键ID" json:"id"`
	StatDate      time.Time       `gorm:"type:date;not null;index:idx_stat_shop,priority:1;comment:统计日期" json:"stat_date"`
	ShopID        uint64          `gorm:"not null;index:idx_stat_shop,priority:2;comment:店铺ID" json:"shop_id"`
	OrderCount    int64           `gorm:"not null;default:0;comment:订单数" json:"order_count"`
	TotalAmount   decimal.Decimal `gorm:"type:decimal(15,2);not null;default:0;comment:订单总额" json:"total_amount"`
	ShippedCount  int64           `gorm:"not null;default:0;comment:已发货数" json:"shipped_count"`
	SettledCount  int64           `gorm:"not null;default:0;comment:已结算数" json:"settled_count"`
	SettledAmount decimal.Decimal `gorm:"type:decimal(15,2);not null;default:0;comment:结算金额" json:"settled_amount"`
	CreatedAt     time.Time       `gorm:"autoCreateTime;comment:创建时间" json:"created_at"`
	UpdatedAt     time.Time       `gorm:"autoUpdateTime;comment:更新时间" json:"updated_at"`
}

func (OrderDailyStat) TableName() string {
	return "order_daily_stats"
}

// FinanceDailyStat 财务每日统计（按店铺统计的每日财务数据）
type FinanceDailyStat struct {
	ID           uint64          `gorm:"primaryKey;comment:主键ID" json:"id"`
	StatDate     time.Time       `gorm:"type:date;not null;index:idx_stat_shop,priority:1;comment:统计日期" json:"stat_date"`
	ShopID       uint64          `gorm:"not null;index:idx_stat_shop,priority:2;comment:店铺ID" json:"shop_id"`
	IncomeCount  int64           `gorm:"not null;default:0;comment:收入笔数" json:"income_count"`
	IncomeAmount decimal.Decimal `gorm:"type:decimal(15,2);not null;default:0;comment:收入金额" json:"income_amount"`
	CreatedAt    time.Time       `gorm:"autoCreateTime;comment:创建时间" json:"created_at"`
	UpdatedAt    time.Time       `gorm:"autoUpdateTime;comment:更新时间" json:"updated_at"`
}

func (FinanceDailyStat) TableName() string {
	return "finance_daily_stats"
}

// PlatformDailyStat 平台每日统计（平台级别的每日汇总数据）
type PlatformDailyStat struct {
	ID            uint64          `gorm:"primaryKey;comment:主键ID" json:"id"`
	StatDate      time.Time       `gorm:"type:date;not null;uniqueIndex;comment:统计日期" json:"stat_date"`
	TotalOrders   int64           `gorm:"not null;default:0;comment:总订单数" json:"total_orders"`
	TotalAmount   decimal.Decimal `gorm:"type:decimal(15,2);not null;default:0;comment:总订单金额" json:"total_amount"`
	SettledAmount decimal.Decimal `gorm:"type:decimal(15,2);not null;default:0;comment:结算金额" json:"settled_amount"`
	PlatformShare decimal.Decimal `gorm:"type:decimal(15,2);not null;default:0;comment:平台分成" json:"platform_share"`
	TotalIncome   decimal.Decimal `gorm:"type:decimal(15,2);not null;default:0;comment:总收入" json:"total_income"`
	ActiveShops   int64           `gorm:"not null;default:0;comment:活跃店铺数" json:"active_shops"`
	CreatedAt     time.Time       `gorm:"autoCreateTime;comment:创建时间" json:"created_at"`
	UpdatedAt     time.Time       `gorm:"autoUpdateTime;comment:更新时间" json:"updated_at"`
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
