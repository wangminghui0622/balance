package operator

import (
	"strconv"

	"balance/backend/internal/consts"
	"balance/backend/internal/services/operator"
	"balance/backend/internal/utils"

	"github.com/gin-gonic/gin"
)

// OrderHandler 订单处理器（运营专用）
type OrderHandler struct {
	orderService *operator.OrderService
}

// NewOrderHandler 创建订单处理器
func NewOrderHandler() *OrderHandler {
	return &OrderHandler{
		orderService: operator.NewOrderService(),
	}
}

// ListOrders 获取订单列表（运营可查看所有订单）
// GET /api/v1/balance/admin/operator/orders
func (h *OrderHandler) ListOrders(c *gin.Context) {
	_, exists := c.Get("user_id")
	if !exists {
		utils.Unauthorized(c, "未登录")
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))
	shopIDStr := c.Query("shop_id")
	ownerIDStr := c.Query("owner_id")
	status := c.Query("status")
	startTime := c.Query("start_time")
	endTime := c.Query("end_time")

	if page < 1 {
		page = consts.DefaultPage
	}
	if pageSize < 1 || pageSize > consts.MaxPageSize {
		pageSize = consts.DefaultPageSize
	}

	var shopID, ownerID int64
	if shopIDStr != "" {
		shopID, _ = strconv.ParseInt(shopIDStr, 10, 64)
	}
	if ownerIDStr != "" {
		ownerID, _ = strconv.ParseInt(ownerIDStr, 10, 64)
	}

	// 运营可以查看所有订单
	list, total, err := h.orderService.ListOrders(c.Request.Context(), ownerID, shopID, status, startTime, endTime, page, pageSize)
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
// GET /api/v1/balance/admin/operator/orders/:shop_id/:order_sn
func (h *OrderHandler) GetOrder(c *gin.Context) {
	_, exists := c.Get("user_id")
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

	order, err := h.orderService.GetOrder(c.Request.Context(), shopID, orderSN)
	if err != nil {
		utils.Error(c, 400, err.Error())
		return
	}

	utils.Success(c, order)
}
