package models

import (
	"time"
)

// ShopOperatorRelation 店铺-运营分配关系（店铺与运营的绑定关系）
type ShopOperatorRelation struct {
	ID           uint64    `gorm:"primaryKey;comment:主键ID" json:"id"`
	ShopID       uint64    `gorm:"not null;uniqueIndex:uk_shop_operator;comment:店铺ID" json:"shop_id"`
	ShopOwnerID  int64     `gorm:"not null;index;comment:店铺老板ID" json:"shop_owner_id"`
	OperatorID   int64     `gorm:"not null;uniqueIndex:uk_shop_operator;index;comment:运营老板ID" json:"operator_id"`
	Status       int8      `gorm:"not null;default:1;comment:状态(1正常/2暂停/3解除)" json:"status"`
	AssignedAt   time.Time `gorm:"not null;comment:分配时间" json:"assigned_at"`
	Remark       string    `gorm:"size:500;not null;default:'';comment:备注" json:"remark"`
	CreatedAt    time.Time `gorm:"autoCreateTime;comment:创建时间" json:"created_at"`
	UpdatedAt    time.Time `gorm:"autoUpdateTime;comment:更新时间" json:"updated_at"`

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
