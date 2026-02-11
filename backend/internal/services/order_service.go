package services

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"balance/backend/internal/consts"
	"balance/backend/internal/database"
	"balance/backend/internal/models"
	"balance/backend/internal/shopee"

	"github.com/redis/go-redis/v9"
	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

// OrderService 订单服务
type OrderService struct {
	db          *gorm.DB
	shopService *ShopService
}

// NewOrderService 创建订单服务
func NewOrderService() *OrderService {
	return &OrderService{
		db:          database.GetDB(),
		shopService: NewShopService(),
	}
}

// SyncOrders 同步订单
func (s *OrderService) SyncOrders(ctx context.Context, shopID uint64, timeFrom, timeTo time.Time, orderStatus string) error {
	rdb := database.GetRedis()
	lockKey := fmt.Sprintf(consts.KeySyncLock, shopID)

	ok, err := rdb.SetNX(ctx, lockKey, time.Now().Unix(), consts.SyncLockExpire).Result()
	if err != nil {
		return fmt.Errorf("获取同步锁失败: %w", err)
	}
	if !ok {
		return fmt.Errorf("正在同步中，请稍后再试")
	}
	defer rdb.Del(ctx, lockKey)

	shop, err := s.shopService.GetShop(ctx, shopID)
	if err != nil {
		return fmt.Errorf("店铺不存在: %w", err)
	}

	accessToken, err := s.shopService.GetAccessToken(ctx, shopID)
	if err != nil {
		return err
	}

	maxRange := int64(consts.ShopeeMaxTimeRange)
	fromTs := timeFrom.Unix()
	toTs := timeTo.Unix()

	for fromTs < toTs {
		endTs := fromTs + maxRange
		if endTs > toTs {
			endTs = toTs
		}

		if err := s.syncOrdersInRange(ctx, shopID, shop.Region, accessToken, fromTs, endTs, orderStatus); err != nil {
			return err
		}

		fromTs = endTs
	}

	return nil
}

func (s *OrderService) syncOrdersInRange(ctx context.Context, shopID uint64, region, accessToken string, timeFrom, timeTo int64, orderStatus string) error {
	client := shopee.NewClient(region)
	limiter := shopee.GetRateLimiter(shopID)

	cursor := ""
	pageSize := consts.ShopeeOrderListPageSize

	for {
		if err := limiter.Wait(ctx); err != nil {
			return fmt.Errorf("限流等待被取消: %w", err)
		}

		var listResp *shopee.OrderListResponse
		err := shopee.RetryWithBackoff(ctx, consts.ShopeeAPIRetryTimes, func() error {
			var err error
			listResp, err = client.GetOrderList(
				accessToken, shopID,
				"create_time",
				timeFrom,
				timeTo,
				pageSize,
				cursor,
				orderStatus,
			)
			return err
		})
		if err != nil {
			return fmt.Errorf("获取订单列表失败: %w", err)
		}

		if len(listResp.Response.OrderList) == 0 {
			break
		}

		orderSNs := make([]string, 0, len(listResp.Response.OrderList))
		for _, o := range listResp.Response.OrderList {
			orderSNs = append(orderSNs, o.OrderSN)
		}

		for i := 0; i < len(orderSNs); i += consts.ShopeeOrderDetailMaxSize {
			end := i + consts.ShopeeOrderDetailMaxSize
			if end > len(orderSNs) {
				end = len(orderSNs)
			}
			batch := orderSNs[i:end]

			if err := limiter.Wait(ctx); err != nil {
				return fmt.Errorf("限流等待被取消: %w", err)
			}

			var detailResp *shopee.OrderDetailResponse
			err := shopee.RetryWithBackoff(ctx, consts.ShopeeAPIRetryTimes, func() error {
				var err error
				detailResp, err = client.GetOrderDetail(accessToken, shopID, batch)
				return err
			})
			if err != nil {
				return fmt.Errorf("获取订单详情失败: %w", err)
			}

			for _, detail := range detailResp.Response.OrderList {
				if err := s.saveOrder(ctx, shopID, &detail); err != nil {
					fmt.Printf("保存订单失败 shop_id=%d order_sn=%s: %v\n", shopID, detail.OrderSN, err)
					continue
				}
			}
		}

		if !listResp.Response.More {
			break
		}
		cursor = listResp.Response.NextCursor
	}

	return nil
}

func (s *OrderService) saveOrder(ctx context.Context, shopID uint64, detail *shopee.OrderDetail) error {
	rdb := database.GetRedis()

	var existingOrder models.Order
	if err := s.db.Select("id", "order_status", "status_locked").
		Where("shop_id = ? AND order_sn = ?", shopID, detail.OrderSN).
		First(&existingOrder).Error; err == nil {
		if existingOrder.StatusLocked {
			return s.saveOrderWithoutStatus(ctx, shopID, detail, existingOrder.ID)
		}

		currentPriority, currentExists := consts.OrderStatusPriority[existingOrder.OrderStatus]
		newPriority, newExists := consts.OrderStatusPriority[detail.OrderStatus]
		if currentExists && newExists && newPriority < currentPriority {
			return s.saveOrderWithoutStatus(ctx, shopID, detail, existingOrder.ID)
		}
	}

	if detail.UpdateTime > 0 {
		updateTimeKey := fmt.Sprintf(consts.KeyOrderUpdateTime, shopID, detail.OrderSN)
		cachedTime, err := rdb.Get(ctx, updateTimeKey).Result()
		if err == nil {
			oldTime, _ := strconv.ParseInt(cachedTime, 10, 64)
			if detail.UpdateTime <= oldTime {
				return nil
			}
		}
		rdb.Set(ctx, updateTimeKey, detail.UpdateTime, consts.OrderUpdateTimeTTL)
	}

	return s.saveOrderFull(ctx, shopID, detail)
}

func (s *OrderService) saveOrderFull(ctx context.Context, shopID uint64, detail *shopee.OrderDetail) error {
	return s.db.Transaction(func(tx *gorm.DB) error {
		var createTime, updateTime, payTime, shipByDate *time.Time
		if detail.CreateTime > 0 {
			t := time.Unix(detail.CreateTime, 0)
			createTime = &t
		}
		if detail.UpdateTime > 0 {
			t := time.Unix(detail.UpdateTime, 0)
			updateTime = &t
		}
		if detail.PayTime > 0 {
			t := time.Unix(detail.PayTime, 0)
			payTime = &t
		}
		if detail.ShipByDate > 0 {
			t := time.Unix(detail.ShipByDate, 0)
			shipByDate = &t
		}

		order := models.Order{
			ShopID:          shopID,
			OrderSN:         detail.OrderSN,
			Region:          detail.Region,
			OrderStatus:     detail.OrderStatus,
			BuyerUserID:     uint64(detail.BuyerUserID),
			BuyerUsername:   detail.BuyerUsername,
			TotalAmount:     decimal.NewFromFloat(detail.TotalAmount),
			Currency:        detail.Currency,
			ShippingCarrier: detail.ShippingCarrier,
			TrackingNumber:  detail.TrackingNo,
			ShipByDate:      shipByDate,
			PayTime:         payTime,
			CreateTime:      createTime,
			UpdateTime:      updateTime,
		}

		if err := tx.Where("shop_id = ? AND order_sn = ?", shopID, detail.OrderSN).
			Assign(order).FirstOrCreate(&order).Error; err != nil {
			return err
		}

		if err := tx.Where("order_id = ?", order.ID).Delete(&models.OrderItem{}).Error; err != nil {
			return err
		}

		for _, item := range detail.ItemList {
			orderItem := models.OrderItem{
				OrderID:   order.ID,
				ShopID:    shopID,
				OrderSN:   detail.OrderSN,
				ItemID:    uint64(item.ItemID),
				ItemName:  item.ItemName,
				ItemSKU:   item.ItemSKU,
				ModelID:   uint64(item.ModelID),
				ModelName: item.ModelName,
				ModelSKU:  item.ModelSKU,
				Quantity:  item.ModelQuantity,
				ItemPrice: decimal.NewFromFloat(item.ModelOriginalPrice),
			}
			if err := tx.Create(&orderItem).Error; err != nil {
				return err
			}
		}

		if err := tx.Where("order_id = ?", order.ID).Delete(&models.OrderAddress{}).Error; err != nil {
			return err
		}

		addr := detail.RecipientAddress
		orderAddress := models.OrderAddress{
			OrderID:     order.ID,
			ShopID:      shopID,
			OrderSN:     detail.OrderSN,
			Name:        addr.Name,
			Phone:       addr.Phone,
			Town:        addr.Town,
			District:    addr.District,
			City:        addr.City,
			State:       addr.State,
			Region:      addr.Region,
			Zipcode:     addr.Zipcode,
			FullAddress: addr.FullAddress,
		}
		if err := tx.Create(&orderAddress).Error; err != nil {
			return err
		}

		rdb := database.GetRedis()
		statusKey := fmt.Sprintf(consts.KeyOrderStatus, shopID, detail.OrderSN)
		rdb.Set(ctx, statusKey, detail.OrderStatus, consts.OrderStatusExpire)

		return nil
	})
}

func (s *OrderService) saveOrderWithoutStatus(ctx context.Context, shopID uint64, detail *shopee.OrderDetail, orderID uint64) error {
	return s.db.Transaction(func(tx *gorm.DB) error {
		var updateTime, shipByDate *time.Time
		if detail.UpdateTime > 0 {
			t := time.Unix(detail.UpdateTime, 0)
			updateTime = &t
		}
		if detail.ShipByDate > 0 {
			t := time.Unix(detail.ShipByDate, 0)
			shipByDate = &t
		}

		updates := map[string]interface{}{
			"shipping_carrier": detail.ShippingCarrier,
			"tracking_number":  detail.TrackingNo,
			"ship_by_date":     shipByDate,
			"update_time":      updateTime,
		}
		if err := tx.Model(&models.Order{}).Where("id = ?", orderID).Updates(updates).Error; err != nil {
			return err
		}

		if err := tx.Where("order_id = ?", orderID).Delete(&models.OrderItem{}).Error; err != nil {
			return err
		}
		for _, item := range detail.ItemList {
			orderItem := models.OrderItem{
				OrderID:   orderID,
				ShopID:    shopID,
				OrderSN:   detail.OrderSN,
				ItemID:    uint64(item.ItemID),
				ItemName:  item.ItemName,
				ItemSKU:   item.ItemSKU,
				ModelID:   uint64(item.ModelID),
				ModelName: item.ModelName,
				ModelSKU:  item.ModelSKU,
				Quantity:  item.ModelQuantity,
				ItemPrice: decimal.NewFromFloat(item.ModelOriginalPrice),
			}
			if err := tx.Create(&orderItem).Error; err != nil {
				return err
			}
		}

		if err := tx.Where("order_id = ?", orderID).Delete(&models.OrderAddress{}).Error; err != nil {
			return err
		}
		addr := detail.RecipientAddress
		orderAddress := models.OrderAddress{
			OrderID:     orderID,
			ShopID:      shopID,
			OrderSN:     detail.OrderSN,
			Name:        addr.Name,
			Phone:       addr.Phone,
			Town:        addr.Town,
			District:    addr.District,
			City:        addr.City,
			State:       addr.State,
			Region:      addr.Region,
			Zipcode:     addr.Zipcode,
			FullAddress: addr.FullAddress,
		}
		if err := tx.Create(&orderAddress).Error; err != nil {
			return err
		}

		return nil
	})
}

// SaveOrderFromWebhook 从Webhook保存订单
func (s *OrderService) SaveOrderFromWebhook(ctx context.Context, shopID uint64, detail *shopee.OrderDetail) error {
	return s.saveOrder(ctx, shopID, detail)
}

// OrderQueryParams 订单查询参数
type OrderQueryParams struct {
	ShopID    uint64
	OrderSN   string
	Status    string
	StartTime string
	EndTime   string
	Page      int
	PageSize  int
	AdminID   int64
}

// ListOrders 获取订单列表
func (s *OrderService) ListOrders(ctx context.Context, params OrderQueryParams) ([]models.Order, int64, error) {
	var orders []models.Order
	var total int64

	query := s.db.Model(&models.Order{})

	if params.AdminID > 0 {
		var shopIDs []uint64
		if err := s.db.Model(&models.Shop{}).Where("admin_id = ?", params.AdminID).Pluck("shop_id", &shopIDs).Error; err != nil {
			return nil, 0, err
		}
		if len(shopIDs) == 0 {
			return orders, 0, nil
		}
		query = query.Where("shop_id IN ?", shopIDs)
	}

	if params.ShopID > 0 {
		query = query.Where("shop_id = ?", params.ShopID)
	}
	if params.Status != "" {
		query = query.Where("order_status = ?", params.Status)
	}
	if params.OrderSN != "" {
		query = query.Where("order_sn LIKE ?", "%"+params.OrderSN+"%")
	}

	if params.StartTime != "" {
		query = query.Where("create_time >= ?", params.StartTime)
	}
	if params.EndTime != "" {
		query = query.Where("create_time <= ?", params.EndTime+" 23:59:59")
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (params.Page - 1) * params.PageSize
	if err := query.Preload("Items").Preload("Address").
		Offset(offset).Limit(params.PageSize).
		Order("id DESC").Find(&orders).Error; err != nil {
		return nil, 0, err
	}

	return orders, total, nil
}

// GetOrder 获取订单详情
func (s *OrderService) GetOrder(ctx context.Context, shopID uint64, orderSN string) (*models.Order, error) {
	var order models.Order
	if err := s.db.Preload("Items").Preload("Address").
		Where("shop_id = ? AND order_sn = ?", shopID, orderSN).
		First(&order).Error; err != nil {
		return nil, err
	}
	return &order, nil
}

// GetReadyToShipOrders 获取待发货订单
func (s *OrderService) GetReadyToShipOrders(ctx context.Context, shopID uint64, page, pageSize int, adminID int64) ([]models.Order, int64, error) {
	params := OrderQueryParams{
		ShopID:   shopID,
		Status:   consts.OrderStatusReadyToShip,
		Page:     page,
		PageSize: pageSize,
		AdminID:  adminID,
	}
	return s.ListOrders(ctx, params)
}

// GetOrderStatus 从缓存或数据库获取订单状态
func (s *OrderService) GetOrderStatus(ctx context.Context, shopID uint64, orderSN string) (string, error) {
	rdb := database.GetRedis()
	statusKey := fmt.Sprintf(consts.KeyOrderStatus, shopID, orderSN)

	status, err := rdb.Get(ctx, statusKey).Result()
	if err == nil {
		return status, nil
	}
	if err != redis.Nil {
		return "", err
	}

	var order models.Order
	if err := s.db.Select("order_status").
		Where("shop_id = ? AND order_sn = ?", shopID, orderSN).
		First(&order).Error; err != nil {
		return "", err
	}

	rdb.Set(ctx, statusKey, order.OrderStatus, consts.OrderStatusExpire)

	return order.OrderStatus, nil
}

// RefreshOrderFromAPI 从API刷新单个订单
func (s *OrderService) RefreshOrderFromAPI(ctx context.Context, shopID uint64, orderSN string) error {
	shop, err := s.shopService.GetShop(ctx, shopID)
	if err != nil {
		return fmt.Errorf("店铺不存在: %w", err)
	}

	accessToken, err := s.shopService.GetAccessToken(ctx, shopID)
	if err != nil {
		return err
	}

	client := shopee.NewClient(shop.Region)

	detailResp, err := client.GetOrderDetail(accessToken, shopID, []string{orderSN})
	if err != nil {
		return fmt.Errorf("获取订单详情失败: %w", err)
	}

	if len(detailResp.Response.OrderList) == 0 {
		return fmt.Errorf("订单不存在")
	}

	return s.saveOrder(ctx, shopID, &detailResp.Response.OrderList[0])
}

// ForceUpdateStatus 强制更新订单状态
func (s *OrderService) ForceUpdateStatus(ctx context.Context, shopID uint64, orderSN string, newStatus string, remark string, lock bool) error {
	if _, ok := consts.OrderStatusPriority[newStatus]; !ok {
		return fmt.Errorf("无效的订单状态: %s", newStatus)
	}

	var order models.Order
	if err := s.db.Where("shop_id = ? AND order_sn = ?", shopID, orderSN).First(&order).Error; err != nil {
		return fmt.Errorf("订单不存在: %w", err)
	}

	oldStatus := order.OrderStatus

	updates := map[string]interface{}{
		"order_status":  newStatus,
		"status_locked": lock,
		"status_remark": remark,
	}
	if err := s.db.Model(&order).Updates(updates).Error; err != nil {
		return fmt.Errorf("更新状态失败: %w", err)
	}

	rdb := database.GetRedis()
	statusKey := fmt.Sprintf(consts.KeyOrderStatus, shopID, orderSN)
	updateTimeKey := fmt.Sprintf(consts.KeyOrderUpdateTime, shopID, orderSN)
	rdb.Del(ctx, statusKey, updateTimeKey)

	log := models.OperationLog{
		ShopID:        shopID,
		OrderSN:       orderSN,
		OperationType: "force_status_update",
		OperationDesc: fmt.Sprintf("强制更新状态: %s -> %s, 锁定: %v, 原因: %s", oldStatus, newStatus, lock, remark),
		Status:        consts.OpStatusSuccess,
	}
	s.db.Create(&log)

	return nil
}

// UnlockStatus 解锁订单状态
func (s *OrderService) UnlockStatus(ctx context.Context, shopID uint64, orderSN string) error {
	result := s.db.Model(&models.Order{}).
		Where("shop_id = ? AND order_sn = ?", shopID, orderSN).
		Updates(map[string]interface{}{
			"status_locked": false,
			"status_remark": "",
		})

	if result.Error != nil {
		return fmt.Errorf("解锁失败: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("订单不存在")
	}

	log := models.OperationLog{
		ShopID:        shopID,
		OrderSN:       orderSN,
		OperationType: "unlock_status",
		OperationDesc: "解锁订单状态，恢复自动更新",
		Status:        consts.OpStatusSuccess,
	}
	s.db.Create(&log)

	return nil
}

// IsStatusLocked 检查订单状态是否被锁定
func (s *OrderService) IsStatusLocked(ctx context.Context, shopID uint64, orderSN string) (bool, error) {
	var order models.Order
	if err := s.db.Select("status_locked").Where("shop_id = ? AND order_sn = ?", shopID, orderSN).First(&order).Error; err != nil {
		return false, err
	}
	return order.StatusLocked, nil
}
