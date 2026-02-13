package models

import (
	"time"
)

// CollectionAccount 收款账户模型
type CollectionAccount struct {
	ID          uint64    `gorm:"primaryKey" json:"id"`
	AdminID     int64     `gorm:"not null;index" json:"admin_id"`
	AccountType string    `gorm:"size:20;not null" json:"account_type"` // wallet/bank
	AccountName string    `gorm:"size:100;not null" json:"account_name"`
	AccountNo   string    `gorm:"size:100;not null" json:"account_no"`
	BankName    string    `gorm:"size:100;not null;default:''" json:"bank_name"`
	BankBranch  string    `gorm:"size:200;not null;default:''" json:"bank_branch"`
	Payee       string    `gorm:"size:100;not null" json:"payee"`
	IsDefault   bool      `gorm:"default:false" json:"is_default"`
	Status      int8      `gorm:"default:1" json:"status"` // 1=正常 2=未激活
	CreatedAt   time.Time `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt   time.Time `gorm:"autoUpdateTime" json:"updated_at"`
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
