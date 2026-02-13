package shopower

import (
	"context"
	"fmt"
	"time"

	"balance/backend/internal/consts"
	"balance/backend/internal/database"
	"balance/backend/internal/models"
	"balance/backend/internal/shopee"
	"balance/backend/internal/utils"

	"gorm.io/gorm"
)

// ShipOrderRequest 发货请求
type ShipOrderRequest struct {
	ShopID   int64  `json:"shop_id"`
	OrderSN  string `json:"order_sn"`
	Pickup   bool   `json:"pickup"`
	Dropoff  bool   `json:"dropoff"`
	NonInteg bool   `json:"non_integrated"`
}

// BatchShipResult 批量发货结果
type BatchShipResult struct {
	OrderSN string `json:"order_sn"`
	Success bool   `json:"success"`
	Message string `json:"message"`
}

// ShipmentService 发货服务（店主专用）
type ShipmentService struct {
	db          *gorm.DB
	shardedDB   *database.ShardedDB
	shopService *ShopService
}

// NewShipmentService 创建发货服务
func NewShipmentService() *ShipmentService {
	db := database.GetDB()
	return &ShipmentService{
		db:          db,
		shardedDB:   database.NewShardedDB(db),
		shopService: NewShopService(),
	}
}

// ShipOrder 发货（带锁、状态检查、日志记录）
func (s *ShipmentService) ShipOrder(ctx context.Context, adminID int64, req *ShipOrderRequest) error {
	rdb := database.GetRedis()

	// 发货锁，防止重复发货
	lockKey := fmt.Sprintf(consts.KeyShipLock, req.ShopID, req.OrderSN)
	ok, err := rdb.SetNX(ctx, lockKey, time.Now().Unix(), consts.ShipLockExpire).Result()
	if err != nil {
		return fmt.Errorf("获取发货锁失败: %w", err)
	}
	if !ok {
		return utils.ErrOrderShipping
	}
	defer rdb.Del(ctx, lockKey)

	shop, err := s.shopService.GetShop(ctx, adminID, req.ShopID)
	if err != nil {
		return err
	}

	// 检查订单状态 - 使用分表
	orderTable := database.GetOrderTableName(uint64(req.ShopID))
	var order models.Order
	if err := s.db.Table(orderTable).Where("shop_id = ? AND order_sn = ?", req.ShopID, req.OrderSN).First(&order).Error; err != nil {
		return fmt.Errorf("订单不存在: %w", err)
	}
	if !order.CanShip() {
		return fmt.Errorf("订单状态不允许发货，当前状态: %s", order.OrderStatus)
	}

	accessToken, err := s.getAccessToken(ctx, uint64(req.ShopID))
	if err != nil {
		return err
	}

	client := shopee.NewClient(shop.Region)

	// 创建发货记录 - 使用分表
	shipmentTable := database.GetShipmentTableName(uint64(req.ShopID))
	now := time.Now()
	shipment := models.Shipment{
		ShopID:          uint64(req.ShopID),
		OrderSN:         req.OrderSN,
		ShippingCarrier: "",
		TrackingNumber:  "",
		ShipStatus:      consts.ShipStatusPending,
	}
	var existingShipment models.Shipment
	if err := s.db.Table(shipmentTable).Where("shop_id = ? AND order_sn = ?", req.ShopID, req.OrderSN).First(&existingShipment).Error; err == nil {
		shipment.ID = existingShipment.ID
		s.db.Table(shipmentTable).Where("id = ?", shipment.ID).Updates(&shipment)
	} else {
		if err := s.db.Table(shipmentTable).Create(&shipment).Error; err != nil {
			return fmt.Errorf("创建发货记录失败: %w", err)
		}
	}

	// 调用Shopee API发货
	_, shipErr := client.ShipOrder(accessToken, uint64(req.ShopID), req.OrderSN, "")

	var logStatus int8 = consts.OpStatusSuccess
	if shipErr != nil {
		shipment.ShipStatus = consts.ShipStatusFailed
		shipment.ErrorMessage = shipErr.Error()
		logStatus = consts.OpStatusFailed
	} else {
		shipment.ShipStatus = consts.ShipStatusShipped
		shipment.ShipTime = &now

		// 更新订单状态 - 使用分表
		s.db.Table(orderTable).
			Where("shop_id = ? AND order_sn = ?", req.ShopID, req.OrderSN).
			Updates(map[string]interface{}{
				"order_status": consts.OrderStatusProcessed,
			})

		// 更新缓存
		statusKey := fmt.Sprintf(consts.KeyOrderStatus, req.ShopID, req.OrderSN)
		rdb.Set(ctx, statusKey, consts.OrderStatusProcessed, consts.OrderStatusExpire)
	}

	s.db.Table(shipmentTable).Where("id = ?", shipment.ID).Updates(&shipment)

	// 记录操作日志
	s.logOperation(uint64(req.ShopID), req.OrderSN, consts.OpTypeOrderShip, "订单发货", logStatus)

	if shipErr != nil {
		return shipErr
	}

	return nil
}

func (s *ShipmentService) logOperation(shopID uint64, orderSN, opType, opDesc string, status int8) {
	logTable := database.GetOperationLogTableName(shopID)
	log := models.OperationLog{
		ShopID:        shopID,
		OrderSN:       orderSN,
		OperationType: opType,
		OperationDesc: opDesc,
		Status:        status,
	}
	s.db.Table(logTable).Create(&log)
}

func (s *ShipmentService) getAccessToken(ctx context.Context, shopID uint64) (string, error) {
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

// BatchShipOrders 批量发货
func (s *ShipmentService) BatchShipOrders(ctx context.Context, adminID int64, orders []ShipOrderRequest) ([]BatchShipResult, error) {
	results := make([]BatchShipResult, len(orders))
	for i, order := range orders {
		err := s.ShipOrder(ctx, adminID, &order)
		results[i] = BatchShipResult{
			OrderSN: order.OrderSN,
			Success: err == nil,
			Message: "",
		}
		if err != nil {
			results[i].Message = err.Error()
		}
	}
	return results, nil
}

// GetShippingParameter 获取发货参数
func (s *ShipmentService) GetShippingParameter(ctx context.Context, adminID int64, shopID int64, orderSN string) (interface{}, error) {
	shop, err := s.shopService.GetShop(ctx, adminID, shopID)
	if err != nil {
		return nil, err
	}

	accessToken, err := s.getAccessToken(ctx, uint64(shopID))
	if err != nil {
		return nil, err
	}

	client := shopee.NewClient(shop.Region)
	return client.GetShippingParameter(accessToken, uint64(shopID), orderSN)
}

// GetTrackingNumber 获取物流单号
func (s *ShipmentService) GetTrackingNumber(ctx context.Context, adminID int64, shopID int64, orderSN string) (string, error) {
	shop, err := s.shopService.GetShop(ctx, adminID, shopID)
	if err != nil {
		return "", err
	}

	accessToken, err := s.getAccessToken(ctx, uint64(shopID))
	if err != nil {
		return "", err
	}

	client := shopee.NewClient(shop.Region)
	resp, err := client.GetTrackingNumber(accessToken, uint64(shopID), orderSN)
	if err != nil {
		return "", err
	}
	return resp.Response.TrackingNumber, nil
}

// ListShipments 获取发货记录列表 - 使用分表
func (s *ShipmentService) ListShipments(ctx context.Context, adminID int64, shopID int64, page, pageSize int) ([]models.Shipment, int64, error) {
	var shopIDs []uint64
	s.db.Model(&models.Shop{}).Where("admin_id = ?", adminID).Pluck("shop_id", &shopIDs)
	if len(shopIDs) == 0 {
		return []models.Shipment{}, 0, nil
	}

	// 如果指定了shopID，直接查询对应分表
	if shopID > 0 {
		shipmentTable := database.GetShipmentTableName(uint64(shopID))
		var shipments []models.Shipment
		var total int64

		query := s.db.Table(shipmentTable).Where("shop_id = ?", shopID)
		query.Count(&total)

		offset := (page - 1) * pageSize
		query.Offset(offset).Limit(pageSize).Order("id DESC").Find(&shipments)
		return shipments, total, nil
	}

	// 按分表索引分组店铺
	shardShops := make(map[int][]uint64)
	for _, sid := range shopIDs {
		idx := database.GetShardIndex(sid)
		shardShops[idx] = append(shardShops[idx], sid)
	}

	var allShipments []models.Shipment
	var total int64

	for idx, sids := range shardShops {
		shipmentTable := fmt.Sprintf("shipments_%d", idx)
		query := s.db.Table(shipmentTable).Where("shop_id IN ?", sids)

		var count int64
		query.Count(&count)
		total += count

		var shipments []models.Shipment
		query.Order("id DESC").Find(&shipments)
		allShipments = append(allShipments, shipments...)
	}

	// 内存分页
	offset := (page - 1) * pageSize
	end := offset + pageSize
	if offset >= len(allShipments) {
		return []models.Shipment{}, total, nil
	}
	if end > len(allShipments) {
		end = len(allShipments)
	}

	return allShipments[offset:end], total, nil
}

// GetShipment 获取发货详情 - 使用分表
func (s *ShipmentService) GetShipment(ctx context.Context, adminID int64, shopID int64, orderSN string) (*models.Shipment, error) {
	if _, err := s.shopService.GetShop(ctx, adminID, shopID); err != nil {
		return nil, err
	}

	shipmentTable := database.GetShipmentTableName(uint64(shopID))
	var shipment models.Shipment
	if err := s.db.Table(shipmentTable).Where("shop_id = ? AND order_sn = ?", shopID, orderSN).First(&shipment).Error; err != nil {
		return nil, utils.ErrShipmentNotFound
	}
	return &shipment, nil
}

// SyncLogisticsChannels 同步物流渠道
func (s *ShipmentService) SyncLogisticsChannels(ctx context.Context, adminID int64, shopID int64) error {
	shop, err := s.shopService.GetShop(ctx, adminID, shopID)
	if err != nil {
		return err
	}

	accessToken, err := s.getAccessToken(ctx, uint64(shopID))
	if err != nil {
		return err
	}

	client := shopee.NewClient(shop.Region)
	channels, err := client.GetLogisticsChannelList(accessToken, uint64(shopID))
	if err != nil {
		return fmt.Errorf("获取物流渠道失败: %w", err)
	}

	for _, ch := range channels.Response.LogisticsChannelList {
		var enabled int8
		if ch.Enabled {
			enabled = 1
		}
		channel := models.LogisticsChannel{
			ShopID:               uint64(shopID),
			LogisticsChannelID:   uint64(ch.LogisticsChannelID),
			LogisticsChannelName: ch.LogisticsChannelName,
			Enabled:              enabled,
		}
		s.db.Where("shop_id = ? AND logistics_channel_id = ?", shopID, ch.LogisticsChannelID).
			Assign(channel).FirstOrCreate(&channel)
	}

	return nil
}

// GetLogisticsChannels 获取物流渠道列表
func (s *ShipmentService) GetLogisticsChannels(ctx context.Context, adminID int64, shopID int64) ([]models.LogisticsChannel, error) {
	if _, err := s.shopService.GetShop(ctx, adminID, shopID); err != nil {
		return nil, err
	}

	var channels []models.LogisticsChannel
	if err := s.db.Where("shop_id = ?", shopID).Find(&channels).Error; err != nil {
		return nil, err
	}
	return channels, nil
}
