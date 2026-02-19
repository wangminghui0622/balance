package shopower

import (
	"strconv"
	"time"

	"balance/backend/internal/consts"
	"balance/backend/internal/services"
	"balance/backend/internal/services/shopower"
	"balance/backend/internal/utils"

	"github.com/gin-gonic/gin"
)

// OrderHandler 订单处理器（店主专用）
type OrderHandler struct {
	orderService *shopower.OrderService
}

// NewOrderHandler 创建订单处理器
func NewOrderHandler() *OrderHandler {
	return &OrderHandler{
		orderService: services.NewOrderServiceWithPrepaymentCheck(),
	}
}

// SyncOrders 同步订单
// POST /api/v1/balance/admin/shopower/orders/sync
func (h *OrderHandler) SyncOrders(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		utils.Unauthorized(c, "未登录")
		return
	}

	var req struct {
		ShopID    int64  `json:"shop_id" binding:"required"`
		StartTime string `json:"start_time"`
		EndTime   string `json:"end_time"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequest(c, "参数错误: "+err.Error())
		return
	}

	var startTime, endTime time.Time
	var err error

	if req.StartTime != "" {
		startTime, err = time.Parse("2006-01-02 15:04:05", req.StartTime)
		if err != nil {
			utils.BadRequest(c, "开始时间格式错误")
			return
		}
	} else {
		startTime = time.Now().AddDate(0, 0, -7)
	}

	if req.EndTime != "" {
		endTime, err = time.Parse("2006-01-02 15:04:05", req.EndTime)
		if err != nil {
			utils.BadRequest(c, "结束时间格式错误")
			return
		}
	} else {
		endTime = time.Now()
	}

	count, err := h.orderService.SyncOrders(c.Request.Context(), userID.(int64), req.ShopID, startTime, endTime)
	if err != nil {
		utils.Error(c, 400, err.Error())
		return
	}

	utils.Success(c, gin.H{"synced_count": count})
}

// ListOrders 获取订单列表
// GET /api/v1/balance/admin/shopower/orders
// 不传 shop_id、不传 status 时即「全部订单」：该 admin 绑定的所有 shop 在 orders_x 下的所有状态订单
func (h *OrderHandler) ListOrders(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		utils.Unauthorized(c, "未登录")
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))
	shopIDStr := c.Query("shop_id")
	status := c.Query("status")
	startTime := c.Query("start_time")
	endTime := c.Query("end_time")

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

	list, total, err := h.orderService.ListOrders(c.Request.Context(), userID.(int64), shopID, status, startTime, endTime, page, pageSize)
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

// GetOrderStats 获取订单统计（四张卡片：全部/未结算/已结算/账款调整，从缓存读，1小时更新）
// GET /api/v1/balance/admin/shopower/orders/stats
func (h *OrderHandler) GetOrderStats(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		utils.Unauthorized(c, "未登录")
		return
	}
	stats, err := h.orderService.GetOrderStatsCached(c.Request.Context(), userID.(int64))
	if err != nil {
		utils.InternalError(c, err.Error())
		return
	}
	utils.Success(c, stats)
}

// GetReadyToShipOrders 获取待发货订单
// GET /api/v1/balance/admin/shopower/orders/ready-to-ship
func (h *OrderHandler) GetReadyToShipOrders(c *gin.Context) {
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

	list, total, err := h.orderService.ListOrders(c.Request.Context(), userID.(int64), shopID, consts.OrderStatusReadyToShip, "", "", page, pageSize)
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

// GetOrder 获取订单详情
// GET /api/v1/balance/admin/shopower/orders/:shop_id/:order_sn
func (h *OrderHandler) GetOrder(c *gin.Context) {
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

	order, err := h.orderService.GetOrder(c.Request.Context(), userID.(int64), shopID, orderSN)
	if err != nil {
		utils.Error(c, 400, err.Error())
		return
	}

	utils.Success(c, order)
}

// RefreshOrder 刷新订单
// POST /api/v1/balance/admin/shopower/orders/:shop_id/:order_sn/refresh
func (h *OrderHandler) RefreshOrder(c *gin.Context) {
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

	order, err := h.orderService.RefreshOrder(c.Request.Context(), userID.(int64), shopID, orderSN)
	if err != nil {
		utils.Error(c, 400, err.Error())
		return
	}

	utils.Success(c, order)
}

// ForceUpdateStatus 强制更新订单状态
// PUT /api/v1/balance/admin/shopower/orders/:shop_id/:order_sn/force-status
func (h *OrderHandler) ForceUpdateStatus(c *gin.Context) {
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

	var req struct {
		Status string `json:"status" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequest(c, "参数错误: "+err.Error())
		return
	}

	if err := h.orderService.ForceUpdateStatus(c.Request.Context(), userID.(int64), shopID, orderSN, req.Status); err != nil {
		utils.Error(c, 400, err.Error())
		return
	}

	utils.Success(c, nil)
}

// UnlockStatus 解锁订单状态
// DELETE /api/v1/balance/admin/shopower/orders/:shop_id/:order_sn/unlock
func (h *OrderHandler) UnlockStatus(c *gin.Context) {
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

	if err := h.orderService.UnlockOrderStatus(c.Request.Context(), userID.(int64), shopID, orderSN); err != nil {
		utils.Error(c, 400, err.Error())
		return
	}

	utils.Success(c, nil)
}
