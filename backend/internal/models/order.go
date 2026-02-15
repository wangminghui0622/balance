package models

import (
	"time"

	"github.com/shopspring/decimal"
)

// Order 订单模型（从 Shopee 同步的订单信息，分表）
type Order struct {
	ID              uint64          `gorm:"primaryKey;comment:主键ID" json:"id"`
	ShopID          uint64          `gorm:"not null;uniqueIndex:uk_shop_order;comment:店铺ID" json:"shop_id"`
	OrderSN         string          `gorm:"size:64;not null;uniqueIndex:uk_shop_order;index;comment:订单编号" json:"order_sn"`
	Region          string          `gorm:"size:10;not null;comment:地区代码" json:"region"`
	OrderStatus     string          `gorm:"size:50;not null;index;comment:订单状态" json:"order_status"`
	StatusLocked    bool            `gorm:"not null;default:false;comment:状态是否锁定" json:"status_locked"`
	StatusRemark    string          `gorm:"size:255;not null;default:'';comment:状态备注" json:"status_remark"`
	BuyerUserID     uint64          `gorm:"not null;default:0;comment:买家用户ID" json:"buyer_user_id"`
	BuyerUsername   string          `gorm:"size:255;not null;default:'';comment:买家用户名" json:"buyer_username"`
	TotalAmount     decimal.Decimal `gorm:"type:decimal(15,2);not null;default:0.00;comment:订单总额" json:"total_amount"`
	Currency        string          `gorm:"size:10;not null;default:'';comment:货币代码" json:"currency"`
	ShippingCarrier string          `gorm:"size:100;not null;default:'';comment:物流承运商" json:"shipping_carrier"`
	TrackingNumber  string          `gorm:"size:100;not null;default:'';comment:物流单号" json:"tracking_number"`
	ShipByDate      *time.Time      `gorm:"index;comment:最晚发货时间" json:"ship_by_date"`
	PayTime         *time.Time      `gorm:"comment:支付时间" json:"pay_time"`
	CreateTime      *time.Time      `gorm:"index;comment:Shopee订单创建时间" json:"create_time"`
	UpdateTime      *time.Time      `gorm:"comment:Shopee订单更新时间" json:"update_time"`
	CreatedAt       time.Time       `gorm:"autoCreateTime;comment:记录创建时间" json:"created_at"`
	UpdatedAt       time.Time       `gorm:"autoUpdateTime;comment:记录更新时间" json:"updated_at"`

	// 关联
	Items   []OrderItem   `gorm:"foreignKey:OrderID" json:"items,omitempty"`
	Address *OrderAddress `gorm:"foreignKey:OrderID" json:"address,omitempty"`

	// 账款调整相关显示字段（非数据库字段，由业务逻辑填充）
	AdjustmentLabel1 string `gorm:"-" json:"adjustment_label_1,omitempty"` // 例如: "账款调整佣金：NT$8.00"
	AdjustmentLabel2 string `gorm:"-" json:"adjustment_label_2,omitempty"` // 例如: "订单账款调整：NT$36.00"
	AdjustmentLabel3 string `gorm:"-" json:"adjustment_label_3,omitempty"` // 例如: "虾皮订单账款调整：NT$46.00"
}

// TableName 指定表名
func (Order) TableName() string {
	return "orders"
}

// CanShip 检查订单是否可以发货
func (o *Order) CanShip() bool {
	return o.OrderStatus == "READY_TO_SHIP"
}

// OrderItem 订单商品模型（订单包含的商品明细，分表）
type OrderItem struct {
	ID        uint64          `gorm:"primaryKey;comment:主键ID" json:"id"`
	OrderID   uint64          `gorm:"not null;index;comment:订单ID" json:"order_id"`
	ShopID    uint64          `gorm:"not null;index:idx_shop_order;comment:店铺ID" json:"shop_id"`
	OrderSN   string          `gorm:"size:64;not null;index:idx_shop_order;comment:订单编号" json:"order_sn"`
	ItemID    uint64          `gorm:"not null;index;comment:商品ID" json:"item_id"`
	ItemName  string          `gorm:"size:512;not null;default:'';comment:商品名称" json:"item_name"`
	ItemSKU   string          `gorm:"size:100;not null;default:'';comment:商品SKU" json:"item_sku"`
	ModelID   uint64          `gorm:"not null;default:0;comment:规格ID" json:"model_id"`
	ModelName string          `gorm:"size:255;not null;default:'';comment:规格名称" json:"model_name"`
	ModelSKU  string          `gorm:"size:100;not null;default:'';comment:规格SKU" json:"model_sku"`
	Quantity  int             `gorm:"not null;default:0;comment:数量" json:"quantity"`
	ItemPrice decimal.Decimal `gorm:"type:decimal(15,2);not null;default:0.00;comment:单价" json:"item_price"`
	CreatedAt time.Time       `gorm:"autoCreateTime;comment:创建时间" json:"created_at"`
	UpdatedAt time.Time       `gorm:"autoUpdateTime;comment:更新时间" json:"updated_at"`
}

// TableName 指定表名
func (OrderItem) TableName() string {
	return "order_items"
}

// OrderAddress 订单收货地址模型（买家收货地址，分表）
type OrderAddress struct {
	ID          uint64    `gorm:"primaryKey;comment:主键ID" json:"id"`
	OrderID     uint64    `gorm:"uniqueIndex;not null;comment:订单ID" json:"order_id"`
	ShopID      uint64    `gorm:"not null;index:idx_shop_order;comment:店铺ID" json:"shop_id"`
	OrderSN     string    `gorm:"size:64;not null;index:idx_shop_order;comment:订单编号" json:"order_sn"`
	Name        string    `gorm:"size:255;not null;default:'';comment:收件人姓名" json:"name"`
	Phone       string    `gorm:"size:50;not null;default:'';comment:收件人电话" json:"phone"`
	Town        string    `gorm:"size:255;not null;default:'';comment:乡镇" json:"town"`
	District    string    `gorm:"size:255;not null;default:'';comment:区县" json:"district"`
	City        string    `gorm:"size:255;not null;default:'';comment:城市" json:"city"`
	State       string    `gorm:"size:255;not null;default:'';comment:省/州" json:"state"`
	Region      string    `gorm:"size:10;not null;default:'';comment:地区代码" json:"region"`
	Zipcode     string    `gorm:"size:20;not null;default:'';comment:邮编" json:"zipcode"`
	FullAddress string    `gorm:"type:text;comment:完整地址" json:"full_address"`
	CreatedAt   time.Time `gorm:"autoCreateTime;comment:创建时间" json:"created_at"`
	UpdatedAt   time.Time `gorm:"autoUpdateTime;comment:更新时间" json:"updated_at"`
}

// TableName 指定表名
func (OrderAddress) TableName() string {
	return "order_addresses"
}
