package models

import (
	"time"
)

// Shipment 发货记录模型（Shopee发货结果记录，分表）
type Shipment struct {
	ID              uint64     `gorm:"primaryKey;comment:主键ID" json:"id"`
	ShopID          uint64     `gorm:"not null;uniqueIndex:uk_shop_order;comment:店铺ID" json:"shop_id"`
	OrderSN         string     `gorm:"size:64;not null;uniqueIndex:uk_shop_order;comment:订单编号" json:"order_sn"`
	PackageNumber   string     `gorm:"size:64;not null;default:'';comment:包裹号" json:"package_number"`
	ShippingCarrier string     `gorm:"size:100;not null;comment:物流承运商" json:"shipping_carrier"`
	TrackingNumber  string     `gorm:"size:100;not null;index;comment:物流单号" json:"tracking_number"`
	ShipStatus      int8       `gorm:"not null;default:0;index;comment:发货状态(0待发货/1已发货/2失败)" json:"ship_status"`
	ShipTime        *time.Time `gorm:"comment:发货时间" json:"ship_time"`
	ErrorMessage    string     `gorm:"size:512;not null;default:'';comment:错误信息" json:"error_message"`
	Remark          string     `gorm:"size:512;not null;default:'';comment:备注" json:"remark"`
	CreatedAt       time.Time  `gorm:"autoCreateTime;comment:创建时间" json:"created_at"`
	UpdatedAt       time.Time  `gorm:"autoUpdateTime;comment:更新时间" json:"updated_at"`
}

// TableName 指定表名
func (Shipment) TableName() string {
	return "shipments"
}

// LogisticsChannel 物流渠道模型（店铺可用的物流渠道）
type LogisticsChannel struct {
	ID                   uint64    `gorm:"primaryKey;comment:主键ID" json:"id"`
	ShopID               uint64    `gorm:"not null;uniqueIndex:uk_shop_channel;index;comment:店铺ID" json:"shop_id"`
	LogisticsChannelID   uint64    `gorm:"not null;uniqueIndex:uk_shop_channel;comment:物流渠道ID" json:"logistics_channel_id"`
	LogisticsChannelName string    `gorm:"size:255;not null;comment:物流渠道名称" json:"logistics_channel_name"`
	CODEnabled           int8      `gorm:"not null;default:0;comment:是否支持货到付款" json:"cod_enabled"`
	Enabled              int8      `gorm:"not null;default:1;comment:是否启用" json:"enabled"`
	CreatedAt            time.Time `gorm:"autoCreateTime;comment:创建时间" json:"created_at"`
	UpdatedAt            time.Time `gorm:"autoUpdateTime;comment:更新时间" json:"updated_at"`
}

// TableName 指定表名
func (LogisticsChannel) TableName() string {
	return "logistics_channels"
}

// OperationLog 操作日志模型（系统操作日志，分表）
type OperationLog struct {
	ID            uint64    `gorm:"primaryKey;comment:主键ID" json:"id"`
	ShopID        uint64    `gorm:"not null;default:0;index;comment:店铺ID" json:"shop_id"`
	OrderSN       string    `gorm:"size:64;not null;default:'';index;comment:订单编号" json:"order_sn"`
	OperationType string    `gorm:"size:50;not null;index;comment:操作类型" json:"operation_type"`
	OperationDesc string    `gorm:"size:512;not null;default:'';comment:操作描述" json:"operation_desc"`
	RequestData   string    `gorm:"type:text;comment:请求数据" json:"request_data"`
	ResponseData  string    `gorm:"type:text;comment:响应数据" json:"response_data"`
	Status        int8      `gorm:"not null;default:1;comment:状态(1成功/0失败)" json:"status"`
	IP            string    `gorm:"size:50;not null;default:'';comment:操作IP" json:"ip"`
	CreatedAt     time.Time `gorm:"autoCreateTime;index;comment:创建时间" json:"created_at"`
}

// TableName 指定表名
func (OperationLog) TableName() string {
	return "operation_logs"
}
