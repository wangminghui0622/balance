package models

import (
	"time"

	"github.com/shopspring/decimal"
)

// PrepaymentAccount 预付款账户
type PrepaymentAccount struct {
	ID              uint64          `gorm:"primaryKey" json:"id"`
	AdminID         int64           `gorm:"not null;uniqueIndex" json:"admin_id"`          // 店铺老板ID
	Balance         decimal.Decimal `gorm:"type:decimal(15,2);not null;default:0.00" json:"balance"`          // 可用余额
	FrozenAmount    decimal.Decimal `gorm:"type:decimal(15,2);not null;default:0.00" json:"frozen_amount"`    // 冻结金额
	TotalRecharge   decimal.Decimal `gorm:"type:decimal(15,2);not null;default:0.00" json:"total_recharge"`   // 累计充值
	TotalConsume    decimal.Decimal `gorm:"type:decimal(15,2);not null;default:0.00" json:"total_consume"`    // 累计消费
	Currency        string          `gorm:"size:10;not null;default:'TWD'" json:"currency"`
	Status          int8            `gorm:"not null;default:1" json:"status"`              // 1=正常 2=冻结
	CreatedAt       time.Time       `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt       time.Time       `gorm:"autoUpdateTime" json:"updated_at"`
}

func (PrepaymentAccount) TableName() string {
	return "prepayment_accounts"
}

// DepositAccount 保证金账户
type DepositAccount struct {
	ID              uint64          `gorm:"primaryKey" json:"id"`
	AdminID         int64           `gorm:"not null;uniqueIndex" json:"admin_id"`          // 店铺老板ID
	Balance         decimal.Decimal `gorm:"type:decimal(15,2);not null;default:0.00" json:"balance"`          // 保证金余额
	RequiredAmount  decimal.Decimal `gorm:"type:decimal(15,2);not null;default:0.00" json:"required_amount"`  // 应缴保证金
	Currency        string          `gorm:"size:10;not null;default:'TWD'" json:"currency"`
	Status          int8            `gorm:"not null;default:1" json:"status"`              // 1=正常 2=不足 3=冻结
	CreatedAt       time.Time       `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt       time.Time       `gorm:"autoUpdateTime" json:"updated_at"`
}

func (DepositAccount) TableName() string {
	return "deposit_accounts"
}

// OperatorAccount 运营老板账户
type OperatorAccount struct {
	ID              uint64          `gorm:"primaryKey" json:"id"`
	AdminID         int64           `gorm:"not null;uniqueIndex" json:"admin_id"`          // 运营老板ID
	Balance         decimal.Decimal `gorm:"type:decimal(15,2);not null;default:0.00" json:"balance"`          // 可用余额
	FrozenAmount    decimal.Decimal `gorm:"type:decimal(15,2);not null;default:0.00" json:"frozen_amount"`    // 冻结金额
	TotalEarnings   decimal.Decimal `gorm:"type:decimal(15,2);not null;default:0.00" json:"total_earnings"`   // 累计收益
	TotalWithdrawn  decimal.Decimal `gorm:"type:decimal(15,2);not null;default:0.00" json:"total_withdrawn"`  // 累计提现
	Currency        string          `gorm:"size:10;not null;default:'TWD'" json:"currency"`
	Status          int8            `gorm:"not null;default:1" json:"status"`              // 1=正常 2=冻结
	CreatedAt       time.Time       `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt       time.Time       `gorm:"autoUpdateTime" json:"updated_at"`
}

func (OperatorAccount) TableName() string {
	return "operator_accounts"
}

// AccountTransaction 账户流水
type AccountTransaction struct {
	ID              uint64          `gorm:"primaryKey" json:"id"`
	TransactionNo   string          `gorm:"size:64;not null;uniqueIndex" json:"transaction_no"`  // 流水号
	AccountType     string          `gorm:"size:20;not null;index" json:"account_type"`          // prepayment/deposit/operator
	AdminID         int64           `gorm:"not null;index" json:"admin_id"`                      // 账户所属用户ID
	TransactionType string          `gorm:"size:30;not null;index" json:"transaction_type"`      // 交易类型
	Amount          decimal.Decimal `gorm:"type:decimal(15,2);not null" json:"amount"`           // 金额 (正=入账 负=出账)
	BalanceBefore   decimal.Decimal `gorm:"type:decimal(15,2);not null" json:"balance_before"`   // 交易前余额
	BalanceAfter    decimal.Decimal `gorm:"type:decimal(15,2);not null" json:"balance_after"`    // 交易后余额
	RelatedOrderSN  string          `gorm:"size:64;not null;default:'';index" json:"related_order_sn"` // 关联订单号
	RelatedID       uint64          `gorm:"not null;default:0" json:"related_id"`                // 关联ID
	Remark          string          `gorm:"size:500;not null;default:''" json:"remark"`
	OperatorID      int64           `gorm:"not null;default:0" json:"operator_id"`               // 操作人ID
	Status          int8            `gorm:"not null;default:1;index" json:"status"`              // 0=待审批 1=已完成 2=已拒绝
	CreatedAt       time.Time       `gorm:"autoCreateTime;index" json:"created_at"`
}

func (AccountTransaction) TableName() string {
	return "account_transactions"
}

// ShopOwnerCommissionAccount 店主佣金账户
type ShopOwnerCommissionAccount struct {
	ID              uint64          `gorm:"primaryKey" json:"id"`
	AdminID         int64           `gorm:"not null;uniqueIndex" json:"admin_id"`          // 店铺老板ID
	Balance         decimal.Decimal `gorm:"type:decimal(15,2);not null;default:0.00" json:"balance"`          // 可用余额
	FrozenAmount    decimal.Decimal `gorm:"type:decimal(15,2);not null;default:0.00" json:"frozen_amount"`    // 冻结金额
	TotalEarnings   decimal.Decimal `gorm:"type:decimal(15,2);not null;default:0.00" json:"total_earnings"`   // 累计收益
	TotalWithdrawn  decimal.Decimal `gorm:"type:decimal(15,2);not null;default:0.00" json:"total_withdrawn"`  // 累计提现
	Currency        string          `gorm:"size:10;not null;default:'TWD'" json:"currency"`
	Status          int8            `gorm:"not null;default:1" json:"status"`              // 1=正常 2=冻结
	CreatedAt       time.Time       `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt       time.Time       `gorm:"autoUpdateTime" json:"updated_at"`
}

func (ShopOwnerCommissionAccount) TableName() string {
	return "shop_owner_commission_accounts"
}

// PlatformCommissionAccount 平台佣金账户
type PlatformCommissionAccount struct {
	ID              uint64          `gorm:"primaryKey" json:"id"`
	Balance         decimal.Decimal `gorm:"type:decimal(15,2);not null;default:0.00" json:"balance"`          // 可用余额
	FrozenAmount    decimal.Decimal `gorm:"type:decimal(15,2);not null;default:0.00" json:"frozen_amount"`    // 冻结金额
	TotalEarnings   decimal.Decimal `gorm:"type:decimal(15,2);not null;default:0.00" json:"total_earnings"`   // 累计收益
	TotalWithdrawn  decimal.Decimal `gorm:"type:decimal(15,2);not null;default:0.00" json:"total_withdrawn"`  // 累计提现
	Currency        string          `gorm:"size:10;not null;default:'TWD'" json:"currency"`
	Status          int8            `gorm:"not null;default:1" json:"status"`              // 1=正常 2=冻结
	CreatedAt       time.Time       `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt       time.Time       `gorm:"autoUpdateTime" json:"updated_at"`
}

func (PlatformCommissionAccount) TableName() string {
	return "platform_commission_accounts"
}

// PenaltyBonusAccount 罚补账户
type PenaltyBonusAccount struct {
	ID              uint64          `gorm:"primaryKey" json:"id"`
	Balance         decimal.Decimal `gorm:"type:decimal(15,2);not null;default:0.00" json:"balance"`          // 余额 (正=待支付罚款/负=待发放补贴)
	TotalPenalty    decimal.Decimal `gorm:"type:decimal(15,2);not null;default:0.00" json:"total_penalty"`    // 累计罚款
	TotalBonus      decimal.Decimal `gorm:"type:decimal(15,2);not null;default:0.00" json:"total_bonus"`      // 累计补贴
	Currency        string          `gorm:"size:10;not null;default:'TWD'" json:"currency"`
	Status          int8            `gorm:"not null;default:1" json:"status"`              // 1=正常 2=冻结
	CreatedAt       time.Time       `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt       time.Time       `gorm:"autoUpdateTime" json:"updated_at"`
}

func (PenaltyBonusAccount) TableName() string {
	return "penalty_bonus_accounts"
}

// EscrowAccount 保证金托管账户 (临时托管店主预付款，待结算时分账)
type EscrowAccount struct {
	ID              uint64          `gorm:"primaryKey" json:"id"`
	Balance         decimal.Decimal `gorm:"type:decimal(15,2);not null;default:0.00" json:"balance"`          // 托管余额
	TotalIn         decimal.Decimal `gorm:"type:decimal(15,2);not null;default:0.00" json:"total_in"`         // 累计转入
	TotalOut        decimal.Decimal `gorm:"type:decimal(15,2);not null;default:0.00" json:"total_out"`        // 累计转出
	Currency        string          `gorm:"size:10;not null;default:'TWD'" json:"currency"`
	Status          int8            `gorm:"not null;default:1" json:"status"`              // 1=正常 2=冻结
	CreatedAt       time.Time       `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt       time.Time       `gorm:"autoUpdateTime" json:"updated_at"`
}

func (EscrowAccount) TableName() string {
	return "escrow_accounts"
}

// WithdrawApplication 提现申请
type WithdrawApplication struct {
	ID              uint64          `gorm:"primaryKey" json:"id"`
	ApplicationNo   string          `gorm:"size:64;not null;uniqueIndex" json:"application_no"`   // 申请单号
	AdminID         int64           `gorm:"not null;index" json:"admin_id"`                       // 申请人ID
	AccountType     string          `gorm:"size:30;not null;index" json:"account_type"`           // 账户类型
	Amount          decimal.Decimal `gorm:"type:decimal(15,2);not null" json:"amount"`            // 提现金额
	Fee             decimal.Decimal `gorm:"type:decimal(15,2);not null;default:0.00" json:"fee"`  // 手续费
	ActualAmount    decimal.Decimal `gorm:"type:decimal(15,2);not null" json:"actual_amount"`     // 实际到账金额
	Currency        string          `gorm:"size:10;not null;default:'TWD'" json:"currency"`
	CollectionAccountID uint64      `gorm:"not null" json:"collection_account_id"`                // 收款账户ID
	Status          int8            `gorm:"not null;default:0;index" json:"status"`               // 0=待审核 1=已通过 2=已拒绝 3=已打款
	AuditRemark     string          `gorm:"size:500;not null;default:''" json:"audit_remark"`     // 审核备注
	AuditBy         int64           `gorm:"not null;default:0" json:"audit_by"`                   // 审核人ID
	AuditAt         *time.Time      `json:"audit_at"`                                             // 审核时间
	PaidAt          *time.Time      `json:"paid_at"`                                              // 打款时间
	Remark          string          `gorm:"size:500;not null;default:''" json:"remark"`           // 申请备注
	CreatedAt       time.Time       `gorm:"autoCreateTime;index" json:"created_at"`
	UpdatedAt       time.Time       `gorm:"autoUpdateTime" json:"updated_at"`
}

func (WithdrawApplication) TableName() string {
	return "withdraw_applications"
}

// RechargeApplication 充值申请 (线下充值审核)
type RechargeApplication struct {
	ID              uint64          `gorm:"primaryKey" json:"id"`
	ApplicationNo   string          `gorm:"size:64;not null;uniqueIndex" json:"application_no"`   // 申请单号
	AdminID         int64           `gorm:"not null;index" json:"admin_id"`                       // 申请人ID
	AccountType     string          `gorm:"size:30;not null;index" json:"account_type"`           // 账户类型: prepayment/deposit
	Amount          decimal.Decimal `gorm:"type:decimal(15,2);not null" json:"amount"`            // 充值金额
	Currency        string          `gorm:"size:10;not null;default:'TWD'" json:"currency"`
	PaymentMethod   string          `gorm:"size:30;not null" json:"payment_method"`               // 支付方式: bank_transfer/cash
	PaymentProof    string          `gorm:"size:500;not null;default:''" json:"payment_proof"`    // 支付凭证(图片URL)
	Status          int8            `gorm:"not null;default:0;index" json:"status"`               // 0=待审核 1=已通过 2=已拒绝
	AuditRemark     string          `gorm:"size:500;not null;default:''" json:"audit_remark"`     // 审核备注
	AuditBy         int64           `gorm:"not null;default:0" json:"audit_by"`                   // 审核人ID
	AuditAt         *time.Time      `json:"audit_at"`                                             // 审核时间
	Remark          string          `gorm:"size:500;not null;default:''" json:"remark"`           // 申请备注
	CreatedAt       time.Time       `gorm:"autoCreateTime;index" json:"created_at"`
	UpdatedAt       time.Time       `gorm:"autoUpdateTime" json:"updated_at"`
}

func (RechargeApplication) TableName() string {
	return "recharge_applications"
}

// 提现/充值申请状态常量
const (
	ApplicationStatusPending  = 0 // 待审核
	ApplicationStatusApproved = 1 // 已通过
	ApplicationStatusRejected = 2 // 已拒绝
	ApplicationStatusPaid     = 3 // 已打款 (仅提现)
)

// 账户类型常量
const (
	AccountTypePrepayment         = "prepayment"          // 预付款账户
	AccountTypeDeposit            = "deposit"             // 保证金账户
	AccountTypeOperator           = "operator"            // 运营账户(回款)
	AccountTypeShopOwnerCommission = "shop_owner_commission" // 店主佣金账户
	AccountTypePlatformCommission = "platform_commission" // 平台佣金账户
	AccountTypePenaltyBonus       = "penalty_bonus"       // 罚补账户
	AccountTypeEscrow             = "escrow"              // 托管账户
)

// 交易类型常量
const (
	TxTypeRecharge       = "recharge"        // 充值
	TxTypeWithdraw       = "withdraw"        // 提现
	TxTypeFreeze         = "freeze"          // 冻结
	TxTypeUnfreeze       = "unfreeze"        // 解冻
	TxTypeOrderPay       = "order_pay"       // 订单支付 (发货时扣款)
	TxTypeOrderRefund    = "order_refund"    // 订单退款
	TxTypeProfitShare    = "profit_share"    // 利润分成
	TxTypeCostSettle     = "cost_settle"     // 成本结算 (给运营)
	TxTypePlatformFee    = "platform_fee"    // 平台服务费
	TxTypeDepositPay     = "deposit_pay"     // 保证金缴纳
	TxTypeDepositRefund  = "deposit_refund"  // 保证金退还
)

// 账户状态常量
const (
	AccountStatusNormal = 1 // 正常
	AccountStatusFrozen = 2 // 冻结
)
