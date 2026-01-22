package models

import (
	"time"

	"gorm.io/gorm"
)

// ShopeeToken Shopee Token 表
// 存储 Shopee API 的所有配置信息（partner_id, partner_key, shop_id, tokens, redirect 等）
type ShopeeToken struct {
	ID                    uint64         `gorm:"column:id;primaryKey;autoIncrement;comment:主键ID" json:"id"`
	ShopID                int64          `gorm:"column:shop_id;type:bigint;not null;uniqueIndex:idx_shop_id;comment:店铺ID" json:"shopId"`
	PartnerID             int64          `gorm:"column:partner_id;type:bigint;not null;comment:Partner ID" json:"partnerId"`
	PartnerKey            string         `gorm:"column:partner_key;type:varchar(255);not null;comment:Partner Key" json:"partnerKey"`
	AccessToken           string         `gorm:"column:access_token;type:varchar(255);not null;comment:Access Token" json:"accessToken"`
	RefreshToken          string         `gorm:"column:refresh_token;type:varchar(255);not null;comment:Refresh Token" json:"refreshToken"`
	TokenExpireAt         *time.Time     `gorm:"column:token_expire_at;type:datetime;comment:Access Token 过期时间" json:"tokenExpireAt"`
	RefreshTokenExpireAt  *time.Time     `gorm:"column:refresh_token_expire_at;type:datetime;comment:Refresh Token 过期时间" json:"refreshTokenExpireAt"`
	IsSandbox             bool           `gorm:"column:is_sandbox;type:tinyint(1);default:1;comment:是否沙箱环境" json:"isSandbox"`
	Redirect              string         `gorm:"column:redirect;type:varchar(255);comment:授权回调地址（域名）" json:"redirect"`
	CreatedAt             time.Time      `gorm:"column:created_at;type:datetime;default:CURRENT_TIMESTAMP;comment:创建时间" json:"createdAt"`
	UpdatedAt             time.Time      `gorm:"column:updated_at;type:datetime;default:CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP;comment:更新时间" json:"updatedAt"`
	DeletedAt             gorm.DeletedAt `gorm:"column:deleted_at;type:datetime;index:idx_deleted_at;comment:删除时间（软删除）" json:"deletedAt"`
}

// TableName 指定表名
func (ShopeeToken) TableName() string {
	return "shopee_token"
}

// ShopeeTokenRepository ShopeeToken 数据访问层
type ShopeeTokenRepository struct {
	db *gorm.DB
}

// NewShopeeTokenRepository 创建 ShopeeTokenRepository 实例
func NewShopeeTokenRepository(db *gorm.DB) *ShopeeTokenRepository {
	return &ShopeeTokenRepository{db: db}
}

// CreateOrUpdate 创建或更新 token（根据 shop_id）
func (r *ShopeeTokenRepository) CreateOrUpdate(token *ShopeeToken) error {
	var existing ShopeeToken
	err := r.db.Where("shop_id = ?", token.ShopID).First(&existing).Error
	
	if err == gorm.ErrRecordNotFound {
		// 不存在，创建新记录
		return r.db.Create(token).Error
	} else if err != nil {
		return err
	}
	
	// 存在，更新记录
	updates := map[string]interface{}{
		"partner_id":            token.PartnerID,
		"access_token":          token.AccessToken,
		"refresh_token":         token.RefreshToken,
		"token_expire_at":       token.TokenExpireAt,
		"refresh_token_expire_at": token.RefreshTokenExpireAt,
		"is_sandbox":            token.IsSandbox,
	}
	// 只有当字段不为空时才更新
	if token.PartnerKey != "" {
		updates["partner_key"] = token.PartnerKey
	}
	if token.Redirect != "" {
		updates["redirect"] = token.Redirect
	}
	return r.db.Model(&existing).Updates(updates).Error
}

// GetByShopID 根据 shop_id 查询 token
func (r *ShopeeTokenRepository) GetByShopID(shopID int64) (*ShopeeToken, error) {
	var token ShopeeToken
	err := r.db.Where("shop_id = ?", shopID).First(&token).Error
	if err != nil {
		return nil, err
	}
	return &token, nil
}

// GetAll 获取所有 Shopee 配置（用于多店铺场景）
func (r *ShopeeTokenRepository) GetAll() ([]*ShopeeToken, error) {
	var tokens []*ShopeeToken
	err := r.db.Find(&tokens).Error
	if err != nil {
		return nil, err
	}
	return tokens, nil
}

// UpdateTokens 更新 access_token 和 refresh_token
func (r *ShopeeTokenRepository) UpdateTokens(shopID int64, accessToken, refreshToken string, tokenExpireAt, refreshTokenExpireAt *time.Time) error {
	updates := map[string]interface{}{
		"access_token": accessToken,
	}
	
	if refreshToken != "" {
		updates["refresh_token"] = refreshToken
	}
	if tokenExpireAt != nil {
		updates["token_expire_at"] = tokenExpireAt
	}
	if refreshTokenExpireAt != nil {
		updates["refresh_token_expire_at"] = refreshTokenExpireAt
	}
	
	return r.db.Model(&ShopeeToken{}).Where("shop_id = ?", shopID).Updates(updates).Error
}

// Delete 软删除 token
func (r *ShopeeTokenRepository) Delete(shopID int64) error {
	return r.db.Where("shop_id = ?", shopID).Delete(&ShopeeToken{}).Error
}
