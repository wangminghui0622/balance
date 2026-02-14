package services

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"balance/backend/internal/consts"
	"balance/backend/internal/database"
	"balance/backend/internal/models"
	"balance/backend/internal/services/shopower"
	"balance/backend/internal/shopee"
	"balance/backend/internal/utils"

	"github.com/go-redsync/redsync/v4"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

// WebhookService Webhook服务
type WebhookService struct {
	db           *gorm.DB
	shardedDB    *database.ShardedDB
	orderService *shopower.OrderService
	shopService  *shopower.ShopService
	rs           *redsync.Redsync
}

// NewWebhookService 创建Webhook服务
func NewWebhookService() *WebhookService {
	db := database.GetDB()
	return &WebhookService{
		db:           db,
		shardedDB:    database.NewShardedDB(db),
		orderService: shopower.NewOrderService(),
		shopService:  shopower.NewShopService(),
		rs:           database.GetRedsync(),
	}
}

// NewBackgroundContext 创建后台上下文
// 注意：调用者应该在使用完毕后调用 cancel()
func NewBackgroundContext() (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.Background(), 30*time.Second)
}

// CheckWebhookDuplicate 检查Webhook是否重复
func (s *WebhookService) CheckWebhookDuplicate(ctx context.Context, shopID uint64, orderSN string, code int, timestamp int64) bool {
	rdb := database.GetRedis()
	dedupKey := fmt.Sprintf(consts.KeyWebhookDedup, shopID, orderSN, code, timestamp)

	ok, err := rdb.SetNX(ctx, dedupKey, 1, consts.WebhookDedupExpire).Result()
	if err != nil {
		return false
	}
	return !ok
}

// AcquireOrderLock 获取订单更新锁，返回释放函数（nil表示失败）
func (s *WebhookService) AcquireOrderLock(ctx context.Context, shopID uint64, orderSN string) func() {
	lockKey := fmt.Sprintf(consts.KeyOrderLock, shopID, orderSN)
	mutex := s.rs.NewMutex(lockKey,
		redsync.WithExpiry(consts.OrderLockExpire),
		redsync.WithTries(1),
	)

	// 尝试获取锁并自动续期
	unlockFunc, acquired := utils.TryLockWithAutoExtend(ctx, mutex, consts.OrderLockExpire/3)
	if !acquired {
		return nil
	}
	return unlockFunc
}

// checkAndSetUpdateTimeLua Lua脚本：原子性地检查并设置更新时间
const checkAndSetUpdateTimeLua = `
	-- KEYS[1]: updateTimeKey
	-- ARGV[1]: newUpdateTime
	-- ARGV[2]: ttl (秒)
	
	local oldTime = redis.call('GET', KEYS[1])
	if oldTime then
		if tonumber(ARGV[1]) <= tonumber(oldTime) then
			return 0
		end
	end
	
	redis.call('SETEX', KEYS[1], ARGV[2], ARGV[1])
	return 1
`

// CheckAndSetUpdateTime 检查并设置更新时间（使用Lua脚本保证原子性）
func (s *WebhookService) CheckAndSetUpdateTime(ctx context.Context, shopID uint64, orderSN string, newUpdateTime int64) bool {
	rdb := database.GetRedis()
	updateTimeKey := fmt.Sprintf(consts.KeyOrderUpdateTime, shopID, orderSN)

	script := redis.NewScript(checkAndSetUpdateTimeLua)
	result, err := script.Run(ctx, rdb,
		[]string{updateTimeKey},
		newUpdateTime, int(consts.OrderUpdateTimeTTL.Seconds()),
	).Int()

	if err != nil {
		// 降级：使用非原子操作
		cachedTime, err := rdb.Get(ctx, updateTimeKey).Result()
		if err == nil {
			oldTime, _ := strconv.ParseInt(cachedTime, 10, 64)
			if newUpdateTime <= oldTime {
				return false
			}
		}
		rdb.Set(ctx, updateTimeKey, newUpdateTime, consts.OrderUpdateTimeTTL)
		return true
	}

	return result == 1
}

// CanStatusTransition 检查状态转换是否合法 - 使用分表
func (s *WebhookService) CanStatusTransition(ctx context.Context, shopID uint64, orderSN string, newStatus string) bool {
	orderTable := database.GetOrderTableName(shopID)
	var order models.Order
	if err := s.db.Table(orderTable).Where("shop_id = ? AND order_sn = ?", shopID, orderSN).First(&order).Error; err != nil {
		return true
	}

	if order.StatusLocked {
		return false
	}

	currentPriority, currentExists := consts.OrderStatusPriority[order.OrderStatus]
	newPriority, newExists := consts.OrderStatusPriority[newStatus]

	if !currentExists || !newExists {
		return true
	}

	return newPriority >= currentPriority
}

// HandleOrderStatusUpdate 处理订单状态更新
func (s *WebhookService) HandleOrderStatusUpdate(ctx context.Context, shopID uint64, data any, timestamp int64) {
	dataBytes, err := json.Marshal(data)
	if err != nil {
		s.logError(ctx, shopID, consts.WebhookOrderStatus, "marshal_error", err)
		return
	}

	var orderData struct {
		OrderSN    string `json:"ordersn"`
		Status     string `json:"status"`
		UpdateTime int64  `json:"update_time"`
	}
	if err := json.Unmarshal(dataBytes, &orderData); err != nil {
		s.logError(ctx, shopID, consts.WebhookOrderStatus, "unmarshal_error", err)
		return
	}

	if s.CheckWebhookDuplicate(ctx, shopID, orderData.OrderSN, consts.WebhookOrderStatus, timestamp) {
		return
	}

	unlockFunc := s.AcquireOrderLock(ctx, shopID, orderData.OrderSN)
	if unlockFunc == nil {
		return
	}
	defer unlockFunc()

	if orderData.UpdateTime > 0 && !s.CheckAndSetUpdateTime(ctx, shopID, orderData.OrderSN, orderData.UpdateTime) {
		return
	}

	if !s.CanStatusTransition(ctx, shopID, orderData.OrderSN, orderData.Status) {
		s.LogWebhook(ctx, shopID, consts.WebhookOrderStatus, "status_rollback_blocked", data)
		return
	}

	s.LogWebhook(ctx, shopID, consts.WebhookOrderStatus, "order_status", data)

	orderTable := database.GetOrderTableName(shopID)
	if err := s.db.Table(orderTable).
		Where("shop_id = ? AND order_sn = ?", shopID, orderData.OrderSN).
		Update("order_status", orderData.Status).Error; err != nil {
		s.logError(ctx, shopID, consts.WebhookOrderStatus, "update_error", err)
		return
	}

	rdb := database.GetRedis()
	cacheKey := fmt.Sprintf(consts.KeyOrderStatus, shopID, orderData.OrderSN)
	rdb.Del(ctx, cacheKey)

	if orderData.Status == consts.OrderStatusReadyToShip {
		go s.refreshOrderDetail(shopID, orderData.OrderSN)
	}
}

// HandleTrackingUpdate 处理物流追踪更新
func (s *WebhookService) HandleTrackingUpdate(ctx context.Context, shopID uint64, data any, timestamp int64) {
	dataBytes, err := json.Marshal(data)
	if err != nil {
		s.logError(ctx, shopID, consts.WebhookTrackingUpdate, "marshal_error", err)
		return
	}

	var trackingData struct {
		OrderSN         string `json:"ordersn"`
		TrackingNumber  string `json:"tracking_no"`
		LogisticsStatus string `json:"logistics_status"`
	}
	if err := json.Unmarshal(dataBytes, &trackingData); err != nil {
		s.logError(ctx, shopID, consts.WebhookTrackingUpdate, "unmarshal_error", err)
		return
	}

	if s.CheckWebhookDuplicate(ctx, shopID, trackingData.OrderSN, consts.WebhookTrackingUpdate, timestamp) {
		return
	}

	s.LogWebhook(ctx, shopID, consts.WebhookTrackingUpdate, "tracking_update", data)

	if trackingData.TrackingNumber != "" {
		shipmentTable := database.GetShipmentTableName(shopID)
		s.db.Table(shipmentTable).
			Where("shop_id = ? AND order_sn = ?", shopID, trackingData.OrderSN).
			Update("tracking_number", trackingData.TrackingNumber)
	}
}

// HandleOrderCancel 处理订单取消
func (s *WebhookService) HandleOrderCancel(ctx context.Context, shopID uint64, data any, timestamp int64) {
	dataBytes, err := json.Marshal(data)
	if err != nil {
		s.logError(ctx, shopID, consts.WebhookBuyerCancelOrder, "marshal_error", err)
		return
	}

	var cancelData struct {
		OrderSN      string `json:"ordersn"`
		CancelBy     string `json:"cancel_by"`
		CancelReason string `json:"cancel_reason"`
	}
	if err := json.Unmarshal(dataBytes, &cancelData); err != nil {
		s.logError(ctx, shopID, consts.WebhookBuyerCancelOrder, "unmarshal_error", err)
		return
	}

	if s.CheckWebhookDuplicate(ctx, shopID, cancelData.OrderSN, consts.WebhookBuyerCancelOrder, timestamp) {
		return
	}

	unlockFunc := s.AcquireOrderLock(ctx, shopID, cancelData.OrderSN)
	if unlockFunc == nil {
		return
	}
	defer unlockFunc()

	s.LogWebhook(ctx, shopID, consts.WebhookBuyerCancelOrder, "order_cancel", data)

	// 更新订单状态 - 使用分表
	orderTable := database.GetOrderTableName(shopID)
	s.db.Table(orderTable).
		Where("shop_id = ? AND order_sn = ?", shopID, cancelData.OrderSN).
		Update("order_status", consts.OrderStatusCancelled)

	shipmentTable := database.GetShipmentTableName(shopID)
	s.db.Table(shipmentTable).
		Where("shop_id = ? AND order_sn = ?", shopID, cancelData.OrderSN).
		Updates(map[string]interface{}{
			"ship_status": consts.ShipStatusFailed,
			"remark":      fmt.Sprintf("订单已取消: %s - %s", cancelData.CancelBy, cancelData.CancelReason),
		})

	// 处理预付款解冻 (如果订单已发货且有冻结金额)
	s.handleOrderCancelRefund(ctx, shopID, cancelData.OrderSN, cancelData.CancelBy, cancelData.CancelReason)

	rdb := database.GetRedis()
	cacheKey := fmt.Sprintf(consts.KeyOrderStatus, shopID, cancelData.OrderSN)
	rdb.Del(ctx, cacheKey)
}

// handleOrderCancelRefund 处理订单取消退款 - 使用分表
func (s *WebhookService) handleOrderCancelRefund(ctx context.Context, shopID uint64, orderSN string, cancelBy string, cancelReason string) {
	// 查找发货记录
	shipmentRecordTable := database.GetOrderShipmentRecordTableName(shopID)
	var shipmentRecord models.OrderShipmentRecord
	err := s.db.Table(shipmentRecordTable).Where("shop_id = ? AND order_sn = ?", shopID, orderSN).First(&shipmentRecord).Error
	if err != nil {
		// 没有发货记录，无需处理
		return
	}

	// 只处理已发货但未结算的订单
	if shipmentRecord.Status != models.ShipmentRecordStatusShipped {
		return
	}

	// 解冻预付款
	accountService := NewAccountService()
	_, err = accountService.UnfreezePrepayment(ctx, shipmentRecord.ShopOwnerID, shipmentRecord.FrozenAmount, orderSN,
		fmt.Sprintf("订单取消退款: %s - %s", cancelBy, cancelReason))
	if err != nil {
		s.logError(ctx, shopID, consts.WebhookBuyerCancelOrder, "unfreeze_error", err)
		return
	}

	// 从托管账户退回
	err = accountService.TransferFromEscrow(ctx, shipmentRecord.FrozenAmount, orderSN, "订单取消退回")
	if err != nil {
		s.logError(ctx, shopID, consts.WebhookBuyerCancelOrder, "escrow_refund_error", err)
	}

	// 更新发货记录状态
	s.db.Table(shipmentRecordTable).
		Where("id = ?", shipmentRecord.ID).
		Updates(map[string]interface{}{
			"status": models.ShipmentRecordStatusCancelled,
			"remark": fmt.Sprintf("订单取消: %s - %s", cancelBy, cancelReason),
		})
}

// HandleReservedOrder 处理保留订单
func (s *WebhookService) HandleReservedOrder(ctx context.Context, shopID uint64, data any) {
	s.LogWebhook(ctx, shopID, consts.WebhookReservedStock, "reserved_order", data)
}

// LogWebhook 记录Webhook日志 - 使用分表
func (s *WebhookService) LogWebhook(ctx context.Context, shopID uint64, code int, eventType string, data any) {
	dataJSON, _ := json.Marshal(data)

	logTable := database.GetOperationLogTableName(shopID)
	log := &models.OperationLog{
		ShopID:        shopID,
		OperationType: consts.OpTypeWebhook,
		OperationDesc: fmt.Sprintf("[%s] code=%d data=%s", eventType, code, string(dataJSON)),
		Status:        consts.OpStatusSuccess,
		CreatedAt:     time.Now(),
	}
	s.db.Table(logTable).Create(log)
}

func (s *WebhookService) refreshOrderDetail(shopID uint64, orderSN string) {
	ctx, cancel := NewBackgroundContext()
	defer cancel()

	var shop models.Shop
	if err := s.db.Where("shop_id = ?", shopID).First(&shop).Error; err != nil {
		return
	}

	var auth models.ShopAuthorization
	if err := s.db.Where("shop_id = ?", shopID).First(&auth).Error; err != nil {
		return
	}

	client := shopee.NewClient(shop.Region)
	limiter := shopee.GetRateLimiter(shopID)

	if err := limiter.Wait(ctx); err != nil {
		return
	}

	var detailResp *shopee.OrderDetailResponse
	err := shopee.RetryWithBackoff(ctx, consts.ShopeeAPIRetryTimes, func() error {
		var err error
		detailResp, err = client.GetOrderDetail(auth.AccessToken, shopID, []string{orderSN})
		return err
	})
	if err != nil || len(detailResp.Response.OrderList) == 0 {
		return
	}

	s.saveOrderFromWebhook(ctx, shopID, &detailResp.Response.OrderList[0])
}

func (s *WebhookService) saveOrderFromWebhook(ctx context.Context, shopID uint64, detail *shopee.OrderDetail) {
	var shop models.Shop
	s.db.Where("shop_id = ?", shopID).First(&shop)

	order := models.Order{
		ShopID:          shopID,
		OrderSN:         detail.OrderSN,
		Region:          shop.Region,
		OrderStatus:     detail.OrderStatus,
		BuyerUserID:     uint64(detail.BuyerUserID),
		BuyerUsername:   detail.BuyerUsername,
		Currency:        detail.Currency,
		ShippingCarrier: detail.ShippingCarrier,
		TrackingNumber:  detail.TrackingNo,
	}

	if detail.PayTime > 0 {
		payTime := time.Unix(detail.PayTime, 0)
		order.PayTime = &payTime
	}
	if detail.CreateTime > 0 {
		createTime := time.Unix(detail.CreateTime, 0)
		order.CreateTime = &createTime
	}

	orderTable := database.GetOrderTableName(shopID)
	var existing models.Order
	if err := s.db.Table(orderTable).Where("shop_id = ? AND order_sn = ?", shopID, detail.OrderSN).First(&existing).Error; err == nil {
		if !existing.StatusLocked {
			order.ID = existing.ID
			s.db.Table(orderTable).Where("id = ?", order.ID).Updates(&order)
		}
	} else {
		s.db.Table(orderTable).Create(&order)
	}
}

func (s *WebhookService) logError(ctx context.Context, shopID uint64, code int, errType string, err error) {
	logTable := database.GetOperationLogTableName(shopID)
	log := &models.OperationLog{
		ShopID:        shopID,
		OperationType: consts.OpTypeWebhook,
		OperationDesc: fmt.Sprintf("[error] code=%d type=%s error=%s", code, errType, err.Error()),
		Status:        consts.OpStatusFailed,
		CreatedAt:     time.Now(),
	}
	s.db.Table(logTable).Create(log)
}
