package utils

import "errors"

// 业务状态码定义
const (
	// 成功
	CodeSuccess = 0

	// 通用错误 (400-499)
	CodeBadRequest   = 400
	CodeUnauthorized = 401
	CodeForbidden    = 403
	CodeNotFound     = 404

	// 服务器错误 (500-599)
	CodeInternalError = 500

	// 认证相关错误码 (1000-1099)
	CodeUsernameExists    = 1001
	CodeEmailExists       = 1002
	CodeInvalidUserType   = 1003
	CodeInvalidCredential = 1004
	CodeAccountDisabled   = 1005
	CodeUserNotFound      = 1006
	CodeEmailNotRegister  = 1007
	CodeInvalidToken      = 1008

	// 邮箱验证码相关错误码 (1100-1199)
	CodeEmailCodeExpired = 1101
	CodeEmailCodeInvalid = 1102

	// 店铺相关错误码 (1200-1299)
	CodeShopNotFound       = 1201
	CodeShopUnauthorized   = 1202
	CodeShopAlreadyBound   = 1203
	CodeShopNoPermission   = 1204
	CodeShopTokenExpired   = 1205
	CodeShopSyncing        = 1206

	// 订单相关错误码 (1300-1399)
	CodeOrderNotFound      = 1301
	CodeOrderCannotShip    = 1302
	CodeOrderShipping      = 1303

	// 发货相关错误码 (1400-1499)
	CodeShipmentNotFound   = 1401
)

// 错误信息定义
var (
	// 认证相关错误
	ErrUsernameExists    = errors.New("用户名已存在")
	ErrEmailExists       = errors.New("邮箱已被注册")
	ErrInvalidUserType   = errors.New("无效的用户类型")
	ErrInvalidCredential = errors.New("用户名或密码错误")
	ErrAccountDisabled   = errors.New("账户已被禁用")
	ErrUserNotFound      = errors.New("用户不存在")
	ErrEmailNotRegister  = errors.New("该邮箱未注册")
	ErrInvalidToken      = errors.New("invalid token")

	// 邮箱验证码相关错误
	ErrEmailCodeExpired = errors.New("验证码已过期或不存在")
	ErrEmailCodeInvalid = errors.New("验证码错误")

	// 店铺相关错误
	ErrShopNotFound     = errors.New("店铺不存在")
	ErrShopUnauthorized = errors.New("店铺未授权")
	ErrShopAlreadyBound = errors.New("该店铺已被其他用户绑定")
	ErrShopNoPermission = errors.New("店铺不存在或无权限访问")
	ErrShopTokenExpired = errors.New("刷新Token已过期，请重新授权")
	ErrShopSyncing      = errors.New("正在同步中，请稍后再试")

	// 订单相关错误
	ErrOrderNotFound   = errors.New("订单不存在")
	ErrOrderCannotShip = errors.New("订单状态不允许发货")
	ErrOrderShipping   = errors.New("订单正在发货中，请勿重复操作")

	// 发货相关错误
	ErrShipmentNotFound = errors.New("发货记录不存在")
)

// 状态码对应的默认消息
var CodeMessages = map[int]string{
	CodeSuccess:       "success",
	CodeBadRequest:    "参数错误",
	CodeUnauthorized:  "未授权",
	CodeForbidden:     "禁止访问",
	CodeNotFound:      "资源不存在",
	CodeInternalError: "内部错误",

	// 认证相关
	CodeUsernameExists:    "用户名已存在",
	CodeEmailExists:       "邮箱已被注册",
	CodeInvalidUserType:   "无效的用户类型",
	CodeInvalidCredential: "用户名或密码错误",
	CodeAccountDisabled:   "账户已被禁用",
	CodeUserNotFound:      "用户不存在",
	CodeEmailNotRegister:  "该邮箱未注册",
	CodeInvalidToken:      "invalid token",

	// 邮箱验证码相关
	CodeEmailCodeExpired: "验证码已过期或不存在",
	CodeEmailCodeInvalid: "验证码错误",

	// 店铺相关
	CodeShopNotFound:     "店铺不存在",
	CodeShopUnauthorized: "店铺未授权",
	CodeShopAlreadyBound: "该店铺已被其他用户绑定",
	CodeShopNoPermission: "店铺不存在或无权限访问",
	CodeShopTokenExpired: "刷新Token已过期，请重新授权",
	CodeShopSyncing:      "正在同步中，请稍后再试",

	// 订单相关
	CodeOrderNotFound:   "订单不存在",
	CodeOrderCannotShip: "订单状态不允许发货",
	CodeOrderShipping:   "订单正在发货中，请勿重复操作",

	// 发货相关
	CodeShipmentNotFound: "发货记录不存在",
}

// GetCodeMessage 获取状态码对应的默认消息
func GetCodeMessage(code int) string {
	if msg, ok := CodeMessages[code]; ok {
		return msg
	}
	return "未知错误"
}
