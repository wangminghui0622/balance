package shopower

import (
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"balance/backend/internal/consts"
	"balance/backend/internal/services/shopower"
	"balance/backend/internal/utils"

	"github.com/gin-gonic/gin"
)

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
func (h *ShopHandler) GetAuthURL(c *gin.Context) {
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
func (h *ShopHandler) AuthCallback(c *gin.Context) {
	code := c.Query("code")
	shopIDStr := c.Query("shop_id")
	stateStr := c.Query("state")

	buildFrontendURL := func(success bool, shopID uint64, errorMsg string) string {
		scheme := "https"
		if c.Request.TLS == nil && c.Request.Header.Get("X-Forwarded-Proto") != "https" {
			scheme = "http"
		}
		host := c.Request.Host
		frontendPath := "/shopee/auth/callback"

		if strings.Contains(host, "localhost") || strings.Contains(host, "127.0.0.1") {
			host = strings.Replace(host, "19090", "3000", 1)
			scheme = "http"
		}

		if success {
			return fmt.Sprintf("%s://%s/balance/app%s?success=true&shop_id=%d", scheme, host, frontendPath, shopID)
		}
		return fmt.Sprintf("%s://%s/balance/app%s?success=false&error=%s", scheme, host, frontendPath, url.QueryEscape(errorMsg))
	}

	if code == "" || shopIDStr == "" {
		frontendURL := buildFrontendURL(false, 0, "缺少必要参数")
		c.Redirect(http.StatusFound, frontendURL)
		return
	}

	shopID, err := strconv.ParseInt(shopIDStr, 10, 64)
	if err != nil {
		frontendURL := buildFrontendURL(false, 0, "店铺ID格式错误")
		c.Redirect(http.StatusFound, frontendURL)
		return
	}

	state, _ := strconv.ParseInt(stateStr, 10, 64)

	if err := h.shopService.HandleAuthCallback(c.Request.Context(), code, shopID, state); err != nil {
		frontendURL := buildFrontendURL(false, uint64(shopID), err.Error())
		c.Redirect(http.StatusFound, frontendURL)
		return
	}

	frontendURL := buildFrontendURL(true, uint64(shopID), "")
	c.Redirect(http.StatusFound, frontendURL)
}

// ListShops 获取店铺列表
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
