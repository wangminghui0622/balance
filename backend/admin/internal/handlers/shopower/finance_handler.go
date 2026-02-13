package shopower

import (
	"strconv"

	"balance/backend/internal/services/shopower"
	"balance/backend/internal/utils"

	"github.com/gin-gonic/gin"
)

// FinanceHandler 财务收入处理器
type FinanceHandler struct {
	financeService *shopower.FinanceService
}

// NewFinanceHandler 创建财务收入处理器
func NewFinanceHandler() *FinanceHandler {
	return &FinanceHandler{
		financeService: shopower.NewFinanceService(),
	}
}

// SyncTransactions 同步钱包交易记录
// POST /finances/:shop_id/sync
func (h *FinanceHandler) SyncTransactions(c *gin.Context) {
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

	count, err := h.financeService.SyncWalletTransactions(c.Request.Context(), adminID, shopID)
	if err != nil {
		utils.InternalError(c, err.Error())
		return
	}

	utils.Success(c, gin.H{
		"synced_count": count,
	})
}

// ListIncomes 获取财务收入列表
// GET /finances?shop_id=xxx&order_sn=xxx&transaction_type=xxx&settlement_status=0&page=1&page_size=20
func (h *FinanceHandler) ListIncomes(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		utils.Unauthorized(c, "未登录")
		return
	}
	adminID := userID.(int64)
	
	shopIDStr := c.Query("shop_id")
	if shopIDStr == "" {
		utils.Error(c, 400, "缺少店铺ID")
		return
	}
	shopID, err := strconv.ParseInt(shopIDStr, 10, 64)
	if err != nil {
		utils.Error(c, 400, "无效的店铺ID")
		return
	}

	orderSN := c.Query("order_sn")
	transactionType := c.Query("transaction_type")
	
	settlementStatus := -1
	if statusStr := c.Query("settlement_status"); statusStr != "" {
		if s, err := strconv.Atoi(statusStr); err == nil {
			settlementStatus = s
		}
	}

	page := 1
	if p, err := strconv.Atoi(c.Query("page")); err == nil && p > 0 {
		page = p
	}
	pageSize := 20
	if ps, err := strconv.Atoi(c.Query("page_size")); err == nil && ps > 0 && ps <= 100 {
		pageSize = ps
	}

	incomes, total, err := h.financeService.ListFinanceIncomes(c.Request.Context(), adminID, shopID, orderSN, transactionType, settlementStatus, page, pageSize)
	if err != nil {
		utils.InternalError(c, err.Error())
		return
	}

	utils.SuccessWithPage(c, incomes, total, page, pageSize)
}

// GetIncomeStats 获取收入统计
// GET /finances/:shop_id/stats
func (h *FinanceHandler) GetIncomeStats(c *gin.Context) {
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

	stats, err := h.financeService.GetShopIncomeStats(c.Request.Context(), adminID, shopID)
	if err != nil {
		utils.InternalError(c, err.Error())
		return
	}

	utils.Success(c, stats)
}
