package models

import (
	"time"

	"github.com/shopspring/decimal"
)

// OrderSettlement 订单结算记录 (利润分成明细)
type OrderSettlement struct {
	ID                  uint64          `gorm:"primaryKey" json:"id"`
	SettlementNo        string          `gorm:"size:64;not null;uniqueIndex" json:"settlement_no"`     // 结算单号
	ShopID              uint64          `gorm:"not null;index" json:"shop_id"`
	OrderSN             string          `gorm:"size:64;not null;uniqueIndex" json:"order_sn"`
	OrderID             uint64          `gorm:"not null;index" json:"order_id"`
	ShopOwnerID         int64           `gorm:"not null;index" json:"shop_owner_id"`                   // 店铺老板ID
	OperatorID          int64           `gorm:"not null;index" json:"operator_id"`                     // 运营老板ID
	Currency            string          `gorm:"size:10;not null;default:'TWD'" json:"currency"`

	// 金额明细
	EscrowAmount        decimal.Decimal `gorm:"type:decimal(15,2);not null;default:0.00" json:"escrow_amount"`        // Shopee结算金额
	GoodsCost           decimal.Decimal `gorm:"type:decimal(15,2);not null;default:0.00" json:"goods_cost"`           // 商品成本
	ShippingCost        decimal.Decimal `gorm:"type:decimal(15,2);not null;default:0.00" json:"shipping_cost"`        // 运费成本
	TotalCost           decimal.Decimal `gorm:"type:decimal(15,2);not null;default:0.00" json:"total_cost"`           // 总成本 = 商品成本 + 运费成本
	Profit              decimal.Decimal `gorm:"type:decimal(15,2);not null;default:0.00" json:"profit"`               // 利润 = 结算金额 - 总成本

	// 分成比例 (百分比)
	PlatformShareRate   decimal.Decimal `gorm:"type:decimal(5,2);not null;default:0.00" json:"platform_share_rate"`   // 平台分成比例
	OperatorShareRate   decimal.Decimal `gorm:"type:decimal(5,2);not null;default:0.00" json:"operator_share_rate"`   // 运营分成比例
	ShopOwnerShareRate  decimal.Decimal `gorm:"type:decimal(5,2);not null;default:0.00" json:"shop_owner_share_rate"` // 店主分成比例

	// 分成金额
	PlatformShare       decimal.Decimal `gorm:"type:decimal(15,2);not null;default:0.00" json:"platform_share"`       // 平台分成
	OperatorShare       decimal.Decimal `gorm:"type:decimal(15,2);not null;default:0.00" json:"operator_share"`       // 运营分成
	ShopOwnerShare      decimal.Decimal `gorm:"type:decimal(15,2);not null;default:0.00" json:"shop_owner_share"`     // 店主分成

	// 运营实际收入 = 成本 + 运营分成
	OperatorIncome      decimal.Decimal `gorm:"type:decimal(15,2);not null;default:0.00" json:"operator_income"`

	// 状态
	Status              int8            `gorm:"not null;default:0;index" json:"status"`                // 0=待结算 1=已结算 2=已取消
	SettledAt           *time.Time      `json:"settled_at"`                                            // 结算时间
	Remark              string          `gorm:"size:500;not null;default:''" json:"remark"`

	CreatedAt           time.Time       `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt           time.Time       `gorm:"autoUpdateTime" json:"updated_at"`
}

func (OrderSettlement) TableName() string {
	return "order_settlements"
}

// 订单结算状态常量
const (
	OrderSettlementPending   = 0 // 待结算
	OrderSettlementCompleted = 1 // 已结算
	OrderSettlementCancelled = 2 // 已取消
)

// ProfitShareConfig 利润分成配置
type ProfitShareConfig struct {
	ID                  uint64          `gorm:"primaryKey" json:"id"`
	ShopID              uint64          `gorm:"not null;uniqueIndex:uk_shop_operator_config" json:"shop_id"`
	OperatorID          int64           `gorm:"not null;uniqueIndex:uk_shop_operator_config" json:"operator_id"`
	PlatformShareRate   decimal.Decimal `gorm:"type:decimal(5,2);not null;default:5.00" json:"platform_share_rate"`   // 平台分成比例 默认5%
	OperatorShareRate   decimal.Decimal `gorm:"type:decimal(5,2);not null;default:45.00" json:"operator_share_rate"`  // 运营分成比例 默认45%
	ShopOwnerShareRate  decimal.Decimal `gorm:"type:decimal(5,2);not null;default:50.00" json:"shop_owner_share_rate"` // 店主分成比例 默认50%
	Status              int8            `gorm:"not null;default:1" json:"status"`                      // 1=生效 2=失效
	EffectiveFrom       time.Time       `gorm:"not null" json:"effective_from"`                        // 生效时间
	EffectiveTo         *time.Time      `json:"effective_to"`                                          // 失效时间
	Remark              string          `gorm:"size:500;not null;default:''" json:"remark"`
	CreatedAt           time.Time       `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt           time.Time       `gorm:"autoUpdateTime" json:"updated_at"`
}

func (ProfitShareConfig) TableName() string {
	return "profit_share_configs"
}

// OrderShipmentRecord 订单发货记录 (运营发货)
type OrderShipmentRecord struct {
	ID                  uint64          `gorm:"primaryKey" json:"id"`
	ShopID              uint64          `gorm:"not null;index" json:"shop_id"`
	OrderSN             string          `gorm:"size:64;not null;uniqueIndex" json:"order_sn"`
	OrderID             uint64          `gorm:"not null;index" json:"order_id"`
	ShopOwnerID         int64           `gorm:"not null;index" json:"shop_owner_id"`
	OperatorID          int64           `gorm:"not null;index" json:"operator_id"`

	// 成本信息
	GoodsCost           decimal.Decimal `gorm:"type:decimal(15,2);not null;default:0.00" json:"goods_cost"`
	ShippingCost        decimal.Decimal `gorm:"type:decimal(15,2);not null;default:0.00" json:"shipping_cost"`
	TotalCost           decimal.Decimal `gorm:"type:decimal(15,2);not null;default:0.00" json:"total_cost"`
	Currency            string          `gorm:"size:10;not null;default:'TWD'" json:"currency"`

	// 预付款冻结信息
	FrozenAmount        decimal.Decimal `gorm:"type:decimal(15,2);not null;default:0.00" json:"frozen_amount"`
	FrozenTransactionID uint64          `gorm:"not null;default:0" json:"frozen_transaction_id"`

	// 发货信息
	ShippingCarrier     string          `gorm:"size:100;not null;default:''" json:"shipping_carrier"`
	TrackingNumber      string          `gorm:"size:100;not null;default:''" json:"tracking_number"`
	ShippedAt           *time.Time      `json:"shipped_at"`

	// 状态
	Status              int8            `gorm:"not null;default:0;index" json:"status"`  // 0=待发货 1=已发货 2=已完成 3=已取消 4=发货失败
	SettlementID        uint64          `gorm:"not null;default:0" json:"settlement_id"` // 关联结算记录
	Remark              string          `gorm:"size:500;not null;default:''" json:"remark"`

	CreatedAt           time.Time       `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt           time.Time       `gorm:"autoUpdateTime" json:"updated_at"`
}

func (OrderShipmentRecord) TableName() string {
	return "order_shipment_records"
}

// 发货记录状态常量
const (
	ShipmentRecordStatusPending   = 0 // 待发货
	ShipmentRecordStatusShipped   = 1 // 已发货
	ShipmentRecordStatusCompleted = 2 // 已完成 (已结算)
	ShipmentRecordStatusCancelled = 3 // 已取消
	ShipmentRecordStatusFailed    = 4 // 发货失败
)
