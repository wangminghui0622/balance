package models

import "time"

// Notification 站内消息通知
type Notification struct {
	ID        uint64    `gorm:"primaryKey;comment:主键ID" json:"id"`
	AdminID   int64     `gorm:"not null;index;comment:接收人ID" json:"admin_id"`
	ShopID    uint64    `gorm:"not null;index;comment:关联店铺ID" json:"shop_id"`
	OrderSN   string    `gorm:"size:64;not null;default:'';index;comment:关联订单号" json:"order_sn"`
	Type      string    `gorm:"size:30;not null;index;comment:消息类型" json:"type"`
	Title     string    `gorm:"size:200;not null;comment:消息标题" json:"title"`
	Content   string    `gorm:"type:text;not null;comment:消息内容" json:"content"`
	IsRead    bool      `gorm:"not null;default:false;index;comment:是否已读" json:"is_read"`
	CreatedAt time.Time `gorm:"autoCreateTime;index;comment:创建时间" json:"created_at"`
}

func (Notification) TableName() string {
	return "notifications"
}

// 消息类型常量
const (
	NotifyTypePrepaymentLow = "prepayment_low" // 预付款不足
)
