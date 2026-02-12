package router

import (
	"balance/backend/admin/internal/handlers"
	"balance/backend/admin/internal/handlers/operator"
	"balance/backend/admin/internal/handlers/platform"
	"balance/backend/admin/internal/handlers/shopower"
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
		// ==================== 公共认证路由（无需登录） ====================
		authHandler := handlers.NewAuthHandler()
		admin.POST(consts.RouteAuthRegister, authHandler.Register)
		admin.POST(consts.RouteAuthLogin, authHandler.Login)
		admin.POST(consts.RouteAuthSendCode, authHandler.SendEmailCode)
		admin.POST(consts.RouteAuthResetPassword, authHandler.ResetPassword)

		// Shopee授权回调（无需登录，放在公共路由下兼容虾皮配置的回调URL）
		shopowerShopHandler := shopower.NewShopHandler()
		admin.GET(consts.RouteAuthCallback, shopowerShopHandler.AuthCallback)

		// 需要登录的公共接口
		authGroup := admin.Group("")
		authGroup.Use(handlers.JWTAuthMiddleware())
		{
			authGroup.GET(consts.RouteAuthMe, authHandler.GetCurrentUser)
		}

		// ==================== 店主路由 (shopower) ====================
		shopowerGroup := admin.Group(consts.ShopowerPrefix)
		{
			// Shopee授权回调（无需登录）- 也在 shopower 前缀下注册一份
			shopowerGroup.GET(consts.RouteShopowerShopCallback, shopowerShopHandler.AuthCallback)

			// 需要登录的店主接口
			shopowerAuth := shopowerGroup.Group("")
			shopowerAuth.Use(handlers.JWTAuthMiddleware())
			shopowerAuth.Use(handlers.UserTypeMiddleware(1)) // userType=1 店主
			{
				// 店铺管理
				shopowerAuth.GET(consts.RouteShopowerShopAuthURL, shopowerShopHandler.GetAuthURL)
				shopowerAuth.GET(consts.RouteShopowerShops, shopowerShopHandler.ListShops)
				shopowerAuth.POST(consts.RouteShopowerShopBind, shopowerShopHandler.BindShop)
				shopowerAuth.GET(consts.RouteShopowerShopDetail, shopowerShopHandler.GetShop)
				shopowerAuth.PUT(consts.RouteShopowerShopStatus, shopowerShopHandler.UpdateShopStatus)
				shopowerAuth.DELETE(consts.RouteShopowerShopDetail, shopowerShopHandler.DeleteShop)
				shopowerAuth.POST(consts.RouteShopowerShopRefreshToken, shopowerShopHandler.RefreshToken)

				// 订单管理
				shopowerOrderHandler := shopower.NewOrderHandler()
				shopowerAuth.POST(consts.RouteShopowerOrderSync, shopowerOrderHandler.SyncOrders)
				shopowerAuth.GET(consts.RouteShopowerOrders, shopowerOrderHandler.ListOrders)
				shopowerAuth.GET(consts.RouteShopowerOrderReadyToShip, shopowerOrderHandler.GetReadyToShipOrders)
				shopowerAuth.GET(consts.RouteShopowerOrderDetail, shopowerOrderHandler.GetOrder)
				shopowerAuth.POST(consts.RouteShopowerOrderRefresh, shopowerOrderHandler.RefreshOrder)
				shopowerAuth.PUT(consts.RouteShopowerOrderForceStatus, shopowerOrderHandler.ForceUpdateStatus)
				shopowerAuth.DELETE(consts.RouteShopowerOrderUnlockStatus, shopowerOrderHandler.UnlockStatus)

				// 发货管理
				shopowerShipmentHandler := shopower.NewShipmentHandler()
				shopowerAuth.POST(consts.RouteShopowerShipmentShip, shopowerShipmentHandler.ShipOrder)
				shopowerAuth.POST(consts.RouteShopowerShipmentBatchShip, shopowerShipmentHandler.BatchShipOrders)
				shopowerAuth.GET(consts.RouteShopowerShipmentParameter, shopowerShipmentHandler.GetShippingParameter)
				shopowerAuth.GET(consts.RouteShopowerShipmentTrackingNo, shopowerShipmentHandler.GetTrackingNumber)
				shopowerAuth.GET(consts.RouteShopowerShipments, shopowerShipmentHandler.ListShipments)
				shopowerAuth.GET(consts.RouteShopowerShipmentDetail, shopowerShipmentHandler.GetShipment)
				shopowerAuth.POST(consts.RouteShopowerShipmentSyncLogistics, shopowerShipmentHandler.SyncLogisticsChannels)
				shopowerAuth.GET(consts.RouteShopowerShipmentLogistics, shopowerShipmentHandler.GetLogisticsChannels)
			}
		}

		// ==================== 运营路由 (operator) ====================
		operatorGroup := admin.Group(consts.OperatorPrefix)
		operatorGroup.Use(handlers.JWTAuthMiddleware())
		operatorGroup.Use(handlers.UserTypeMiddleware(5)) // userType=5 运营
		{
			// 店铺管理
			operatorShopHandler := operator.NewShopHandler()
			operatorGroup.GET(consts.RouteOperatorShops, operatorShopHandler.ListShops)
			operatorGroup.GET(consts.RouteOperatorShopDetail, operatorShopHandler.GetShop)

			// 订单管理
			operatorOrderHandler := operator.NewOrderHandler()
			operatorGroup.GET(consts.RouteOperatorOrders, operatorOrderHandler.ListOrders)
			operatorGroup.GET(consts.RouteOperatorOrderDetail, operatorOrderHandler.GetOrder)
		}

		// ==================== 平台路由 (platform) ====================
		platformGroup := admin.Group(consts.PlatformPrefix)
		platformGroup.Use(handlers.JWTAuthMiddleware())
		platformGroup.Use(handlers.UserTypeMiddleware(9)) // userType=9 平台
		{
			// 用户管理
			platformUserHandler := platform.NewUserHandler()
			platformGroup.GET(consts.RoutePlatformUsers, platformUserHandler.ListUsers)
			platformGroup.GET(consts.RoutePlatformUserDetail, platformUserHandler.GetUser)
			platformGroup.PUT(consts.RoutePlatformUserStatus, platformUserHandler.UpdateUserStatus)

			// 店铺管理
			platformShopHandler := platform.NewShopHandler()
			platformGroup.GET(consts.RoutePlatformShops, platformShopHandler.ListShops)
			platformGroup.GET(consts.RoutePlatformShopDetail, platformShopHandler.GetShop)
			platformGroup.PUT(consts.RoutePlatformShopStatus, platformShopHandler.UpdateShopStatus)
		}

	}

	return r
}
