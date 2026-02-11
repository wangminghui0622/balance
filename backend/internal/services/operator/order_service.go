package operator

import (
	"context"
	"fmt"

	"balance/backend/internal/database"
	"balance/backend/internal/models"

	"gorm.io/gorm"
)

// OrderService 订单服务（运营专用）
type OrderService struct {
	db *gorm.DB
}

// NewOrderService 创建订单服务
func NewOrderService() *OrderService {
	return &OrderService{
		db: database.GetDB(),
	}
}

// ListOrders 获取订单列表（运营可查看所有订单）
func (s *OrderService) ListOrders(ctx context.Context, ownerID, shopID int64, status, startTime, endTime string, page, pageSize int) ([]models.Order, int64, error) {
	var orders []models.Order
	var total int64

	query := s.db.Model(&models.Order{})

	if ownerID > 0 {
		var shopIDs []uint64
		s.db.Model(&models.Shop{}).Where("admin_id = ?", ownerID).Pluck("shop_id", &shopIDs)
		if len(shopIDs) > 0 {
			query = query.Where("shop_id IN ?", shopIDs)
		} else {
			return []models.Order{}, 0, nil
		}
	}

	if shopID > 0 {
		query = query.Where("shop_id = ?", shopID)
	}
	if status != "" {
		query = query.Where("order_status = ?", status)
	}
	if startTime != "" {
		query = query.Where("create_time >= ?", startTime)
	}
	if endTime != "" {
		query = query.Where("create_time <= ?", endTime)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	if err := query.Offset(offset).Limit(pageSize).Order("id DESC").Find(&orders).Error; err != nil {
		return nil, 0, err
	}

	return orders, total, nil
}

// GetOrder 获取订单详情
func (s *OrderService) GetOrder(ctx context.Context, shopID int64, orderSN string) (*models.Order, error) {
	var order models.Order
	if err := s.db.Where("shop_id = ? AND order_sn = ?", shopID, orderSN).First(&order).Error; err != nil {
		return nil, fmt.Errorf("订单不存在")
	}
	return &order, nil
}
