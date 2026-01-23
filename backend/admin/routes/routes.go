package routes

import (
	"balance/admin/controllers"
	"balance/admin/services"
	"balance/internal/config"
	"balance/internal/middleware"
	"balance/internal/models"
	shareUtils "balance/internal/utils"
	"log"
	"time"

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

	// 初始化Shopee API客户端（从数据库读取所有配置）
	tokenRepo := models.NewShopeeTokenRepository(db)
	// 优先从数据库读取，如果没有则尝试从配置文件读取（兼容旧配置）
	var shopeeToken *models.ShopeeToken
	var useConfigFile bool

	// 尝试从数据库读取
	if cfg.ShopeeShopID > 0 {
		token, err := tokenRepo.GetByShopID(cfg.ShopeeShopID)
		if err == nil && token != nil {
			shopeeToken = token
			log.Printf("✅ 已从数据库加载 Shopee 配置 (shop_id=%d)", token.ShopID)
		}
	}

	// 如果数据库中没有，尝试从配置文件读取（兼容旧配置）
	if shopeeToken == nil && cfg.ShopeePartnerID > 0 && cfg.ShopeePartnerKey != "" && cfg.ShopeeShopID > 0 {
		useConfigFile = true
		log.Printf("⚠️  数据库中没有找到 Shopee 配置，使用配置文件中的配置（建议迁移到数据库）")
		// 创建临时 token 对象用于初始化客户端
		tokenExpireAt := time.Now().Add(4 * time.Hour)
		if cfg.ShopeeAccessToken == "" {
			log.Printf("⚠️  配置文件中的 access_token 为空，Shopee API 客户端将无法使用")
		} else {
			shopeeToken = &models.ShopeeToken{
				ShopID:        cfg.ShopeeShopID,
				PartnerID:     cfg.ShopeePartnerID,
				PartnerKey:    cfg.ShopeePartnerKey,
				AccessToken:   cfg.ShopeeAccessToken,
				RefreshToken:  cfg.ShopeeRefreshToken,
				TokenExpireAt: &tokenExpireAt,
				IsSandbox:     cfg.ShopeeIsSandbox,
				Redirect:      cfg.ShopeeRedirect,
			}
		}
	}

	// 初始化 Shopee API 客户端
	if shopeeToken != nil && shopeeToken.AccessToken != "" {
		tokenExpireAt := time.Time{}
		if shopeeToken.TokenExpireAt != nil {
			tokenExpireAt = *shopeeToken.TokenExpireAt
		}

		shopeeClient := shareUtils.NewShopeeAPIClientWithRefresh(
			shopeeToken.PartnerID,
			shopeeToken.PartnerKey,
			shopeeToken.ShopID,
			shopeeToken.AccessToken,
			shopeeToken.RefreshToken,
			tokenExpireAt,
			shopeeToken.IsSandbox,
			// Token 刷新回调：当 token 自动刷新时，保存到数据库
			func(accessToken, refreshToken string, expireIn int64) {
				tokenExpireAt := time.Now().Add(time.Duration(expireIn) * time.Second)
				err := tokenRepo.UpdateTokens(shopeeToken.ShopID, accessToken, refreshToken, &tokenExpireAt, nil)
				if err != nil {
					log.Printf("❌ 保存刷新后的 token 到数据库失败: %v", err)
				} else {
					log.Printf("✅ Shopee access_token 已自动刷新并保存到数据库")
				}
			},
		)
		orderService.SetShopeeClient(shopeeClient)

		// 如果是从配置文件读取的，尝试保存到数据库
		if useConfigFile {
			err := tokenRepo.CreateOrUpdate(shopeeToken)
			if err != nil {
				log.Printf("⚠️  尝试将配置文件中的配置保存到数据库失败: %v", err)
			} else {
				log.Printf("✅ 已将配置文件中的 Shopee 配置保存到数据库")
			}
		}
	}

	// 初始化控制器
	authController := controllers.NewAuthController(authService)
	orderController := controllers.NewOrderController(orderService, db)
	shopeeAuthController := controllers.NewShopeeAuthController(cfg, db)
	// 认证路由
	auth := r.Group("/api/v1/balance/admin/auth")
	{
		auth.POST("/login", authController.Login)
		auth.POST("/register", authController.Register)
	}
	// Shopee 授权回调（用于换取 access_token）
	r.GET("/api/v1/balance/admin/shopee/auth/callback", shopeeAuthController.AuthCallback)
	// Shopee 授权链接生成（方便前端/浏览器获取授权URL）
	r.POST("/api/v1/balance/admin/shopee/auth/url", shopeeAuthController.GenerateAuthURL)

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
