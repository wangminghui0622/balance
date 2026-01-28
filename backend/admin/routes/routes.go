package routes

import (
	"balance/admin/controllers"
	"balance/admin/services"
	"balance/internal/config"
	"balance/internal/middleware"
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

// SetupRoutes 设置路由
func SetupRoutes(db *gorm.DB, redisClient *redis.Client, cfg *config.Config) *gin.Engine {
	r := gin.Default()
	r.Use(middleware.CORSMiddleware())
	// 初始化服务
	loginService := services.NewLoginService(db, redisClient, []byte(cfg.JWTSecret), cfg.JWTExpiration)
	orderService := services.NewOrderService(db)
	authService := services.NewAuthService(db)
	// 初始化控制器
	loginController := controllers.NewLoginController(loginService)
	orderController := controllers.NewOrderController(orderService, db)
	shopeeAuthController := controllers.NewShopeeAuthController(cfg, db, authService, orderService, redisClient)
	authController := controllers.NewAuthController(authService)
	// 认证路由
	auth := r.Group("/api/v1/balance/admin/auth")
	{
		auth.POST("/register", loginController.Register)
		auth.POST("/login", loginController.Login)
		auth.GET("/me", loginController.GetCurrentUser) // 获取当前用户信息
	}
	baseUrl := "/api/v1/balance/admin/"

	r.POST(baseUrl+"shopee/auth/cfg", authController.GetByPartnerId)
	// Shopee 授权链接生成（方便前端/浏览器获取授权URL）
	r.POST(baseUrl+"shopee/auth/url", shopeeAuthController.GenerateAuthURL)
	// Shopee 授权回调（用于换取 access_token）
	r.GET(baseUrl+"shopee/auth/callback", shopeeAuthController.AuthCallback)
	//
	r.POST(baseUrl+"shopee/auth/bind", shopeeAuthController.AuthBind)
	// Shopee 刷新令牌接口
	r.POST(baseUrl+"shopee/auth/refresh", shopeeAuthController.RefreshToken)
	// Shopee 发送换绑验证码接口
	r.POST(baseUrl+"shopee/auth/rebind/send-code", shopeeAuthController.SendRebindCode)
	// Shopee 换绑验证码验证接口
	r.POST(baseUrl+"shopee/auth/rebind/verify", shopeeAuthController.VerifyRebindCode)
	// Shopee 确认换绑接口
	r.POST(baseUrl+"shopee/auth/rebind/confirm", shopeeAuthController.ConfirmRebind)
	// Shopee 取消换绑接口
	r.POST(baseUrl+"shopee/auth/rebind/cancel", shopeeAuthController.CancelRebind)

	// Shopee 店铺列表
	r.POST(baseUrl+"shopee/shop/list", shopeeAuthController.ShopList)
	// Shopee 订单状态回调（对外给虾皮配置的回调地址）
	// 示例： https://你的域名/balance/orderStatusSync/callback
	r.POST(baseUrl+"orderStatusSync/callback", orderController.ShopeeCallback)

	// Shopee 订单拉取接口（需要认证）
	order := r.Group(baseUrl + "order")
	{
		order.GET("/list", orderController.FetchOrders)        // 拉取订单列表
		order.GET("/detail", orderController.FetchOrderDetail) // 拉取订单详情
	}
	//shopee := r.Group(baseUrl + "shop")
	//{
	//	shopee.GET("/list", orderController.FetchShoplist)     // 拉取店铺列表
	//	shopee.GET("/detail", orderController.FetchShopdetail) // 拉取店铺详情
	//}
	return r
}
