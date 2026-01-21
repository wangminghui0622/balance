package constants

// ==================== 用户类型 ====================
const (
	UserTypeShopOwner int8 = 1 // 店铺
	UserTypeOperator  int8 = 5 // 运营
	UserTypePlatform  int8 = 9 // 平台
)

// IsValidUserType 验证用户类型是否有效（注册时只允许店铺和运营）
func IsValidUserType(userType int8) bool {
	return userType == UserTypeShopOwner || userType == UserTypeOperator
}

// IsValidUserTypeForRegister 验证注册时的用户类型（只允许店铺和运营）
func IsValidUserTypeForRegister(userType int8) bool {
	return userType == UserTypeShopOwner || userType == UserTypeOperator
}

// ==================== 用户状态 ====================
const (
	UserStatusNormal   int8 = 1 // 正常
	UserStatusDisabled int8 = 2 // 禁用
)

// ==================== HTTP 状态码 ====================
const (
	HTTPStatusOK                  = 200 // 成功
	HTTPStatusBadRequest          = 400 // 请求参数错误
	HTTPStatusUnauthorized        = 401 // 未授权
	HTTPStatusForbidden           = 403 // 禁止访问
	HTTPStatusInternalServerError = 500 // 服务器内部错误
)

// ==================== 响应码 ====================
const (
	ResponseCodeSuccess = 200 // 成功
	ResponseCodeError   = 400 // 错误
)

// ==================== 语言 ====================
const (
	LanguageZh = "zh" // 中文
	LanguageEn = "en" // 英文
)

// ==================== 默认值 ====================
const (
	DefaultLanguage = LanguageZh       // 默认语言
	DefaultStatus   = UserStatusNormal // 默认状态
)

// ==================== ID生成器初始值 ====================
const (
	IDInitialShopOwner int64 = 19906070668 // 店铺ID初始值
	IDInitialOperator  int64 = 58608109796 // 运营ID初始值
	IDInitialPlatform  int64 = 91609051906 // 平台ID初始值
)

// ==================== ID生成器间隔 ====================
const (
	// 店铺ID间隔：100-500之间的随机值
	IDIncrementShopOwnerMin int64 = 100 // 店铺ID最小增量
	IDIncrementShopOwnerMax int64 = 500 // 店铺ID最大增量

	// 运营ID间隔：30-50之间的随机值
	IDIncrementOperatorMin int64 = 30 // 运营ID最小增量
	IDIncrementOperatorMax int64 = 50 // 运营ID最大增量

	// 平台ID间隔：10-20之间的随机值
	IDIncrementPlatformMin int64 = 10 // 平台ID最小增量
	IDIncrementPlatformMax int64 = 20 // 平台ID最大增量
)

// ==================== 降级ID生成器（时间戳ID） ====================
const (
	// 降级ID前缀：8开头，表示使用时间戳降级方案
	// 格式：8 + 毫秒时间戳（12位） + 随机数（0-999，3位）
	// 例如：81705747200000123 = 8 + 1705747200000 + 123
	FallbackIDPrefix  int64 = 80000000000 // 降级ID前缀
	FallbackRandomMax int64 = 999         // 降级ID随机数最大值（3位）
)
