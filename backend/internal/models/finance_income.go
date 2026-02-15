package models

import (
	"time"

	"github.com/shopspring/decimal"
)

// FinanceIncome Shopee钱包交易记录（从Shopee同步的财务流水，分表）
type FinanceIncome struct {
	ID                     uint64          `gorm:"primaryKey;comment:主键ID" json:"id"`
	ShopID                 uint64          `gorm:"not null;index;comment:店铺ID" json:"shop_id"`
	TransactionID          int64           `gorm:"not null;uniqueIndex;comment:Shopee交易ID" json:"transaction_id"`
	OrderSN                string          `gorm:"size:64;not null;index;comment:关联订单号" json:"order_sn"`
	RefundSN               string          `gorm:"size:64;not null;default:'';comment:退款单号" json:"refund_sn"`
	Status                 string          `gorm:"size:20;not null;default:'';comment:交易状态" json:"status"`
	WalletType             string          `gorm:"size:20;not null;default:'';comment:钱包类型" json:"wallet_type"`
	TransactionType        string          `gorm:"size:50;not null;index;comment:交易类型" json:"transaction_type"`
	Amount                 decimal.Decimal `gorm:"type:decimal(15,2);not null;default:0.00;comment:交易金额" json:"amount"`
	CurrentBalance         decimal.Decimal `gorm:"type:decimal(15,2);not null;default:0.00;comment:交易后余额" json:"current_balance"`
	TransactionTime        int64           `gorm:"not null;index;comment:交易时间戳" json:"transaction_time"`
	TransactionFee         decimal.Decimal `gorm:"type:decimal(15,2);not null;default:0.00;comment:交易手续费" json:"transaction_fee"`
	Description            string          `gorm:"size:500;not null;default:'';comment:交易描述" json:"description"`
	BuyerName              string          `gorm:"size:100;not null;default:'';comment:买家名称" json:"buyer_name"`
	Reason                 string          `gorm:"size:255;not null;default:'';comment:交易原因" json:"reason"`
	WithdrawalID           int64           `gorm:"not null;default:0;comment:提现ID" json:"withdrawal_id"`
	WithdrawalType         string          `gorm:"size:20;not null;default:'';comment:提现类型" json:"withdrawal_type"`
	TransactionTabType     string          `gorm:"size:50;not null;default:'';comment:交易标签类型" json:"transaction_tab_type"`
	MoneyFlow              string          `gorm:"size:20;not null;default:'';comment:资金流向" json:"money_flow"`
	SettlementHandleStatus int8            `gorm:"not null;default:0;comment:结算处理状态(0待处理/1已处理)" json:"settlement_handle_status"`
	CreatedAt              time.Time       `gorm:"autoCreateTime;comment:创建时间" json:"created_at"`
	UpdatedAt              time.Time       `gorm:"autoUpdateTime;comment:更新时间" json:"updated_at"`
}

// TableName 指定表名
func (FinanceIncome) TableName() string {
	return "finance_incomes"
}

// 结算处理状态
const (
	SettlementStatusPending   = 0 // 待结算
	SettlementStatusCompleted = 1 // 已结算
)

// 交易类型常量
const (
	TransactionTypeEscrowVerifiedAdd   = "ESCROW_VERIFIED_ADD"   // 托管收入（订单结算/打款）
	TransactionTypeWithdrawalCreated   = "WITHDRAWAL_CREATED"    // 提现创建
	TransactionTypeWithdrawalCompleted = "WITHDRAWAL_COMPLETED"  // 提现完成
	TransactionTypeRefund              = "REFUND"                // 退款
	TransactionTypeEscrowAdjustment    = "ESCROW_ADJUSTMENT"     // 托管调账
	TransactionTypeSellerAdjustment    = "SELLER_ADJUSTMENT"     // 卖家调账
	TransactionTypeCommissionAdjust    = "COMMISSION_ADJUSTMENT" // 佣金调账
	TransactionTypeServiceFeeAdjust    = "SERVICE_FEE_ADJUSTMENT" // 服务费调账
	TransactionTypeShippingFeeAdjust   = "SHIPPING_FEE_ADJUSTMENT" // 运费调账
)

// IsOrderIncome 是否为订单收入
func (f *FinanceIncome) IsOrderIncome() bool {
	return f.TransactionType == TransactionTypeEscrowVerifiedAdd && f.OrderSN != ""
}

// IsWithdrawal 是否为提现
func (f *FinanceIncome) IsWithdrawal() bool {
	return f.TransactionType == TransactionTypeWithdrawalCreated ||
		f.TransactionType == TransactionTypeWithdrawalCompleted
}

// IsAdjustment 是否为调账（退款、扣款等）
func (f *FinanceIncome) IsAdjustment() bool {
	return f.TransactionType == TransactionTypeRefund ||
		f.TransactionType == TransactionTypeEscrowAdjustment ||
		f.TransactionType == TransactionTypeSellerAdjustment ||
		f.TransactionType == TransactionTypeCommissionAdjust ||
		f.TransactionType == TransactionTypeServiceFeeAdjust ||
		f.TransactionType == TransactionTypeShippingFeeAdjust
}

// NeedsSettlementHandling 是否需要结算处理（打款或调账）
func (f *FinanceIncome) NeedsSettlementHandling() bool {
	return f.IsOrderIncome() || (f.IsAdjustment() && f.OrderSN != "")
}
