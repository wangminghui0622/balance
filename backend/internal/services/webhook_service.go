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
	db            *gorm.DB
	shardedDB     *database.ShardedDB
	orderService  *shopower.OrderService
	shopService   *shopower.ShopService
	returnService *ReturnService
	rs            *redsync.Redsync
	idGenerator   *utils.IDGenerator
}

// NewWebhookService 创建Webhook服务
func NewWebhookService() *WebhookService {
	db := database.GetDB()
	return &WebhookService{
		db:            db,
		shardedDB:     database.NewShardedDB(db),
		orderService:  NewOrderServiceWithPrepaymentCheck(),
		shopService:   shopower.NewShopService(),
		returnService: NewReturnService(),
		rs:            database.GetRedsync(),
		idGenerator:   utils.NewIDGenerator(database.GetRedis()),
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
func (s *WebhookService) CanStatusTransition(ctx context.Context, orderTable string, shopID uint64, orderSN string, newStatus string) bool {
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
	//1.重复推送同一条消息（网络重试等原因），如果这个 key 已经存在，说明这条 Webhook 之前已经处理过了，直接丢弃
	if s.CheckWebhookDuplicate(ctx, shopID, orderData.OrderSN, consts.WebhookOrderStatus, timestamp) {
		return
	}
	//2.防止同一时刻有多个 Webhook 同时修改同一个订单,（比如状态更新和取消同时到达）。获取锁失败说明有其他协程正在处理这个订单，直接放弃
	unlockFunc := s.AcquireOrderLock(ctx, shopID, orderData.OrderSN)
	if unlockFunc == nil {
		return
	}
	defer unlockFunc()
	//3.防止乱序消息。Shopee 的 Webhook 可能不按时间顺序到达（比如旧的更新晚于新的更新到达）。
	//用 Redis Lua 脚本原子性地检查：如果这条消息的 update_time ≤ 已经处理过的最新时间，说明是过期消息，丢弃。否则更新为新的最新时间
	if orderData.UpdateTime > 0 && !s.CheckAndSetUpdateTime(ctx, shopID, orderData.OrderSN, orderData.UpdateTime) {
		return
	}
	orderTable := database.GetOrderTableName(shopID)
	//4.防止状态回退。订单状态有优先级（比如 COMPLETED > SHIPPED > READY_TO_SHIP），如果新状态的优先级低于当前状态（比如已经是 COMPLETED 了又来了一个 SHIPPED），
	//就阻止这次更新。同时如果订单被店主手动锁定了状态（StatusLocked），也会被拦截。被拦截的事件会记录日志，方便排查。
	if !s.CanStatusTransition(ctx, orderTable, shopID, orderData.OrderSN, orderData.Status) {
		s.LogWebhook(ctx, shopID, consts.WebhookOrderStatus, "status_rollback_blocked", data)
		return
	}
	if orderData.Status == consts.OrderStatusReadyToShip {
		// 待发货：拉取完整订单详情并全量更新（包含状态、买家、物流、金额等所有字段）
		go s.refreshOrderDetail(shopID, orderData.OrderSN)
	} else {
		// 其他状态：只需更新状态字段（兜底检查 status_locked，防止覆盖锁定状态）
		if err := s.db.Table(orderTable).
			Where("shop_id = ? AND order_sn = ? AND status_locked = false", shopID, orderData.OrderSN).
			Update("order_status", orderData.Status).Error; err != nil {
			s.logError(ctx, shopID, consts.WebhookOrderStatus, "update_error", err)
			return
		}
	}

	s.LogWebhook(ctx, shopID, consts.WebhookOrderStatus, "order_status", data)

	// 清除状态缓存，下次查询从数据库读取最新值
	rdb := database.GetRedis()
	cacheKey := fmt.Sprintf(consts.KeyOrderStatus, shopID, orderData.OrderSN)
	rdb.Del(ctx, cacheKey)
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

	// 获取订单锁，防止与同一订单的状态更新/取消并发冲突
	unlockFunc := s.AcquireOrderLock(ctx, shopID, trackingData.OrderSN)
	if unlockFunc == nil {
		return
	}
	defer unlockFunc()

	s.LogWebhook(ctx, shopID, consts.WebhookTrackingUpdate, "tracking_update", data)

	if trackingData.TrackingNumber != "" {
		// 同时更新 shipments 表和 orders 表的物流单号
		shipmentTable := database.GetShipmentTableName(shopID)
		s.db.Table(shipmentTable).
			Where("shop_id = ? AND order_sn = ?", shopID, trackingData.OrderSN).
			Update("tracking_number", trackingData.TrackingNumber)

		orderTable := database.GetOrderTableName(shopID)
		s.db.Table(orderTable).
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

	// 更新订单状态（检查 status_locked，防止覆盖店主手动锁定的状态）
	orderTable := database.GetOrderTableName(shopID)
	s.db.Table(orderTable).
		Where("shop_id = ? AND order_sn = ? AND status_locked = false", shopID, cancelData.OrderSN).
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
	err = accountService.TransferFromEscrow(ctx, shipmentRecord.ShopOwnerID, shipmentRecord.FrozenAmount, orderSN, "订单取消退回")
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

// HandleReturn 处理退货退款事件（退货创建 / 退货状态变更）
//
// Shopee 推送退货事件时调用，从 data 中提取 return_sn 和 order_sn，
// 然后调用 GetReturnDetail 获取完整退货信息并保存到数据库。
// 如果退货状态已确认退款，会自动解冻预付款。
func (s *WebhookService) HandleReturn(ctx context.Context, shopID uint64, data any, timestamp int64, code int) {
	dataBytes, err := json.Marshal(data)
	if err != nil {
		s.logError(ctx, shopID, code, "marshal_error", err)
		return
	}

	var returnData struct {
		ReturnSN   string `json:"returnsn"`
		OrderSN    string `json:"ordersn"`
		Status     string `json:"status"`
		UpdateTime int64  `json:"update_time"`
	}
	if err := json.Unmarshal(dataBytes, &returnData); err != nil {
		s.logError(ctx, shopID, code, "unmarshal_error", err)
		return
	}

	if returnData.ReturnSN == "" {
		return
	}

	// Webhook 去重（使用 eventCode 而非 timestamp，防止虾皮重试时不同 timestamp 绕过去重）
	rdb := database.GetRedis()
	dedupKey := fmt.Sprintf(consts.KeyReturnDedup, shopID, returnData.ReturnSN, code)
	ok, err := rdb.SetNX(ctx, dedupKey, 1, consts.WebhookDedupExpire).Result()
	if err != nil || !ok {
		return // 已处理过或 Redis 错误
	}

	// 记录日志
	s.LogWebhook(ctx, shopID, code, "return_event", data)

	// 异步拉取退货详情并处理退款
	go s.refreshReturnDetail(shopID, returnData.ReturnSN)
}

// refreshReturnDetail 从 Shopee 拉取退货详情并调用 ReturnService 处理
func (s *WebhookService) refreshReturnDetail(shopID uint64, returnSN string) {
	ctx, cancel := NewBackgroundContext()
	defer cancel()

	// 查询店铺信息
	var shop models.Shop
	if err := s.db.Where("shop_id = ?", shopID).First(&shop).Error; err != nil {
		return
	}

	// 查询授权信息
	var auth models.ShopAuthorization
	if err := s.db.Where("shop_id = ?", shopID).First(&auth).Error; err != nil {
		return
	}

	client := shopee.NewClient(shop.Region)

	// 等待限流令牌
	if err := shopee.WaitForRateLimit(ctx, shopID); err != nil {
		return
	}

	// 调用 Shopee GetReturnDetail 获取完整退货信息
	var detailResp *shopee.ReturnDetailResponse
	err := shopee.RetryWithBackoff(ctx, consts.ShopeeAPIRetryTimes, func() error {
		var err error
		detailResp, err = client.GetReturnDetail(auth.AccessToken, shopID, returnSN)
		return err
	})
	if err != nil {
		fmt.Printf("[Webhook] 店铺=%d 退货=%s 获取退货详情失败: %v\n", shopID, returnSN, err)
		return
	}

	// 通过 ReturnService 写入 DB 并处理退款（复用实例，避免每次创建新服务）
	if err := s.returnService.UpsertReturn(ctx, shopID, detailResp); err != nil {
		fmt.Printf("[Webhook] 店铺=%d 退货=%s 写入失败: %v\n", shopID, returnSN, err)
	}
}

// HandleReservedOrder 处理保留订单
func (s *WebhookService) HandleReservedOrder(ctx context.Context, shopID uint64, data any) {
	s.LogWebhook(ctx, shopID, consts.WebhookReservedStock, "reserved_order", data)
}

// LogWebhook 记录Webhook日志 - 使用分表
func (s *WebhookService) LogWebhook(ctx context.Context, shopID uint64, code int, eventType string, data any) {
	dataJSON, _ := json.Marshal(data)

	logID, _ := s.idGenerator.GenerateOperationLogID(ctx)
	logTable := database.GetOperationLogTableName(shopID)
	log := &models.OperationLog{
		ID:            uint64(logID),
		ShopID:        shopID,
		OperationType: consts.OpTypeWebhook,
		OperationDesc: fmt.Sprintf("[%s] code=%d data=%s", eventType, code, string(dataJSON)),
		Status:        consts.OpStatusSuccess,
		CreatedAt:     time.Now(),
	}
	s.db.Table(logTable).Create(log)
}

// refreshOrderDetail 从 Shopee 拉取完整订单详情并写入 DB
// 此方法在独立 goroutine 中运行（HandleOrderStatusUpdate 中 go 调用），
// 因此需要自己获取订单锁，防止与 Patrol 巡检并发写入同一订单时产生冲突。
func (s *WebhookService) refreshOrderDetail(shopID uint64, orderSN string) {
	// 独立的 30 秒超时上下文（不依赖调用方的 ctx）
	ctx, cancel := NewBackgroundContext()
	defer cancel()

	// 获取订单级分布式锁，防止与 Patrol 或其他 Webhook 并发写同一订单
	lockKey := fmt.Sprintf(consts.KeyOrderLock, shopID, orderSN)
	mutex := s.rs.NewMutex(lockKey,
		redsync.WithExpiry(consts.OrderLockExpire),
		redsync.WithTries(3),                        // 允许重试 3 次（上层 HandleOrderStatusUpdate 已释放锁）
		redsync.WithRetryDelay(200*time.Millisecond), // 每次重试间隔 200ms
	)
	unlockFunc, acquired := utils.TryLockWithAutoExtend(ctx, mutex, consts.OrderLockExpire/3)
	if !acquired {
		// 获取锁失败，说明有其他协程正在处理该订单，放弃（Patrol 或下次 Webhook 会补）
		return
	}
	defer unlockFunc()

	// 查询店铺信息（region 用于选择 API Host）
	var shop models.Shop
	if err := s.db.Where("shop_id = ?", shopID).First(&shop).Error; err != nil {
		return
	}

	// 查询授权信息（access_token 用于调 Shopee API）
	var auth models.ShopAuthorization
	if err := s.db.Where("shop_id = ?", shopID).First(&auth).Error; err != nil {
		return
	}

	client := shopee.NewClient(shop.Region)

	// 等待限流令牌
	if err := shopee.WaitForRateLimit(ctx, shopID); err != nil {
		return
	}

	// 调用 Shopee GetOrderDetail 获取完整订单信息（含商品、地址、金额等）
	var detailResp *shopee.OrderDetailResponse
	err := shopee.RetryWithBackoff(ctx, consts.ShopeeAPIRetryTimes, func() error {
		var err error
		detailResp, err = client.GetOrderDetail(auth.AccessToken, shopID, []string{orderSN})
		return err
	})
	if err != nil || len(detailResp.Response.OrderList) == 0 {
		return
	}

	// 通过统一的 UpsertOrder 写入 DB（内含事务 + 行锁 + 状态优先级 + 预付款检查）
	s.saveOrderFromWebhook(ctx, shopID, shop.Region, &detailResp.Response.OrderList[0])
}

func (s *WebhookService) saveOrderFromWebhook(ctx context.Context, shopID uint64, region string, detail *shopee.OrderDetail) {
	// 统一走 OrderService.UpsertOrder（唯一写入入口，含预付款检查）
	if err := s.orderService.UpsertOrder(ctx, shopID, region, detail); err != nil {
		fmt.Printf("[Webhook] 店铺=%d 订单=%s 写入失败: %v\n", shopID, detail.OrderSN, err)
	}
}

func (s *WebhookService) logError(ctx context.Context, shopID uint64, code int, errType string, err error) {
	logID, _ := s.idGenerator.GenerateOperationLogID(ctx)
	logTable := database.GetOperationLogTableName(shopID)
	log := &models.OperationLog{
		ID:            uint64(logID),
		ShopID:        shopID,
		OperationType: consts.OpTypeWebhook,
		OperationDesc: fmt.Sprintf("[error] code=%d type=%s error=%s", code, errType, err.Error()),
		Status:        consts.OpStatusFailed,
		CreatedAt:     time.Now(),
	}
	s.db.Table(logTable).Create(log)
}
