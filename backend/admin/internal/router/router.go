package router

import (
	"balance/backend/admin/internal/handlers"
	"balance/backend/internal/consts"

	"github.com/gin-gonic/gin"
)

// SetupRouter 设置路由
func SetupRouter(mode string) *gin.Engine {
	gin.SetMode(mode)
	r := gin.New()
	r.Use(gin.Logger())
	r.Use(gin.Recovery())

	// 健康检查
	r.GET(consts.RouteHealth, func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	// Webhook接收（放在API前缀外，方便虾皮配置）
	webhookHandler := handlers.NewWebhookHandler()
	r.POST(consts.RouteWebhook, webhookHandler.HandleWebhook)

	admin := r.Group(consts.AdminPrefix)
	{
		// 用户认证（无需登录）
		authHandler := handlers.NewAuthHandler()
		admin.POST(consts.RouteAuthRegister, authHandler.Register)
		admin.POST(consts.RouteAuthLogin, authHandler.Login)
		admin.POST(consts.RouteAuthSendCode, authHandler.SendEmailCode)

		// 授权回调（Shopee）
		admin.GET(consts.RouteAuthCallback, handlers.NewShopHandler().AuthCallback)

		// 需要登录的接口
		authGroup := admin.Group("")
		authGroup.Use(handlers.JWTAuthMiddleware())
		{
			// 获取当前用户信息
			authGroup.GET(consts.RouteAuthMe, authHandler.GetCurrentUser)

			// 店铺管理
			shopHandler := handlers.NewShopHandler()
			authGroup.GET(consts.RouteShopAuthURL, shopHandler.GetAuthURL)
			authGroup.GET(consts.RouteShops, shopHandler.ListShops)
			authGroup.POST(consts.RouteShopBind, shopHandler.BindShop)
			authGroup.GET(consts.RouteShopDetail, shopHandler.GetShop)
			authGroup.PUT(consts.RouteShopStatus, shopHandler.UpdateShopStatus)
			authGroup.DELETE(consts.RouteShopDetail, shopHandler.DeleteShop)
			authGroup.POST(consts.RouteShopRefreshToken, shopHandler.RefreshToken)

			// 订单管理
			orderHandler := handlers.NewOrderHandler()
			authGroup.POST(consts.RouteOrderSync, orderHandler.SyncOrders)
			authGroup.GET(consts.RouteOrders, orderHandler.ListOrders)
			authGroup.GET(consts.RouteOrderReadyToShip, orderHandler.GetReadyToShipOrders)
			authGroup.GET(consts.RouteOrderDetail, orderHandler.GetOrder)
			authGroup.POST(consts.RouteOrderRefresh, orderHandler.RefreshOrder)
			authGroup.PUT(consts.RouteOrderForceStatus, orderHandler.ForceUpdateStatus)
			authGroup.DELETE(consts.RouteOrderUnlockStatus, orderHandler.UnlockStatus)

			// 发货管理
			shipmentHandler := handlers.NewShipmentHandler()
			authGroup.POST(consts.RouteShipmentShip, shipmentHandler.ShipOrder)
			authGroup.POST(consts.RouteShipmentBatchShip, shipmentHandler.BatchShipOrders)
			authGroup.GET(consts.RouteShipmentParameter, shipmentHandler.GetShippingParameter)
			authGroup.GET(consts.RouteShipmentTrackingNo, shipmentHandler.GetTrackingNumber)
			authGroup.GET(consts.RouteShipments, shipmentHandler.ListShipments)
			authGroup.GET(consts.RouteShipmentDetail, shipmentHandler.GetShipment)
			authGroup.POST(consts.RouteShipmentSyncLogistics, shipmentHandler.SyncLogisticsChannels)
			authGroup.GET(consts.RouteShipmentLogistics, shipmentHandler.GetLogisticsChannels)
		}
	}

	return r
}
