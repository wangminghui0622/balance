package operator

import (
	"context"
	"fmt"
	"time"

	"balance/backend/internal/database"
	"balance/backend/internal/models"
	"balance/backend/internal/services"
	"balance/backend/internal/services/shopower"
	"balance/backend/internal/shopee"
	"balance/backend/internal/utils"

	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

// ShipmentService 运营发货服务
type ShipmentService struct {
	db              *gorm.DB
	accountService  *services.AccountService
	shopService     *shopower.ShopService
	shipmentService *shopower.ShipmentService
	shardedDB       *database.ShardedDB
	idGenerator     *utils.IDGenerator
}

// NewShipmentService 创建运营发货服务
func NewShipmentService() *ShipmentService {
	db := database.GetDB()
	return &ShipmentService{
		db:              db,
		accountService:  services.NewAccountService(),
		shopService:     shopower.NewShopService(),
		shipmentService: shopower.NewShipmentService(),
		shardedDB:       database.NewShardedDB(db),
		idGenerator:     utils.NewIDGenerator(database.GetRedis()),
	}
}

// ShipOrderRequest 发货请求
type ShipOrderRequest struct {
	ShopID       uint64          `json:"shop_id" binding:"required"`
	OrderSN      string          `json:"order_sn" binding:"required"`
	GoodsCost    decimal.Decimal `json:"goods_cost" binding:"required"`    // 商品成本
	ShippingCost decimal.Decimal `json:"shipping_cost"`                    // 运费成本
	PickupInfo   *PickupInfo     `json:"pickup_info"`                      // 取件信息
}

// PickupInfo 取件信息
type PickupInfo struct {
	AddressID     uint64 `json:"address_id"`
	PickupTimeID  string `json:"pickup_time_id"`
	TrackingNo    string `json:"tracking_no"`
}

// ShipOrder 运营发货 (含预付款检查)
func (s *ShipmentService) ShipOrder(ctx context.Context, operatorID int64, req *ShipOrderRequest) (*models.OrderShipmentRecord, error) {
	// 1. 检查运营是否有权限操作该店铺
	relation, err := s.checkOperatorPermission(operatorID, req.ShopID)
	if err != nil {
		return nil, err
	}

	// 2. 获取订单信息 - 使用分表
	orderTable := database.GetOrderTableName(req.ShopID)
	var order models.Order
	if err := s.db.Table(orderTable).Where("shop_id = ? AND order_sn = ?", req.ShopID, req.OrderSN).First(&order).Error; err != nil {
		return nil, fmt.Errorf("订单不存在")
	}

	// 3. 检查订单状态
	if order.OrderStatus != "READY_TO_SHIP" {
		return nil, fmt.Errorf("订单状态不是待发货，当前状态: %s", order.OrderStatus)
	}

	// 4. 检查是否已有发货记录 - 使用分表
	shipmentRecordTable := database.GetOrderShipmentRecordTableName(req.ShopID)
	var existingRecord models.OrderShipmentRecord
	if err := s.db.Table(shipmentRecordTable).Where("order_sn = ?", req.OrderSN).First(&existingRecord).Error; err == nil {
		if existingRecord.Status == models.ShipmentRecordStatusShipped {
			return nil, fmt.Errorf("订单已发货")
		}
		if existingRecord.Status == models.ShipmentRecordStatusPending {
			return nil, fmt.Errorf("订单正在处理中")
		}
	}

	// 5. 计算总成本
	totalCost := req.GoodsCost.Add(req.ShippingCost)

	// 6. 检查店铺老板预付款余额
	balance, _, err := s.accountService.GetPrepaymentBalance(ctx, relation.ShopOwnerID)
	if err != nil {
		return nil, fmt.Errorf("获取预付款余额失败: %w", err)
	}
	if balance.LessThan(totalCost) {
		return nil, fmt.Errorf("店铺老板预付款余额不足，当前余额: %s, 需要: %s", balance.String(), totalCost.String())
	}

	// 7. 冻结预付款
	freezeTx, err := s.accountService.FreezePrepayment(ctx, relation.ShopOwnerID, totalCost, req.OrderSN, fmt.Sprintf("发货冻结-订单%s", req.OrderSN))
	if err != nil {
		return nil, fmt.Errorf("冻结预付款失败: %w", err)
	}

	// 8. 创建发货记录
	recordID, _ := s.idGenerator.GenerateShipmentRecordID(ctx)
	record := &models.OrderShipmentRecord{
		ID:                  uint64(recordID),
		ShopID:              req.ShopID,
		OrderSN:             req.OrderSN,
		OrderID:             order.ID,
		ShopOwnerID:         relation.ShopOwnerID,
		OperatorID:          operatorID,
		GoodsCost:           req.GoodsCost,
		ShippingCost:        req.ShippingCost,
		TotalCost:           totalCost,
		Currency:            order.Currency,
		FrozenAmount:        totalCost,
		FrozenTransactionID: freezeTx.ID,
		Status:              models.ShipmentRecordStatusPending,
	}
	if err := s.db.Table(shipmentRecordTable).Create(record).Error; err != nil {
		// 回滚冻结
		s.accountService.UnfreezePrepayment(ctx, relation.ShopOwnerID, totalCost, req.OrderSN, "发货失败回滚")
		return nil, fmt.Errorf("创建发货记录失败: %w", err)
	}

	// 9. 调用 Shopee 发货 API
	err = s.callShopeeShipOrder(ctx, req.ShopID, req.OrderSN, req.PickupInfo)
	if err != nil {
		// 发货失败，更新记录状态
		record.Status = models.ShipmentRecordStatusFailed
		record.Remark = err.Error()
		s.db.Table(shipmentRecordTable).Where("id = ?", record.ID).Updates(map[string]interface{}{
			"status": record.Status,
			"remark": record.Remark,
		})

		// 解冻预付款
		s.accountService.UnfreezePrepayment(ctx, relation.ShopOwnerID, totalCost, req.OrderSN, "发货失败回滚")
		return nil, fmt.Errorf("调用Shopee发货失败: %w", err)
	}

	// 10. 发货成功，更新记录
	now := time.Now()
	record.Status = models.ShipmentRecordStatusShipped
	record.ShippedAt = &now
	if err := s.db.Table(shipmentRecordTable).Where("id = ?", record.ID).Updates(map[string]interface{}{
		"status":     record.Status,
		"shipped_at": record.ShippedAt,
	}).Error; err != nil {
		return nil, fmt.Errorf("更新发货记录失败: %w", err)
	}

	// 11. 转入托管账户 (冻结金额转入托管，待结算时分账)
	err = s.accountService.TransferToEscrow(ctx, relation.ShopOwnerID, totalCost, req.OrderSN, fmt.Sprintf("发货托管-订单%s", req.OrderSN))
	if err != nil {
		// 托管失败不影响发货流程，记录日志即可
		record.Remark = fmt.Sprintf("托管转入失败: %s", err.Error())
		s.db.Table(shipmentRecordTable).Where("id = ?", record.ID).Update("remark", record.Remark)
	}

	// 12. 获取物流单号
	go s.fetchTrackingNumber(req.ShopID, req.OrderSN, record.ID)

	return record, nil
}

// checkOperatorPermission 检查运营权限
func (s *ShipmentService) checkOperatorPermission(operatorID int64, shopID uint64) (*models.ShopOperatorRelation, error) {
	var relation models.ShopOperatorRelation
	err := s.db.Where("operator_id = ? AND shop_id = ? AND status = ?", operatorID, shopID, models.RelationStatusActive).First(&relation).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("无权操作该店铺")
		}
		return nil, err
	}
	return &relation, nil
}

// callShopeeShipOrder 调用 Shopee 发货 API
func (s *ShipmentService) callShopeeShipOrder(ctx context.Context, shopID uint64, orderSN string, pickupInfo *PickupInfo) error {
	// 获取店铺信息
	var shop models.Shop
	if err := s.db.Where("shop_id = ?", shopID).First(&shop).Error; err != nil {
		return fmt.Errorf("店铺不存在")
	}

	// 获取 access token
	accessToken, err := s.getAccessToken(ctx, shopID)
	if err != nil {
		return fmt.Errorf("获取访问令牌失败: %w", err)
	}

	// 调用 Shopee API
	client := shopee.NewClient(shop.Region)
	
	trackingNo := ""
	if pickupInfo != nil && pickupInfo.TrackingNo != "" {
		trackingNo = pickupInfo.TrackingNo
	}

	_, err = client.ShipOrder(accessToken, shopID, orderSN, trackingNo)
	return err
}

// getAccessToken 获取访问令牌
func (s *ShipmentService) getAccessToken(ctx context.Context, shopID uint64) (string, error) {
	rdb := database.GetRedis()
	cacheKey := fmt.Sprintf("shop:token:%d", shopID)

	token, err := rdb.Get(ctx, cacheKey).Result()
	if err == nil && token != "" {
		return token, nil
	}

	var auth models.ShopAuthorization
	if err := s.db.Where("shop_id = ?", shopID).First(&auth).Error; err != nil {
		return "", fmt.Errorf("店铺未授权")
	}

	if time.Now().After(auth.ExpiresAt) {
		return "", fmt.Errorf("访问令牌已过期")
	}

	rdb.Set(ctx, cacheKey, auth.AccessToken, time.Until(auth.ExpiresAt))
	return auth.AccessToken, nil
}

// fetchTrackingNumber 异步获取物流单号 - 使用分表
func (s *ShipmentService) fetchTrackingNumber(shopID uint64, orderSN string, recordID uint64) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	time.Sleep(3 * time.Second) // 等待 Shopee 处理

	var shop models.Shop
	if err := s.db.Where("shop_id = ?", shopID).First(&shop).Error; err != nil {
		return
	}

	accessToken, err := s.getAccessToken(ctx, shopID)
	if err != nil {
		return
	}

	client := shopee.NewClient(shop.Region)
	resp, err := client.GetTrackingNumber(accessToken, shopID, orderSN)
	if err != nil {
		return
	}

	shipmentRecordTable := database.GetOrderShipmentRecordTableName(shopID)
	s.db.Table(shipmentRecordTable).Where("id = ?", recordID).Updates(map[string]interface{}{
		"tracking_number": resp.Response.TrackingNumber,
	})
}

// GetOperatorOrders 获取运营的待发货订单 - 使用分表
func (s *ShipmentService) GetOperatorOrders(ctx context.Context, operatorID int64, status string, page, pageSize int) ([]models.Order, int64, error) {
	// 获取运营关联的店铺
	var relations []models.ShopOperatorRelation
	if err := s.db.Where("operator_id = ? AND status = ?", operatorID, models.RelationStatusActive).Find(&relations).Error; err != nil {
		return nil, 0, err
	}

	if len(relations) == 0 {
		return []models.Order{}, 0, nil
	}

	// 按分表索引分组店铺
	shardShops := make(map[int][]uint64)
	for _, r := range relations {
		idx := database.GetShardIndex(r.ShopID)
		shardShops[idx] = append(shardShops[idx], r.ShopID)
	}

	var allOrders []models.Order
	var total int64

	orderStatus := status
	if orderStatus == "" {
		orderStatus = "READY_TO_SHIP"
	}

	// 查询每个分表
	for idx, sids := range shardShops {
		orderTable := fmt.Sprintf("orders_%d", idx)
		query := s.db.Table(orderTable).Where("shop_id IN ?", sids).Where("order_status = ?", orderStatus)

		var count int64
		query.Count(&count)
		total += count

		var orders []models.Order
		query.Order("create_time DESC").Find(&orders)
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

// GetShipmentRecords 获取发货记录 - 遍历所有分表
func (s *ShipmentService) GetShipmentRecords(ctx context.Context, operatorID int64, status int8, page, pageSize int) ([]models.OrderShipmentRecord, int64, error) {
	var allRecords []models.OrderShipmentRecord
	var total int64

	// 遍历所有分表
	for i := 0; i < database.ShardCount; i++ {
		shipmentRecordTable := fmt.Sprintf("order_shipment_records_%d", i)

		query := s.db.Table(shipmentRecordTable).Where("operator_id = ?", operatorID)
		if status >= 0 {
			query = query.Where("status = ?", status)
		}

		var count int64
		query.Count(&count)
		total += count

		var records []models.OrderShipmentRecord
		query.Order("created_at DESC").Find(&records)
		allRecords = append(allRecords, records...)
	}

	// 内存分页
	offset := (page - 1) * pageSize
	end := offset + pageSize
	if offset >= len(allRecords) {
		return []models.OrderShipmentRecord{}, total, nil
	}
	if end > len(allRecords) {
		end = len(allRecords)
	}

	return allRecords[offset:end], total, nil
}
