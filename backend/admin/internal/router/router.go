package router

import (
	"time"

	"balance/backend/admin/internal/handlers"
	"balance/backend/admin/internal/handlers/operator"
	"balance/backend/admin/internal/handlers/platform"
	"balance/backend/admin/internal/handlers/shopower"
	"balance/backend/internal/config"
	"balance/backend/internal/consts"
	"balance/backend/internal/middleware"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// SetupRouter 设置路由
func SetupRouter(mode string, cfg *config.Config) *gin.Engine {
	gin.SetMode(mode)
	r := gin.New()
	r.Use(gin.Logger())
	r.Use(gin.Recovery())

	// CORS配置
	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"*"}, // 生产环境应配置具体域名
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization", "X-Requested-With"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	// Prometheus监控中间件
	r.Use(middleware.PrometheusMiddleware())

	// Prometheus指标端点
	r.GET("/metrics", gin.WrapH(promhttp.Handler()))

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

				// 结算明细管理
				shopowerEscrowHandler := shopower.NewEscrowHandler()
				shopowerAuth.GET(consts.RouteShopowerEscrows, shopowerEscrowHandler.ListPendingEscrows)
				shopowerAuth.POST(consts.RouteShopowerEscrowSync, shopowerEscrowHandler.SyncEscrow)
				shopowerAuth.GET(consts.RouteShopowerEscrowDetail, shopowerEscrowHandler.GetEscrow)
				shopowerAuth.POST("/escrows/batch-sync", shopowerEscrowHandler.BatchSyncEscrows)

				// 财务收入管理
				shopowerFinanceHandler := shopower.NewFinanceHandler()
				shopowerAuth.GET(consts.RouteShopowerFinances, shopowerFinanceHandler.ListIncomes)
				shopowerAuth.POST(consts.RouteShopowerFinanceSync, shopowerFinanceHandler.SyncTransactions)
				shopowerAuth.GET(consts.RouteShopowerFinanceStats, shopowerFinanceHandler.GetIncomeStats)

				// 账户管理
				shopowerAccountHandler := shopower.NewAccountHandler()
				shopowerAuth.GET("/account/prepayment", shopowerAccountHandler.GetPrepaymentAccount)
				shopowerAuth.GET("/account/deposit", shopowerAccountHandler.GetDepositAccount)
				shopowerAuth.GET("/account/commission", shopowerAccountHandler.GetCommissionAccount)
				shopowerAuth.GET("/account/summary", shopowerAccountHandler.GetAllAccounts)
				shopowerAuth.GET("/account/prepayment/transactions", shopowerAccountHandler.GetPrepaymentTransactions)
				shopowerAuth.GET("/account/deposit/transactions", shopowerAccountHandler.GetDepositTransactions)
				shopowerAuth.GET("/account/commission/transactions", shopowerAccountHandler.GetCommissionTransactions)

				// 结算管理
				shopowerAuth.GET("/settlements", shopowerAccountHandler.GetSettlements)
				shopowerAuth.GET("/settlements/stats", shopowerAccountHandler.GetSettlementStats)

				// 提现管理
				shopowerAuth.POST("/withdraw/apply", shopowerAccountHandler.ApplyWithdraw)
				shopowerAuth.GET("/withdraw/list", shopowerAccountHandler.GetWithdrawApplications)

				// 充值管理
				shopowerAuth.POST("/recharge/apply", shopowerAccountHandler.ApplyRecharge)
				shopowerAuth.GET("/recharge/list", shopowerAccountHandler.GetRechargeApplications)
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

			// 发货管理 (运营发货)
			operatorShipmentHandler := operator.NewShipmentHandler()
			operatorGroup.POST("/shipments/ship", operatorShipmentHandler.ShipOrder)
			operatorGroup.GET("/orders/pending", operatorShipmentHandler.GetPendingOrders)
			operatorGroup.GET("/shipments", operatorShipmentHandler.GetShipmentRecords)

			// 结算管理
			operatorSettlementHandler := operator.NewSettlementHandler()
			operatorGroup.GET("/settlements", operatorSettlementHandler.GetSettlements)
			operatorGroup.GET("/settlements/stats", operatorSettlementHandler.GetSettlementStats)

			// 账户管理
			operatorAccountHandler := operator.NewAccountHandler()
			operatorGroup.GET("/account", operatorAccountHandler.GetAccount)
			operatorGroup.GET("/account/transactions", operatorAccountHandler.GetTransactions)

			// 提现管理
			operatorGroup.POST("/withdraw/apply", operatorAccountHandler.ApplyWithdraw)
			operatorGroup.GET("/withdraw/list", operatorAccountHandler.GetWithdrawApplications)
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

			// 同步管理
			platformSyncHandler := platform.NewSyncHandler()
			platformGroup.GET(consts.RoutePlatformSyncStats, platformSyncHandler.GetSyncStats)
			platformGroup.GET(consts.RoutePlatformSyncRecords, platformSyncHandler.ListSyncRecords)
			platformGroup.POST(consts.RoutePlatformSyncRecordReset, platformSyncHandler.ResetSyncRecord)

			// 结算管理
			platformSettlementHandler := platform.NewSettlementHandler()
			platformGroup.GET("/settlements", platformSettlementHandler.GetSettlements)
			platformGroup.GET("/settlements/stats", platformSettlementHandler.GetSettlementStats)
			platformGroup.GET("/settlements/pending", platformSettlementHandler.GetPendingSettlements)
			platformGroup.POST("/settlements/process", platformSettlementHandler.ProcessSettlement)

			// 合作管理（店铺-运营分配）
			platformCooperationHandler := platform.NewCooperationHandler()
			platformGroup.GET("/cooperations", platformCooperationHandler.ListCooperations)
			platformGroup.POST("/cooperations", platformCooperationHandler.CreateCooperation)
			platformGroup.PUT("/cooperations/:id", platformCooperationHandler.UpdateCooperation)
			platformGroup.DELETE("/cooperations/:id", platformCooperationHandler.CancelCooperation)
			platformGroup.GET("/cooperations/stats", platformCooperationHandler.GetCooperationStats)
			platformGroup.GET("/operators", platformCooperationHandler.GetOperatorList)
			platformGroup.GET("/shop-owners", platformCooperationHandler.GetShopOwnerList)

			// 账户管理
			platformAccountHandler := platform.NewAccountHandler()
			platformGroup.GET("/accounts/prepayment", platformAccountHandler.ListPrepaymentAccounts)
			platformGroup.GET("/accounts/deposit", platformAccountHandler.ListDepositAccounts)
			platformGroup.GET("/accounts/operator", platformAccountHandler.ListOperatorAccounts)
			platformGroup.POST("/accounts/prepayment/recharge", platformAccountHandler.RechargePrepayment)
			platformGroup.POST("/accounts/deposit/pay", platformAccountHandler.PayDeposit)
			platformGroup.GET("/accounts/transactions", platformAccountHandler.GetAccountTransactions)
			platformGroup.GET("/accounts/stats", platformAccountHandler.GetAccountStats)
			platformGroup.GET("/account/commission", platformAccountHandler.GetPlatformCommissionAccount)
			platformGroup.GET("/account/commission/transactions", platformAccountHandler.GetPlatformCommissionTransactions)

			// 提现审核
			platformGroup.GET("/withdraw/list", platformAccountHandler.GetWithdrawApplications)
			platformGroup.POST("/withdraw/approve", platformAccountHandler.ApproveWithdraw)
			platformGroup.POST("/withdraw/reject", platformAccountHandler.RejectWithdraw)
			platformGroup.POST("/withdraw/confirm_paid", platformAccountHandler.ConfirmWithdrawPaid)

			// 充值审核
			platformGroup.GET("/recharge/list", platformAccountHandler.GetRechargeApplications)
			platformGroup.POST("/recharge/approve", platformAccountHandler.ApproveRecharge)
			platformGroup.POST("/recharge/reject", platformAccountHandler.RejectRecharge)

			// 佣金管理
			platformCommissionHandler := platform.NewCommissionHandler()
			platformGroup.GET("/commission/stats", platformCommissionHandler.GetCommissionStats)
			platformGroup.GET("/commission/list", platformCommissionHandler.GetCommissionList)

			// 财务审核
			platformFinanceAuditHandler := platform.NewFinanceAuditHandler()
			platformGroup.GET("/finance/audit/stats", platformFinanceAuditHandler.GetAuditStats)
			platformGroup.GET("/finance/audit/withdraw", platformFinanceAuditHandler.GetWithdrawAuditList)
			platformGroup.GET("/finance/audit/recharge", platformFinanceAuditHandler.GetRechargeAuditList)
			platformGroup.POST("/finance/audit/approve", platformFinanceAuditHandler.ApproveAudit)
			platformGroup.POST("/finance/withdraw/apply", platformFinanceAuditHandler.CreateWithdrawApplication)

			// 罚补账户
			platformPenaltyHandler := platform.NewPenaltyHandler()
			platformGroup.GET("/penalty/stats", platformPenaltyHandler.GetPenaltyStats)
			platformGroup.GET("/penalty/list", platformPenaltyHandler.GetPenaltyList)
			platformGroup.POST("/penalty/create", platformPenaltyHandler.CreatePenalty)

			// 收款账户
			platformCollectionHandler := platform.NewCollectionHandler()
			platformGroup.GET("/collection/accounts", platformCollectionHandler.GetCollectionAccounts)
			platformGroup.POST("/collection/accounts", platformCollectionHandler.CreateCollectionAccount)
			platformGroup.PUT("/collection/accounts/:id", platformCollectionHandler.UpdateCollectionAccount)
			platformGroup.DELETE("/collection/accounts/:id", platformCollectionHandler.DeleteCollectionAccount)
			platformGroup.POST("/collection/accounts/:id/default", platformCollectionHandler.SetDefaultAccount)

			// 模拟数据生成器（仅沙箱环境可用）
			platformMockHandler := platform.NewMockHandler(cfg)
			platformGroup.POST("/mock/orders", platformMockHandler.GenerateMockOrders)
			platformGroup.DELETE("/mock/clean", platformMockHandler.CleanMockData)
		}

	}

	return r
}
