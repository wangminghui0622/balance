package models

import (
	"time"
)

// ShopOperatorRelation 店铺-运营分配关系
type ShopOperatorRelation struct {
	ID           uint64    `gorm:"primaryKey" json:"id"`
	ShopID       uint64    `gorm:"not null;uniqueIndex:uk_shop_operator" json:"shop_id"`
	ShopOwnerID  int64     `gorm:"not null;index" json:"shop_owner_id"`  // 店铺老板ID
	OperatorID   int64     `gorm:"not null;uniqueIndex:uk_shop_operator;index" json:"operator_id"` // 运营老板ID
	Status       int8      `gorm:"not null;default:1" json:"status"`     // 1=正常 2=暂停 3=解除
	AssignedAt   time.Time `gorm:"not null" json:"assigned_at"`          // 分配时间
	Remark       string    `gorm:"size:500;not null;default:''" json:"remark"`
	CreatedAt    time.Time `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt    time.Time `gorm:"autoUpdateTime" json:"updated_at"`

	// 关联
	Shop     *Shop  `gorm:"foreignKey:ShopID;references:ShopID" json:"shop,omitempty"`
	Operator *Admin `gorm:"foreignKey:OperatorID;references:ID" json:"operator,omitempty"`
}

func (ShopOperatorRelation) TableName() string {
	return "shop_operator_relations"
}

// 关系状态常量
const (
	RelationStatusActive   = 1 // 正常
	RelationStatusPaused   = 2 // 暂停
	RelationStatusReleased = 3 // 解除
)
