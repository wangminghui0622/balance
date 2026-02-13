package platform

import (
	"strconv"

	"balance/backend/internal/services"
	"balance/backend/internal/utils"

	"github.com/gin-gonic/gin"
)

// SettlementHandler 平台结算处理器
type SettlementHandler struct {
	settlementService *services.SettlementService
}

// NewSettlementHandler 创建平台结算处理器
func NewSettlementHandler() *SettlementHandler {
	return &SettlementHandler{
		settlementService: services.NewSettlementService(),
	}
}

// GetSettlements 获取结算记录列表
// GET /platform/settlements?shop_owner_id=1&operator_id=2&status=1&page=1&page_size=20
func (h *SettlementHandler) GetSettlements(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))

	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	// 平台可查看所有结算记录
	settlements, total, err := h.settlementService.GetSettlements(c.Request.Context(), 0, "platform", page, pageSize)
	if err != nil {
		utils.Error(c, 500, err.Error())
		return
	}

	utils.SuccessWithPage(c, settlements, total, page, pageSize)
}

// GetSettlementStats 获取结算统计
// GET /platform/settlements/stats
func (h *SettlementHandler) GetSettlementStats(c *gin.Context) {
	stats, err := h.settlementService.GetSettlementStats(c.Request.Context(), 0, "platform")
	if err != nil {
		utils.Error(c, 500, err.Error())
		return
	}

	utils.Success(c, stats)
}

// ProcessSettlement 手动触发结算处理
// POST /platform/settlements/process
func (h *SettlementHandler) ProcessSettlement(c *gin.Context) {
	count, err := h.settlementService.ProcessShopeeSettlement(c.Request.Context())
	if err != nil {
		utils.Error(c, 500, err.Error())
		return
	}

	utils.Success(c, gin.H{
		"processed_count": count,
		"message":         "结算处理完成",
	})
}

// GetPendingSettlements 获取待结算订单
// GET /platform/settlements/pending?page=1&page_size=20
func (h *SettlementHandler) GetPendingSettlements(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))

	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	records, total, err := h.settlementService.GetPendingSettlements(c.Request.Context(), page, pageSize)
	if err != nil {
		utils.Error(c, 500, err.Error())
		return
	}

	utils.SuccessWithPage(c, records, total, page, pageSize)
}
