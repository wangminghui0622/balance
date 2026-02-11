package router

import (
	"balance/backend/app/internal/handlers"
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
		// 用户认证（无需登录）
		appAuthHandler := handlers.NewAuthHandler()
		app.POST(consts.RouteAuthRegister, appAuthHandler.Register)
		app.POST(consts.RouteAuthLogin, appAuthHandler.Login)
		app.POST(consts.RouteAuthSendCode, appAuthHandler.SendEmailCode)
		app.POST(consts.RouteAuthResetPassword, appAuthHandler.ResetPassword)

		// 需要登录的接口
		appAuthGroup := app.Group("")
		appAuthGroup.Use(handlers.JWTAuthMiddleware())
		{
			// 获取当前用户信息
			appAuthGroup.GET(consts.RouteAuthMe, appAuthHandler.GetCurrentUser)

			// 店铺管理
			appShopHandler := handlers.NewShopHandler()
			appAuthGroup.GET(consts.RouteShopAuthURL, appShopHandler.GetAuthURL)
			appAuthGroup.GET(consts.RouteShops, appShopHandler.ListShops)
			appAuthGroup.POST(consts.RouteShopBind, appShopHandler.BindShop)
			appAuthGroup.GET(consts.RouteShopDetail, appShopHandler.GetShop)
			appAuthGroup.PUT(consts.RouteShopStatus, appShopHandler.UpdateShopStatus)
			appAuthGroup.DELETE(consts.RouteShopDetail, appShopHandler.DeleteShop)
			appAuthGroup.POST(consts.RouteShopRefreshToken, appShopHandler.RefreshToken)

			// 订单管理
			appOrderHandler := handlers.NewOrderHandler()
			appAuthGroup.POST(consts.RouteOrderSync, appOrderHandler.SyncOrders)
			appAuthGroup.GET(consts.RouteOrders, appOrderHandler.ListOrders)
			appAuthGroup.GET(consts.RouteOrderReadyToShip, appOrderHandler.GetReadyToShipOrders)
			appAuthGroup.GET(consts.RouteOrderDetail, appOrderHandler.GetOrder)
			appAuthGroup.POST(consts.RouteOrderRefresh, appOrderHandler.RefreshOrder)

			// 发货管理
			appShipmentHandler := handlers.NewShipmentHandler()
			appAuthGroup.POST(consts.RouteShipmentShip, appShipmentHandler.ShipOrder)
			appAuthGroup.POST(consts.RouteShipmentBatchShip, appShipmentHandler.BatchShipOrders)
			appAuthGroup.GET(consts.RouteShipmentParameter, appShipmentHandler.GetShippingParameter)
			appAuthGroup.GET(consts.RouteShipmentTrackingNo, appShipmentHandler.GetTrackingNumber)
			appAuthGroup.GET(consts.RouteShipments, appShipmentHandler.ListShipments)
			appAuthGroup.GET(consts.RouteShipmentDetail, appShipmentHandler.GetShipment)
		}
	}
	return r
}
