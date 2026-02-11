package handlers

import (
	"strconv"
	"time"

	"balance/backend/internal/services"

	"github.com/gin-gonic/gin"
)

// OrderHandler 订单处理器
type OrderHandler struct {
	orderService *services.OrderService
}

// NewOrderHandler 创建订单处理器
func NewOrderHandler() *OrderHandler {
	return &OrderHandler{
		orderService: services.NewOrderService(),
	}
}

// SyncOrdersRequest 同步订单请求
type SyncOrdersRequest struct {
	ShopID      uint64 `json:"shop_id" binding:"required"`
	TimeFrom    int64  `json:"time_from"`    // Unix时间戳，默认7天前
	TimeTo      int64  `json:"time_to"`      // Unix时间戳，默认当前
	OrderStatus string `json:"order_status"` // 可选，指定同步的订单状态
}

// SyncOrders 同步订单
// POST /api/v1/orders/sync
func (h *OrderHandler) SyncOrders(c *gin.Context) {
	var req SyncOrdersRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		BadRequest(c, "参数错误")
		return
	}

	// 默认时间范围：最近7天
	now := time.Now()
	timeFrom := now.AddDate(0, 0, -7)
	timeTo := now

	if req.TimeFrom > 0 {
		timeFrom = time.Unix(req.TimeFrom, 0)
	}
	if req.TimeTo > 0 {
		timeTo = time.Unix(req.TimeTo, 0)
	}

	if err := h.orderService.SyncOrders(c.Request.Context(), req.ShopID, timeFrom, timeTo, req.OrderStatus); err != nil {
		InternalError(c, err.Error())
		return
	}

	Success(c, gin.H{"message": "同步完成"})
}

// ListOrders 获取订单列表
// GET /api/v1/orders?shop_id=xxx&status=READY_TO_SHIP&page=1&page_size=10&order_sn=xxx&start_time=2024-01-01&end_time=2024-01-31
func (h *OrderHandler) ListOrders(c *gin.Context) {
	shopID, _ := strconv.ParseUint(c.Query("shop_id"), 10, 64)
	status := c.Query("status")
	orderSN := c.Query("order_sn")
	startTime := c.Query("start_time")
	endTime := c.Query("end_time")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))

	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 10
	}

	// 从上下文获取当前登录用户ID
	adminID, _ := c.Get("user_id")
	var userID int64 = 0
	if adminID != nil {
		userID = adminID.(int64)
	}

	params := services.OrderQueryParams{
		ShopID:    shopID,
		Status:    status,
		OrderSN:   orderSN,
		StartTime: startTime,
		EndTime:   endTime,
		Page:      page,
		PageSize:  pageSize,
		AdminID:   userID,
	}

	orders, total, err := h.orderService.ListOrders(c.Request.Context(), params)
	if err != nil {
		InternalError(c, err.Error())
		return
	}

	SuccessWithPage(c, orders, total, page, pageSize)
}

// GetOrder 获取订单详情
// GET /api/v1/orders/:shop_id/:order_sn
func (h *OrderHandler) GetOrder(c *gin.Context) {
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

	order, err := h.orderService.GetOrder(c.Request.Context(), shopID, orderSN)
	if err != nil {
		NotFound(c, "订单不存在")
		return
	}

	Success(c, order)
}

// GetReadyToShipOrders 获取待发货订单
// GET /api/v1/orders/ready-to-ship?shop_id=xxx&page=1&page_size=10
func (h *OrderHandler) GetReadyToShipOrders(c *gin.Context) {
	shopID, _ := strconv.ParseUint(c.Query("shop_id"), 10, 64)
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))

	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 10
	}

	// 从上下文获取当前登录用户ID
	adminID, _ := c.Get("user_id")
	var userID int64 = 0
	if adminID != nil {
		userID = adminID.(int64)
	}

	orders, total, err := h.orderService.GetReadyToShipOrders(c.Request.Context(), shopID, page, pageSize, userID)
	if err != nil {
		InternalError(c, err.Error())
		return
	}

	SuccessWithPage(c, orders, total, page, pageSize)
}

// RefreshOrder 刷新单个订单
// POST /api/v1/orders/:shop_id/:order_sn/refresh
func (h *OrderHandler) RefreshOrder(c *gin.Context) {
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

	if err := h.orderService.RefreshOrderFromAPI(c.Request.Context(), shopID, orderSN); err != nil {
		InternalError(c, err.Error())
		return
	}

	Success(c, gin.H{"message": "刷新成功"})
}

// ForceUpdateStatusRequest 强制更新状态请求
type ForceUpdateStatusRequest struct {
	Status string `json:"status" binding:"required"` // 目标状态
	Remark string `json:"remark"`                    // 操作原因
	Lock   bool   `json:"lock"`                      // 是否锁定状态（锁定后不再接受自动更新）
}

// ForceUpdateStatus 强制更新订单状态（人工操作）
// PUT /api/v1/orders/:shop_id/:order_sn/force-status
func (h *OrderHandler) ForceUpdateStatus(c *gin.Context) {
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

	var req ForceUpdateStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		BadRequest(c, "参数错误: "+err.Error())
		return
	}

	if err := h.orderService.ForceUpdateStatus(c.Request.Context(), shopID, orderSN, req.Status, req.Remark, req.Lock); err != nil {
		InternalError(c, err.Error())
		return
	}

	Success(c, gin.H{
		"message": "状态更新成功",
		"status":  req.Status,
		"locked":  req.Lock,
	})
}

// UnlockStatus 解锁订单状态
// DELETE /api/v1/orders/:shop_id/:order_sn/unlock
func (h *OrderHandler) UnlockStatus(c *gin.Context) {
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

	if err := h.orderService.UnlockStatus(c.Request.Context(), shopID, orderSN); err != nil {
		InternalError(c, err.Error())
		return
	}

	Success(c, gin.H{"message": "解锁成功，订单状态恢复自动更新"})
}
