package models

import (
	"time"

	"github.com/shopspring/decimal"
)

// Order 订单模型
type Order struct {
	ID              uint64          `gorm:"primaryKey;autoIncrement" json:"id"`
	ShopID          uint64          `gorm:"not null;uniqueIndex:uk_shop_order" json:"shop_id"`
	OrderSN         string          `gorm:"size:64;not null;uniqueIndex:uk_shop_order;index" json:"order_sn"`
	Region          string          `gorm:"size:10;not null" json:"region"`
	OrderStatus     string          `gorm:"size:50;not null;index" json:"order_status"`
	StatusLocked    bool            `gorm:"not null;default:false" json:"status_locked"`
	StatusRemark    string          `gorm:"size:255;not null;default:''" json:"status_remark"`
	BuyerUserID     uint64          `gorm:"not null;default:0" json:"buyer_user_id"`
	BuyerUsername   string          `gorm:"size:255;not null;default:''" json:"buyer_username"`
	TotalAmount     decimal.Decimal `gorm:"type:decimal(15,2);not null;default:0.00" json:"total_amount"`
	Currency        string          `gorm:"size:10;not null;default:''" json:"currency"`
	ShippingCarrier string          `gorm:"size:100;not null;default:''" json:"shipping_carrier"`
	TrackingNumber  string          `gorm:"size:100;not null;default:''" json:"tracking_number"`
	ShipByDate      *time.Time      `gorm:"index" json:"ship_by_date"`
	PayTime         *time.Time      `json:"pay_time"`
	CreateTime      *time.Time      `gorm:"index" json:"create_time"`
	UpdateTime      *time.Time      `json:"update_time"`
	CreatedAt       time.Time       `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt       time.Time       `gorm:"autoUpdateTime" json:"updated_at"`

	// 关联
	Items   []OrderItem   `gorm:"foreignKey:OrderID" json:"items,omitempty"`
	Address *OrderAddress `gorm:"foreignKey:OrderID" json:"address,omitempty"`
}

// TableName 指定表名
func (Order) TableName() string {
	return "orders"
}

// CanShip 检查订单是否可以发货
func (o *Order) CanShip() bool {
	return o.OrderStatus == "READY_TO_SHIP"
}

// OrderItem 订单商品模型
type OrderItem struct {
	ID        uint64          `gorm:"primaryKey;autoIncrement" json:"id"`
	OrderID   uint64          `gorm:"not null;index" json:"order_id"`
	ShopID    uint64          `gorm:"not null;index:idx_shop_order" json:"shop_id"`
	OrderSN   string          `gorm:"size:64;not null;index:idx_shop_order" json:"order_sn"`
	ItemID    uint64          `gorm:"not null;index" json:"item_id"`
	ItemName  string          `gorm:"size:512;not null;default:''" json:"item_name"`
	ItemSKU   string          `gorm:"size:100;not null;default:''" json:"item_sku"`
	ModelID   uint64          `gorm:"not null;default:0" json:"model_id"`
	ModelName string          `gorm:"size:255;not null;default:''" json:"model_name"`
	ModelSKU  string          `gorm:"size:100;not null;default:''" json:"model_sku"`
	Quantity  int             `gorm:"not null;default:0" json:"quantity"`
	ItemPrice decimal.Decimal `gorm:"type:decimal(15,2);not null;default:0.00" json:"item_price"`
	CreatedAt time.Time       `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt time.Time       `gorm:"autoUpdateTime" json:"updated_at"`
}

// TableName 指定表名
func (OrderItem) TableName() string {
	return "order_items"
}

// OrderAddress 订单收货地址模型
type OrderAddress struct {
	ID          uint64    `gorm:"primaryKey;autoIncrement" json:"id"`
	OrderID     uint64    `gorm:"uniqueIndex;not null" json:"order_id"`
	ShopID      uint64    `gorm:"not null;index:idx_shop_order" json:"shop_id"`
	OrderSN     string    `gorm:"size:64;not null;index:idx_shop_order" json:"order_sn"`
	Name        string    `gorm:"size:255;not null;default:''" json:"name"`
	Phone       string    `gorm:"size:50;not null;default:''" json:"phone"`
	Town        string    `gorm:"size:255;not null;default:''" json:"town"`
	District    string    `gorm:"size:255;not null;default:''" json:"district"`
	City        string    `gorm:"size:255;not null;default:''" json:"city"`
	State       string    `gorm:"size:255;not null;default:''" json:"state"`
	Region      string    `gorm:"size:10;not null;default:''" json:"region"`
	Zipcode     string    `gorm:"size:20;not null;default:''" json:"zipcode"`
	FullAddress string    `gorm:"type:text" json:"full_address"`
	CreatedAt   time.Time `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt   time.Time `gorm:"autoUpdateTime" json:"updated_at"`
}

// TableName 指定表名
func (OrderAddress) TableName() string {
	return "order_addresses"
}
