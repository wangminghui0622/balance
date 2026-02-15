package models

import (
	"time"

	"github.com/shopspring/decimal"
)

// OrderEscrow 订单结算明细模型（从Shopee同步的订单结算明细，分表）
type OrderEscrow struct {
	ID                       uint64          `gorm:"primaryKey;comment:主键ID" json:"id"`
	ShopID                   uint64          `gorm:"not null;uniqueIndex:uk_shop_order_escrow;comment:店铺ID" json:"shop_id"`
	OrderSN                  string          `gorm:"size:64;not null;uniqueIndex:uk_shop_order_escrow;comment:订单编号" json:"order_sn"`
	OrderID                  uint64          `gorm:"not null;index;comment:订单ID" json:"order_id"`
	Currency                 string          `gorm:"size:10;not null;default:'';comment:货币代码" json:"currency"`

	// 核心结算金额
	EscrowAmount             decimal.Decimal `gorm:"type:decimal(15,2);not null;default:0.00;comment:最终结算金额" json:"escrow_amount"`
	BuyerTotalAmount         decimal.Decimal `gorm:"type:decimal(15,2);not null;default:0.00;comment:买家支付总额" json:"buyer_total_amount"`
	OriginalPrice            decimal.Decimal `gorm:"type:decimal(15,2);not null;default:0.00;comment:商品原价" json:"original_price"`

	// 折扣相关
	SellerDiscount           decimal.Decimal `gorm:"type:decimal(15,2);not null;default:0.00;comment:卖家折扣" json:"seller_discount"`
	ShopeeDiscount           decimal.Decimal `gorm:"type:decimal(15,2);not null;default:0.00;comment:平台折扣" json:"shopee_discount"`
	VoucherFromSeller        decimal.Decimal `gorm:"type:decimal(15,2);not null;default:0.00;comment:卖家优惠券" json:"voucher_from_seller"`
	VoucherFromShopee        decimal.Decimal `gorm:"type:decimal(15,2);not null;default:0.00;comment:平台优惠券" json:"voucher_from_shopee"`
	Coins                    decimal.Decimal `gorm:"type:decimal(15,2);not null;default:0.00;comment:虾皮币抵扣" json:"coins"`

	// 运费相关
	BuyerPaidShippingFee     decimal.Decimal `gorm:"type:decimal(15,2);not null;default:0.00;comment:买家支付运费" json:"buyer_paid_shipping_fee"`
	FinalShippingFee         decimal.Decimal `gorm:"type:decimal(15,2);not null;default:0.00;comment:最终运费" json:"final_shipping_fee"`
	ActualShippingFee        decimal.Decimal `gorm:"type:decimal(15,2);not null;default:0.00;comment:实际运费" json:"actual_shipping_fee"`
	EstimatedShippingFee     decimal.Decimal `gorm:"type:decimal(15,2);not null;default:0.00;comment:预估运费" json:"estimated_shipping_fee"`
	ShippingFeeDiscount      decimal.Decimal `gorm:"type:decimal(15,2);not null;default:0.00;comment:运费折扣" json:"shipping_fee_discount"`
	SellerShippingDiscount   decimal.Decimal `gorm:"type:decimal(15,2);not null;default:0.00;comment:卖家运费折扣" json:"seller_shipping_discount"`
	ReverseShippingFee       decimal.Decimal `gorm:"type:decimal(15,2);not null;default:0.00;comment:退货运费" json:"reverse_shipping_fee"`

	// 费用相关
	CommissionFee            decimal.Decimal `gorm:"type:decimal(15,2);not null;default:0.00;comment:平台佣金" json:"commission_fee"`
	ServiceFee               decimal.Decimal `gorm:"type:decimal(15,2);not null;default:0.00;comment:服务费" json:"service_fee"`
	SellerTransactionFee     decimal.Decimal `gorm:"type:decimal(15,2);not null;default:0.00;comment:卖家交易手续费" json:"seller_transaction_fee"`
	BuyerTransactionFee      decimal.Decimal `gorm:"type:decimal(15,2);not null;default:0.00;comment:买家交易手续费" json:"buyer_transaction_fee"`
	CreditCardTransactionFee decimal.Decimal `gorm:"type:decimal(15,2);not null;default:0.00;comment:信用卡手续费" json:"credit_card_transaction_fee"`
	EscrowTax                decimal.Decimal `gorm:"type:decimal(15,2);not null;default:0.00;comment:托管税费" json:"escrow_tax"`
	CrossBorderTax           decimal.Decimal `gorm:"type:decimal(15,2);not null;default:0.00;comment:跨境税费" json:"cross_border_tax"`

	// 补贴/返还
	PaymentPromotion         decimal.Decimal `gorm:"type:decimal(15,2);not null;default:0.00;comment:支付促销" json:"payment_promotion"`
	CreditCardPromotion      decimal.Decimal `gorm:"type:decimal(15,2);not null;default:0.00;comment:信用卡促销" json:"credit_card_promotion"`
	SellerLostCompensation   decimal.Decimal `gorm:"type:decimal(15,2);not null;default:0.00;comment:卖家丢失补偿" json:"seller_lost_compensation"`
	SellerCoinCashBack       decimal.Decimal `gorm:"type:decimal(15,2);not null;default:0.00;comment:卖家虾皮币返现" json:"seller_coin_cash_back"`
	SellerReturnRefund       decimal.Decimal `gorm:"type:decimal(15,2);not null;default:0.00;comment:卖家退货退款" json:"seller_return_refund"`
	FinalProductProtection   decimal.Decimal `gorm:"type:decimal(15,2);not null;default:0.00;comment:商品保护费" json:"final_product_protection"`

	// 成本相关
	CostOfGoodsSold          decimal.Decimal `gorm:"type:decimal(15,2);not null;default:0.00;comment:商品成本" json:"cost_of_goods_sold"`
	OriginalCostOfGoodsSold  decimal.Decimal `gorm:"type:decimal(15,2);not null;default:0.00;comment:原始商品成本" json:"original_cost_of_goods_sold"`

	// 其他
	DrcAdjustableRefund      decimal.Decimal `gorm:"type:decimal(15,2);not null;default:0.00;comment:DRC可调整退款" json:"drc_adjustable_refund"`
	ItemsCount               int             `gorm:"not null;default:0;comment:商品数量" json:"items_count"`

	// 状态
	SyncStatus               int8            `gorm:"not null;default:0;comment:同步状态(0未同步/1已同步/2失败)" json:"sync_status"`
	SyncTime                 *time.Time      `gorm:"comment:同步时间" json:"sync_time"`
	SyncError                string          `gorm:"size:500;not null;default:'';comment:同步错误信息" json:"sync_error"`

	CreatedAt                time.Time       `gorm:"autoCreateTime;comment:创建时间" json:"created_at"`
	UpdatedAt                time.Time       `gorm:"autoUpdateTime;comment:更新时间" json:"updated_at"`
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
