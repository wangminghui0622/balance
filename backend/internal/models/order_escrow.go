package models

import (
	"time"

	"github.com/shopspring/decimal"
)

// OrderEscrow 订单结算明细模型
type OrderEscrow struct {
	ID                       uint64          `gorm:"primaryKey" json:"id"`
	ShopID                   uint64          `gorm:"not null;uniqueIndex:uk_shop_order_escrow" json:"shop_id"`
	OrderSN                  string          `gorm:"size:64;not null;uniqueIndex:uk_shop_order_escrow" json:"order_sn"`
	OrderID                  uint64          `gorm:"not null;index" json:"order_id"`
	Currency                 string          `gorm:"size:10;not null;default:''" json:"currency"`

	// 核心结算金额
	EscrowAmount             decimal.Decimal `gorm:"type:decimal(15,2);not null;default:0.00" json:"escrow_amount"`              // 最终结算金额
	BuyerTotalAmount         decimal.Decimal `gorm:"type:decimal(15,2);not null;default:0.00" json:"buyer_total_amount"`         // 买家支付总额
	OriginalPrice            decimal.Decimal `gorm:"type:decimal(15,2);not null;default:0.00" json:"original_price"`             // 商品原价

	// 折扣相关
	SellerDiscount           decimal.Decimal `gorm:"type:decimal(15,2);not null;default:0.00" json:"seller_discount"`            // 卖家折扣
	ShopeeDiscount           decimal.Decimal `gorm:"type:decimal(15,2);not null;default:0.00" json:"shopee_discount"`            // 平台折扣
	VoucherFromSeller        decimal.Decimal `gorm:"type:decimal(15,2);not null;default:0.00" json:"voucher_from_seller"`        // 卖家优惠券
	VoucherFromShopee        decimal.Decimal `gorm:"type:decimal(15,2);not null;default:0.00" json:"voucher_from_shopee"`        // 平台优惠券
	Coins                    decimal.Decimal `gorm:"type:decimal(15,2);not null;default:0.00" json:"coins"`                      // 虾皮币抵扣

	// 运费相关
	BuyerPaidShippingFee     decimal.Decimal `gorm:"type:decimal(15,2);not null;default:0.00" json:"buyer_paid_shipping_fee"`    // 买家支付运费
	FinalShippingFee         decimal.Decimal `gorm:"type:decimal(15,2);not null;default:0.00" json:"final_shipping_fee"`         // 最终运费
	ActualShippingFee        decimal.Decimal `gorm:"type:decimal(15,2);not null;default:0.00" json:"actual_shipping_fee"`        // 实际运费
	EstimatedShippingFee     decimal.Decimal `gorm:"type:decimal(15,2);not null;default:0.00" json:"estimated_shipping_fee"`     // 预估运费
	ShippingFeeDiscount      decimal.Decimal `gorm:"type:decimal(15,2);not null;default:0.00" json:"shipping_fee_discount"`      // 运费折扣
	SellerShippingDiscount   decimal.Decimal `gorm:"type:decimal(15,2);not null;default:0.00" json:"seller_shipping_discount"`   // 卖家运费折扣
	ReverseShippingFee       decimal.Decimal `gorm:"type:decimal(15,2);not null;default:0.00" json:"reverse_shipping_fee"`       // 退货运费

	// 费用相关
	CommissionFee            decimal.Decimal `gorm:"type:decimal(15,2);not null;default:0.00" json:"commission_fee"`             // 平台佣金
	ServiceFee               decimal.Decimal `gorm:"type:decimal(15,2);not null;default:0.00" json:"service_fee"`                // 服务费
	SellerTransactionFee     decimal.Decimal `gorm:"type:decimal(15,2);not null;default:0.00" json:"seller_transaction_fee"`     // 卖家交易手续费
	BuyerTransactionFee      decimal.Decimal `gorm:"type:decimal(15,2);not null;default:0.00" json:"buyer_transaction_fee"`      // 买家交易手续费
	CreditCardTransactionFee decimal.Decimal `gorm:"type:decimal(15,2);not null;default:0.00" json:"credit_card_transaction_fee"` // 信用卡手续费
	EscrowTax                decimal.Decimal `gorm:"type:decimal(15,2);not null;default:0.00" json:"escrow_tax"`                 // 托管税费
	CrossBorderTax           decimal.Decimal `gorm:"type:decimal(15,2);not null;default:0.00" json:"cross_border_tax"`           // 跨境税费

	// 补贴/返还
	PaymentPromotion         decimal.Decimal `gorm:"type:decimal(15,2);not null;default:0.00" json:"payment_promotion"`          // 支付促销
	CreditCardPromotion      decimal.Decimal `gorm:"type:decimal(15,2);not null;default:0.00" json:"credit_card_promotion"`      // 信用卡促销
	SellerLostCompensation   decimal.Decimal `gorm:"type:decimal(15,2);not null;default:0.00" json:"seller_lost_compensation"`   // 卖家丢失补偿
	SellerCoinCashBack       decimal.Decimal `gorm:"type:decimal(15,2);not null;default:0.00" json:"seller_coin_cash_back"`      // 卖家虾皮币返现
	SellerReturnRefund       decimal.Decimal `gorm:"type:decimal(15,2);not null;default:0.00" json:"seller_return_refund"`       // 卖家退货退款
	FinalProductProtection   decimal.Decimal `gorm:"type:decimal(15,2);not null;default:0.00" json:"final_product_protection"`   // 商品保护费

	// 成本相关
	CostOfGoodsSold          decimal.Decimal `gorm:"type:decimal(15,2);not null;default:0.00" json:"cost_of_goods_sold"`         // 商品成本
	OriginalCostOfGoodsSold  decimal.Decimal `gorm:"type:decimal(15,2);not null;default:0.00" json:"original_cost_of_goods_sold"` // 原始商品成本

	// 其他
	DrcAdjustableRefund      decimal.Decimal `gorm:"type:decimal(15,2);not null;default:0.00" json:"drc_adjustable_refund"`      // DRC可调整退款
	ItemsCount               int             `gorm:"not null;default:0" json:"items_count"`                                       // 商品数量

	// 状态
	SyncStatus               int8            `gorm:"not null;default:0" json:"sync_status"`                                       // 同步状态: 0-未同步 1-已同步 2-同步失败
	SyncTime                 *time.Time      `json:"sync_time"`                                                                   // 同步时间
	SyncError                string          `gorm:"size:500;not null;default:''" json:"sync_error"`                              // 同步错误信息

	CreatedAt                time.Time       `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt                time.Time       `gorm:"autoUpdateTime" json:"updated_at"`
}

// TableName 指定表名
func (OrderEscrow) TableName() string {
	return "order_escrows"
}

// 同步状态常量
const (
	EscrowSyncStatusPending = 0 // 未同步
	EscrowSyncStatusSuccess = 1 // 已同步
	EscrowSyncStatusFailed  = 2 // 同步失败
)

// OrderEscrowItem 订单结算商品明细
type OrderEscrowItem struct {
	ID                        uint64          `gorm:"primaryKey" json:"id"`
	EscrowID                  uint64          `gorm:"not null;index" json:"escrow_id"`
	ShopID                    uint64          `gorm:"not null;index" json:"shop_id"`
	OrderSN                   string          `gorm:"size:64;not null;index" json:"order_sn"`
	ItemID                    uint64          `gorm:"not null;index" json:"item_id"`
	ItemName                  string          `gorm:"size:512;not null;default:''" json:"item_name"`
	ItemSKU                   string          `gorm:"size:100;not null;default:''" json:"item_sku"`
	ModelID                   uint64          `gorm:"not null;default:0" json:"model_id"`
	ModelName                 string          `gorm:"size:255;not null;default:''" json:"model_name"`
	ModelSKU                  string          `gorm:"size:100;not null;default:''" json:"model_sku"`
	QuantityPurchased         int             `gorm:"not null;default:0" json:"quantity_purchased"`
	OriginalPrice             decimal.Decimal `gorm:"type:decimal(15,2);not null;default:0.00" json:"original_price"`
	DiscountedPrice           decimal.Decimal `gorm:"type:decimal(15,2);not null;default:0.00" json:"discounted_price"`
	SellerDiscount            decimal.Decimal `gorm:"type:decimal(15,2);not null;default:0.00" json:"seller_discount"`
	ShopeeDiscount            decimal.Decimal `gorm:"type:decimal(15,2);not null;default:0.00" json:"shopee_discount"`
	DiscountFromCoin          decimal.Decimal `gorm:"type:decimal(15,2);not null;default:0.00" json:"discount_from_coin"`
	DiscountFromVoucher       decimal.Decimal `gorm:"type:decimal(15,2);not null;default:0.00" json:"discount_from_voucher"`
	DiscountFromVoucherSeller decimal.Decimal `gorm:"type:decimal(15,2);not null;default:0.00" json:"discount_from_voucher_seller"`
	DiscountFromVoucherShopee decimal.Decimal `gorm:"type:decimal(15,2);not null;default:0.00" json:"discount_from_voucher_shopee"`
	ActivityType              string          `gorm:"size:50;not null;default:''" json:"activity_type"`
	ActivityID                uint64          `gorm:"not null;default:0" json:"activity_id"`
	CreatedAt                 time.Time       `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt                 time.Time       `gorm:"autoUpdateTime" json:"updated_at"`
}

// TableName 指定表名
func (OrderEscrowItem) TableName() string {
	return "order_escrow_items"
}
