package models

import (
	"time"

	"github.com/shopspring/decimal"
)

// OrderSettlement 订单结算记录（利润分成明细，分表）
type OrderSettlement struct {
	ID                  uint64          `gorm:"primaryKey;comment:主键ID" json:"id"`
	SettlementNo        string          `gorm:"size:64;not null;uniqueIndex;comment:结算单号" json:"settlement_no"`
	ShopID              uint64          `gorm:"not null;index;comment:店铺ID" json:"shop_id"`
	OrderSN             string          `gorm:"size:64;not null;uniqueIndex;comment:订单编号" json:"order_sn"`
	OrderID             uint64          `gorm:"not null;index;comment:订单ID" json:"order_id"`
	ShopOwnerID         int64           `gorm:"not null;index;comment:店铺老板ID" json:"shop_owner_id"`
	OperatorID          int64           `gorm:"not null;index;comment:运营老板ID" json:"operator_id"`
	Currency            string          `gorm:"size:10;not null;default:'TWD';comment:货币代码" json:"currency"`

	// 金额明细
	EscrowAmount        decimal.Decimal `gorm:"type:decimal(15,2);not null;default:0.00;comment:Shopee结算金额" json:"escrow_amount"`
	GoodsCost           decimal.Decimal `gorm:"type:decimal(15,2);not null;default:0.00;comment:商品成本" json:"goods_cost"`
	ShippingCost        decimal.Decimal `gorm:"type:decimal(15,2);not null;default:0.00;comment:运费成本" json:"shipping_cost"`
	TotalCost           decimal.Decimal `gorm:"type:decimal(15,2);not null;default:0.00;comment:总成本" json:"total_cost"`
	Profit              decimal.Decimal `gorm:"type:decimal(15,2);not null;default:0.00;comment:利润" json:"profit"`

	// 分成比例 (百分比)
	PlatformShareRate   decimal.Decimal `gorm:"type:decimal(5,2);not null;default:0.00;comment:平台分成比例%" json:"platform_share_rate"`
	OperatorShareRate   decimal.Decimal `gorm:"type:decimal(5,2);not null;default:0.00;comment:运营分成比例%" json:"operator_share_rate"`
	ShopOwnerShareRate  decimal.Decimal `gorm:"type:decimal(5,2);not null;default:0.00;comment:店主分成比例%" json:"shop_owner_share_rate"`

	// 分成金额
	PlatformShare       decimal.Decimal `gorm:"type:decimal(15,2);not null;default:0.00;comment:平台分成金额" json:"platform_share"`
	OperatorShare       decimal.Decimal `gorm:"type:decimal(15,2);not null;default:0.00;comment:运营分成金额" json:"operator_share"`
	ShopOwnerShare      decimal.Decimal `gorm:"type:decimal(15,2);not null;default:0.00;comment:店主分成金额" json:"shop_owner_share"`

	// 运营实际收入 = 成本 + 运营分成
	OperatorIncome      decimal.Decimal `gorm:"type:decimal(15,2);not null;default:0.00;comment:运营实际收入" json:"operator_income"`

	// 状态
	Status              int8            `gorm:"not null;default:0;index;comment:状态(0待结算/1已结算/2已取消)" json:"status"`
	SettledAt           *time.Time      `gorm:"comment:结算时间" json:"settled_at"`
	Remark              string          `gorm:"size:500;not null;default:'';comment:备注" json:"remark"`

	CreatedAt           time.Time       `gorm:"autoCreateTime;comment:创建时间" json:"created_at"`
	UpdatedAt           time.Time       `gorm:"autoUpdateTime;comment:更新时间" json:"updated_at"`
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

// ProfitShareConfig 利润分成配置（店铺与运营的分成比例配置）
type ProfitShareConfig struct {
	ID                  uint64          `gorm:"primaryKey;comment:主键ID" json:"id"`
	ShopID              uint64          `gorm:"not null;uniqueIndex:uk_shop_operator_config;comment:店铺ID" json:"shop_id"`
	OperatorID          int64           `gorm:"not null;uniqueIndex:uk_shop_operator_config;comment:运营ID" json:"operator_id"`
	PlatformShareRate   decimal.Decimal `gorm:"type:decimal(5,2);not null;default:5.00;comment:平台分成比例%" json:"platform_share_rate"`
	OperatorShareRate   decimal.Decimal `gorm:"type:decimal(5,2);not null;default:45.00;comment:运营分成比例%" json:"operator_share_rate"`
	ShopOwnerShareRate  decimal.Decimal `gorm:"type:decimal(5,2);not null;default:50.00;comment:店主分成比例%" json:"shop_owner_share_rate"`
	Status              int8            `gorm:"not null;default:1;comment:状态(1生效/2失效)" json:"status"`
	EffectiveFrom       time.Time       `gorm:"not null;comment:生效时间" json:"effective_from"`
	EffectiveTo         *time.Time      `gorm:"comment:失效时间" json:"effective_to"`
	Remark              string          `gorm:"size:500;not null;default:'';comment:备注" json:"remark"`
	CreatedAt           time.Time       `gorm:"autoCreateTime;comment:创建时间" json:"created_at"`
	UpdatedAt           time.Time       `gorm:"autoUpdateTime;comment:更新时间" json:"updated_at"`
}

func (ProfitShareConfig) TableName() string {
	return "profit_share_configs"
}

// OrderShipmentRecord 订单发货记录（运营发货时创建，分表）
type OrderShipmentRecord struct {
	ID                  uint64          `gorm:"primaryKey;comment:主键ID" json:"id"`
	ShopID              uint64          `gorm:"not null;index;comment:店铺ID" json:"shop_id"`
	OrderSN             string          `gorm:"size:64;not null;uniqueIndex;comment:订单编号" json:"order_sn"`
	OrderID             uint64          `gorm:"not null;index;comment:订单ID" json:"order_id"`
	ShopOwnerID         int64           `gorm:"not null;index;comment:店铺老板ID" json:"shop_owner_id"`
	OperatorID          int64           `gorm:"not null;index;comment:运营老板ID" json:"operator_id"`

	// 成本信息
	GoodsCost           decimal.Decimal `gorm:"type:decimal(15,2);not null;default:0.00;comment:商品成本" json:"goods_cost"`
	ShippingCost        decimal.Decimal `gorm:"type:decimal(15,2);not null;default:0.00;comment:运费成本" json:"shipping_cost"`
	TotalCost           decimal.Decimal `gorm:"type:decimal(15,2);not null;default:0.00;comment:总成本" json:"total_cost"`
	Currency            string          `gorm:"size:10;not null;default:'TWD';comment:货币代码" json:"currency"`

	// 预付款冻结信息
	FrozenAmount        decimal.Decimal `gorm:"type:decimal(15,2);not null;default:0.00;comment:冻结金额" json:"frozen_amount"`
	FrozenTransactionID uint64          `gorm:"not null;default:0;comment:冻结流水ID" json:"frozen_transaction_id"`

	// 发货信息
	ShippingCarrier     string          `gorm:"size:100;not null;default:'';comment:物流承运商" json:"shipping_carrier"`
	TrackingNumber      string          `gorm:"size:100;not null;default:'';comment:物流单号" json:"tracking_number"`
	ShippedAt           *time.Time      `gorm:"comment:发货时间" json:"shipped_at"`

	// 状态
	Status              int8            `gorm:"not null;default:0;index;comment:状态(0待发货/1已发货/2已完成/3已取消/4发货失败)" json:"status"`
	SettlementID        uint64          `gorm:"not null;default:0;comment:关联结算记录ID" json:"settlement_id"`
	Remark              string          `gorm:"size:500;not null;default:'';comment:备注" json:"remark"`

	CreatedAt           time.Time       `gorm:"autoCreateTime;comment:创建时间" json:"created_at"`
	UpdatedAt           time.Time       `gorm:"autoUpdateTime;comment:更新时间" json:"updated_at"`
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
