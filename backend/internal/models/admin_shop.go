package models

import (
	"time"

	"gorm.io/gorm"
)

// AdminShop admin 和 shop 的关联表（多对多关系）
type AdminShop struct {
	ID        uint64         `gorm:"column:id;primaryKey;autoIncrement;comment:主键ID" json:"id"`
	AdminID   int64          `gorm:"column:admin_id;type:bigint;not null;index:idx_admin_id;comment:管理员ID" json:"adminId"`
	ShopID    int64          `gorm:"column:shop_id;type:bigint;not null;index:idx_shop_id;comment:店铺ID" json:"shopId"`
	IsPrimary bool           `gorm:"column:is_primary;type:tinyint(1);default:0;comment:是否为主店铺" json:"isPrimary"`
	Status    int8           `gorm:"column:status;type:tinyint;default:1;comment:关联状态: 1正常 2禁用" json:"status"`
	Remark    string         `gorm:"column:remark;type:varchar(500);comment:备注" json:"remark"`
	CreatedAt time.Time      `gorm:"column:created_at;type:datetime;default:CURRENT_TIMESTAMP;comment:创建时间" json:"createdAt"`
	UpdatedAt time.Time      `gorm:"column:updated_at;type:datetime;default:CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP;comment:更新时间" json:"updatedAt"`
	DeletedAt gorm.DeletedAt `gorm:"column:deleted_at;type:datetime;index:idx_deleted_at;comment:删除时间（软删除）" json:"deletedAt"`
}

// TableName 指定表名
func (AdminShop) TableName() string {
	return "admin_shop"
}

// AdminShopRepository AdminShop 数据访问层
type AdminShopRepository struct {
	db *gorm.DB
}

// NewAdminShopRepository 创建 AdminShopRepository 实例
func NewAdminShopRepository(db *gorm.DB) *AdminShopRepository {
	return &AdminShopRepository{db: db}
}

// Create 创建关联关系
func (r *AdminShopRepository) Create(adminShop *AdminShop) error {
	return r.db.Create(adminShop).Error
}

// CreateOrUpdate 创建或更新关联关系（如果已存在则更新，不存在则创建）
func (r *AdminShopRepository) CreateOrUpdate(adminShop *AdminShop) error {
	var existing AdminShop
	err := r.db.Where("admin_id = ? AND shop_id = ?", adminShop.AdminID, adminShop.ShopID).First(&existing).Error

	if err == gorm.ErrRecordNotFound {
		// 不存在，创建新记录
		return r.db.Create(adminShop).Error
	} else if err != nil {
		return err
	}

	// 存在，更新记录
	return r.db.Model(&existing).Updates(map[string]interface{}{
		"is_primary": adminShop.IsPrimary,
		"status":     adminShop.Status,
		"remark":     adminShop.Remark,
	}).Error
}

// GetByAdminID 根据 admin_id 查询所有关联的店铺
func (r *AdminShopRepository) GetByAdminID(adminID int64) ([]*AdminShop, error) {
	var adminShops []*AdminShop
	err := r.db.Where("admin_id = ?", adminID).Find(&adminShops).Error
	if err != nil {
		return nil, err
	}
	return adminShops, nil
}

// GetByShopID 根据 shop_id 查询所有关联的管理员
func (r *AdminShopRepository) GetByShopID(shopID int64) ([]*AdminShop, error) {
	var adminShops []*AdminShop
	err := r.db.Where("shop_id = ? and status = ?", shopID, 1).Find(&adminShops).Error
	if err != nil {
		return nil, err
	}
	return adminShops, nil
}

// GetByAdminIDAndShopID 根据 admin_id 和 shop_id 查询关联关系
func (r *AdminShopRepository) GetByAdminIDAndShopID(adminID, shopID int64) (*AdminShop, error) {
	var adminShop AdminShop
	err := r.db.Where("admin_id = ? AND shop_id = ?", adminID, shopID).First(&adminShop).Error
	if err != nil {
		return nil, err
	}
	return &adminShop, nil
}

// Exists 检查关联关系是否存在
func (r *AdminShopRepository) Exists(adminID, shopID int64) (bool, error) {
	var count int64
	err := r.db.Model(&AdminShop{}).Where("admin_id = ? AND shop_id = ?", adminID, shopID).Count(&count).Error
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

// Delete 软删除关联关系
func (r *AdminShopRepository) Delete(adminID, shopID int64) error {
	return r.db.Where("admin_id = ? AND shop_id = ?", adminID, shopID).Delete(&AdminShop{}).Error
}

// SetPrimary 设置主店铺（一个 admin 只能有一个主店铺）
func (r *AdminShopRepository) SetPrimary(adminID, shopID int64) error {
	// 先将该 admin 的所有店铺都设置为非主店铺
	err := r.db.Model(&AdminShop{}).Where("admin_id = ?", adminID).Update("is_primary", false).Error
	if err != nil {
		return err
	}
	// 然后设置指定的店铺为主店铺
	return r.db.Model(&AdminShop{}).Where("admin_id = ? AND shop_id = ?", adminID, shopID).Update("is_primary", true).Error
}

// GetPrimaryShop 获取 admin 的主店铺
func (r *AdminShopRepository) GetPrimaryShop(adminID int64) (*AdminShop, error) {
	var adminShop AdminShop
	err := r.db.Where("admin_id = ? AND is_primary = ?", adminID, true).First(&adminShop).Error
	if err != nil {
		return nil, err
	}
	return &adminShop, nil
}

// UpdateStatus 更新关联状态
func (r *AdminShopRepository) UpdateStatus(adminID, shopID int64, status int8) error {
	return r.db.Model(&AdminShop{}).Where("admin_id = ? AND shop_id = ?", adminID, shopID).Update("status", status).Error
}
