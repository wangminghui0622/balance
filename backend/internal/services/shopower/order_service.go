package shopower

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"balance/backend/internal/consts"
	"balance/backend/internal/database"
	"balance/backend/internal/models"
	"balance/backend/internal/shopee"
	"balance/backend/internal/utils"

	"github.com/go-redsync/redsync/v4"
	"github.com/redis/go-redis/v9"
	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

// OrderService 订单服务（店主专用）
type OrderService struct {
	db          *gorm.DB
	shopService *ShopService
	shardedDB   *database.ShardedDB
	rs          *redsync.Redsync
}

// NewOrderService 创建订单服务
func NewOrderService() *OrderService {
	db := database.GetDB()
	return &OrderService{
		db:          db,
		shopService: NewShopService(),
		shardedDB:   database.NewShardedDB(db),
		rs:          database.GetRedsync(),
	}
}

// SyncOrders 同步订单（支持分时间段、分页、限流、重试）
func (s *OrderService) SyncOrders(ctx context.Context, adminID int64, shopID int64, timeFrom, timeTo time.Time) (int, error) {
	shop, err := s.shopService.GetShop(ctx, adminID, shopID)
	if err != nil {
		return 0, err
	}

	// 使用 redsync 创建分布式锁
	lockKey := fmt.Sprintf(consts.KeySyncLock, shopID)
	mutex := s.rs.NewMutex(lockKey,
		redsync.WithExpiry(consts.SyncLockExpire),
		redsync.WithTries(1),
	)

	// 尝试获取锁并自动续期
	unlockFunc, acquired := utils.TryLockWithAutoExtend(ctx, mutex, consts.SyncLockExpire/3)
	if !acquired {
		return 0, utils.ErrShopSyncing
	}
	defer unlockFunc()

	accessToken, err := s.getAccessToken(ctx, uint64(shopID))
	if err != nil {
		return 0, fmt.Errorf("获取访问令牌失败: %w", err)
	}

	// 分时间段同步（Shopee API限制15天）
	maxRange := int64(consts.ShopeeMaxTimeRange)
	fromTs := timeFrom.Unix()
	toTs := timeTo.Unix()
	totalCount := 0

	for fromTs < toTs {
		endTs := fromTs + maxRange
		if endTs > toTs {
			endTs = toTs
		}

		count, err := s.syncOrdersInRange(ctx, uint64(shopID), shop.Region, accessToken, fromTs, endTs, "")
		if err != nil {
			return totalCount, err
		}
		totalCount += count
		fromTs = endTs
	}

	s.shopService.UpdateLastSyncTime(uint64(shopID))
	return totalCount, nil
}

// syncOrdersInRange 在指定时间范围内同步订单（支持分页、限流、重试）
func (s *OrderService) syncOrdersInRange(ctx context.Context, shopID uint64, region, accessToken string, timeFrom, timeTo int64, orderStatus string) (int, error) {
	client := shopee.NewClient(region)
	limiter := shopee.GetRateLimiter(shopID)

	cursor := ""
	pageSize := consts.ShopeeOrderListPageSize
	totalCount := 0

	for {
		if err := limiter.Wait(ctx); err != nil {
			return totalCount, fmt.Errorf("限流等待被取消: %w", err)
		}

		var listResp *shopee.OrderListResponse
		err := shopee.RetryWithBackoff(ctx, consts.ShopeeAPIRetryTimes, func() error {
			var err error
			listResp, err = client.GetOrderList(accessToken, shopID, "create_time", timeFrom, timeTo, pageSize, cursor, orderStatus)
			return err
		})
		if err != nil {
			return totalCount, fmt.Errorf("获取订单列表失败: %w", err)
		}

		if len(listResp.Response.OrderList) == 0 {
			break
		}

		orderSNs := make([]string, 0, len(listResp.Response.OrderList))
		for _, o := range listResp.Response.OrderList {
			orderSNs = append(orderSNs, o.OrderSN)
		}

		// 分批获取订单详情
		for i := 0; i < len(orderSNs); i += consts.ShopeeOrderDetailMaxSize {
			end := i + consts.ShopeeOrderDetailMaxSize
			if end > len(orderSNs) {
				end = len(orderSNs)
			}
			batch := orderSNs[i:end]

			if err := limiter.Wait(ctx); err != nil {
				return totalCount, fmt.Errorf("限流等待被取消: %w", err)
			}

			var detailResp *shopee.OrderDetailResponse
			err := shopee.RetryWithBackoff(ctx, consts.ShopeeAPIRetryTimes, func() error {
				var err error
				detailResp, err = client.GetOrderDetail(accessToken, shopID, batch)
				return err
			})
			if err != nil {
				return totalCount, fmt.Errorf("获取订单详情失败: %w", err)
			}

			for _, detail := range detailResp.Response.OrderList {
				if err := s.saveOrderFull(ctx, shopID, &detail); err != nil {
					fmt.Printf("保存订单失败 shop_id=%d order_sn=%s: %v\n", shopID, detail.OrderSN, err)
					continue
				}
				totalCount++
			}
		}

		if !listResp.Response.More {
			break
		}
		cursor = listResp.Response.NextCursor
	}

	return totalCount, nil
}

func (s *OrderService) getAccessToken(ctx context.Context, shopID uint64) (string, error) {
	var auth models.ShopAuthorization
	if err := s.db.Where("shop_id = ?", shopID).First(&auth).Error; err != nil {
		return "", utils.ErrShopUnauthorized
	}
	if auth.IsAccessTokenExpired() {
		if err := s.shopService.RefreshToken(ctx, shopID); err != nil {
			return "", err
		}
		s.db.Where("shop_id = ?", shopID).First(&auth)
	}
	return auth.AccessToken, nil
}

func (s *OrderService) saveOrder(ctx context.Context, shopID uint64, detail *shopee.OrderDetail) error {
	rdb := database.GetRedis()
	orderTable := database.GetOrderTableName(shopID)

	var existingOrder models.Order
	if err := s.db.Table(orderTable).Select("id", "order_status", "status_locked").
		Where("shop_id = ? AND order_sn = ?", shopID, detail.OrderSN).
		First(&existingOrder).Error; err == nil {
		// 状态锁定时只更新部分字段
		if existingOrder.StatusLocked {
			return s.saveOrderWithoutStatus(ctx, shopID, detail, existingOrder.ID)
		}

		// 状态优先级检查
		currentPriority, currentExists := consts.OrderStatusPriority[existingOrder.OrderStatus]
		newPriority, newExists := consts.OrderStatusPriority[detail.OrderStatus]
		if currentExists && newExists && newPriority < currentPriority {
			return s.saveOrderWithoutStatus(ctx, shopID, detail, existingOrder.ID)
		}
	}

	// 检查更新时间缓存（使用Lua脚本保证原子性）
	if detail.UpdateTime > 0 {
		updateTimeKey := fmt.Sprintf(consts.KeyOrderUpdateTime, shopID, detail.OrderSN)
		if !s.checkAndSetUpdateTime(ctx, rdb, updateTimeKey, detail.UpdateTime) {
			return nil
		}
	}

	return s.saveOrderFull(ctx, shopID, detail)
}

// saveOrderFull 完整保存订单（包含商品和地址）- 使用分表
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

		// 使用分表
		orderTable := database.GetOrderTableName(shopID)
		orderItemTable := database.GetOrderItemTableName(shopID)
		orderAddressTable := database.GetOrderAddressTableName(shopID)

		if err := tx.Table(orderTable).Where("shop_id = ? AND order_sn = ?", shopID, detail.OrderSN).
			Assign(order).FirstOrCreate(&order).Error; err != nil {
			return err
		}

		// 保存订单商品
		if err := tx.Table(orderItemTable).Where("order_id = ?", order.ID).Delete(&models.OrderItem{}).Error; err != nil {
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
			if err := tx.Table(orderItemTable).Create(&orderItem).Error; err != nil {
				return err
			}
		}

		// 保存订单地址
		if err := tx.Table(orderAddressTable).Where("order_id = ?", order.ID).Delete(&models.OrderAddress{}).Error; err != nil {
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
		if err := tx.Table(orderAddressTable).Create(&orderAddress).Error; err != nil {
			return err
		}

		// 缓存订单状态
		rdb := database.GetRedis()
		statusKey := fmt.Sprintf(consts.KeyOrderStatus, shopID, detail.OrderSN)
		rdb.Set(ctx, statusKey, detail.OrderStatus, consts.OrderStatusExpire)

		return nil
	})
}

// saveOrderWithoutStatus 保存订单但不更新状态 - 使用分表
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

		// 使用分表
		orderTable := database.GetOrderTableName(shopID)
		orderItemTable := database.GetOrderItemTableName(shopID)
		orderAddressTable := database.GetOrderAddressTableName(shopID)

		updates := map[string]interface{}{
			"shipping_carrier": detail.ShippingCarrier,
			"tracking_number":  detail.TrackingNo,
			"ship_by_date":     shipByDate,
			"update_time":      updateTime,
		}
		if err := tx.Table(orderTable).Where("id = ?", orderID).Updates(updates).Error; err != nil {
			return err
		}

		// 更新订单商品
		if err := tx.Table(orderItemTable).Where("order_id = ?", orderID).Delete(&models.OrderItem{}).Error; err != nil {
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
			if err := tx.Table(orderItemTable).Create(&orderItem).Error; err != nil {
				return err
			}
		}

		// 更新订单地址
		if err := tx.Table(orderAddressTable).Where("order_id = ?", orderID).Delete(&models.OrderAddress{}).Error; err != nil {
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
		if err := tx.Table(orderAddressTable).Create(&orderAddress).Error; err != nil {
			return err
		}

		return nil
	})
}

// SaveOrderFromWebhook 从Webhook保存订单
func (s *OrderService) SaveOrderFromWebhook(ctx context.Context, shopID uint64, detail *shopee.OrderDetail) error {
	return s.saveOrder(ctx, shopID, detail)
}

// ListOrders 获取订单列表 - 使用分表
func (s *OrderService) ListOrders(ctx context.Context, adminID int64, shopID int64, status, startTime, endTime string, page, pageSize int) ([]models.Order, int64, error) {
	var shopIDs []uint64
	s.db.Model(&models.Shop{}).Where("admin_id = ?", adminID).Pluck("shop_id", &shopIDs)
	if len(shopIDs) == 0 {
		return []models.Order{}, 0, nil
	}

	// 如果指定了shopID，只查询对应分表
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
		if err := query.Count(&total).Error; err != nil {
			return nil, 0, err
		}

		var orders []models.Order
		offset := (page - 1) * pageSize
		if err := query.Offset(offset).Limit(pageSize).Order("id DESC").Find(&orders).Error; err != nil {
			return nil, 0, err
		}

		for i := range orders {
			s.fillOrderLabels(&orders[i])
		}
		return orders, total, nil
	}

	// 未指定shopID，需要查询多个分表（按店铺分组查询）
	var allOrders []models.Order
	var total int64

	// 按分表索引分组店铺
	shardShops := make(map[int][]uint64)
	for _, sid := range shopIDs {
		idx := database.GetShardIndex(sid)
		shardShops[idx] = append(shardShops[idx], sid)
	}

	// 查询每个分表
	for idx, sids := range shardShops {
		orderTable := fmt.Sprintf("orders_%d", idx)
		query := s.db.Table(orderTable).Where("shop_id IN ?", sids)
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

	// 内存排序和分页
	// 按ID降序排序
	for i := 0; i < len(allOrders)-1; i++ {
		for j := i + 1; j < len(allOrders); j++ {
			if allOrders[i].ID < allOrders[j].ID {
				allOrders[i], allOrders[j] = allOrders[j], allOrders[i]
			}
		}
	}

	// 分页
	offset := (page - 1) * pageSize
	end := offset + pageSize
	if offset >= len(allOrders) {
		return []models.Order{}, total, nil
	}
	if end > len(allOrders) {
		end = len(allOrders)
	}
	result := allOrders[offset:end]

	for i := range result {
		s.fillOrderLabels(&result[i])
	}

	return result, total, nil
}

// fillOrderLabels 填充订单显示标签
func (s *OrderService) fillOrderLabels(order *models.Order) {
	currency := order.Currency
	if currency == "" {
		currency = "NT$"
	}
	amount := order.TotalAmount.StringFixed(2)

	// 根据订单状态设置不同的显示标签
	switch order.OrderStatus {
	case "COMPLETED":
		// 已结算订单
		order.AdjustmentLabel1 = fmt.Sprintf("已结算佣金：%s0.00", currency)
		order.AdjustmentLabel2 = fmt.Sprintf("订单金额：%s%s", currency, amount)
		order.AdjustmentLabel3 = fmt.Sprintf("虾皮订单金额：%s%s", currency, amount)
	case "CANCELLED", "IN_CANCEL":
		// 账款调整订单
		order.AdjustmentLabel1 = fmt.Sprintf("账款调整佣金：%s0.00", currency)
		order.AdjustmentLabel2 = fmt.Sprintf("订单账款调整：%s%s", currency, amount)
		order.AdjustmentLabel3 = fmt.Sprintf("虾皮订单账款调整：%s%s", currency, amount)
	default:
		// 未结算订单（待发货、已发货等）
		order.AdjustmentLabel1 = fmt.Sprintf("未结算佣金：%s0.00", currency)
		order.AdjustmentLabel2 = fmt.Sprintf("订单金额：%s%s", currency, amount)
		order.AdjustmentLabel3 = fmt.Sprintf("虾皮订单金额：%s%s", currency, amount)
	}
}

// GetOrder 获取订单详情 - 使用分表
func (s *OrderService) GetOrder(ctx context.Context, adminID int64, shopID int64, orderSN string) (*models.Order, error) {
	if _, err := s.shopService.GetShop(ctx, adminID, shopID); err != nil {
		return nil, err
	}
	orderTable := database.GetOrderTableName(uint64(shopID))
	var order models.Order
	if err := s.db.Table(orderTable).Where("shop_id = ? AND order_sn = ?", shopID, orderSN).First(&order).Error; err != nil {
		return nil, utils.ErrOrderNotFound
	}
	return &order, nil
}

// RefreshOrder 刷新订单 - 使用分表
func (s *OrderService) RefreshOrder(ctx context.Context, adminID int64, shopID int64, orderSN string) (*models.Order, error) {
	shop, err := s.shopService.GetShop(ctx, adminID, shopID)
	if err != nil {
		return nil, err
	}

	accessToken, err := s.getAccessToken(ctx, uint64(shopID))
	if err != nil {
		return nil, err
	}

	client := shopee.NewClient(shop.Region)
	orderDetailsResp, err := client.GetOrderDetail(accessToken, uint64(shopID), []string{orderSN})
	if err != nil {
		return nil, fmt.Errorf("获取订单详情失败: %w", err)
	}

	if len(orderDetailsResp.Response.OrderList) == 0 {
		return nil, utils.ErrOrderNotFound
	}

	detail := &orderDetailsResp.Response.OrderList[0]
	if err := s.saveOrder(ctx, uint64(shopID), detail); err != nil {
		return nil, err
	}

	orderTable := database.GetOrderTableName(uint64(shopID))
	var order models.Order
	s.db.Table(orderTable).Where("shop_id = ? AND order_sn = ?", shopID, orderSN).First(&order)
	return &order, nil
}

// ForceUpdateStatus 强制更新订单状态 - 使用分表
func (s *OrderService) ForceUpdateStatus(ctx context.Context, adminID int64, shopID int64, orderSN, newStatus string) error {
	if _, err := s.shopService.GetShop(ctx, adminID, shopID); err != nil {
		return err
	}

	orderTable := database.GetOrderTableName(uint64(shopID))
	var order models.Order
	if err := s.db.Table(orderTable).Where("shop_id = ? AND order_sn = ?", shopID, orderSN).First(&order).Error; err != nil {
		return utils.ErrOrderNotFound
	}

	updates := map[string]interface{}{
		"order_status":  newStatus,
		"status_locked": true,
		"status_remark": "店主手动更新",
	}
	return s.db.Table(orderTable).Where("id = ?", order.ID).Updates(updates).Error
}

// UnlockOrderStatus 解锁订单状态 - 使用分表
func (s *OrderService) UnlockOrderStatus(ctx context.Context, adminID int64, shopID int64, orderSN string) error {
	if _, err := s.shopService.GetShop(ctx, adminID, shopID); err != nil {
		return err
	}

	orderTable := database.GetOrderTableName(uint64(shopID))
	result := s.db.Table(orderTable).
		Where("shop_id = ? AND order_sn = ?", shopID, orderSN).
		Updates(map[string]interface{}{
			"status_locked": false,
			"status_remark": "",
		})

	if result.RowsAffected == 0 {
		return utils.ErrOrderNotFound
	}
	return result.Error
}

// GetReadyToShipOrders 获取待发货订单
func (s *OrderService) GetReadyToShipOrders(ctx context.Context, adminID int64, shopID int64, page, pageSize int) ([]models.Order, int64, error) {
	return s.ListOrders(ctx, adminID, shopID, consts.OrderStatusReadyToShip, "", "", page, pageSize)
}

// GetOrderStatus 从缓存或数据库获取订单状态 - 使用分表
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

	orderTable := database.GetOrderTableName(shopID)
	var order models.Order
	if err := s.db.Table(orderTable).Select("order_status").
		Where("shop_id = ? AND order_sn = ?", shopID, orderSN).
		First(&order).Error; err != nil {
		return "", err
	}

	rdb.Set(ctx, statusKey, order.OrderStatus, consts.OrderStatusExpire)
	return order.OrderStatus, nil
}

// IsStatusLocked 检查订单状态是否被锁定 - 使用分表
func (s *OrderService) IsStatusLocked(ctx context.Context, shopID uint64, orderSN string) (bool, error) {
	orderTable := database.GetOrderTableName(shopID)
	var order models.Order
	if err := s.db.Table(orderTable).Select("status_locked").Where("shop_id = ? AND order_sn = ?", shopID, orderSN).First(&order).Error; err != nil {
		return false, err
	}
	return order.StatusLocked, nil
}

// RefreshOrderFromAPI 从API刷新单个订单（无权限验证，供内部使用）
func (s *OrderService) RefreshOrderFromAPI(ctx context.Context, shopID uint64, orderSN string) error {
	var shop models.Shop
	if err := s.db.Where("shop_id = ?", shopID).First(&shop).Error; err != nil {
		return fmt.Errorf("店铺不存在: %w", err)
	}

	accessToken, err := s.getAccessToken(ctx, shopID)
	if err != nil {
		return err
	}

	client := shopee.NewClient(shop.Region)
	detailResp, err := client.GetOrderDetail(accessToken, shopID, []string{orderSN})
	if err != nil {
		return fmt.Errorf("获取订单详情失败: %w", err)
	}

	if len(detailResp.Response.OrderList) == 0 {
		return utils.ErrOrderNotFound
	}

	return s.saveOrder(ctx, shopID, &detailResp.Response.OrderList[0])
}

// checkAndSetUpdateTimeLua Lua脚本：原子性地检查并设置更新时间
const checkAndSetUpdateTimeLua = `
	local oldTime = redis.call('GET', KEYS[1])
	if oldTime then
		if tonumber(ARGV[1]) <= tonumber(oldTime) then
			return 0
		end
	end
	redis.call('SETEX', KEYS[1], ARGV[2], ARGV[1])
	return 1
`

// checkAndSetUpdateTime 原子性地检查并设置更新时间
func (s *OrderService) checkAndSetUpdateTime(ctx context.Context, rdb *redis.Client, key string, newTime int64) bool {
	script := redis.NewScript(checkAndSetUpdateTimeLua)
	result, err := script.Run(ctx, rdb, []string{key}, newTime, int(consts.OrderUpdateTimeTTL.Seconds())).Int()
	if err != nil {
		// 降级：使用非原子操作
		cachedTime, err := rdb.Get(ctx, key).Result()
		if err == nil {
			oldTime, _ := strconv.ParseInt(cachedTime, 10, 64)
			if newTime <= oldTime {
				return false
			}
		}
		rdb.Set(ctx, key, newTime, consts.OrderUpdateTimeTTL)
		return true
	}
	return result == 1
}
