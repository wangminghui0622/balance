package operator

import (
	"strconv"

	"balance/backend/internal/services"
	"balance/backend/internal/utils"

	"github.com/gin-gonic/gin"
)

// SettlementHandler 结算处理器
type SettlementHandler struct {
	settlementService *services.SettlementService
}

// NewSettlementHandler 创建结算处理器
func NewSettlementHandler() *SettlementHandler {
	return &SettlementHandler{
		settlementService: services.NewSettlementService(),
	}
}

// GetSettlements 获取结算记录
// GET /operator/settlements?page=1&page_size=20
func (h *SettlementHandler) GetSettlements(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		utils.Unauthorized(c, "未登录")
		return
	}
	operatorID := userID.(int64)

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))

	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	settlements, total, err := h.settlementService.GetSettlements(c.Request.Context(), operatorID, "operator", page, pageSize)
	if err != nil {
		utils.Error(c, 500, err.Error())
		return
	}

	utils.SuccessWithPage(c, settlements, total, page, pageSize)
}

// GetSettlementStats 获取结算统计
// GET /operator/settlements/stats
func (h *SettlementHandler) GetSettlementStats(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		utils.Unauthorized(c, "未登录")
		return
	}
	operatorID := userID.(int64)

	stats, err := h.settlementService.GetSettlementStats(c.Request.Context(), operatorID, "operator")
	if err != nil {
		utils.Error(c, 500, err.Error())
		return
	}

	utils.Success(c, stats)
}
