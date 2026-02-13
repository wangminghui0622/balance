package operator

import (
	"strconv"

	"balance/backend/internal/services/operator"
	"balance/backend/internal/utils"

	"github.com/gin-gonic/gin"
	"github.com/shopspring/decimal"
)

// ShipmentHandler 运营发货处理器
type ShipmentHandler struct {
	shipmentService *operator.ShipmentService
}

// NewShipmentHandler 创建运营发货处理器
func NewShipmentHandler() *ShipmentHandler {
	return &ShipmentHandler{
		shipmentService: operator.NewShipmentService(),
	}
}

// ShipOrderRequest 发货请求
type ShipOrderRequest struct {
	ShopID       uint64  `json:"shop_id" binding:"required"`
	OrderSN      string  `json:"order_sn" binding:"required"`
	GoodsCost    float64 `json:"goods_cost" binding:"required"`
	ShippingCost float64 `json:"shipping_cost"`
	PickupInfo   *struct {
		AddressID    uint64 `json:"address_id"`
		PickupTimeID string `json:"pickup_time_id"`
		TrackingNo   string `json:"tracking_no"`
	} `json:"pickup_info"`
}

// ShipOrder 运营发货
// POST /operator/shipments/ship
func (h *ShipmentHandler) ShipOrder(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		utils.Unauthorized(c, "未登录")
		return
	}
	operatorID := userID.(int64)

	var req ShipOrderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Error(c, 400, "参数错误: "+err.Error())
		return
	}

	shipReq := &operator.ShipOrderRequest{
		ShopID:       req.ShopID,
		OrderSN:      req.OrderSN,
		GoodsCost:    decimal.NewFromFloat(req.GoodsCost),
		ShippingCost: decimal.NewFromFloat(req.ShippingCost),
	}
	if req.PickupInfo != nil {
		shipReq.PickupInfo = &operator.PickupInfo{
			AddressID:    req.PickupInfo.AddressID,
			PickupTimeID: req.PickupInfo.PickupTimeID,
			TrackingNo:   req.PickupInfo.TrackingNo,
		}
	}

	record, err := h.shipmentService.ShipOrder(c.Request.Context(), operatorID, shipReq)
	if err != nil {
		utils.Error(c, 500, err.Error())
		return
	}

	utils.Success(c, record)
}

// GetPendingOrders 获取待发货订单
// GET /operator/orders/pending?status=READY_TO_SHIP&page=1&page_size=20
func (h *ShipmentHandler) GetPendingOrders(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		utils.Unauthorized(c, "未登录")
		return
	}
	operatorID := userID.(int64)

	status := c.DefaultQuery("status", "READY_TO_SHIP")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))

	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	orders, total, err := h.shipmentService.GetOperatorOrders(c.Request.Context(), operatorID, status, page, pageSize)
	if err != nil {
		utils.Error(c, 500, err.Error())
		return
	}

	utils.SuccessWithPage(c, orders, total, page, pageSize)
}

// GetShipmentRecords 获取发货记录
// GET /operator/shipments?status=1&page=1&page_size=20
func (h *ShipmentHandler) GetShipmentRecords(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		utils.Unauthorized(c, "未登录")
		return
	}
	operatorID := userID.(int64)

	statusStr := c.DefaultQuery("status", "-1")
	status, _ := strconv.ParseInt(statusStr, 10, 8)
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))

	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	records, total, err := h.shipmentService.GetShipmentRecords(c.Request.Context(), operatorID, int8(status), page, pageSize)
	if err != nil {
		utils.Error(c, 500, err.Error())
		return
	}

	utils.SuccessWithPage(c, records, total, page, pageSize)
}
