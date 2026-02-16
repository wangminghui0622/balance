package handlers

import (
	"fmt"
	"net/http"

	"balance/backend/internal/consts"
	"balance/backend/internal/services"

	"github.com/gin-gonic/gin"
)

// WebhookHandler Webhook处理器
type WebhookHandler struct {
	webhookService *services.WebhookService
}

// NewWebhookHandler 创建Webhook处理器
func NewWebhookHandler() *WebhookHandler {
	return &WebhookHandler{
		webhookService: services.NewWebhookService(),
	}
}

// WebhookRequest 虾皮Webhook请求
type WebhookRequest struct {
	Code      int    `json:"code"`      // 事件类型
	ShopID    uint64 `json:"shop_id"`   // 店铺ID
	Timestamp int64  `json:"timestamp"` // 时间戳
	Data      any    `json:"data"`      // 事件数据（不同事件类型数据结构不同）
}

// OrderStatusData 订单状态更新数据
type OrderStatusData struct {
	OrderSN    string `json:"ordersn"`
	Status     string `json:"status"`
	UpdateTime int64  `json:"update_time"`
}

// TrackingData 物流追踪更新数据
type TrackingData struct {
	OrderSN         string `json:"ordersn"`
	TrackingNumber  string `json:"tracking_no"`
	LogisticsStatus string `json:"logistics_status"`
}

// HandleWebhook 处理Webhook推送
// @Summary 处理虾皮Webhook推送
// @Description 接收虾皮平台的Webhook推送通知
// @Tags Webhook
// @Accept json
// @Produce json
// @Param body body WebhookRequest true "Webhook请求"
// @Success 200 {object} map[string]interface{}
// @Router /webhook [post]
func (h *WebhookHandler) HandleWebhook(c *gin.Context) {
	fmt.Println("*********************************这里是webhook事件*************************************************")
	var req WebhookRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		// Webhook需要快速返回200，否则虾皮会重试
		c.JSON(http.StatusOK, gin.H{"code": 0, "message": "received"})
		return
	}

	// 异步处理，快速响应
	go h.processWebhook(req)

	c.JSON(http.StatusOK, gin.H{"code": 0, "message": "received"})
}

// processWebhook 异步处理Webhook
func (h *WebhookHandler) processWebhook(req WebhookRequest) {
	ctx, cancel := services.NewBackgroundContext()
	defer cancel()

	switch req.Code {
	case consts.WebhookShopAuth:
		// 店铺授权事件（通常不需要处理，授权走回调）
		h.webhookService.LogWebhook(ctx, req.ShopID, req.Code, "shop_auth", req.Data)

	case consts.WebhookOrderStatus:
		// 订单状态更新
		fmt.Println("****************订单状态变更:**************************",req.ShopID,"****************",req.Data)
		h.webhookService.HandleOrderStatusUpdate(ctx, req.ShopID, req.Data, req.Timestamp)

	case consts.WebhookTrackingUpdate:
		// 物流追踪更新
		h.webhookService.HandleTrackingUpdate(ctx, req.ShopID, req.Data, req.Timestamp)

	case consts.WebhookBuyerCancelOrder:
		// 买家/卖家取消订单
		h.webhookService.HandleOrderCancel(ctx, req.ShopID, req.Data, req.Timestamp)

	case consts.WebhookBannedItem:
		// 商品禁止销售
		h.webhookService.LogWebhook(ctx, req.ShopID, req.Code, "banned_item", req.Data)

	case consts.WebhookPromotionUpdate:
		// 促销更新
		h.webhookService.LogWebhook(ctx, req.ShopID, req.Code, "promotion_update", req.Data)

	case consts.WebhookReservedStock:
		// 保留订单
		h.webhookService.HandleReservedOrder(ctx, req.ShopID, req.Data)

	default:
		// 未知事件类型，记录日志
		h.webhookService.LogWebhook(ctx, req.ShopID, req.Code, "unknown", req.Data)
	}
}
