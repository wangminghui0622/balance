package models

import (
	"time"

	"github.com/shopspring/decimal"
)

// FinanceIncome Shopee钱包交易记录（财务收入）
type FinanceIncome struct {
	ID                     uint64          `gorm:"primaryKey" json:"id"`
	ShopID                 uint64          `gorm:"not null;index" json:"shop_id"`
	TransactionID          int64           `gorm:"not null;uniqueIndex" json:"transaction_id"`
	OrderSN                string          `gorm:"size:64;not null;index" json:"order_sn"`
	RefundSN               string          `gorm:"size:64;not null;default:''" json:"refund_sn"`
	Status                 string          `gorm:"size:20;not null;default:''" json:"status"`
	WalletType             string          `gorm:"size:20;not null;default:''" json:"wallet_type"`
	TransactionType        string          `gorm:"size:50;not null;index" json:"transaction_type"`
	Amount                 decimal.Decimal `gorm:"type:decimal(15,2);not null;default:0.00" json:"amount"`
	CurrentBalance         decimal.Decimal `gorm:"type:decimal(15,2);not null;default:0.00" json:"current_balance"`
	TransactionTime        int64           `gorm:"not null;index" json:"transaction_time"`
	TransactionFee         decimal.Decimal `gorm:"type:decimal(15,2);not null;default:0.00" json:"transaction_fee"`
	Description            string          `gorm:"size:500;not null;default:''" json:"description"`
	BuyerName              string          `gorm:"size:100;not null;default:''" json:"buyer_name"`
	Reason                 string          `gorm:"size:255;not null;default:''" json:"reason"`
	WithdrawalID           int64           `gorm:"not null;default:0" json:"withdrawal_id"`
	WithdrawalType         string          `gorm:"size:20;not null;default:''" json:"withdrawal_type"`
	TransactionTabType     string          `gorm:"size:50;not null;default:''" json:"transaction_tab_type"`
	MoneyFlow              string          `gorm:"size:20;not null;default:''" json:"money_flow"`
	SettlementHandleStatus int8            `gorm:"not null;default:0" json:"settlement_handle_status"`
	CreatedAt              time.Time       `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt              time.Time       `gorm:"autoUpdateTime" json:"updated_at"`
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
	TransactionTypeEscrowVerifiedAdd = "ESCROW_VERIFIED_ADD"  // 托管收入（订单结算）
	TransactionTypeWithdrawalCreated = "WITHDRAWAL_CREATED"   // 提现创建
	TransactionTypeWithdrawalCompleted = "WITHDRAWAL_COMPLETED" // 提现完成
	TransactionTypeRefund            = "REFUND"               // 退款
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
