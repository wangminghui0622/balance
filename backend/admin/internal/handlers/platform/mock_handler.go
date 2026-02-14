package platform

import (
	"net/http"
	"strconv"

	"balance/backend/internal/config"
	"balance/backend/internal/services"
	"balance/backend/internal/utils"

	"github.com/gin-gonic/gin"
)

// MockHandler 模拟数据处理器（仅沙箱环境可用）
type MockHandler struct {
	generator *services.MockDataGenerator
	cfg       *config.Config
}

// NewMockHandler 创建模拟数据处理器
func NewMockHandler(cfg *config.Config) *MockHandler {
	return &MockHandler{
		generator: services.NewMockDataGenerator(),
		cfg:       cfg,
	}
}

// GenerateMockOrders 生成模拟订单数据
func (h *MockHandler) GenerateMockOrders(c *gin.Context) {
	// 检查是否为沙箱环境
	if h.cfg.Shopee.IsProduction {
		utils.Error(c, http.StatusForbidden, "此接口仅在沙箱环境可用")
		return
	}

	var req services.MockOrderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Error(c, http.StatusBadRequest, "参数错误: "+err.Error())
		return
	}

	if req.ShopID == 0 {
		utils.Error(c, http.StatusBadRequest, "shop_id不能为空")
		return
	}

	result, err := h.generator.GenerateMockOrders(c.Request.Context(), &req)
	if err != nil {
		utils.Error(c, http.StatusInternalServerError, "生成模拟数据失败: "+err.Error())
		return
	}

	utils.Success(c, result)
}

// CleanMockData 清理模拟数据
func (h *MockHandler) CleanMockData(c *gin.Context) {
	// 检查是否为沙箱环境
	if h.cfg.Shopee.IsProduction {
		utils.Error(c, http.StatusForbidden, "此接口仅在沙箱环境可用")
		return
	}

	shopIDStr := c.Query("shop_id")
	shopID, err := strconv.ParseUint(shopIDStr, 10, 64)
	if err != nil || shopID == 0 {
		utils.Error(c, http.StatusBadRequest, "shop_id不能为空")
		return
	}

	deleted, err := h.generator.CleanMockData(c.Request.Context(), shopID)
	if err != nil {
		utils.Error(c, http.StatusInternalServerError, "清理模拟数据失败: "+err.Error())
		return
	}

	utils.Success(c, gin.H{
		"deleted_count": deleted,
	})
}
