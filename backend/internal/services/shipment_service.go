package services

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"balance/backend/internal/consts"
	"balance/backend/internal/database"
	"balance/backend/internal/models"
	"balance/backend/internal/shopee"

	"gorm.io/gorm"
)

// ShipmentService 发货服务
type ShipmentService struct {
	db           *gorm.DB
	shopService  *ShopService
	orderService *OrderService
}

// NewShipmentService 创建发货服务
func NewShipmentService() *ShipmentService {
	return &ShipmentService{
		db:           database.GetDB(),
		shopService:  NewShopService(),
		orderService: NewOrderService(),
	}
}

// ShipOrderRequest 发货请求
type ShipOrderRequest struct {
	ShopID          uint64 `json:"shop_id" binding:"required"`
	OrderSN         string `json:"order_sn" binding:"required"`
	TrackingNumber  string `json:"tracking_number" binding:"required"`
	ShippingCarrier string `json:"shipping_carrier"`
}

// ShipOrder 订单发货
func (s *ShipmentService) ShipOrder(ctx context.Context, req *ShipOrderRequest) error {
	rdb := database.GetRedis()

	lockKey := fmt.Sprintf(consts.KeyShipLock, req.ShopID, req.OrderSN)
	ok, err := rdb.SetNX(ctx, lockKey, time.Now().Unix(), consts.ShipLockExpire).Result()
	if err != nil {
		return fmt.Errorf("获取发货锁失败: %w", err)
	}
	if !ok {
		return fmt.Errorf("订单正在发货中，请勿重复操作")
	}
	defer rdb.Del(ctx, lockKey)

	order, err := s.orderService.GetOrder(ctx, req.ShopID, req.OrderSN)
	if err != nil {
		return fmt.Errorf("订单不存在: %w", err)
	}

	if !order.CanShip() {
		return fmt.Errorf("订单状态不允许发货，当前状态: %s", order.OrderStatus)
	}

	shop, err := s.shopService.GetShop(ctx, req.ShopID)
	if err != nil {
		return fmt.Errorf("店铺不存在: %w", err)
	}

	accessToken, err := s.shopService.GetAccessToken(ctx, req.ShopID)
	if err != nil {
		return err
	}

	client := shopee.NewClient(shop.Region)

	reqData, _ := json.Marshal(req)

	shipment := models.Shipment{
		ShopID:          req.ShopID,
		OrderSN:         req.OrderSN,
		ShippingCarrier: req.ShippingCarrier,
		TrackingNumber:  req.TrackingNumber,
		ShipStatus:      consts.ShipStatusPending,
	}

	if err := s.db.Create(&shipment).Error; err != nil {
		if err := s.db.Where("shop_id = ? AND order_sn = ?", req.ShopID, req.OrderSN).
			Assign(shipment).FirstOrCreate(&shipment).Error; err != nil {
			return fmt.Errorf("创建发货记录失败: %w", err)
		}
	}

	_, shipErr := client.ShipOrder(accessToken, req.ShopID, req.OrderSN, req.TrackingNumber)

	now := time.Now()
	var respData []byte
	var logStatus int8 = consts.OpStatusSuccess

	if shipErr != nil {
		shipment.ShipStatus = consts.ShipStatusFailed
		shipment.ErrorMessage = shipErr.Error()
		logStatus = consts.OpStatusFailed
		respData = []byte(shipErr.Error())
	} else {
		shipment.ShipStatus = consts.ShipStatusShipped
		shipment.ShipTime = &now
		respData = []byte("success")

		s.db.Model(&models.Order{}).
			Where("shop_id = ? AND order_sn = ?", req.ShopID, req.OrderSN).
			Updates(map[string]interface{}{
				"order_status":     consts.OrderStatusProcessed,
				"tracking_number":  req.TrackingNumber,
				"shipping_carrier": req.ShippingCarrier,
			})

		statusKey := fmt.Sprintf(consts.KeyOrderStatus, req.ShopID, req.OrderSN)
		rdb.Set(ctx, statusKey, consts.OrderStatusProcessed, consts.OrderStatusExpire)
	}

	s.db.Save(&shipment)

	s.logOperation(req.ShopID, req.OrderSN, consts.OpTypeOrderShip, "订单发货",
		string(reqData), string(respData), logStatus, "")

	if shipErr != nil {
		return shipErr
	}

	return nil
}

// BatchShipOrders 批量发货
func (s *ShipmentService) BatchShipOrders(ctx context.Context, requests []*ShipOrderRequest) []BatchShipResult {
	results := make([]BatchShipResult, len(requests))

	for i, req := range requests {
		results[i] = BatchShipResult{
			ShopID:  req.ShopID,
			OrderSN: req.OrderSN,
		}

		if err := s.ShipOrder(ctx, req); err != nil {
			results[i].Success = false
			results[i].Error = err.Error()
		} else {
			results[i].Success = true
		}
	}

	return results
}

// BatchShipResult 批量发货结果
type BatchShipResult struct {
	ShopID  uint64 `json:"shop_id"`
	OrderSN string `json:"order_sn"`
	Success bool   `json:"success"`
	Error   string `json:"error,omitempty"`
}

// GetShippingParameter 获取发货参数
func (s *ShipmentService) GetShippingParameter(ctx context.Context, shopID uint64, orderSN string) (*shopee.GetShippingParameterResponse, error) {
	shop, err := s.shopService.GetShop(ctx, shopID)
	if err != nil {
		return nil, fmt.Errorf("店铺不存在: %w", err)
	}

	accessToken, err := s.shopService.GetAccessToken(ctx, shopID)
	if err != nil {
		return nil, err
	}

	client := shopee.NewClient(shop.Region)
	return client.GetShippingParameter(accessToken, shopID, orderSN)
}

// GetTrackingNumber 获取运单号
func (s *ShipmentService) GetTrackingNumber(ctx context.Context, shopID uint64, orderSN string) (*shopee.GetTrackingNumberResponse, error) {
	shop, err := s.shopService.GetShop(ctx, shopID)
	if err != nil {
		return nil, fmt.Errorf("店铺不存在: %w", err)
	}

	accessToken, err := s.shopService.GetAccessToken(ctx, shopID)
	if err != nil {
		return nil, err
	}

	client := shopee.NewClient(shop.Region)
	return client.GetTrackingNumber(accessToken, shopID, orderSN)
}

// GetShipment 获取发货记录
func (s *ShipmentService) GetShipment(ctx context.Context, shopID uint64, orderSN string) (*models.Shipment, error) {
	var shipment models.Shipment
	if err := s.db.Where("shop_id = ? AND order_sn = ?", shopID, orderSN).
		First(&shipment).Error; err != nil {
		return nil, err
	}
	return &shipment, nil
}

// ListShipments 获取发货记录列表
func (s *ShipmentService) ListShipments(ctx context.Context, shopID uint64, status *int8, page, pageSize int) ([]models.Shipment, int64, error) {
	var shipments []models.Shipment
	var total int64

	query := s.db.Model(&models.Shipment{})
	if shopID > 0 {
		query = query.Where("shop_id = ?", shopID)
	}
	if status != nil {
		query = query.Where("ship_status = ?", *status)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	if err := query.Offset(offset).Limit(pageSize).Order("id DESC").Find(&shipments).Error; err != nil {
		return nil, 0, err
	}

	return shipments, total, nil
}

func (s *ShipmentService) logOperation(shopID uint64, orderSN, opType, opDesc, reqData, respData string, status int8, ip string) {
	log := models.OperationLog{
		ShopID:        shopID,
		OrderSN:       orderSN,
		OperationType: opType,
		OperationDesc: opDesc,
		RequestData:   reqData,
		ResponseData:  respData,
		Status:        status,
		IP:            ip,
	}
	s.db.Create(&log)
}

// SyncLogisticsChannels 同步物流渠道
func (s *ShipmentService) SyncLogisticsChannels(ctx context.Context, shopID uint64) error {
	shop, err := s.shopService.GetShop(ctx, shopID)
	if err != nil {
		return fmt.Errorf("店铺不存在: %w", err)
	}

	accessToken, err := s.shopService.GetAccessToken(ctx, shopID)
	if err != nil {
		return err
	}

	client := shopee.NewClient(shop.Region)
	resp, err := client.GetLogisticsChannelList(accessToken, shopID)
	if err != nil {
		return fmt.Errorf("获取物流渠道失败: %w", err)
	}

	for _, ch := range resp.Response.LogisticsChannelList {
		channel := models.LogisticsChannel{
			ShopID:               shopID,
			LogisticsChannelID:   uint64(ch.LogisticsChannelID),
			LogisticsChannelName: ch.LogisticsChannelName,
			CODEnabled:           boolToInt8(ch.CODEnabled),
			Enabled:              boolToInt8(ch.Enabled),
		}

		if err := s.db.Where("shop_id = ? AND logistics_channel_id = ?", shopID, ch.LogisticsChannelID).
			Assign(channel).FirstOrCreate(&channel).Error; err != nil {
			return fmt.Errorf("保存物流渠道失败: %w", err)
		}
	}

	rdb := database.GetRedis()
	cacheKey := fmt.Sprintf(consts.KeyLogistics, shopID)
	cacheData, _ := json.Marshal(resp.Response.LogisticsChannelList)
	rdb.Set(ctx, cacheKey, cacheData, consts.LogisticsExpire)

	return nil
}

// GetLogisticsChannels 获取物流渠道列表
func (s *ShipmentService) GetLogisticsChannels(ctx context.Context, shopID uint64) ([]models.LogisticsChannel, error) {
	var channels []models.LogisticsChannel
	if err := s.db.Where("shop_id = ? AND enabled = 1", shopID).Find(&channels).Error; err != nil {
		return nil, err
	}
	return channels, nil
}

func boolToInt8(b bool) int8 {
	if b {
		return 1
	}
	return 0
}
