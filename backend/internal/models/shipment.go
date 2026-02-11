package models

import (
	"time"
)

// Shipment 发货记录模型
type Shipment struct {
	ID              uint64     `gorm:"primaryKey;autoIncrement" json:"id"`
	ShopID          uint64     `gorm:"not null;uniqueIndex:uk_shop_order" json:"shop_id"`
	OrderSN         string     `gorm:"size:64;not null;uniqueIndex:uk_shop_order" json:"order_sn"`
	PackageNumber   string     `gorm:"size:64;not null;default:''" json:"package_number"`
	ShippingCarrier string     `gorm:"size:100;not null" json:"shipping_carrier"`
	TrackingNumber  string     `gorm:"size:100;not null;index" json:"tracking_number"`
	ShipStatus      int8       `gorm:"not null;default:0;index" json:"ship_status"`
	ShipTime        *time.Time `json:"ship_time"`
	ErrorMessage    string     `gorm:"size:512;not null;default:''" json:"error_message"`
	CreatedAt       time.Time  `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt       time.Time  `gorm:"autoUpdateTime" json:"updated_at"`
}

// TableName 指定表名
func (Shipment) TableName() string {
	return "shipments"
}

// LogisticsChannel 物流渠道模型
type LogisticsChannel struct {
	ID                   uint64    `gorm:"primaryKey;autoIncrement" json:"id"`
	ShopID               uint64    `gorm:"not null;uniqueIndex:uk_shop_channel;index" json:"shop_id"`
	LogisticsChannelID   uint64    `gorm:"not null;uniqueIndex:uk_shop_channel" json:"logistics_channel_id"`
	LogisticsChannelName string    `gorm:"size:255;not null" json:"logistics_channel_name"`
	CODEnabled           int8      `gorm:"not null;default:0" json:"cod_enabled"`
	Enabled              int8      `gorm:"not null;default:1" json:"enabled"`
	CreatedAt            time.Time `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt            time.Time `gorm:"autoUpdateTime" json:"updated_at"`
}

// TableName 指定表名
func (LogisticsChannel) TableName() string {
	return "logistics_channels"
}

// OperationLog 操作日志模型
type OperationLog struct {
	ID            uint64    `gorm:"primaryKey;autoIncrement" json:"id"`
	ShopID        uint64    `gorm:"not null;default:0;index" json:"shop_id"`
	OrderSN       string    `gorm:"size:64;not null;default:'';index" json:"order_sn"`
	OperationType string    `gorm:"size:50;not null;index" json:"operation_type"`
	OperationDesc string    `gorm:"size:512;not null;default:''" json:"operation_desc"`
	RequestData   string    `gorm:"type:text" json:"request_data"`
	ResponseData  string    `gorm:"type:text" json:"response_data"`
	Status        int8      `gorm:"not null;default:1" json:"status"`
	IP            string    `gorm:"size:50;not null;default:''" json:"ip"`
	CreatedAt     time.Time `gorm:"autoCreateTime;index" json:"created_at"`
}

// TableName 指定表名
func (OperationLog) TableName() string {
	return "operation_logs"
}
