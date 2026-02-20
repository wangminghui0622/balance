package models

import (
	"time"

	"github.com/shopspring/decimal"
)

// 退货退款状态常量（对应 Shopee ReturnStatus）
const (
	ReturnStatusRequested        = "REQUESTED"          // 买家发起退货
	ReturnStatusAccepted         = "ACCEPTED"           // 卖家同意退货
	ReturnStatusCancelled        = "CANCELLED"          // 退货取消
	ReturnStatusJudging          = "JUDGING"            // 平台仲裁中
	ReturnStatusRefundPaid       = "REFUND_PAID"        // 退款已到账
	ReturnStatusClosed           = "CLOSED"             // 退货关闭
	ReturnStatusProcessing       = "PROCESSING"         // 处理中
	ReturnStatusSellerDispute    = "SELLER_DISPUTE"     // 卖家申诉
	ReturnStatusSellerCompensate = "SELLER_COMPENSATION" // 卖家赔付
)

// 退货退款处理状态（系统内部）
const (
	ReturnRefundUnprocessed = 0 // 未处理（预付款未返还）
	ReturnRefundProcessed   = 1 // 已处理（预付款已返还）
	ReturnRefundSkipped     = 2 // 跳过（无需返还，如订单未扣除预付款）
	ReturnRefundProcessing  = 3 // 处理中（已标记，正在返还预付款）
	ReturnRefundFailed      = 4 // 处理失败（返还出错，需人工介入）
)

// Return 退货退款记录（从 Shopee 同步的退货退款信息，分表）
type Return struct {
	ID                uint64          `gorm:"primaryKey;comment:主键ID" json:"id"`
	ShopID            uint64          `gorm:"not null;uniqueIndex:uk_shop_return;index;comment:店铺ID" json:"shop_id"`
	ReturnSN          string          `gorm:"size:64;not null;uniqueIndex:uk_shop_return;comment:退货单号" json:"return_sn"`
	OrderSN           string          `gorm:"size:64;not null;index;comment:关联订单号" json:"order_sn"`
	Reason            string          `gorm:"size:100;not null;default:'';comment:退货原因" json:"reason"`
	TextReason        string          `gorm:"size:500;not null;default:'';comment:买家退货说明" json:"text_reason"`
	RefundAmount      decimal.Decimal `gorm:"type:decimal(15,2);not null;default:0.00;comment:退款金额" json:"refund_amount"`
	AmountBeforeDisc  decimal.Decimal `gorm:"type:decimal(15,2);not null;default:0.00;comment:折扣前金额" json:"amount_before_discount"`
	Currency          string          `gorm:"size:10;not null;default:'';comment:币种" json:"currency"`
	Status            string          `gorm:"size:50;not null;index;comment:退货状态(Shopee)" json:"status"`
	NeedsLogistics    bool            `gorm:"not null;default:false;comment:是否需要退回商品" json:"needs_logistics"`
	TrackingNumber    string          `gorm:"size:100;not null;default:'';comment:退货物流单号" json:"tracking_number"`
	LogisticsStatus   string          `gorm:"size:50;not null;default:'';comment:退货物流状态" json:"logistics_status"`
	BuyerUsername     string          `gorm:"size:255;not null;default:'';comment:买家用户名" json:"buyer_username"`
	ShopeeCreateTime  *time.Time      `gorm:"comment:Shopee退货创建时间" json:"shopee_create_time"`
	ShopeeUpdateTime  *time.Time      `gorm:"comment:Shopee退货更新时间" json:"shopee_update_time"`
	DueDate           *time.Time      `gorm:"comment:卖家处理截止时间" json:"due_date"`
	RefundStatus      int8            `gorm:"not null;default:0;index;comment:退款处理状态(0未处理/1已返还预付款/2跳过)" json:"refund_status"`
	RefundProcessedAt *time.Time      `gorm:"comment:退款处理时间" json:"refund_processed_at"`
	CreatedAt         time.Time       `gorm:"autoCreateTime;comment:记录创建时间" json:"created_at"`
	UpdatedAt         time.Time       `gorm:"autoUpdateTime;comment:记录更新时间" json:"updated_at"`
}

// TableName 指定表名
func (Return) TableName() string {
	return "returns"
}

// IsRefundConfirmed 退款是否已确认（ACCEPTED / REFUND_PAID 等状态表示退款已确定）
func (r *Return) IsRefundConfirmed() bool {
	return r.Status == ReturnStatusAccepted ||
		r.Status == ReturnStatusRefundPaid
}
