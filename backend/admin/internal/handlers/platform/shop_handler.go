package platform

import (
	"strconv"

	"balance/backend/internal/consts"
	"balance/backend/internal/services/platform"
	"balance/backend/internal/utils"

	"github.com/gin-gonic/gin"
)

// ShopHandler 店铺管理处理器（平台专用）
type ShopHandler struct {
	shopService *platform.ShopService
}

// NewShopHandler 创建店铺管理处理器
func NewShopHandler() *ShopHandler {
	return &ShopHandler{
		shopService: platform.NewShopService(),
	}
}

// ListShops 获取店铺列表（平台可查看所有店铺）
// GET /api/v1/balance/admin/platform/shops
func (h *ShopHandler) ListShops(c *gin.Context) {
	_, exists := c.Get("user_id")
	if !exists {
		utils.Unauthorized(c, "未登录")
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))
	ownerIDStr := c.Query("owner_id")
	keyword := c.Query("keyword")
	status := c.Query("status")

	if page < 1 {
		page = consts.DefaultPage
	}
	if pageSize < 1 || pageSize > consts.MaxPageSize {
		pageSize = consts.DefaultPageSize
	}

	var ownerID int64
	if ownerIDStr != "" {
		ownerID, _ = strconv.ParseInt(ownerIDStr, 10, 64)
	}

	list, total, err := h.shopService.ListShops(c.Request.Context(), ownerID, keyword, status, page, pageSize)
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

// GetShop 获取店铺详情
// GET /api/v1/balance/admin/platform/shops/:shop_id
func (h *ShopHandler) GetShop(c *gin.Context) {
	_, exists := c.Get("user_id")
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

	shop, err := h.shopService.GetShop(c.Request.Context(), shopID)
	if err != nil {
		utils.Error(c, 400, err.Error())
		return
	}

	utils.Success(c, shop)
}

// UpdateShopStatus 更新店铺状态
// PUT /api/v1/balance/admin/platform/shops/:shop_id/status
func (h *ShopHandler) UpdateShopStatus(c *gin.Context) {
	_, exists := c.Get("user_id")
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

	var req struct {
		Status int `json:"status" binding:"oneof=0 1"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequest(c, "参数错误: "+err.Error())
		return
	}

	if err := h.shopService.UpdateShopStatus(c.Request.Context(), shopID, req.Status); err != nil {
		utils.Error(c, 400, err.Error())
		return
	}

	utils.Success(c, nil)
}
