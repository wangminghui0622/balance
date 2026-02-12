package router

import (
	"balance/backend/app/internal/handlers"
	"balance/backend/app/internal/handlers/shopower"
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

	app := r.Group(consts.AppPrefix)
	{
		// ==================== 公共认证路由（无需登录） ====================
		authHandler := handlers.NewAuthHandler()
		app.POST(consts.RouteAuthRegister, authHandler.Register)
		app.POST(consts.RouteAuthLogin, authHandler.Login)
		app.POST(consts.RouteAuthSendCode, authHandler.SendEmailCode)
		app.POST(consts.RouteAuthResetPassword, authHandler.ResetPassword)

		// 需要登录的公共接口
		authGroup := app.Group("")
		authGroup.Use(handlers.JWTAuthMiddleware())
		{
			authGroup.GET(consts.RouteAuthMe, authHandler.GetCurrentUser)
		}

		// ==================== 店主路由 (shopower) ====================
		shopowerGroup := app.Group(consts.ShopowerPrefix)
		{
			// Shopee授权回调（无需登录）
			shopowerShopHandler := shopower.NewShopHandler()
			shopowerGroup.GET(consts.RouteShopowerShopCallback, shopowerShopHandler.AuthCallback)

			// 需要登录的店主接口
			shopowerAuth := shopowerGroup.Group("")
			shopowerAuth.Use(handlers.JWTAuthMiddleware())
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
	}
	return r
}
