package routes

import (
	"balance/admin/controllers"
	"balance/admin/services"
	"balance/internal/config"
	"balance/internal/middleware"
	shareUtils "balance/internal/utils"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

// SetupRoutes 设置路由
func SetupRoutes(db *gorm.DB, redisClient *redis.Client, cfg *config.Config) *gin.Engine {
	r := gin.Default()
	r.Use(middleware.CORSMiddleware())
	// 初始化服务
	authService := services.NewAuthService(db, redisClient, []byte(cfg.JWTSecret), cfg.JWTExpiration)
	orderService := services.NewOrderService("")

	// 初始化Shopee API客户端（如果配置了）
	if cfg.ShopeePartnerID > 0 && cfg.ShopeePartnerKey != "" && cfg.ShopeeShopID > 0 && cfg.ShopeeAccessToken != "" {
		shopeeClient := shareUtils.NewShopeeAPIClient(
			cfg.ShopeePartnerID,
			cfg.ShopeePartnerKey,
			cfg.ShopeeShopID,
			cfg.ShopeeAccessToken,
			cfg.ShopeeIsSandbox,
		)
		orderService.SetShopeeClient(shopeeClient)
	}

	// 初始化控制器
	authController := controllers.NewAuthController(authService)
	orderController := controllers.NewOrderController(orderService, db)
	shopeeAuthController := controllers.NewShopeeAuthController(cfg)
	// 认证路由
	auth := r.Group("/api/v1/balance/admin/auth")
	{
		auth.POST("/login", authController.Login)
		auth.POST("/register", authController.Register)
	}
	// Shopee 授权回调（用于换取 access_token）
	r.GET("/api/v1/balance/admin/shopee/auth/callback", shopeeAuthController.AuthCallback)
	// Shopee 授权链接生成（方便前端/浏览器获取授权URL）
	r.GET("/api/v1/balance/admin/shopee/auth/url", shopeeAuthController.GenerateAuthURL)

	// Shopee 订单状态回调（对外给虾皮配置的回调地址）
	// 示例： https://你的域名/balance/orderStatusSync/callback
	r.POST("/api/v1/balance/admin/orderStatusSync/callback", orderController.ShopeeCallback)

	// Shopee 订单拉取接口（需要认证）
	shopee := r.Group("/api/v1/balance/admin/order")
	{
		shopee.GET("/list", orderController.FetchOrders)        // 拉取订单列表
		shopee.GET("/detail", orderController.FetchOrderDetail) // 拉取订单详情
	}

	return r
}
