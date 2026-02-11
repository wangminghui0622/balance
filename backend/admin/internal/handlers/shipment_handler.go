package handlers

import (
	"strconv"

	"balance/backend/internal/services"

	"github.com/gin-gonic/gin"
)

// ShipmentHandler 发货处理器
type ShipmentHandler struct {
	shipmentService *services.ShipmentService
}

// NewShipmentHandler 创建发货处理器
func NewShipmentHandler() *ShipmentHandler {
	return &ShipmentHandler{
		shipmentService: services.NewShipmentService(),
	}
}

// ShipOrder 订单发货
// POST /api/v1/shipments/ship
func (h *ShipmentHandler) ShipOrder(c *gin.Context) {
	var req services.ShipOrderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		BadRequest(c, "参数错误: "+err.Error())
		return
	}

	if err := h.shipmentService.ShipOrder(c.Request.Context(), &req); err != nil {
		InternalError(c, err.Error())
		return
	}

	Success(c, gin.H{"message": "发货成功"})
}

// BatchShipRequest 批量发货请求
type BatchShipRequest struct {
	Orders []*services.ShipOrderRequest `json:"orders" binding:"required,min=1,max=50"`
}

// BatchShipOrders 批量发货
// POST /api/v1/shipments/batch-ship
func (h *ShipmentHandler) BatchShipOrders(c *gin.Context) {
	var req BatchShipRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		BadRequest(c, "参数错误: "+err.Error())
		return
	}

	results := h.shipmentService.BatchShipOrders(c.Request.Context(), req.Orders)
	Success(c, results)
}

// GetShippingParameter 获取发货参数
// GET /api/v1/shipments/shipping-parameter?shop_id=xxx&order_sn=xxx
func (h *ShipmentHandler) GetShippingParameter(c *gin.Context) {
	shopID, err := strconv.ParseUint(c.Query("shop_id"), 10, 64)
	if err != nil {
		BadRequest(c, "店铺ID格式错误")
		return
	}

	orderSN := c.Query("order_sn")
	if orderSN == "" {
		BadRequest(c, "订单号不能为空")
		return
	}

	result, err := h.shipmentService.GetShippingParameter(c.Request.Context(), shopID, orderSN)
	if err != nil {
		InternalError(c, err.Error())
		return
	}

	Success(c, result.Response)
}

// GetTrackingNumber 获取运单号
// GET /api/v1/shipments/tracking-number?shop_id=xxx&order_sn=xxx
func (h *ShipmentHandler) GetTrackingNumber(c *gin.Context) {
	shopID, err := strconv.ParseUint(c.Query("shop_id"), 10, 64)
	if err != nil {
		BadRequest(c, "店铺ID格式错误")
		return
	}

	orderSN := c.Query("order_sn")
	if orderSN == "" {
		BadRequest(c, "订单号不能为空")
		return
	}

	result, err := h.shipmentService.GetTrackingNumber(c.Request.Context(), shopID, orderSN)
	if err != nil {
		InternalError(c, err.Error())
		return
	}

	Success(c, result.Response)
}

// GetShipment 获取发货记录
// GET /api/v1/shipments/:shop_id/:order_sn
func (h *ShipmentHandler) GetShipment(c *gin.Context) {
	shopID, err := strconv.ParseUint(c.Param("shop_id"), 10, 64)
	if err != nil {
		BadRequest(c, "店铺ID格式错误")
		return
	}

	orderSN := c.Param("order_sn")
	if orderSN == "" {
		BadRequest(c, "订单号不能为空")
		return
	}

	shipment, err := h.shipmentService.GetShipment(c.Request.Context(), shopID, orderSN)
	if err != nil {
		NotFound(c, "发货记录不存在")
		return
	}

	Success(c, shipment)
}

// ListShipments 获取发货记录列表
// GET /api/v1/shipments?shop_id=xxx&status=1&page=1&page_size=10
func (h *ShipmentHandler) ListShipments(c *gin.Context) {
	shopID, _ := strconv.ParseUint(c.Query("shop_id"), 10, 64)
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))

	var status *int8
	if statusStr := c.Query("status"); statusStr != "" {
		s, _ := strconv.ParseInt(statusStr, 10, 8)
		st := int8(s)
		status = &st
	}

	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 10
	}

	shipments, total, err := h.shipmentService.ListShipments(c.Request.Context(), shopID, status, page, pageSize)
	if err != nil {
		InternalError(c, err.Error())
		return
	}

	SuccessWithPage(c, shipments, total, page, pageSize)
}

// SyncLogisticsChannels 同步物流渠道
// POST /api/v1/shipments/sync-logistics/:shop_id
func (h *ShipmentHandler) SyncLogisticsChannels(c *gin.Context) {
	shopID, err := strconv.ParseUint(c.Param("shop_id"), 10, 64)
	if err != nil {
		BadRequest(c, "店铺ID格式错误")
		return
	}

	if err := h.shipmentService.SyncLogisticsChannels(c.Request.Context(), shopID); err != nil {
		InternalError(c, err.Error())
		return
	}

	Success(c, gin.H{"message": "同步成功"})
}

// GetLogisticsChannels 获取物流渠道列表
// GET /api/v1/shipments/logistics/:shop_id
func (h *ShipmentHandler) GetLogisticsChannels(c *gin.Context) {
	shopID, err := strconv.ParseUint(c.Param("shop_id"), 10, 64)
	if err != nil {
		BadRequest(c, "店铺ID格式错误")
		return
	}

	channels, err := h.shipmentService.GetLogisticsChannels(c.Request.Context(), shopID)
	if err != nil {
		InternalError(c, err.Error())
		return
	}

	Success(c, channels)
}
