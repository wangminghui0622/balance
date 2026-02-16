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
	idGenerator *utils.IDGenerator
}

// NewOrderService 创建订单服务
func NewOrderService() *OrderService {
	db := database.GetDB()
	return &OrderService{
		db:          db,
		shopService: NewShopService(),
		shardedDB:   database.NewShardedDB(db),
		rs:          database.GetRedsync(),
		idGenerator: utils.NewIDGenerator(database.GetRedis()),
	}
}

// SyncOrders 同步订单（支持分时间段、分页、限流、重试）
// 此方法会先查询店铺信息进行权限校验，适用于 Handler 调用
func (s *OrderService) SyncOrders(ctx context.Context, adminID int64, shopID int64, timeFrom, timeTo time.Time) (int, error) {
	shop, err := s.shopService.GetShop(ctx, adminID, shopID)
	if err != nil {
		return 0, err
	}
	return s.SyncOrdersWithShop(ctx, shop, timeFrom, timeTo)
}

// SyncOrdersWithShop 同步订单（传入已有的 shop 对象，带分布式锁）
// 此方法适用于 Handler 手动触发的场景
func (s *OrderService) SyncOrdersWithShop(ctx context.Context, shop *models.Shop, timeFrom, timeTo time.Time) (int, error) {
	shopID := shop.ShopID

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

	return s.syncOrdersInternal(ctx, shop, timeFrom, timeTo)
}

// SyncOrdersWithShopNoLock 同步订单（不加锁，调用方已持有锁）
// 此方法适用于调度器等已在外层加锁的场景
func (s *OrderService) SyncOrdersWithShopNoLock(ctx context.Context, shop *models.Shop, timeFrom, timeTo time.Time) (int, error) {
	return s.syncOrdersInternal(ctx, shop, timeFrom, timeTo)
}

// syncOrdersInternal 同步订单内部实现（不加锁）
//
// 同步流程分三层循环：
//  1. 外层循环（本函数）：将大时间范围拆分成多个 ≤15天 的小段（Shopee API 限制）
//  2. 中层循环（syncOrdersInRange）：分页获取该时间段内的所有订单号（每页最多100个）
//  3. 内层循环（syncOrdersInRange）：分批获取订单详情（每批最多50个）
func (s *OrderService) syncOrdersInternal(ctx context.Context, shop *models.Shop, timeFrom, timeTo time.Time) (int, error) {
	shopID := shop.ShopID

	// [调试日志] 打印同步参数
	fmt.Printf("[SyncDebug] 店铺=%d Region=%s timeFrom=%s timeTo=%s\n",
		shopID, shop.Region, timeFrom.Format("2006-01-02 15:04:05"), timeTo.Format("2006-01-02 15:04:05"))

	accessToken, err := s.getAccessToken(ctx, shopID)
	if err != nil {
		return 0, fmt.Errorf("获取访问令牌失败: %w", err)
	}
	fmt.Printf("[SyncDebug] 店铺=%d AccessToken=%s...(前20字符)\n", shopID, truncateStr(accessToken, 20))

	// === 外层循环：分时间段同步 ===
	// Shopee API 限制单次查询最多 15 天，所以需要将大时间范围拆分
	// 例如：同步 60 天数据 -> 拆分为 4 段（Day1-15, Day16-30, Day31-45, Day46-60）
	maxRange := int64(consts.ShopeeMaxTimeRange)
	fromTs := timeFrom.Unix()
	toTs := timeTo.Unix()
	totalCount := 0
	chunkIdx := 0

	for fromTs < toTs {
		endTs := fromTs + maxRange // 15天
		if endTs > toTs {
			endTs = toTs
		}
		chunkIdx++
		fmt.Printf("[SyncDebug] 店铺=%d 时间段#%d: %s ~ %s (unix: %d ~ %d)\n",
			shopID, chunkIdx,
			time.Unix(fromTs, 0).Format("2006-01-02 15:04:05"),
			time.Unix(endTs, 0).Format("2006-01-02 15:04:05"),
			fromTs, endTs)

		count, err := s.syncOrdersInRange(ctx, shopID, shop.Region, accessToken, fromTs, endTs, "")
		if err != nil {
			return totalCount, err
		}
		fmt.Printf("[SyncDebug] 店铺=%d 时间段#%d 获取到 %d 条订单\n", shopID, chunkIdx, count)
		totalCount += count
		fromTs = endTs
	}

	fmt.Printf("[SyncDebug] 店铺=%d 同步总计 %d 条订单, 共 %d 个时间段\n", shopID, totalCount, chunkIdx)
	s.shopService.UpdateLastSyncTime(shopID)
	return totalCount, nil
}

// syncOrdersInRange 在指定时间范围内同步订单（支持分页、限流、重试）
func (s *OrderService) syncOrdersInRange(ctx context.Context, shopID uint64, region, accessToken string, timeFrom, timeTo int64, orderStatus string) (int, error) {
	client := shopee.NewClient(region)

	cursor := ""
	pageSize := consts.ShopeeOrderListPageSize
	totalCount := 0
	pageIdx := 0

	fmt.Printf("[SyncDebug] syncOrdersInRange 店铺=%d region=%s host=%s timeFrom=%d timeTo=%d orderStatus=%q\n",
		shopID, region, client.GetHost(), timeFrom, timeTo, orderStatus)

	// === 中层循环：分页获取订单列表 ===
	// Shopee API 单次最多返回 100 个订单号，通过 cursor 分页获取所有订单
	// 例如：该时间段有 500 个订单 -> 分 5 页获取
	for {
		if err := shopee.WaitForRateLimit(ctx, shopID); err != nil {
			return totalCount, fmt.Errorf("限流等待被取消: %w", err)
		}

		pageIdx++
		var listResp *shopee.OrderListResponse
		err := shopee.RetryWithBackoff(ctx, consts.ShopeeAPIRetryTimes, func() error {
			var err error
			listResp, err = client.GetOrderList(accessToken, shopID, "create_time", timeFrom, timeTo, pageSize, cursor, orderStatus)
			return err
		})
		if err != nil {
			fmt.Printf("[SyncDebug] 店铺=%d 第%d页获取订单列表失败: %v\n", shopID, pageIdx, err)
			return totalCount, fmt.Errorf("获取订单列表失败: %w", err)
		}

		fmt.Printf("[SyncDebug] 店铺=%d 第%d页: 返回 %d 条订单, more=%v, nextCursor=%q\n",
			shopID, pageIdx, len(listResp.Response.OrderList), listResp.Response.More, listResp.Response.NextCursor)

		if len(listResp.Response.OrderList) == 0 {
			fmt.Printf("[SyncDebug] 店铺=%d 第%d页: 订单列表为空，结束分页\n", shopID, pageIdx)
			break
		}

		// 提取本页所有订单号
		orderSNs := make([]string, 0, len(listResp.Response.OrderList))
		for _, o := range listResp.Response.OrderList {
			orderSNs = append(orderSNs, o.OrderSN)
		}

		// === 内层循环：分批获取订单详情 ===
		// Shopee API 单次最多查询 50 个订单详情，所以需要分批
		// 例如：本页 100 个订单号 -> 分 2 批获取详情
		for i := 0; i < len(orderSNs); i += consts.ShopeeOrderDetailMaxSize { // 50个
			end := i + consts.ShopeeOrderDetailMaxSize
			if end > len(orderSNs) {
				end = len(orderSNs)
			}
			batch := orderSNs[i:end]

			if err := shopee.WaitForRateLimit(ctx, shopID); err != nil {
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

		// 使用分表
		orderTable := database.GetOrderTableName(shopID)
		orderItemTable := database.GetOrderItemTableName(shopID)
		orderAddressTable := database.GetOrderAddressTableName(shopID)

		// 查询是否已存在
		var existingOrder models.Order
		isNew := tx.Table(orderTable).Where("shop_id = ? AND order_sn = ?", shopID, detail.OrderSN).
			First(&existingOrder).Error != nil

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

		if isNew {
			// 新订单：生成ID后插入
			orderID, err := s.idGenerator.GenerateOrderID(ctx)
			if err != nil {
				return fmt.Errorf("生成订单ID失败: %w", err)
			}
			order.ID = uint64(orderID)
			if err := tx.Table(orderTable).Create(&order).Error; err != nil {
				return err
			}
		} else {
			// 已存在：更新
			order.ID = existingOrder.ID
			if err := tx.Table(orderTable).Where("id = ?", existingOrder.ID).Save(&order).Error; err != nil {
				return err
			}
		}

		// 更新订单商品（先删除旧数据，再插入新数据）
		if err := tx.Table(orderItemTable).Where("order_id = ?", order.ID).Delete(&models.OrderItem{}).Error; err != nil {
			return err
		}
		for _, item := range detail.ItemList {
			itemID, err := s.idGenerator.GenerateOrderItemID(ctx)
			if err != nil {
				return fmt.Errorf("生成订单商品ID失败: %w", err)
			}
			orderItem := models.OrderItem{
				ID:        uint64(itemID),
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

		// 更新订单地址（先删除旧数据，再插入新数据）
		if err := tx.Table(orderAddressTable).Where("order_id = ?", order.ID).Delete(&models.OrderAddress{}).Error; err != nil {
			return err
		}
		addrID, err := s.idGenerator.GenerateOrderAddressID(ctx)
		if err != nil {
			return fmt.Errorf("生成订单地址ID失败: %w", err)
		}
		addr := detail.RecipientAddress
		orderAddress := models.OrderAddress{
			ID:          uint64(addrID),
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
			itemID, err := s.idGenerator.GenerateOrderItemID(ctx)
			if err != nil {
				return fmt.Errorf("生成订单商品ID失败: %w", err)
			}
			orderItem := models.OrderItem{
				ID:        uint64(itemID),
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
		addrID, err := s.idGenerator.GenerateOrderAddressID(ctx)
		if err != nil {
			return fmt.Errorf("生成订单地址ID失败: %w", err)
		}
		addr := detail.RecipientAddress
		orderAddress := models.OrderAddress{
			ID:          uint64(addrID),
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

// truncateStr 截断字符串用于日志输出
func truncateStr(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen]
}
