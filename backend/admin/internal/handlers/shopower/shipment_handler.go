package shopower

import (
	"strconv"

	"balance/backend/internal/consts"
	"balance/backend/internal/services/shopower"
	"balance/backend/internal/utils"

	"github.com/gin-gonic/gin"
)

// ShipmentHandler 发货处理器（店主专用）
type ShipmentHandler struct {
	shipmentService *shopower.ShipmentService
}

// NewShipmentHandler 创建发货处理器
func NewShipmentHandler() *ShipmentHandler {
	return &ShipmentHandler{
		shipmentService: shopower.NewShipmentService(),
	}
}

// ShipOrder 发货
// POST /api/v1/balance/admin/shopower/shipments/ship
func (h *ShipmentHandler) ShipOrder(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		utils.Unauthorized(c, "未登录")
		return
	}

	var req shopower.ShipOrderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequest(c, "参数错误: "+err.Error())
		return
	}

	if err := h.shipmentService.ShipOrder(c.Request.Context(), userID.(int64), &req); err != nil {
		utils.Error(c, 400, err.Error())
		return
	}

	utils.Success(c, nil)
}

// BatchShipOrders 批量发货
// POST /api/v1/balance/admin/shopower/shipments/batch-ship
func (h *ShipmentHandler) BatchShipOrders(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		utils.Unauthorized(c, "未登录")
		return
	}

	var req struct {
		Orders []shopower.ShipOrderRequest `json:"orders" binding:"required,min=1"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequest(c, "参数错误: "+err.Error())
		return
	}

	results, err := h.shipmentService.BatchShipOrders(c.Request.Context(), userID.(int64), req.Orders)
	if err != nil {
		utils.Error(c, 400, err.Error())
		return
	}

	utils.Success(c, gin.H{"results": results})
}

// GetShippingParameter 获取发货参数
// GET /api/v1/balance/admin/shopower/shipments/shipping-parameter
func (h *ShipmentHandler) GetShippingParameter(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		utils.Unauthorized(c, "未登录")
		return
	}

	shopIDStr := c.Query("shop_id")
	orderSN := c.Query("order_sn")

	shopID, err := strconv.ParseInt(shopIDStr, 10, 64)
	if err != nil {
		utils.BadRequest(c, "店铺ID格式错误")
		return
	}

	param, err := h.shipmentService.GetShippingParameter(c.Request.Context(), userID.(int64), shopID, orderSN)
	if err != nil {
		utils.Error(c, 400, err.Error())
		return
	}

	utils.Success(c, param)
}

// GetTrackingNumber 获取物流单号
// GET /api/v1/balance/admin/shopower/shipments/tracking-number
func (h *ShipmentHandler) GetTrackingNumber(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		utils.Unauthorized(c, "未登录")
		return
	}

	shopIDStr := c.Query("shop_id")
	orderSN := c.Query("order_sn")

	shopID, err := strconv.ParseInt(shopIDStr, 10, 64)
	if err != nil {
		utils.BadRequest(c, "店铺ID格式错误")
		return
	}

	trackingNo, err := h.shipmentService.GetTrackingNumber(c.Request.Context(), userID.(int64), shopID, orderSN)
	if err != nil {
		utils.Error(c, 400, err.Error())
		return
	}

	utils.Success(c, gin.H{"tracking_number": trackingNo})
}

// ListShipments 获取发货记录列表
// GET /api/v1/balance/admin/shopower/shipments
func (h *ShipmentHandler) ListShipments(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		utils.Unauthorized(c, "未登录")
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))
	shopIDStr := c.Query("shop_id")

	if page < 1 {
		page = consts.DefaultPage
	}
	if pageSize < 1 || pageSize > consts.MaxPageSize {
		pageSize = consts.DefaultPageSize
	}

	var shopID int64
	if shopIDStr != "" {
		shopID, _ = strconv.ParseInt(shopIDStr, 10, 64)
	}

	list, total, err := h.shipmentService.ListShipments(c.Request.Context(), userID.(int64), shopID, page, pageSize)
	if err != nil {
		utils.InternalError(c, err.Error())
		return
	}

	utils.Success(c, gin.H{
		"list":  list,
		"total": total,
		"page":  page,
		"size":  pageSize,
	})
}

// GetShipment 获取发货详情
// GET /api/v1/balance/admin/shopower/shipments/:shop_id/:order_sn
func (h *ShipmentHandler) GetShipment(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		utils.Unauthorized(c, "未登录")
		return
	}

	shopIDStr := c.Param("shop_id")
	orderSN := c.Param("order_sn")

	shopID, err := strconv.ParseInt(shopIDStr, 10, 64)
	if err != nil {
		utils.BadRequest(c, "店铺ID格式错误")
		return
	}

	shipment, err := h.shipmentService.GetShipment(c.Request.Context(), userID.(int64), shopID, orderSN)
	if err != nil {
		utils.Error(c, 400, err.Error())
		return
	}

	utils.Success(c, shipment)
}

// SyncLogisticsChannels 同步物流渠道
// POST /api/v1/balance/admin/shopower/shipments/sync-logistics/:shop_id
func (h *ShipmentHandler) SyncLogisticsChannels(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		utils.Unauthorized(c, "未登录")
		return
	}

	shopIDStr := c.Param("shop_id")
	shopID, err := strconv.ParseInt(shopIDStr, 10, 64)
	if err != nil {
		utils.BadRequest(c, "店铺ID格式错误")
		return
	}

	if err := h.shipmentService.SyncLogisticsChannels(c.Request.Context(), userID.(int64), shopID); err != nil {
		utils.Error(c, 400, err.Error())
		return
	}

	utils.Success(c, nil)
}

// GetLogisticsChannels 获取物流渠道列表
// GET /api/v1/balance/admin/shopower/shipments/logistics/:shop_id
func (h *ShipmentHandler) GetLogisticsChannels(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		utils.Unauthorized(c, "未登录")
		return
	}

	shopIDStr := c.Param("shop_id")
	shopID, err := strconv.ParseInt(shopIDStr, 10, 64)
	if err != nil {
		utils.BadRequest(c, "店铺ID格式错误")
		return
	}

	channels, err := h.shipmentService.GetLogisticsChannels(c.Request.Context(), userID.(int64), shopID)
	if err != nil {
		utils.Error(c, 400, err.Error())
		return
	}

	utils.Success(c, gin.H{"channels": channels})
}
