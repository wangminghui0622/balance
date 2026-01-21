package models

import "gorm.io/gorm"

// PushData 原始数据表(非订单)
// 对应表: push_data
type PushData struct {
	ID        uint64         `gorm:"column:id;primaryKey;autoIncrement;comment:主键ID" json:"id"`
	Data      string         `gorm:"column:data;type:json;comment:原始数据(非订单)-JSON格式" json:"data"`
	CreatedAt gorm.DeletedAt `gorm:"column:created_at;type:datetime;default:CURRENT_TIMESTAMP;comment:创建时间" json:"createdAt"`
	UpdatedAt gorm.DeletedAt `gorm:"column:updated_at;type:datetime;default:CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP;comment:更新时间" json:"updatedAt"`
	DeletedAt gorm.DeletedAt `gorm:"column:deleted_at;type:datetime;index:idx_deleted_at;comment:删除时间（软删除）" json:"deletedAt"`
}

// TableName 指定表名
func (PushData) TableName() string {
	return "push_data"
}

