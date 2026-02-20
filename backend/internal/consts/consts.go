package consts

import "time"

// ==================== 店铺状态 ====================

const (
	ShopStatusDisabled = 0 // 禁用
	ShopStatusEnabled  = 1 // 正常
)

// ==================== 订单状态 (虾皮订单状态) ====================

const (
	OrderStatusUnpaid         = "UNPAID"          // 未付款
	OrderStatusReadyToShip    = "READY_TO_SHIP"   // 待发货
	OrderStatusProcessed      = "PROCESSED"       // 已处理
	OrderStatusShipped        = "SHIPPED"         // 已发货
	OrderStatusCompleted      = "COMPLETED"       // 已完成
	OrderStatusInCancel       = "IN_CANCEL"       // 取消中
	OrderStatusCancelled        = "CANCELLED"           // 已取消
	OrderStatusCancelledBeforeShip = "CANCELLED_BEFORE_SHIP" // 发货前取消（全额退款/退货/取消且运营未发货）
	OrderStatusInvoicePending   = "INVOICE_PENDING"     // 待开票
)

// OrderStatusPriority 订单状态优先级（数字越大越靠后，不可逆）
var OrderStatusPriority = map[string]int{
	OrderStatusUnpaid:            1,
	OrderStatusInvoicePending:    2,
	OrderStatusReadyToShip:       3,
	OrderStatusProcessed:         4,
	OrderStatusShipped:           5,
	OrderStatusCompleted:         6,
	OrderStatusInCancel:          7,
	OrderStatusCancelledBeforeShip: 8,
	OrderStatusCancelled:         9,
}

// ==================== 发货状态 ====================

const (
	ShipStatusPending = 0 // 待发货
	ShipStatusShipped = 1 // 已发货
	ShipStatusFailed  = 2 // 发货失败
)

// ==================== 操作类型 ====================

const (
	OpTypeAuthCallback  = "auth_callback"
	OpTypeTokenRefresh  = "token_refresh"
	OpTypeOrderSync     = "order_sync"
	OpTypeOrderShip     = "order_ship"
	OpTypeLogisticsSync = "logistics_sync"
	OpTypeWebhook       = "webhook"
)

// ==================== Webhook事件类型 ====================

const (
	WebhookShopAuth          = 0
	WebhookOrderStatus       = 3
	WebhookTrackingUpdate    = 4
	WebhookBannedItem        = 5
	WebhookReturnCreated     = 6  // 退货退款创建
	WebhookPromotionUpdate   = 7
	WebhookReservedStock     = 8
	WebhookBuyerCancelOrder  = 9
	WebhookSellerCancelOrder = 9
	WebhookReturnStatusChange = 15 // 退货退款状态变更
)

// ==================== 操作状态 ====================

const (
	OpStatusFailed  = 0
	OpStatusSuccess = 1
)

// ==================== Redis Key ====================

const (
	KeyShopToken       = "shopee:token:%d"
	KeyShopInfo        = "shopee:shop:%d"
	KeyOrderStatus     = "shopee:order:status:%d:%s"
	KeySyncLock        = "shopee:lock:sync:%d"
	KeyShipLock        = "shopee:lock:ship:%d:%s"
	KeyRateLimit       = "shopee:ratelimit:%d:%s"
	KeyShipQueue       = "shopee:queue:ship"
	KeyLogistics       = "shopee:logistics:%d"
	KeyWebhookDedup    = "shopee:webhook:dedup:%d:%s:%d:%d"
	KeyOrderLock       = "shopee:lock:order:%d:%s"
	KeyOrderUpdateTime      = "shopee:order:update_time:%d:%s"
	KeyPrepaymentNotified = "balance:prepayment:notified:%d"     // shopID — 预付款不足通知去重（避免短时间重复通知）
	KeyReturnLock         = "shopee:lock:return:%d:%s"           // shopID, returnSN — 退货退款处理锁
	KeyReturnDedup        = "shopee:webhook:return:dedup:%d:%s:%d" // shopID, returnSN, eventCode — 退货Webhook去重（不含timestamp，防止虾皮重试绕过去重）
	KeyShopowerOrderStats = "balance:shopower:order_stats:%d"     // admin_id — 店主订单统计缓存（全部/未结算/已结算/账款调整）
)

// ==================== 缓存过期时间 ====================

const (
	TokenExpireBuffer  = 5 * time.Minute
	ShopInfoExpire     = 1 * time.Hour
	OrderStatusExpire  = 30 * time.Minute
	SyncLockExpire     = 15 * time.Minute // 订单巡检8min + 退货巡检3min + 余量，防止锁在任务完成前过期
	ShipLockExpire     = 2 * time.Minute  // 增加到2分钟，防止API调用超时
	RateLimitExpire    = 1 * time.Minute
	LogisticsExpire    = 1 * time.Hour
	WebhookDedupExpire = 5 * time.Minute
	OrderLockExpire    = 30 * time.Second // 增加到30秒，防止数据库操作超时
	ReturnLockExpire   = 30 * time.Second // 退货退款处理锁过期时间
	OrderUpdateTimeTTL         = 24 * time.Hour
	PrepaymentNotifiedCooldown = 30 * time.Minute // 预付款不足通知冷却：30分钟内不重复通知
	ShopowerOrderStatsExpire  = 1 * time.Hour     // 店主订单统计缓存 TTL：1 小时
)

// ==================== 分页默认值 ====================

const (
	DefaultPage     = 1
	DefaultPageSize = 10
	MaxPageSize     = 100
)

// ==================== 虾皮API相关 ====================

const (
	ShopeeAPITimeout         = 30
	ShopeeOrderListPageSize  = 100
	ShopeeOrderDetailMaxSize = 50
	ShopeeMaxTimeRange       = 15 * 24 * 3600
	ShopeeAPIRateLimit       = 10
	ShopeeAPIRetryTimes      = 3
	ShopeeAPIRetryTimesSync  = 5  // 巡检/同步类操作，对临时错误重试更多次
	ShopeeAPIRetryInterval   = 1000
)

// ==================== 路由前缀 ====================
const (
	RouteWebhook = "/api/v1/balance/admin/webhook"
)
const (
	AdminPrefix    = "/api/v1/balance/admin"
	AppPrefix      = "/api/v1/balance/app"
	ShopowerPrefix = "/shopower"
	OperatorPrefix = "/operator"
	PlatformPrefix = "/platform"
)

// ==================== 路由路径 ====================
const (
	RouteHealth = "/health"
)

// ==================== 公共认证路由 ====================

const (
	RouteAuthCallback      = "/auth/callback"
	RouteAuthRegister      = "/auth/register"
	RouteAuthLogin         = "/auth/login"
	RouteAuthMe            = "/auth/me"
	RouteAuthSendCode      = "/auth/send-code"
	RouteAuthResetPassword = "/auth/reset-password"
)

// ==================== 店主路由 (shopower) ====================

const (
	RouteShopowerShops            = "/shops"
	RouteShopowerShopAuthURL      = "/shops/auth-url"
	RouteShopowerShopCallback     = "/shops/callback"
	RouteShopowerShopBind         = "/shops/bind"
	RouteShopowerShopDetail       = "/shops/:shop_id"
	RouteShopowerShopStatus       = "/shops/:shop_id/status"
	RouteShopowerShopRefreshToken = "/shops/:shop_id/refresh-token"
)

const (
	RouteShopowerOrders            = "/orders"
	RouteShopowerOrderStats        = "/orders/stats"
	RouteShopowerOrderSync         = "/orders/sync"
	RouteShopowerOrderReadyToShip  = "/orders/ready-to-ship"
	RouteShopowerOrderDetail       = "/orders/:shop_id/:order_sn"
	RouteShopowerOrderRefresh      = "/orders/:shop_id/:order_sn/refresh"
	RouteShopowerOrderForceStatus  = "/orders/:shop_id/:order_sn/force-status"
	RouteShopowerOrderUnlockStatus = "/orders/:shop_id/:order_sn/unlock"
)

const (
	RouteShopowerShipments             = "/shipments"
	RouteShopowerShipmentShip          = "/shipments/ship"
	RouteShopowerShipmentBatchShip     = "/shipments/batch-ship"
	RouteShopowerShipmentParameter     = "/shipments/shipping-parameter"
	RouteShopowerShipmentTrackingNo    = "/shipments/tracking-number"
	RouteShopowerShipmentDetail        = "/shipments/:shop_id/:order_sn"
	RouteShopowerShipmentSyncLogistics = "/shipments/sync-logistics/:shop_id"
	RouteShopowerShipmentLogistics     = "/shipments/logistics/:shop_id"
)

const (
	RouteShopowerEscrows      = "/escrows"
	RouteShopowerEscrowSync   = "/escrows/:shop_id/:order_sn/sync"
	RouteShopowerEscrowDetail = "/escrows/:shop_id/:order_sn"
)

const (
	RouteShopowerFinances     = "/finances"
	RouteShopowerFinanceSync  = "/finances/:shop_id/sync"
	RouteShopowerFinanceStats = "/finances/:shop_id/stats"
)

// ==================== 运营路由 (operator) ====================

const (
	RouteOperatorShops      = "/shops"
	RouteOperatorShopDetail = "/shops/:shop_id"
)

const (
	RouteOperatorOrders      = "/orders"
	RouteOperatorOrderDetail = "/orders/:shop_id/:order_sn"
)

// ==================== 平台路由 (platform) ====================

const (
	RoutePlatformUsers      = "/users"
	RoutePlatformUserDetail = "/users/:user_id"
	RoutePlatformUserStatus = "/users/:user_id/status"
)

const (
	RoutePlatformSyncStats       = "/sync/stats"
	RoutePlatformSyncRecords     = "/sync/records"
	RoutePlatformSyncRecordReset = "/sync/records/:shop_id/reset"
)

const (
	RoutePlatformShops      = "/shops"
	RoutePlatformShopDetail = "/shops/:shop_id"
	RoutePlatformShopStatus = "/shops/:shop_id/status"
)

const (
	RoutePlatformOrders      = "/orders"
	RoutePlatformOrderDetail = "/orders/:shop_id/:order_sn"
)
