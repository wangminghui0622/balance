package models

import (
	"time"

	"github.com/shopspring/decimal"
)

// PrepaymentAccount 预付款账户（店主发货前预付成本）
type PrepaymentAccount struct {
	ID              uint64          `gorm:"primaryKey;comment:主键ID" json:"id"`
	AdminID         int64           `gorm:"not null;uniqueIndex;comment:店铺老板ID" json:"admin_id"`
	Balance         decimal.Decimal `gorm:"type:decimal(15,2);not null;default:0.00;comment:可用余额" json:"balance"`
	FrozenAmount    decimal.Decimal `gorm:"type:decimal(15,2);not null;default:0.00;comment:冻结金额" json:"frozen_amount"`
	TotalRecharge   decimal.Decimal `gorm:"type:decimal(15,2);not null;default:0.00;comment:累计充值" json:"total_recharge"`
	TotalConsume    decimal.Decimal `gorm:"type:decimal(15,2);not null;default:0.00;comment:累计消费" json:"total_consume"`
	Currency        string          `gorm:"size:10;not null;default:'TWD';comment:货币代码" json:"currency"`
	Status          int8            `gorm:"not null;default:1;comment:状态(1正常/2冻结)" json:"status"`
	CreatedAt       time.Time       `gorm:"autoCreateTime;comment:创建时间" json:"created_at"`
	UpdatedAt       time.Time       `gorm:"autoUpdateTime;comment:更新时间" json:"updated_at"`
}

func (PrepaymentAccount) TableName() string {
	return "prepayment_accounts"
}

// DepositAccount 保证金账户（店主缴纳的保证金）
type DepositAccount struct {
	ID              uint64          `gorm:"primaryKey;comment:主键ID" json:"id"`
	AdminID         int64           `gorm:"not null;uniqueIndex;comment:店铺老板ID" json:"admin_id"`
	Balance         decimal.Decimal `gorm:"type:decimal(15,2);not null;default:0.00;comment:保证金余额" json:"balance"`
	RequiredAmount  decimal.Decimal `gorm:"type:decimal(15,2);not null;default:0.00;comment:应缴保证金" json:"required_amount"`
	Currency        string          `gorm:"size:10;not null;default:'TWD';comment:货币代码" json:"currency"`
	Status          int8            `gorm:"not null;default:1;comment:状态(1正常/2不足/3冻结)" json:"status"`
	CreatedAt       time.Time       `gorm:"autoCreateTime;comment:创建时间" json:"created_at"`
	UpdatedAt       time.Time       `gorm:"autoUpdateTime;comment:更新时间" json:"updated_at"`
}

func (DepositAccount) TableName() string {
	return "deposit_accounts"
}

// OperatorAccount 运营老板账户（运营收到的成本+分成）
type OperatorAccount struct {
	ID              uint64          `gorm:"primaryKey;comment:主键ID" json:"id"`
	AdminID         int64           `gorm:"not null;uniqueIndex;comment:运营老板ID" json:"admin_id"`
	Balance         decimal.Decimal `gorm:"type:decimal(15,2);not null;default:0.00;comment:可用余额" json:"balance"`
	FrozenAmount    decimal.Decimal `gorm:"type:decimal(15,2);not null;default:0.00;comment:冻结金额" json:"frozen_amount"`
	TotalEarnings   decimal.Decimal `gorm:"type:decimal(15,2);not null;default:0.00;comment:累计收益" json:"total_earnings"`
	TotalWithdrawn  decimal.Decimal `gorm:"type:decimal(15,2);not null;default:0.00;comment:累计提现" json:"total_withdrawn"`
	Currency        string          `gorm:"size:10;not null;default:'TWD';comment:货币代码" json:"currency"`
	Status          int8            `gorm:"not null;default:1;comment:状态(1正常/2冻结)" json:"status"`
	CreatedAt       time.Time       `gorm:"autoCreateTime;comment:创建时间" json:"created_at"`
	UpdatedAt       time.Time       `gorm:"autoUpdateTime;comment:更新时间" json:"updated_at"`
}

func (OperatorAccount) TableName() string {
	return "operator_accounts"
}

// AccountTransaction 账户流水（记录所有账户资金变动）
type AccountTransaction struct {
	ID              uint64          `gorm:"primaryKey;comment:主键ID" json:"id"`
	TransactionNo   string          `gorm:"size:64;not null;uniqueIndex;comment:流水号" json:"transaction_no"`
	AccountType     string          `gorm:"size:20;not null;index;comment:账户类型" json:"account_type"`
	AdminID         int64           `gorm:"not null;index;comment:账户所属用户ID" json:"admin_id"`
	TransactionType string          `gorm:"size:30;not null;index;comment:交易类型" json:"transaction_type"`
	Amount          decimal.Decimal `gorm:"type:decimal(15,2);not null;comment:金额(正入账/负出账)" json:"amount"`
	BalanceBefore   decimal.Decimal `gorm:"type:decimal(15,2);not null;comment:交易前余额" json:"balance_before"`
	BalanceAfter    decimal.Decimal `gorm:"type:decimal(15,2);not null;comment:交易后余额" json:"balance_after"`
	RelatedOrderSN  string          `gorm:"size:64;not null;default:'';index;comment:关联订单号" json:"related_order_sn"`
	RelatedID       uint64          `gorm:"not null;default:0;comment:关联ID" json:"related_id"`
	Remark          string          `gorm:"size:500;not null;default:'';comment:备注" json:"remark"`
	OperatorID      int64           `gorm:"not null;default:0;comment:操作人ID" json:"operator_id"`
	Status          int8            `gorm:"not null;default:1;index;comment:状态(0待审批/1已完成/2已拒绝)" json:"status"`
	CreatedAt       time.Time       `gorm:"autoCreateTime;index;comment:创建时间" json:"created_at"`
}

func (AccountTransaction) TableName() string {
	return "account_transactions"
}

// ShopOwnerCommissionAccount 店主佣金账户（店主利润分成）
type ShopOwnerCommissionAccount struct {
	ID              uint64          `gorm:"primaryKey;comment:主键ID" json:"id"`
	AdminID         int64           `gorm:"not null;uniqueIndex;comment:店铺老板ID" json:"admin_id"`
	Balance         decimal.Decimal `gorm:"type:decimal(15,2);not null;default:0.00;comment:可用余额" json:"balance"`
	FrozenAmount    decimal.Decimal `gorm:"type:decimal(15,2);not null;default:0.00;comment:冻结金额" json:"frozen_amount"`
	TotalEarnings   decimal.Decimal `gorm:"type:decimal(15,2);not null;default:0.00;comment:累计收益" json:"total_earnings"`
	TotalWithdrawn  decimal.Decimal `gorm:"type:decimal(15,2);not null;default:0.00;comment:累计提现" json:"total_withdrawn"`
	Currency        string          `gorm:"size:10;not null;default:'TWD';comment:货币代码" json:"currency"`
	Status          int8            `gorm:"not null;default:1;comment:状态(1正常/2冻结)" json:"status"`
	CreatedAt       time.Time       `gorm:"autoCreateTime;comment:创建时间" json:"created_at"`
	UpdatedAt       time.Time       `gorm:"autoUpdateTime;comment:更新时间" json:"updated_at"`
}

func (ShopOwnerCommissionAccount) TableName() string {
	return "shop_owner_commission_accounts"
}

// PlatformCommissionAccount 平台佣金账户（平台利润分成，单例）
type PlatformCommissionAccount struct {
	ID              uint64          `gorm:"primaryKey;comment:主键ID" json:"id"`
	Balance         decimal.Decimal `gorm:"type:decimal(15,2);not null;default:0.00;comment:可用余额" json:"balance"`
	FrozenAmount    decimal.Decimal `gorm:"type:decimal(15,2);not null;default:0.00;comment:冻结金额" json:"frozen_amount"`
	TotalEarnings   decimal.Decimal `gorm:"type:decimal(15,2);not null;default:0.00;comment:累计收益" json:"total_earnings"`
	TotalWithdrawn  decimal.Decimal `gorm:"type:decimal(15,2);not null;default:0.00;comment:累计提现" json:"total_withdrawn"`
	Currency        string          `gorm:"size:10;not null;default:'TWD';comment:货币代码" json:"currency"`
	Status          int8            `gorm:"not null;default:1;comment:状态(1正常/2冻结)" json:"status"`
	CreatedAt       time.Time       `gorm:"autoCreateTime;comment:创建时间" json:"created_at"`
	UpdatedAt       time.Time       `gorm:"autoUpdateTime;comment:更新时间" json:"updated_at"`
}

func (PlatformCommissionAccount) TableName() string {
	return "platform_commission_accounts"
}

// PenaltyBonusAccount 罚补账户（运营罚款和补贴）
type PenaltyBonusAccount struct {
	ID              uint64          `gorm:"primaryKey;autoIncrement;comment:主键ID" json:"id"`
	AdminID         int64           `gorm:"not null;uniqueIndex;comment:用户ID" json:"admin_id"`
	Balance         decimal.Decimal `gorm:"type:decimal(15,2);not null;default:0.00;comment:余额(正待付罚款/负待发补贴)" json:"balance"`
	TotalPenalty    decimal.Decimal `gorm:"type:decimal(15,2);not null;default:0.00;comment:累计罚款" json:"total_penalty"`
	TotalBonus      decimal.Decimal `gorm:"type:decimal(15,2);not null;default:0.00;comment:累计补贴" json:"total_bonus"`
	Currency        string          `gorm:"size:10;not null;default:'TWD';comment:货币代码" json:"currency"`
	Status          int8            `gorm:"not null;default:1;comment:状态(1正常/2冻结)" json:"status"`
	CreatedAt       time.Time       `gorm:"autoCreateTime;comment:创建时间" json:"created_at"`
	UpdatedAt       time.Time       `gorm:"autoUpdateTime;comment:更新时间" json:"updated_at"`
}

func (PenaltyBonusAccount) TableName() string {
	return "penalty_bonus_accounts"
}

// EscrowAccount 托管账户（临时托管店主预付款，待结算时分账）
type EscrowAccount struct {
	ID              uint64          `gorm:"primaryKey;autoIncrement;comment:主键ID" json:"id"`
	AdminID         int64           `gorm:"not null;uniqueIndex;comment:用户ID(店主)" json:"admin_id"`
	Balance         decimal.Decimal `gorm:"type:decimal(15,2);not null;default:0.00;comment:托管余额" json:"balance"`
	TotalIn         decimal.Decimal `gorm:"type:decimal(15,2);not null;default:0.00;comment:累计转入" json:"total_in"`
	TotalOut        decimal.Decimal `gorm:"type:decimal(15,2);not null;default:0.00;comment:累计转出" json:"total_out"`
	Currency        string          `gorm:"size:10;not null;default:'TWD';comment:货币代码" json:"currency"`
	Status          int8            `gorm:"not null;default:1;comment:状态(1正常/2冻结)" json:"status"`
	CreatedAt       time.Time       `gorm:"autoCreateTime;comment:创建时间" json:"created_at"`
	UpdatedAt       time.Time       `gorm:"autoUpdateTime;comment:更新时间" json:"updated_at"`
}

func (EscrowAccount) TableName() string {
	return "escrow_accounts"
}

// WithdrawApplication 提现申请
type WithdrawApplication struct {
	ID              uint64          `gorm:"primaryKey;comment:主键ID" json:"id"`
	ApplicationNo   string          `gorm:"size:64;not null;uniqueIndex;comment:申请单号" json:"application_no"`
	AdminID         int64           `gorm:"not null;index;comment:申请人ID" json:"admin_id"`
	AccountType     string          `gorm:"size:30;not null;index;comment:账户类型" json:"account_type"`
	Amount          decimal.Decimal `gorm:"type:decimal(15,2);not null;comment:提现金额" json:"amount"`
	Fee             decimal.Decimal `gorm:"type:decimal(15,2);not null;default:0.00;comment:手续费" json:"fee"`
	ActualAmount    decimal.Decimal `gorm:"type:decimal(15,2);not null;comment:实际到账金额" json:"actual_amount"`
	Currency        string          `gorm:"size:10;not null;default:'TWD';comment:货币代码" json:"currency"`
	CollectionAccountID uint64      `gorm:"not null;comment:收款账户ID" json:"collection_account_id"`
	Status          int8            `gorm:"not null;default:0;index;comment:状态(0待审核/1已通过/2已拒绝/3已打款)" json:"status"`
	AuditRemark     string          `gorm:"size:500;not null;default:'';comment:审核备注" json:"audit_remark"`
	AuditBy         int64           `gorm:"not null;default:0;comment:审核人ID" json:"audit_by"`
	AuditAt         *time.Time      `gorm:"comment:审核时间" json:"audit_at"`
	PaidAt          *time.Time      `gorm:"comment:打款时间" json:"paid_at"`
	Remark          string          `gorm:"size:500;not null;default:'';comment:申请备注" json:"remark"`
	CreatedAt       time.Time       `gorm:"autoCreateTime;index;comment:创建时间" json:"created_at"`
	UpdatedAt       time.Time       `gorm:"autoUpdateTime;comment:更新时间" json:"updated_at"`
}

func (WithdrawApplication) TableName() string {
	return "withdraw_applications"
}

// RechargeApplication 充值申请（线下充值审核）
type RechargeApplication struct {
	ID              uint64          `gorm:"primaryKey;comment:主键ID" json:"id"`
	ApplicationNo   string          `gorm:"size:64;not null;uniqueIndex;comment:申请单号" json:"application_no"`
	AdminID         int64           `gorm:"not null;index;comment:申请人ID" json:"admin_id"`
	AccountType     string          `gorm:"size:30;not null;index;comment:账户类型(prepayment/deposit)" json:"account_type"`
	Amount          decimal.Decimal `gorm:"type:decimal(15,2);not null;comment:充值金额" json:"amount"`
	Currency        string          `gorm:"size:10;not null;default:'TWD';comment:货币代码" json:"currency"`
	PaymentMethod   string          `gorm:"size:30;not null;comment:支付方式(bank_transfer/cash)" json:"payment_method"`
	PaymentProof    string          `gorm:"size:500;not null;default:'';comment:支付凭证图片URL" json:"payment_proof"`
	Status          int8            `gorm:"not null;default:0;index;comment:状态(0待审核/1已通过/2已拒绝)" json:"status"`
	AuditRemark     string          `gorm:"size:500;not null;default:'';comment:审核备注" json:"audit_remark"`
	AuditBy         int64           `gorm:"not null;default:0;comment:审核人ID" json:"audit_by"`
	AuditAt         *time.Time      `gorm:"comment:审核时间" json:"audit_at"`
	Remark          string          `gorm:"size:500;not null;default:'';comment:申请备注" json:"remark"`
	CreatedAt       time.Time       `gorm:"autoCreateTime;index;comment:创建时间" json:"created_at"`
	UpdatedAt       time.Time       `gorm:"autoUpdateTime;comment:更新时间" json:"updated_at"`
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
	TxTypeAdjustment     = "adjustment"      // 调账（虾皮退款/扣款）
)

// 账户状态常量
const (
	AccountStatusNormal = 1 // 正常
	AccountStatusFrozen = 2 // 冻结
)
