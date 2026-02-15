package models

import (
	"time"
)

// CollectionAccount 收款账户模型（用户提现收款账户）
type CollectionAccount struct {
	ID          uint64    `gorm:"primaryKey;comment:主键ID" json:"id"`
	AdminID     int64     `gorm:"not null;index;comment:用户ID" json:"admin_id"`
	AccountType string    `gorm:"size:20;not null;comment:账户类型(wallet/bank)" json:"account_type"`
	AccountName string    `gorm:"size:100;not null;comment:账户名称" json:"account_name"`
	AccountNo   string    `gorm:"size:100;not null;comment:账户号码" json:"account_no"`
	BankName    string    `gorm:"size:100;not null;default:'';comment:银行名称" json:"bank_name"`
	BankBranch  string    `gorm:"size:200;not null;default:'';comment:银行支行" json:"bank_branch"`
	Payee       string    `gorm:"size:100;not null;comment:收款人姓名" json:"payee"`
	IsDefault   bool      `gorm:"default:false;comment:是否默认账户" json:"is_default"`
	Status      int8      `gorm:"default:1;comment:状态(1正常/2未激活)" json:"status"`
	CreatedAt   time.Time `gorm:"autoCreateTime;comment:创建时间" json:"created_at"`
	UpdatedAt   time.Time `gorm:"autoUpdateTime;comment:更新时间" json:"updated_at"`
}

func (CollectionAccount) TableName() string {
	return "collection_accounts"
}

// 收款账户类型常量
const (
	CollectionAccountTypeWallet = "wallet" // 电子钱包
	CollectionAccountTypeBank   = "bank"   // 银行账户
)

// 收款账户状态常量
const (
	CollectionAccountStatusActive   = 1 // 正常
	CollectionAccountStatusInactive = 2 // 未激活
)
