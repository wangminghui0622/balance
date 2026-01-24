package models

import (
	"gorm.io/gorm"
	"time"
)

type AuthConfig struct {
	ID         uint64         `gorm:"column:id;primaryKey;autoIncrement;comment:主键ID" json:"id"`
	PartnerID  int64          `gorm:"column:partner_id;type:bigint;not null;comment:Partner ID" json:"partnerId"`
	PartnerKey string         `gorm:"column:partner_key;type:varchar(255);not null;comment:Partner Key" json:"partnerKey"`
	IsSandbox  bool           `gorm:"column:is_sandbox;type:tinyint(1);default:1;comment:是否沙箱环境" json:"isSandbox"`
	Redirect   string         `gorm:"column:redirect;type:varchar(255);comment:授权回调地址（域名）" json:"redirect"`
	CreatedAt  time.Time      `gorm:"column:created_at;type:datetime;default:CURRENT_TIMESTAMP;comment:创建时间" json:"createdAt"`
	UpdatedAt  time.Time      `gorm:"column:updated_at;type:datetime;default:CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP;comment:更新时间" json:"updatedAt"`
	DeletedAt  gorm.DeletedAt `gorm:"column:deleted_at;type:datetime;index:idx_deleted_at;comment:删除时间（软删除）" json:"deletedAt"`
}

func (AuthConfig) TableName() string {
	return "auth_config"
}

type AuthRepository struct {
	db *gorm.DB
}

func NewAuthRepository(db *gorm.DB) *AuthRepository {
	return &AuthRepository{db: db}
}

func (r *AuthRepository) GetByPartnerId() (*AuthConfig, error) {
	var existing AuthConfig
	err := r.db.Where("partner_id = ?", 1203446).First(&existing).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return &existing, nil
		} else {
			return &existing, err
		}
	}
	return &existing, nil
}
