package shopower

import (
	"strconv"

	"balance/backend/internal/services/shopower"
	"balance/backend/internal/utils"

	"github.com/gin-gonic/gin"
)

// EscrowHandler 结算明细处理器
type EscrowHandler struct {
	escrowService *shopower.EscrowService
}

// NewEscrowHandler 创建结算明细处理器
func NewEscrowHandler() *EscrowHandler {
	return &EscrowHandler{
		escrowService: shopower.NewEscrowService(),
	}
}

// SyncEscrow 同步订单结算明细
// POST /escrows/:shop_id/:order_sn/sync
func (h *EscrowHandler) SyncEscrow(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		utils.Unauthorized(c, "未登录")
		return
	}
	adminID := userID.(int64)
	shopID, err := strconv.ParseInt(c.Param("shop_id"), 10, 64)
	if err != nil {
		utils.Error(c, 400, "无效的店铺ID")
		return
	}
	orderSN := c.Param("order_sn")

	escrow, err := h.escrowService.SyncOrderEscrow(c.Request.Context(), adminID, shopID, orderSN)
	if err != nil {
		utils.Error(c, 500, err.Error())
		return
	}

	utils.Success(c, escrow)
}

// GetEscrow 获取订单结算明细
// GET /escrows/:shop_id/:order_sn
func (h *EscrowHandler) GetEscrow(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		utils.Unauthorized(c, "未登录")
		return
	}
	adminID := userID.(int64)
	shopID, err := strconv.ParseInt(c.Param("shop_id"), 10, 64)
	if err != nil {
		utils.Error(c, 400, "无效的店铺ID")
		return
	}
	orderSN := c.Param("order_sn")

	// 获取是否包含商品明细
	withItems := c.Query("with_items") == "true"

	if withItems {
		escrow, items, err := h.escrowService.GetOrderEscrowWithItems(c.Request.Context(), adminID, shopID, orderSN)
		if err != nil {
			utils.Error(c, 404, err.Error())
			return
		}
		utils.Success(c, gin.H{
			"escrow": escrow,
			"items":  items,
		})
		return
	}

	escrow, err := h.escrowService.GetOrderEscrow(c.Request.Context(), adminID, shopID, orderSN)
	if err != nil {
		utils.Error(c, 404, err.Error())
		return
	}

	utils.Success(c, escrow)
}

// ListPendingEscrows 获取待同步结算明细的订单列表
// GET /escrows?shop_id=xxx&limit=100
func (h *EscrowHandler) ListPendingEscrows(c *gin.Context) {
	shopIDStr := c.Query("shop_id")
	if shopIDStr == "" {
		utils.Error(c, 400, "缺少店铺ID")
		return
	}
	shopID, err := strconv.ParseUint(shopIDStr, 10, 64)
	if err != nil {
		utils.Error(c, 400, "无效的店铺ID")
		return
	}

	limit := 100
	if limitStr := c.Query("limit"); limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 && l <= 500 {
			limit = l
		}
	}

	orderSNs, err := h.escrowService.ListPendingEscrows(c.Request.Context(), shopID, limit)
	if err != nil {
		utils.InternalError(c, err.Error())
		return
	}

	utils.Success(c, gin.H{
		"order_sns": orderSNs,
		"count":     len(orderSNs),
	})
}

// BatchSyncEscrows 批量同步结算明细
// POST /escrows/batch-sync
func (h *EscrowHandler) BatchSyncEscrows(c *gin.Context) {
	var req struct {
		ShopID   uint64   `json:"shop_id" binding:"required"`
		OrderSNs []string `json:"order_sns" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Error(c, 400, "参数错误")
		return
	}

	if len(req.OrderSNs) > 50 {
		utils.Error(c, 400, "单次最多同步50个订单")
		return
	}

	successCount, failCount, err := h.escrowService.BatchSyncEscrows(c.Request.Context(), req.ShopID, req.OrderSNs)
	if err != nil {
		utils.InternalError(c, err.Error())
		return
	}

	utils.Success(c, gin.H{
		"success_count": successCount,
		"fail_count":    failCount,
	})
}
