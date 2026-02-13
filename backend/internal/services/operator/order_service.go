package operator

import (
	"context"
	"fmt"

	"balance/backend/internal/database"
	"balance/backend/internal/models"
	"balance/backend/internal/utils"

	"gorm.io/gorm"
)

// OrderService 订单服务（运营专用）
type OrderService struct {
	db        *gorm.DB
	shardedDB *database.ShardedDB
}

// NewOrderService 创建订单服务
func NewOrderService() *OrderService {
	db := database.GetDB()
	return &OrderService{
		db:        db,
		shardedDB: database.NewShardedDB(db),
	}
}

// ListOrders 获取订单列表（运营可查看所有订单）- 使用分表
func (s *OrderService) ListOrders(ctx context.Context, ownerID, shopID int64, status, startTime, endTime string, page, pageSize int) ([]models.Order, int64, error) {
	// 如果指定了shopID，直接查询对应分表
	if shopID > 0 {
		orderTable := database.GetOrderTableName(uint64(shopID))
		query := s.db.Table(orderTable).Where("shop_id = ?", shopID)
		if status != "" {
			query = query.Where("order_status = ?", status)
		}
		if startTime != "" {
			query = query.Where("create_time >= ?", startTime)
		}
		if endTime != "" {
			query = query.Where("create_time <= ?", endTime)
		}

		var total int64
		query.Count(&total)

		var orders []models.Order
		offset := (page - 1) * pageSize
		query.Offset(offset).Limit(pageSize).Order("id DESC").Find(&orders)
		return orders, total, nil
	}

	// 获取店铺ID列表
	var shopIDs []uint64
	if ownerID > 0 {
		s.db.Model(&models.Shop{}).Where("admin_id = ?", ownerID).Pluck("shop_id", &shopIDs)
		if len(shopIDs) == 0 {
			return []models.Order{}, 0, nil
		}
	}

	// 按分表索引分组店铺
	shardShops := make(map[int][]uint64)
	if len(shopIDs) > 0 {
		for _, sid := range shopIDs {
			idx := database.GetShardIndex(sid)
			shardShops[idx] = append(shardShops[idx], sid)
		}
	} else {
		// 遍历所有分表
		for i := 0; i < database.ShardCount; i++ {
			shardShops[i] = nil
		}
	}

	var allOrders []models.Order
	var total int64

	for idx, sids := range shardShops {
		orderTable := fmt.Sprintf("orders_%d", idx)
		query := s.db.Table(orderTable)

		if len(sids) > 0 {
			query = query.Where("shop_id IN ?", sids)
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

		var count int64
		query.Count(&count)
		total += count

		var orders []models.Order
		query.Order("id DESC").Find(&orders)
		allOrders = append(allOrders, orders...)
	}

	// 内存分页
	offset := (page - 1) * pageSize
	end := offset + pageSize
	if offset >= len(allOrders) {
		return []models.Order{}, total, nil
	}
	if end > len(allOrders) {
		end = len(allOrders)
	}

	return allOrders[offset:end], total, nil
}

// GetOrder 获取订单详情 - 使用分表
func (s *OrderService) GetOrder(ctx context.Context, shopID int64, orderSN string) (*models.Order, error) {
	orderTable := database.GetOrderTableName(uint64(shopID))
	var order models.Order
	if err := s.db.Table(orderTable).Where("shop_id = ? AND order_sn = ?", shopID, orderSN).First(&order).Error; err != nil {
		return nil, utils.ErrOrderNotFound
	}
	return &order, nil
}
