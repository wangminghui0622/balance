package handlers

import (
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"balance/backend/internal/services"

	"github.com/gin-gonic/gin"
)

// ShopHandler 店铺处理器
type ShopHandler struct {
	shopService *services.ShopService
}

// NewShopHandler 创建店铺处理器
func NewShopHandler() *ShopHandler {
	return &ShopHandler{
		shopService: services.NewShopService(),
	}
}

// GetAuthURL 获取授权链接
// GET /api/v1/shops/auth-url?region=SG
func (h *ShopHandler) GetAuthURL(c *gin.Context) {
	region := c.DefaultQuery("region", "SG")

	// 从上下文获取当前登录用户ID
	adminID, _ := c.Get("user_id")
	var userID int64 = 0
	if adminID != nil {
		userID = adminID.(int64)
	}

	url := h.shopService.GetAuthURL(region, userID)
	Success(c, gin.H{"url": url})
}

// AuthCallback 授权回调
// GET /api/v1/auth/callback?code=xxx&shop_id=xxx&state=xxx
// Shopee 授权完成后会回调此接口，处理完成后重定向到前端页面
func (h *ShopHandler) AuthCallback(c *gin.Context) {
	code := c.Query("code")
	shopIDStr := c.Query("shop_id")
	region := c.DefaultQuery("region", "SG")
	state := c.Query("state") // state参数包含adminID

	// 构建前端回调URL的辅助函数
	buildFrontendURL := func(success bool, shopID uint64, errorMsg string) string {
		scheme := "https"
		if c.Request.TLS == nil && c.Request.Header.Get("X-Forwarded-Proto") != "https" {
			scheme = "http"
		}
		host := c.Request.Host

		// 前端回调页面路径
		frontendPath := "/shopee/auth/callback"

		// 开发环境处理：如果是后端端口，替换为前端端口
		if strings.Contains(host, "localhost") || strings.Contains(host, "127.0.0.1") {
			// 开发环境，替换端口
			host = strings.Replace(host, "19090", "3000", 1)
			scheme = "http"
		}

		if success {
			return fmt.Sprintf("%s://%s/balance/admin%s?success=true&shop_id=%d", scheme, host, frontendPath, shopID)
		}
		return fmt.Sprintf("%s://%s/balance/admin%s?success=false&error=%s", scheme, host, frontendPath, url.QueryEscape(errorMsg))
	}

	if code == "" || shopIDStr == "" {
		frontendURL := buildFrontendURL(false, 0, "缺少必要参数")
		c.Redirect(http.StatusFound, frontendURL)
		return
	}

	shopID, err := strconv.ParseUint(shopIDStr, 10, 64)
	if err != nil {
		frontendURL := buildFrontendURL(false, 0, "店铺ID格式错误")
		c.Redirect(http.StatusFound, frontendURL)
		return
	}

	// 解析state参数中的adminID
	var adminID int64 = 0
	if state != "" {
		adminID, _ = strconv.ParseInt(state, 10, 64)
	}
	fmt.Printf("[DEBUG] AuthCallback: shop_id=%d, state=%s, adminID=%d\n", shopID, state, adminID)

	if err := h.shopService.HandleAuthCallback(c.Request.Context(), code, shopID, region, adminID); err != nil {
		fmt.Printf("[DEBUG] AuthCallback error: %v\n", err)
		frontendURL := buildFrontendURL(false, shopID, err.Error())
		c.Redirect(http.StatusFound, frontendURL)
		return
	}

	fmt.Printf("[DEBUG] AuthCallback success, redirecting to frontend\n")
	// 授权成功，重定向到前端成功页面
	frontendURL := buildFrontendURL(true, shopID, "")
	c.Redirect(http.StatusFound, frontendURL)
}

// ListShops 获取店铺列表
// GET /api/v1/shops?page=1&page_size=10&status=1
func (h *ShopHandler) ListShops(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))

	var status *int8
	if statusStr := c.Query("status"); statusStr != "" {
		s, _ := strconv.ParseInt(statusStr, 10, 8)
		st := int8(s)
		status = &st
	}

	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 10
	}

	// 从上下文获取当前登录用户ID，只返回该用户的店铺
	adminID, _ := c.Get("user_id")
	var userID int64 = 0
	if adminID != nil {
		userID = adminID.(int64)
	}

	shops, total, err := h.shopService.ListShops(c.Request.Context(), page, pageSize, status, userID)
	if err != nil {
		InternalError(c, err.Error())
		return
	}

	SuccessWithPage(c, shops, total, page, pageSize)
}

// GetShop 获取店铺详情
// GET /api/v1/shops/:shop_id
func (h *ShopHandler) GetShop(c *gin.Context) {
	shopID, err := strconv.ParseUint(c.Param("shop_id"), 10, 64)
	if err != nil {
		BadRequest(c, "店铺ID格式错误")
		return
	}

	shop, err := h.shopService.GetShop(c.Request.Context(), shopID)
	if err != nil {
		NotFound(c, "店铺不存在")
		return
	}

	Success(c, shop)
}

// UpdateShopStatusRequest 更新店铺状态请求
type UpdateShopStatusRequest struct {
	Status int8 `json:"status" binding:"required,oneof=0 1"`
}

// UpdateShopStatus 更新店铺状态
// PUT /api/v1/shops/:shop_id/status
func (h *ShopHandler) UpdateShopStatus(c *gin.Context) {
	shopID, err := strconv.ParseUint(c.Param("shop_id"), 10, 64)
	if err != nil {
		BadRequest(c, "店铺ID格式错误")
		return
	}

	var req UpdateShopStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		BadRequest(c, "参数错误")
		return
	}

	if err := h.shopService.UpdateShopStatus(c.Request.Context(), shopID, req.Status); err != nil {
		InternalError(c, err.Error())
		return
	}

	Success(c, nil)
}

// DeleteShop 删除店铺
// DELETE /api/v1/shops/:shop_id
func (h *ShopHandler) DeleteShop(c *gin.Context) {
	shopID, err := strconv.ParseUint(c.Param("shop_id"), 10, 64)
	if err != nil {
		BadRequest(c, "店铺ID格式错误")
		return
	}

	if err := h.shopService.DeleteShop(c.Request.Context(), shopID); err != nil {
		InternalError(c, err.Error())
		return
	}

	Success(c, nil)
}

// RefreshToken 刷新Token
// POST /api/v1/shops/:shop_id/refresh-token
func (h *ShopHandler) RefreshToken(c *gin.Context) {
	shopID, err := strconv.ParseUint(c.Param("shop_id"), 10, 64)
	if err != nil {
		BadRequest(c, "店铺ID格式错误")
		return
	}

	if err := h.shopService.RefreshToken(c.Request.Context(), shopID); err != nil {
		InternalError(c, err.Error())
		return
	}

	Success(c, gin.H{"message": "Token刷新成功"})
}

// BindShopRequest 绑定店铺请求
type BindShopRequest struct {
	ShopID uint64 `json:"shop_id" binding:"required"`
}

// BindShop 将店铺绑定到当前用户
// POST /api/v1/shops/bind
func (h *ShopHandler) BindShop(c *gin.Context) {
	var req BindShopRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		fmt.Printf("[DEBUG] BindShop: 参数错误 - %v\n", err)
		BadRequest(c, "参数错误: "+err.Error())
		return
	}

	// 从上下文获取当前登录用户ID
	adminID, exists := c.Get("user_id")
	if !exists {
		fmt.Printf("[DEBUG] BindShop: 未登录，user_id不存在\n")
		Unauthorized(c, "未登录")
		return
	}

	fmt.Printf("[DEBUG] BindShop: shop_id=%d, adminID=%v\n", req.ShopID, adminID)

	if err := h.shopService.BindShopToAdmin(c.Request.Context(), req.ShopID, adminID.(int64)); err != nil {
		fmt.Printf("[DEBUG] BindShop error: %v\n", err)
		Error(c, 400, err.Error())
		return
	}

	fmt.Printf("[DEBUG] BindShop success\n")
	Success(c, gin.H{"message": "绑定成功"})
}
