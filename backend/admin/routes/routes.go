package routes

import (
	"balance/admin/controllers"
	"balance/admin/services"
	"balance/internal/config"
	"balance/internal/middleware"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

// SetupRoutes 设置路由
func SetupRoutes(db *gorm.DB, redisClient *redis.Client, cfg *config.Config) *gin.Engine {
	r := gin.Default()

	// CORS中间件
	r.Use(middleware.CORSMiddleware())

	// 健康检查
	r.GET("/api/v1/balance/admin/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":  "ok",
			"message": "Admin service is running",
		})
	})
	// 初始化服务
	authService := services.NewAuthService(db, redisClient, []byte(cfg.JWTSecret), cfg.JWTExpiration)
	// 初始化控制器
	authController := controllers.NewAuthController(authService)
	// 认证路由
	auth := r.Group("/api/v1/balance/admin/auth")
	{
		auth.POST("/login", authController.Login)
		auth.POST("/register", authController.Register)
	}
	// 受保护接口
	api := r.Group("/api/v1/balance/admin/balance", middleware.AuthMiddleware([]byte(cfg.JWTSecret)))
	{
		api.GET("/me", func(c *gin.Context) {
			userId, _ := c.Get("userId")
			c.JSON(http.StatusOK, gin.H{
				"code":    200,
				"message": "success",
				"data": gin.H{
					"userId": userId,
				},
			})
		})
	}
	return r
}
