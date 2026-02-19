package shopower

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"
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
	"gorm.io/gorm/clause"
)

// PrepaymentCheckFunc 预付款检查回调函数类型
// 参数：shopID, orderSN, prepaymentAmount(预付款扣除金额), orderTable
type PrepaymentCheckFunc func(ctx context.Context, shopID uint64, orderSN string, prepaymentAmount decimal.Decimal, orderTable string)

// EscrowFetchFunc 获取订单结算明细的回调函数类型（由上层注入，避免循环依赖）
// 返回值：escrow detail response, error
type EscrowFetchFunc func(ctx context.Context, shopID uint64, orderSN string) (*shopee.EscrowDetailResponse, error)

// OrderService 订单服务（店主专用）
type OrderService struct {
	db          *gorm.DB
	shopService *ShopService
	shardedDB   *database.ShardedDB
	rs          *redsync.Redsync
	idGenerator *utils.IDGenerator

	// 预付款检查回调（由上层注入，避免循环依赖）
	onPrepaymentCheck PrepaymentCheckFunc
	// 结算明细获取回调（READY_TO_SHIP 时调用，获取费用明细用于预付款计算）
	onEscrowFetch EscrowFetchFunc
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

// SetPrepaymentCheckFunc 设置预付款检查回调（由上层在初始化时注入）
func (s *OrderService) SetPrepaymentCheckFunc(fn PrepaymentCheckFunc) {
	s.onPrepaymentCheck = fn
}

// SetEscrowFetchFunc 设置结算明细获取回调（由上层在初始化时注入）
func (s *OrderService) SetEscrowFetchFunc(fn EscrowFetchFunc) {
	s.onEscrowFetch = fn
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

// syncOrdersInternal 同步订单内部实现（不加锁，供手动触发的全量同步使用）
//
// 同步流程分三层循环：
//  1. 外层循环（本函数）：将大时间范围拆分成多个 ≤15天 的小段（Shopee API 限制）
//  2. 中层循环（syncOrdersInRange）：分页获取该时间段内的所有订单号（每页最多100个）
//  3. 内层循环（syncOrdersInRange）：分批获取订单详情（每批最多50个）
func (s *OrderService) syncOrdersInternal(ctx context.Context, shop *models.Shop, timeFrom, timeTo time.Time) (int, error) {
	shopID := shop.ShopID

	accessToken, err := s.getAccessToken(ctx, shopID)
	if err != nil {
		return 0, fmt.Errorf("获取访问令牌失败: %w", err)
	}

	maxRange := int64(consts.ShopeeMaxTimeRange)
	fromTs := timeFrom.Unix()
	toTs := timeTo.Unix()
	totalCount := 0

	for fromTs < toTs {
		endTs := fromTs + maxRange
		if endTs > toTs {
			endTs = toTs
		}

		count, err := s.syncOrdersInRange(ctx, shopID, shop.Region, accessToken, fromTs, endTs, "")
		if err != nil {
			return totalCount, err
		}
		totalCount += count
		fromTs = endTs
	}

	s.shopService.UpdateLastSyncTime(shopID)
	return totalCount, nil
}

// syncOrdersInRange 在指定时间范围内同步订单（支持分页、限流、重试）
func (s *OrderService) syncOrdersInRange(ctx context.Context, shopID uint64, region, accessToken string, timeFrom, timeTo int64, orderStatus string) (int, error) {
	client := shopee.NewClient(region)

	cursor := ""
	pageSize := consts.ShopeeOrderListPageSize
	totalCount := 0

	for {
		if err := shopee.WaitForRateLimit(ctx, shopID); err != nil {
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
				if err := s.UpsertOrder(ctx, shopID, region, &detail); err != nil {
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




// ListOrders 获取订单列表 - 使用分表
// 语义：「全部订单」= 不传 shop_id、不传 status 时，返回 orders_x 下该 admin 绑定的所有 shop 的所有状态订单（可带 start_time/end_time 筛选）
func (s *OrderService) ListOrders(ctx context.Context, adminID int64, shopID int64, status, startTime, endTime string, page, pageSize int) ([]models.Order, int64, error) {
	var shopIDs []uint64
	s.db.Model(&models.Shop{}).Where("admin_id = ?", adminID).Pluck("shop_id", &shopIDs)
	if len(shopIDs) == 0 {
		return []models.Order{}, 0, nil
	}

	// 指定了 shop_id 时，只查该店铺对应分表
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

		s.batchFillOrderLabels(orders)
		s.batchFillOrderItems(orders)
		return orders, total, nil
	}

	// 未指定 shop_id：查该 admin 下所有绑定店铺的订单（orders_x 多表），status 为空表示所有状态
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

	// 内存排序和分页（按ID降序排序）
	sort.Slice(allOrders, func(i, j int) bool {
		return allOrders[i].ID > allOrders[j].ID
	})

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

	s.batchFillOrderLabels(result)
	s.batchFillOrderItems(result)

	return result, total, nil
}

// OrderStats 店主订单统计（全部/未结算/已结算/账款调整，不限制时间）
type OrderStats struct {
	AllOrdersCount    int64   `json:"all_orders_count"`
	AllOrdersAmount   float64 `json:"all_orders_amount"`
	UnsettledCount    int64   `json:"unsettled_count"`
	UnsettledAmount   float64 `json:"unsettled_amount"`
	SettledCount      int64   `json:"settled_count"`
	SettledAmount     float64 `json:"settled_amount"`
	AdjustmentCount   int64   `json:"adjustment_count"`
	AdjustmentAmount  float64 `json:"adjustment_amount"`
}

// ComputeOrderStats 计算该 admin 绑定所有店铺的订单统计（不限制时间，金额为 total_amount 之和）
func (s *OrderService) ComputeOrderStats(ctx context.Context, adminID int64) (*OrderStats, error) {
	var shopIDs []uint64
	if err := s.db.Model(&models.Shop{}).Where("admin_id = ?", adminID).Pluck("shop_id", &shopIDs).Error; err != nil {
		return nil, err
	}
	if len(shopIDs) == 0 {
		return &OrderStats{}, nil
	}

	shardShops := make(map[int][]uint64)
	for _, sid := range shopIDs {
		idx := database.GetShardIndex(sid)
		shardShops[idx] = append(shardShops[idx], sid)
	}

	stats := &OrderStats{}
	for idx, sids := range shardShops {
		orderTable := fmt.Sprintf("orders_%d", idx)
		escrowTable := fmt.Sprintf("order_escrows_%d", idx)

		// 全部：COUNT(*), SUM(total_amount)
		var allCount int64
		var allSum float64
		s.db.Table(orderTable).Where("shop_id IN ?", sids).Count(&allCount)
		s.db.Table(orderTable).Where("shop_id IN ?", sids).Select("COALESCE(SUM(total_amount),0)").Scan(&allSum)
		stats.AllOrdersCount += allCount
		stats.AllOrdersAmount += allSum

		// 未结算：READY_TO_SHIP, PROCESSED, SHIPPED
		var unsettledCount int64
		var unsettledSum float64
		s.db.Table(orderTable).Where("shop_id IN ? AND order_status IN ?", sids, []string{"READY_TO_SHIP", "PROCESSED", "SHIPPED"}).Count(&unsettledCount)
		s.db.Table(orderTable).Where("shop_id IN ? AND order_status IN ?", sids, []string{"READY_TO_SHIP", "PROCESSED", "SHIPPED"}).Select("COALESCE(SUM(total_amount),0)").Scan(&unsettledSum)
		stats.UnsettledCount += unsettledCount
		stats.UnsettledAmount += unsettledSum

		// 已结算：COMPLETED
		var settledCount int64
		var settledSum float64
		s.db.Table(orderTable).Where("shop_id IN ? AND order_status = ?", sids, "COMPLETED").Count(&settledCount)
		s.db.Table(orderTable).Where("shop_id IN ? AND order_status = ?", sids, "COMPLETED").Select("COALESCE(SUM(total_amount),0)").Scan(&settledSum)
		stats.SettledCount += settledCount
		stats.SettledAmount += settledSum

		// 账款调整：有 order_escrow 且 drc_adjustable_refund / seller_return_refund / reverse_shipping_fee 任一非 0
		var adjCount int64
		var adjSum float64
		s.db.Table(orderTable+" AS o").
			Joins("INNER JOIN "+escrowTable+" e ON o.shop_id = e.shop_id AND o.order_sn = e.order_sn").
			Where("o.shop_id IN ? AND (e.drc_adjustable_refund != 0 OR e.seller_return_refund != 0 OR e.reverse_shipping_fee != 0)", sids).
			Count(&adjCount)
		s.db.Table(orderTable+" AS o").
			Select("COALESCE(SUM(o.total_amount),0)").
			Joins("INNER JOIN "+escrowTable+" e ON o.shop_id = e.shop_id AND o.order_sn = e.order_sn").
			Where("o.shop_id IN ? AND (e.drc_adjustable_refund != 0 OR e.seller_return_refund != 0 OR e.reverse_shipping_fee != 0)", sids).
			Scan(&adjSum)
		stats.AdjustmentCount += adjCount
		stats.AdjustmentAmount += adjSum
	}

	return stats, nil
}

// GetOrderStatsCached 从缓存读取订单统计；miss 时计算并写入缓存后返回
func (s *OrderService) GetOrderStatsCached(ctx context.Context, adminID int64) (*OrderStats, error) {
	rdb := database.GetRedis()
	key := fmt.Sprintf(consts.KeyShopowerOrderStats, adminID)
	data, err := rdb.Get(ctx, key).Bytes()
	if err == nil {
		var stats OrderStats
		if err := json.Unmarshal(data, &stats); err != nil {
			return nil, err
		}
		return &stats, nil
	}
	if err != redis.Nil {
		return nil, err
	}
	stats, err := s.ComputeOrderStats(ctx, adminID)
	if err != nil {
		return nil, err
	}
	if err := s.SetOrderStatsCache(ctx, adminID, stats); err != nil {
		// 写缓存失败不影响返回
	}
	return stats, nil
}

// SetOrderStatsCache 将订单统计写入 Redis 缓存（供定时器使用）
func (s *OrderService) SetOrderStatsCache(ctx context.Context, adminID int64, stats *OrderStats) error {
	data, err := json.Marshal(stats)
	if err != nil {
		return err
	}
	key := fmt.Sprintf(consts.KeyShopowerOrderStats, adminID)
	return database.GetRedis().Set(ctx, key, data, consts.ShopowerOrderStatsExpire).Err()
}

// batchFillOrderLabels 批量填充订单显示标签（解决 N+1 查询问题）
// 将同一分表的订单分组后用 IN 查询一次拿到所有 escrow 数据，避免每个订单查一次
func (s *OrderService) batchFillOrderLabels(orders []models.Order) {
	if len(orders) == 0 {
		return
	}

	// 按 escrow 分表索引分组
	shardOrders := make(map[int][]string) // shardIdx → []orderSN
	shardShopID := make(map[int]uint64)   // shardIdx → 任一 shopID（用于表名）
	for _, o := range orders {
		idx := database.GetShardIndex(o.ShopID)
		shardOrders[idx] = append(shardOrders[idx], o.OrderSN)
		shardShopID[idx] = o.ShopID
	}

	// 批量查询每个分表的 escrow 数据
	escrowMap := make(map[string]models.OrderEscrow) // key = "shopID:orderSN"
	for idx, orderSNs := range shardOrders {
		escrowTable := database.GetOrderEscrowTableName(shardShopID[idx])
		var escrows []models.OrderEscrow
		s.db.Table(escrowTable).Where("order_sn IN ?", orderSNs).Find(&escrows)
		for _, e := range escrows {
			key := fmt.Sprintf("%d:%s", e.ShopID, e.OrderSN)
			escrowMap[key] = e
		}
	}

	// 填充每个订单的标签
	for i := range orders {
		key := fmt.Sprintf("%d:%s", orders[i].ShopID, orders[i].OrderSN)
		if escrow, ok := escrowMap[key]; ok {
			s.fillOrderLabelsWithEscrow(&orders[i], &escrow)
		} else {
			s.fillOrderLabelsWithEscrow(&orders[i], nil)
		}
	}
}

// batchFillOrderItems 批量填充订单商品明细（从 order_items_x 按 order_sn 关联）
func (s *OrderService) batchFillOrderItems(orders []models.Order) {
	if len(orders) == 0 {
		return
	}

	// 按分表分组：(shardIdx -> []orderSN)
	shardOrderSNs := make(map[int][]string)
	shardShopID := make(map[int]uint64)
	for _, o := range orders {
		idx := database.GetShardIndex(o.ShopID)
		shardOrderSNs[idx] = append(shardOrderSNs[idx], o.OrderSN)
		shardShopID[idx] = o.ShopID
	}

	// 批量查询 order_items：order_sn 在 Shopee 中全局唯一，按 order_sn 关联即可
	// key = order_sn -> []OrderItem
	itemsMap := make(map[string][]models.OrderItem)
	for idx, orderSNs := range shardOrderSNs {
		if len(orderSNs) == 0 {
			continue
		}
		itemTable := database.GetOrderItemTableName(shardShopID[idx])
		var items []models.OrderItem
		s.db.Table(itemTable).Where("order_sn IN ?", orderSNs).Order("id ASC").Find(&items)
		for _, it := range items {
			itemsMap[it.OrderSN] = append(itemsMap[it.OrderSN], it)
		}
	}

	// 填充每个订单的 Items
	for i := range orders {
		orders[i].Items = itemsMap[orders[i].OrderSN]
		if orders[i].Items == nil {
			orders[i].Items = []models.OrderItem{}
		}
	}
}

// fillOrderLabels 填充订单显示标签（单条查询版本，供 GetOrder 等单条场景使用）
func (s *OrderService) fillOrderLabels(order *models.Order) {
	escrowTable := database.GetOrderEscrowTableName(order.ShopID)
	var escrow models.OrderEscrow
	if err := s.db.Table(escrowTable).
		Where("shop_id = ? AND order_sn = ?", order.ShopID, order.OrderSN).
		First(&escrow).Error; err == nil {
		s.fillOrderLabelsWithEscrow(order, &escrow)
	} else {
		s.fillOrderLabelsWithEscrow(order, nil)
	}
}

// fillOrderLabelsWithEscrow 填充订单显示标签（核心逻辑，escrow 可为 nil）
//
// Label 含义：
//   - AdjustmentLabel1: 佣金信息（头部右上角）
//   - AdjustmentLabel2: 订单金额/账款调整（头部右上角）
//   - AdjustmentLabel3: 虾皮订单金额/账款调整（商品列表下方）
func (s *OrderService) fillOrderLabelsWithEscrow(order *models.Order, escrow *models.OrderEscrow) {
	currency := order.Currency
	if currency == "" {
		currency = "NT$"
	}
	amount := order.TotalAmount.StringFixed(2)
	hasEscrow := escrow != nil

	switch order.OrderStatus {
	case consts.OrderStatusCompleted:
		if hasEscrow {
			commission := escrow.CommissionFee.Add(escrow.ServiceFee).Abs()
			escrowAmt := escrow.EscrowAmount
			order.AdjustmentLabel1 = fmt.Sprintf("已结算佣金：%s%s", currency, commission.StringFixed(2))
			order.AdjustmentLabel2 = fmt.Sprintf("订单金额：%s%s", currency, amount)
			order.AdjustmentLabel3 = fmt.Sprintf("虾皮订单结算：%s%s", currency, escrowAmt.StringFixed(2))
		} else {
			order.AdjustmentLabel1 = fmt.Sprintf("已结算佣金：%s--", currency)
			order.AdjustmentLabel2 = fmt.Sprintf("订单金额：%s%s", currency, amount)
			order.AdjustmentLabel3 = fmt.Sprintf("虾皮订单金额：%s%s", currency, amount)
		}

	case consts.OrderStatusCancelled, consts.OrderStatusCancelledBeforeShip, consts.OrderStatusInCancel:
		if hasEscrow {
			adjustTotal := escrow.SellerReturnRefund.Add(escrow.DrcAdjustableRefund).Add(escrow.ReverseShippingFee)
			commission := escrow.CommissionFee.Add(escrow.ServiceFee).Abs()
			escrowAmt := escrow.EscrowAmount
			order.AdjustmentLabel1 = fmt.Sprintf("账款调整佣金：%s%s", currency, commission.StringFixed(2))
			if adjustTotal.IsZero() {
				order.AdjustmentLabel2 = fmt.Sprintf("订单账款调整：%s%s", currency, amount)
			} else {
				order.AdjustmentLabel2 = fmt.Sprintf("订单账款调整：%s%s", currency, adjustTotal.Abs().StringFixed(2))
			}
			order.AdjustmentLabel3 = fmt.Sprintf("虾皮订单账款调整：%s%s", currency, escrowAmt.StringFixed(2))
		} else {
			order.AdjustmentLabel1 = fmt.Sprintf("账款调整佣金：%s0.00", currency)
			order.AdjustmentLabel2 = fmt.Sprintf("订单账款调整：%s%s", currency, amount)
			order.AdjustmentLabel3 = fmt.Sprintf("虾皮订单账款调整：%s%s", currency, amount)
		}

	default:
		order.AdjustmentLabel1 = fmt.Sprintf("未结算佣金：%s--", currency)
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
	if err := s.UpsertOrder(ctx, uint64(shopID), shop.Region, detail); err != nil {
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

	return s.UpsertOrder(ctx, shopID, shop.Region, &detailResp.Response.OrderList[0])
}


// ==================== 统一订单写入入口 ====================

// UpsertOrder 统一的订单写入入口（Webhook 和 Patrol 巡检共用）
//
// 并发安全：
//   - 事务内使用 SELECT ... FOR UPDATE 对已有订单行加行锁，串行化并发写入
//   - 新订单插入依赖 (shop_id, order_sn) 唯一索引兜底，重复插入会被数据库拒绝
//   - 状态回退由 OrderStatusPriority 拦截，StatusLocked 由店主手动锁定
//   - READY_TO_SHIP 预付款检查在事务提交后异步触发，CheckAndDeductForOrder 自带行锁幂等
//
// 返回 nil 表示写入成功（含 "无需更新" 的情况）
func (s *OrderService) UpsertOrder(ctx context.Context, shopID uint64, region string, detail *shopee.OrderDetail) error {
	orderTable := database.GetOrderTableName(shopID)
	orderItemTable := database.GetOrderItemTableName(shopID)
	orderAddressTable := database.GetOrderAddressTableName(shopID)
	totalAmount := decimal.NewFromFloat(detail.TotalAmount)

	// 解析 Shopee 时间戳为 Go time 指针
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

	// 仅更新非状态字段的通用 map（StatusLocked 或状态优先级低时使用）
	partialUpdates := map[string]interface{}{
		"shipping_carrier": detail.ShippingCarrier,
		"tracking_number":  detail.TrackingNo,
		"ship_by_date":     shipByDate,
		"update_time":      updateTime,
	}

	err := s.db.Transaction(func(tx *gorm.DB) error {
		// 尝试用 FOR UPDATE 锁住已有的订单行
		// 如果订单不存在，First 返回 ErrRecordNotFound，不会加锁
		// 如果订单存在，行锁保证并发的 Webhook 和 Patrol 串行执行后续逻辑
		var existing models.Order
		found := tx.Table(orderTable).
			Clauses(clause.Locking{Strength: "UPDATE"}).
			Where("shop_id = ? AND order_sn = ?", shopID, detail.OrderSN).
			First(&existing).Error == nil

		// 构造完整的订单对象
		order := models.Order{
			ShopID:          shopID,
			OrderSN:         detail.OrderSN,
			Region:          region,
			OrderStatus:     detail.OrderStatus,
			BuyerUserID:     uint64(detail.BuyerUserID),
			BuyerUsername:   detail.BuyerUsername,
			TotalAmount:     totalAmount,
			Currency:        detail.Currency,
			ShippingCarrier: detail.ShippingCarrier,
			TrackingNumber:  detail.TrackingNo,
			ShipByDate:      shipByDate,
			PayTime:         payTime,
			CreateTime:      createTime,
			UpdateTime:      updateTime,
		}

		if !found {
			// 新订单：生成分布式 ID 后插入
			// 并发场景下若两个协程同时走到这里，唯一索引 (shop_id, order_sn) 会让第二个 Create 失败
			orderID, err := s.idGenerator.GenerateOrderID(ctx)
			if err != nil {
				return fmt.Errorf("生成订单ID失败: %w", err)
			}
			order.ID = uint64(orderID)
			if err := tx.Table(orderTable).Create(&order).Error; err != nil {
				return err
			}
		} else {
			// 已存在：状态锁定 → 仅更新物流等非状态字段
			if existing.StatusLocked {
				return tx.Table(orderTable).Where("id = ?", existing.ID).Updates(partialUpdates).Error
			}
			// 状态优先级检查：不允许回退（如 COMPLETED → SHIPPED 被拦截）
			currentP, cOK := consts.OrderStatusPriority[existing.OrderStatus]
			newP, nOK := consts.OrderStatusPriority[detail.OrderStatus]
			if cOK && nOK && newP < currentP {
				return tx.Table(orderTable).Where("id = ?", existing.ID).Updates(partialUpdates).Error
			}
			// 状态允许推进 → 全量更新
			order.ID = existing.ID
			if err := tx.Table(orderTable).Where("id = ?", existing.ID).Save(&order).Error; err != nil {
				return err
			}
		}

		// 写入/更新商品（先删后插，保证与 Shopee 一致）
		tx.Table(orderItemTable).Where("order_id = ?", order.ID).Delete(&models.OrderItem{})
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

		// 写入/更新地址（先删后插）
		tx.Table(orderAddressTable).Where("order_id = ?", order.ID).Delete(&models.OrderAddress{})
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
		return tx.Table(orderAddressTable).Create(&orderAddress).Error
	})

	if err != nil {
		return err
	}

	// 事务提交成功后，READY_TO_SHIP 触发：
	//   1. 调 get_escrow_detail 获取费用明细并写入订单
	//   2. 用 escrow_amount 作为预付款扣除金额（而非 total_amount）
	if detail.OrderStatus == consts.OrderStatusReadyToShip {
		prepaymentAmount := totalAmount // 兜底：如果获取不到 escrow_detail，使用 total_amount

		// 尝试获取 Shopee 结算明细（READY_TO_SHIP 时已可获取预估值）
		if s.onEscrowFetch != nil {
			escrowResp, err := s.onEscrowFetch(ctx, shopID, detail.OrderSN)
			if err == nil && escrowResp != nil {
				income := escrowResp.Response.OrderIncome
				escrowAmount := decimal.NewFromFloat(income.EscrowAmount)

				// 将费用明细写入订单表
				feeUpdates := map[string]interface{}{
					"escrow_amount_snapshot":       escrowAmount,
					"buyer_paid_shipping_fee":      decimal.NewFromFloat(income.BuyerPaidShippingFee),
					"original_cost_of_goods_sold":  decimal.NewFromFloat(income.OriginalCostOfGoodsSold),
					"commission_fee":               decimal.NewFromFloat(income.CommissionFee),
					"seller_transaction_fee":       decimal.NewFromFloat(income.SellerTransactionFee),
					"credit_card_transaction_fee":  decimal.NewFromFloat(income.CreditCardTransactionFee),
					"service_fee":                  decimal.NewFromFloat(income.ServiceFee),
					"prepayment_amount":            escrowAmount, // 预付款扣除金额 = escrow_amount
				}
				if err := s.db.Table(orderTable).
					Where("shop_id = ? AND order_sn = ?", shopID, detail.OrderSN).
					Updates(feeUpdates).Error; err != nil {
					fmt.Printf("[UpsertOrder] 店铺=%d 订单=%s 写入escrow费用明细失败: %v\n", shopID, detail.OrderSN, err)
				} else {
					prepaymentAmount = escrowAmount // 使用 escrow_amount 作为预付款金额
				}
			} else if err != nil {
				fmt.Printf("[UpsertOrder] 店铺=%d 订单=%s 获取escrow明细失败(降级用total_amount): %v\n", shopID, detail.OrderSN, err)
				// 降级：将 total_amount 写入 prepayment_amount
				s.db.Table(orderTable).
					Where("shop_id = ? AND order_sn = ?", shopID, detail.OrderSN).
					Update("prepayment_amount", totalAmount)
			}
		}

		// 触发预付款检查
		if s.onPrepaymentCheck != nil {
			s.onPrepaymentCheck(ctx, shopID, detail.OrderSN, prepaymentAmount, orderTable)
		}
	}

	return nil
}

// ==================== Sync 巡检模式 ====================

// PatrolOrders 巡检订单（轻量比对模式，只对遗漏/不一致的订单拉详情写入）
// 返回值：found = Shopee 侧订单总数, patched = 实际补录写入的订单数
func (s *OrderService) PatrolOrders(ctx context.Context, shop *models.Shop, timeFrom, timeTo time.Time) (found int, patched int, err error) {
	shopID := shop.ShopID
	region := shop.Region

	// 获取当前有效的 access_token（过期则自动刷新）
	accessToken, err := s.getAccessToken(ctx, shopID)
	if err != nil {
		return 0, 0, fmt.Errorf("获取访问令牌失败: %w", err)
	}

	// Shopee API 单次查询最多覆盖 15 天，超过则拆分为多个时间段依次巡检
	maxRange := int64(consts.ShopeeMaxTimeRange)
	fromTs := timeFrom.Unix()
	toTs := timeTo.Unix()

	// 按 15 天为一段，循环处理每段时间范围
	for fromTs < toTs {
		// 计算本段结束时间，不超过总截止时间
		endTs := fromTs + maxRange
		if endTs > toTs {
			endTs = toTs
		}

		// 执行单段巡检，累加 found 和 patched 计数
		f, p, segErr := s.patrolOrdersInRange(ctx, shopID, region, accessToken, fromTs, endTs)
		found += f
		patched += p
		// 任何一段出错则提前返回，已累加的计数仍然返回
		if segErr != nil {
			return found, patched, segErr
		}
		// 推进到下一段
		fromTs = endTs
	}

	// 全部段完成后，更新 shops 表的 last_sync_at，下次巡检从此时间开始
	s.shopService.UpdateLastSyncTime(shopID)
	return found, patched, nil
}

// patrolOrdersInRange 在单个时间段内巡检订单（分页拉取 → 比对 → 补录）
func (s *OrderService) patrolOrdersInRange(ctx context.Context, shopID uint64, region, accessToken string, timeFrom, timeTo int64) (found int, patched int, err error) {
	// 创建 Shopee API 客户端（根据 region 选择对应的 API Host）
	client := shopee.NewClient(region)
	// 游标分页：初始为空字符串，后续使用 Shopee 返回的 next_cursor
	cursor := ""
	// 每页最多拉取 100 条订单号（Shopee API 上限）
	pageSize := consts.ShopeeOrderListPageSize

	// 分页循环：持续拉取直到 Shopee 返回 more=false 或列表为空
	for {
		// 调用前等待限流令牌，防止触发 Shopee API 频率限制
		if err := shopee.WaitForRateLimit(ctx, shopID); err != nil {
			return found, patched, fmt.Errorf("限流等待被取消: %w", err)
		}

		// 调用 Shopee GetOrderList，获取本页的订单号 + 订单状态（轻量接口，不含详情）
		var listResp *shopee.OrderListResponse
		apiErr := shopee.RetryWithBackoff(ctx, consts.ShopeeAPIRetryTimes, func() error {
			var err error
			listResp, err = client.GetOrderList(accessToken, shopID, "create_time", timeFrom, timeTo, pageSize, cursor, "")
			return err
		})
		if apiErr != nil {
			return found, patched, fmt.Errorf("获取订单列表失败: %w", apiErr)
		}

		// 本页无数据，结束分页
		if len(listResp.Response.OrderList) == 0 {
			break
		}

		// 将 Shopee 返回的订单号和状态收集到结构体切片中
		type shopeeOrder struct {
			OrderSN string
			Status  string
		}
		var shopeeOrders []shopeeOrder
		for _, o := range listResp.Response.OrderList {
			shopeeOrders = append(shopeeOrders, shopeeOrder{OrderSN: o.OrderSN, Status: o.OrderStatus})
		}
		// 累加 Shopee 侧订单总数
		found += len(shopeeOrders)

		// 提取本页所有订单号，用于批量查询 DB
		orderSNs := make([]string, len(shopeeOrders))
		for i, o := range shopeeOrders {
			orderSNs[i] = o.OrderSN
		}

		// 根据 shopID 定位分表，批量查出 DB 中已有的订单号和状态
		orderTable := database.GetOrderTableName(shopID)
		var dbOrders []models.Order
		s.db.Table(orderTable).
			Select("order_sn", "order_status").
			Where("shop_id = ? AND order_sn IN ?", shopID, orderSNs).
			Find(&dbOrders)

		// 构建 DB 订单状态 map：order_sn → order_status，用于快速比对
		dbMap := make(map[string]string, len(dbOrders))
		for _, o := range dbOrders {
			dbMap[o.OrderSN] = o.OrderStatus
		}

		// 逐条比对 Shopee 和 DB，筛出需要补录的订单号
		var needPatch []string
		for _, so := range shopeeOrders {
			dbStatus, exists := dbMap[so.OrderSN]
			if !exists {
				// 情况 1：DB 中不存在该订单（Webhook 遗漏），需要新建
				needPatch = append(needPatch, so.OrderSN)
			} else if dbStatus != so.Status {
				// 情况 2：DB 中的状态落后于 Shopee（如 Webhook 丢失状态更新），需要更新
				// UpsertOrder 内部有状态优先级判断，不会回退到更低状态
				needPatch = append(needPatch, so.OrderSN)
			}
		}

		// 如果有需要补录的订单，分批拉取详情并写入 DB
		if len(needPatch) > 0 {
			fmt.Printf("[Patrol] 店铺=%d 发现 %d 条遗漏/不一致订单，补录中...\n", shopID, len(needPatch))

			// 每批最多 50 条（Shopee GetOrderDetail API 上限）
			for i := 0; i < len(needPatch); i += consts.ShopeeOrderDetailMaxSize {
				// 计算本批的结束下标
				end := i + consts.ShopeeOrderDetailMaxSize
				if end > len(needPatch) {
					end = len(needPatch)
				}
				// 截取本批订单号
				batch := needPatch[i:end]

				// 每批请求前等待限流令牌
				if err := shopee.WaitForRateLimit(ctx, shopID); err != nil {
					return found, patched, fmt.Errorf("限流等待被取消: %w", err)
				}

				// 调用 Shopee GetOrderDetail 获取完整订单信息（含商品、地址、金额等）
				var detailResp *shopee.OrderDetailResponse
				apiErr := shopee.RetryWithBackoff(ctx, consts.ShopeeAPIRetryTimes, func() error {
					var err error
					detailResp, err = client.GetOrderDetail(accessToken, shopID, batch)
					return err
				})
				if apiErr != nil {
					return found, patched, fmt.Errorf("获取订单详情失败: %w", apiErr)
				}

				// 逐条调用 UpsertOrder 写入 DB（内含事务、状态优先级控制、预付款检查）
				for _, detail := range detailResp.Response.OrderList {
					if err := s.UpsertOrder(ctx, shopID, region, &detail); err != nil {
						fmt.Printf("[Patrol] 店铺=%d 订单=%s 补录失败: %v\n", shopID, detail.OrderSN, err)
						continue // 单条失败不影响其他订单
					}
					// 补录成功，计数 +1
					patched++
				}
			}
		}

		// 检查是否还有下一页
		if !listResp.Response.More {
			break // Shopee 返回 more=false，本段巡检结束
		}
		// 用 Shopee 返回的游标请求下一页
		cursor = listResp.Response.NextCursor
	}

	return found, patched, nil
}

