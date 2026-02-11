package shopower

import (
	"strconv"

	"balance/backend/internal/consts"
	"balance/backend/internal/services/shopower"
	"balance/backend/internal/utils"

	"github.com/gin-gonic/gin"
)

// 注意：service已移到 internal/services/shopower 目录

// ShopHandler 店铺处理器（店主专用）
type ShopHandler struct {
	shopService *shopower.ShopService
}

// NewShopHandler 创建店铺处理器
func NewShopHandler() *ShopHandler {
	return &ShopHandler{
		shopService: shopower.NewShopService(),
	}
}

// GetAuthURL 获取Shopee授权URL
// GET /api/v1/balance/admin/shopower/shops/auth-url
func (h *ShopHandler) GetAuthURL(c *gin.Context) {
	// 从上下文获取用户ID
	userID, exists := c.Get("user_id")
	if !exists {
		utils.Unauthorized(c, "未登录")
		return
	}

	url, err := h.shopService.GetAuthURL(c.Request.Context(), userID.(int64))
	if err != nil {
		utils.Error(c, 400, err.Error())
		return
	}

	utils.Success(c, gin.H{"auth_url": url})
}

// AuthCallback Shopee授权回调
// GET /api/v1/balance/admin/shopower/shops/callback
func (h *ShopHandler) AuthCallback(c *gin.Context) {
	code := c.Query("code")
	shopIDStr := c.Query("shop_id")
	stateStr := c.Query("state")

	if code == "" || shopIDStr == "" {
		utils.BadRequest(c, "缺少必要参数")
		return
	}

	shopID, err := strconv.ParseInt(shopIDStr, 10, 64)
	if err != nil {
		utils.BadRequest(c, "店铺ID格式错误")
		return
	}

	state, _ := strconv.ParseInt(stateStr, 10, 64)

	if err := h.shopService.HandleAuthCallback(c.Request.Context(), code, shopID, state); err != nil {
		utils.Error(c, 400, err.Error())
		return
	}

	// 授权成功，重定向到前端页面
	c.Redirect(302, "/admin/shopower/stores?auth=success")
}

// ListShops 获取店铺列表
// GET /api/v1/balance/admin/shopower/shops
func (h *ShopHandler) ListShops(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		utils.Unauthorized(c, "未登录")
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))

	if page < 1 {
		page = consts.DefaultPage
	}
	if pageSize < 1 || pageSize > consts.MaxPageSize {
		pageSize = consts.DefaultPageSize
	}

	list, total, err := h.shopService.ListShops(c.Request.Context(), userID.(int64), page, pageSize)
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

// BindShop 绑定店铺
// POST /api/v1/balance/admin/shopower/shops/bind
func (h *ShopHandler) BindShop(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		utils.Unauthorized(c, "未登录")
		return
	}

	var req struct {
		ShopID int64 `json:"shop_id" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequest(c, "参数错误: "+err.Error())
		return
	}

	if err := h.shopService.BindShop(c.Request.Context(), userID.(int64), req.ShopID); err != nil {
		utils.Error(c, 400, err.Error())
		return
	}

	utils.Success(c, nil)
}

// GetShop 获取店铺详情
// GET /api/v1/balance/admin/shopower/shops/:shop_id
func (h *ShopHandler) GetShop(c *gin.Context) {
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

	shop, err := h.shopService.GetShop(c.Request.Context(), userID.(int64), shopID)
	if err != nil {
		utils.Error(c, 400, err.Error())
		return
	}

	utils.Success(c, shop)
}

// UpdateShopStatus 更新店铺状态
// PUT /api/v1/balance/admin/shopower/shops/:shop_id/status
func (h *ShopHandler) UpdateShopStatus(c *gin.Context) {
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

	var req struct {
		Status int `json:"status" binding:"oneof=0 1"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequest(c, "参数错误: "+err.Error())
		return
	}

	if err := h.shopService.UpdateShopStatus(c.Request.Context(), userID.(int64), shopID, req.Status); err != nil {
		utils.Error(c, 400, err.Error())
		return
	}

	utils.Success(c, nil)
}

// DeleteShop 删除店铺
// DELETE /api/v1/balance/admin/shopower/shops/:shop_id
func (h *ShopHandler) DeleteShop(c *gin.Context) {
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

	if err := h.shopService.DeleteShop(c.Request.Context(), userID.(int64), shopID); err != nil {
		utils.Error(c, 400, err.Error())
		return
	}

	utils.Success(c, nil)
}

// RefreshToken 刷新店铺Token
// POST /api/v1/balance/admin/shopower/shops/:shop_id/refresh-token
func (h *ShopHandler) RefreshToken(c *gin.Context) {
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

	if err := h.shopService.RefreshShopToken(c.Request.Context(), userID.(int64), shopID); err != nil {
		utils.Error(c, 400, err.Error())
		return
	}

	utils.Success(c, nil)
}
