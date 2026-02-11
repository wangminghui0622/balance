package operator

import (
	"context"
	"fmt"

	"balance/backend/internal/database"
	"balance/backend/internal/models"

	"gorm.io/gorm"
)

// ShopService 店铺服务（运营专用）
type ShopService struct {
	db *gorm.DB
}

// NewShopService 创建店铺服务
func NewShopService() *ShopService {
	return &ShopService{
		db: database.GetDB(),
	}
}

// ListShops 获取店铺列表（运营可查看所有店铺）
func (s *ShopService) ListShops(ctx context.Context, ownerID int64, keyword string, page, pageSize int) ([]models.ShopWithAuth, int64, error) {
	var shops []models.Shop
	var total int64

	query := s.db.Model(&models.Shop{})
	if ownerID > 0 {
		query = query.Where("admin_id = ?", ownerID)
	}
	if keyword != "" {
		query = query.Where("shop_name LIKE ? OR shop_id_str LIKE ?", "%"+keyword+"%", "%"+keyword+"%")
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	if err := query.Offset(offset).Limit(pageSize).Order("id DESC").Find(&shops).Error; err != nil {
		return nil, 0, err
	}

	result := make([]models.ShopWithAuth, len(shops))
	for i := range shops {
		shop := &shops[i]
		if shop.ShopIDStr == "" {
			shop.ShopIDStr = fmt.Sprintf("%d", shop.ShopID)
		}
		result[i] = models.ShopWithAuth{Shop: *shop}
	}

	return result, total, nil
}

// GetShop 获取店铺详情
func (s *ShopService) GetShop(ctx context.Context, shopID int64) (*models.Shop, error) {
	var shop models.Shop
	if err := s.db.Where("shop_id = ?", shopID).First(&shop).Error; err != nil {
		return nil, fmt.Errorf("店铺不存在")
	}
	return &shop, nil
}
