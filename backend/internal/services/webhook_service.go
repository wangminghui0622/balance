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

	"gorm.io/gorm"
)

// WebhookService Webhook服务
type WebhookService struct {
	db           *gorm.DB
	orderService *shopower.OrderService
	shopService  *shopower.ShopService
}

// NewWebhookService 创建Webhook服务
func NewWebhookService() *WebhookService {
	return &WebhookService{
		db:           database.GetDB(),
		orderService: shopower.NewOrderService(),
		shopService:  shopower.NewShopService(),
	}
}

// NewBackgroundContext 创建后台上下文
func NewBackgroundContext() context.Context {
	ctx, _ := context.WithTimeout(context.Background(), 30*time.Second)
	return ctx
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

// AcquireOrderLock 获取订单更新锁
func (s *WebhookService) AcquireOrderLock(ctx context.Context, shopID uint64, orderSN string) bool {
	rdb := database.GetRedis()
	lockKey := fmt.Sprintf(consts.KeyOrderLock, shopID, orderSN)

	ok, err := rdb.SetNX(ctx, lockKey, time.Now().Unix(), consts.OrderLockExpire).Result()
	if err != nil {
		return false
	}
	return ok
}

// ReleaseOrderLock 释放订单更新锁
func (s *WebhookService) ReleaseOrderLock(ctx context.Context, shopID uint64, orderSN string) {
	rdb := database.GetRedis()
	lockKey := fmt.Sprintf(consts.KeyOrderLock, shopID, orderSN)
	rdb.Del(ctx, lockKey)
}

// CheckAndSetUpdateTime 检查并设置更新时间
func (s *WebhookService) CheckAndSetUpdateTime(ctx context.Context, shopID uint64, orderSN string, newUpdateTime int64) bool {
	rdb := database.GetRedis()
	updateTimeKey := fmt.Sprintf(consts.KeyOrderUpdateTime, shopID, orderSN)

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

// CanStatusTransition 检查状态转换是否合法
func (s *WebhookService) CanStatusTransition(ctx context.Context, shopID uint64, orderSN string, newStatus string) bool {
	var order models.Order
	if err := s.db.Where("shop_id = ? AND order_sn = ?", shopID, orderSN).First(&order).Error; err != nil {
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

	if !s.AcquireOrderLock(ctx, shopID, orderData.OrderSN) {
		return
	}
	defer s.ReleaseOrderLock(ctx, shopID, orderData.OrderSN)

	if orderData.UpdateTime > 0 && !s.CheckAndSetUpdateTime(ctx, shopID, orderData.OrderSN, orderData.UpdateTime) {
		return
	}

	if !s.CanStatusTransition(ctx, shopID, orderData.OrderSN, orderData.Status) {
		s.LogWebhook(ctx, shopID, consts.WebhookOrderStatus, "status_rollback_blocked", data)
		return
	}

	s.LogWebhook(ctx, shopID, consts.WebhookOrderStatus, "order_status", data)

	if err := s.db.Model(&models.Order{}).
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
		s.db.Model(&models.Shipment{}).
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

	if !s.AcquireOrderLock(ctx, shopID, cancelData.OrderSN) {
		return
	}
	defer s.ReleaseOrderLock(ctx, shopID, cancelData.OrderSN)

	s.LogWebhook(ctx, shopID, consts.WebhookBuyerCancelOrder, "order_cancel", data)

	s.db.Model(&models.Order{}).
		Where("shop_id = ? AND order_sn = ?", shopID, cancelData.OrderSN).
		Update("order_status", consts.OrderStatusCancelled)

	s.db.Model(&models.Shipment{}).
		Where("shop_id = ? AND order_sn = ?", shopID, cancelData.OrderSN).
		Updates(map[string]interface{}{
			"ship_status": consts.ShipStatusFailed,
			"remark":      fmt.Sprintf("订单已取消: %s - %s", cancelData.CancelBy, cancelData.CancelReason),
		})

	rdb := database.GetRedis()
	cacheKey := fmt.Sprintf(consts.KeyOrderStatus, shopID, cancelData.OrderSN)
	rdb.Del(ctx, cacheKey)
}

// HandleReservedOrder 处理保留订单
func (s *WebhookService) HandleReservedOrder(ctx context.Context, shopID uint64, data any) {
	s.LogWebhook(ctx, shopID, consts.WebhookReservedStock, "reserved_order", data)
}

// LogWebhook 记录Webhook日志
func (s *WebhookService) LogWebhook(ctx context.Context, shopID uint64, code int, eventType string, data any) {
	dataJSON, _ := json.Marshal(data)

	log := &models.OperationLog{
		ShopID:        shopID,
		OperationType: consts.OpTypeWebhook,
		OperationDesc: fmt.Sprintf("[%s] code=%d data=%s", eventType, code, string(dataJSON)),
		Status:        consts.OpStatusSuccess,
		CreatedAt:     time.Now(),
	}
	s.db.Create(log)
}

func (s *WebhookService) refreshOrderDetail(shopID uint64, orderSN string) {
	ctx := NewBackgroundContext()

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

	var existing models.Order
	if err := s.db.Where("shop_id = ? AND order_sn = ?", shopID, detail.OrderSN).First(&existing).Error; err == nil {
		if !existing.StatusLocked {
			order.ID = existing.ID
			s.db.Save(&order)
		}
	} else {
		s.db.Create(&order)
	}
}

func (s *WebhookService) logError(ctx context.Context, shopID uint64, code int, errType string, err error) {
	log := &models.OperationLog{
		ShopID:        shopID,
		OperationType: consts.OpTypeWebhook,
		OperationDesc: fmt.Sprintf("[error] code=%d type=%s error=%s", code, errType, err.Error()),
		Status:        consts.OpStatusFailed,
		CreatedAt:     time.Now(),
	}
	s.db.Create(log)
}
