package utils

import (
	"context"
	"fmt"

	"github.com/shopspring/decimal"
)

// ==================== 支付网关统一接口 ====================
//
// 所有第三方支付（PayPal、支付宝、LINE Pay、VISA 等）统一抽象为 PaymentGateway 接口。
// 当前为预留封装，各实现返回 ErrPaymentNotImplemented，
// 待正式对接时替换为真实 SDK 调用即可，业务层无需改动。

// ErrPaymentNotImplemented 支付渠道尚未对接
var ErrPaymentNotImplemented = fmt.Errorf("该支付渠道暂未开通，请使用线下充值")

// PaymentMethod 支付方式常量
const (
	PayMethodBankTransfer = "bank_transfer" // 银行转账（线下）
	PayMethodCash         = "cash"          // 现金（线下）
	PayMethodPayPal       = "paypal"        // PayPal
	PayMethodAlipay       = "alipay"        // 支付宝
	PayMethodLinePay      = "linepay"       // LINE Pay
	PayMethodVisa         = "visa"          // VISA 信用卡
	PayMethodWechat       = "wechat"        // 微信支付
)

// PaymentMethodInfo 支付方式元信息
type PaymentMethodInfo struct {
	Code    string `json:"code"`    // 支付方式编码
	Name    string `json:"name"`    // 显示名称
	Enabled bool   `json:"enabled"` // 是否启用
	Online  bool   `json:"online"`  // 是否为线上支付
	Icon    string `json:"icon"`    // 图标URL（前端用）
}

// AllPaymentMethods 返回所有支付方式及其启用状态
func AllPaymentMethods() []PaymentMethodInfo {
	return []PaymentMethodInfo{
		{Code: PayMethodBankTransfer, Name: "银行转账", Enabled: true, Online: false, Icon: ""},
		{Code: PayMethodCash, Name: "现金", Enabled: true, Online: false, Icon: ""},
		{Code: PayMethodPayPal, Name: "PayPal", Enabled: false, Online: true, Icon: ""},
		{Code: PayMethodAlipay, Name: "支付宝", Enabled: false, Online: true, Icon: ""},
		{Code: PayMethodLinePay, Name: "LINE Pay", Enabled: false, Online: true, Icon: ""},
		{Code: PayMethodVisa, Name: "VISA", Enabled: false, Online: true, Icon: ""},
		{Code: PayMethodWechat, Name: "微信支付", Enabled: false, Online: true, Icon: ""},
	}
}

// EnabledPaymentMethods 返回当前启用的支付方式
func EnabledPaymentMethods() []PaymentMethodInfo {
	var result []PaymentMethodInfo
	for _, m := range AllPaymentMethods() {
		if m.Enabled {
			result = append(result, m)
		}
	}
	return result
}

// IsOnlinePayment 判断是否为线上支付方式
func IsOnlinePayment(method string) bool {
	for _, m := range AllPaymentMethods() {
		if m.Code == method {
			return m.Online
		}
	}
	return false
}

// IsPaymentMethodEnabled 判断支付方式是否启用
func IsPaymentMethodEnabled(method string) bool {
	for _, m := range AllPaymentMethods() {
		if m.Code == method {
			return m.Enabled
		}
	}
	return false
}

// ==================== 支付网关接口定义 ====================

// CreatePaymentRequest 创建支付请求参数
type CreatePaymentRequest struct {
	OrderNo     string          // 业务订单号（充值申请单号）
	Amount      decimal.Decimal // 支付金额
	Currency    string          // 货币代码 TWD/USD 等
	Description string          // 支付描述
	ReturnURL   string          // 支付成功跳转URL
	CancelURL   string          // 支付取消跳转URL
	NotifyURL   string          // 异步回调通知URL
	Extra       map[string]string // 扩展参数
}

// CreatePaymentResponse 创建支付响应
type CreatePaymentResponse struct {
	PaymentID  string // 第三方支付单号
	PaymentURL string // 跳转支付页面URL
	QRCodeURL  string // 二维码支付URL（适用于扫码支付）
	Status     string // 状态: created/pending
}

// PaymentNotification 支付回调通知
type PaymentNotification struct {
	PaymentID   string          // 第三方支付单号
	OrderNo     string          // 业务订单号
	Amount      decimal.Decimal // 实际支付金额
	Currency    string          // 货币
	Status      string          // paid/failed/cancelled
	RawData     string          // 原始回调数据（JSON）
}

// QueryPaymentResponse 查询支付结果
type QueryPaymentResponse struct {
	PaymentID string          // 第三方支付单号
	OrderNo   string          // 业务订单号
	Amount    decimal.Decimal // 支付金额
	Currency  string          // 货币
	Status    string          // paid/pending/failed/cancelled
}

// PaymentGateway 第三方支付网关统一接口
type PaymentGateway interface {
	// Name 返回支付渠道名称
	Name() string

	// CreatePayment 创建支付订单，返回支付链接
	CreatePayment(ctx context.Context, req *CreatePaymentRequest) (*CreatePaymentResponse, error)

	// QueryPayment 查询支付状态
	QueryPayment(ctx context.Context, paymentID string) (*QueryPaymentResponse, error)

	// VerifyNotification 验证并解析回调通知（验签）
	VerifyNotification(ctx context.Context, rawBody []byte) (*PaymentNotification, error)
}

// ==================== 各渠道占位实现 ====================

// PayPalGateway PayPal 支付（占位）
type PayPalGateway struct{}

func NewPayPalGateway() *PayPalGateway { return &PayPalGateway{} }

func (g *PayPalGateway) Name() string { return "PayPal" }

func (g *PayPalGateway) CreatePayment(_ context.Context, _ *CreatePaymentRequest) (*CreatePaymentResponse, error) {
	// TODO: 对接 PayPal REST API v2
	// https://developer.paypal.com/docs/api/orders/v2/
	return nil, ErrPaymentNotImplemented
}

func (g *PayPalGateway) QueryPayment(_ context.Context, _ string) (*QueryPaymentResponse, error) {
	return nil, ErrPaymentNotImplemented
}

func (g *PayPalGateway) VerifyNotification(_ context.Context, _ []byte) (*PaymentNotification, error) {
	return nil, ErrPaymentNotImplemented
}

// AlipayGateway 支付宝（占位）
type AlipayGateway struct{}

func NewAlipayGateway() *AlipayGateway { return &AlipayGateway{} }

func (g *AlipayGateway) Name() string { return "Alipay" }

func (g *AlipayGateway) CreatePayment(_ context.Context, _ *CreatePaymentRequest) (*CreatePaymentResponse, error) {
	// TODO: 对接支付宝国际版 / 支付宝当面付
	// https://global.alipay.com/docs/ac/web/create
	return nil, ErrPaymentNotImplemented
}

func (g *AlipayGateway) QueryPayment(_ context.Context, _ string) (*QueryPaymentResponse, error) {
	return nil, ErrPaymentNotImplemented
}

func (g *AlipayGateway) VerifyNotification(_ context.Context, _ []byte) (*PaymentNotification, error) {
	return nil, ErrPaymentNotImplemented
}

// LinePayGateway LINE Pay（占位）
type LinePayGateway struct{}

func NewLinePayGateway() *LinePayGateway { return &LinePayGateway{} }

func (g *LinePayGateway) Name() string { return "LINE Pay" }

func (g *LinePayGateway) CreatePayment(_ context.Context, _ *CreatePaymentRequest) (*CreatePaymentResponse, error) {
	// TODO: 对接 LINE Pay API v3
	// https://pay.line.me/documents/online_v3.html
	return nil, ErrPaymentNotImplemented
}

func (g *LinePayGateway) QueryPayment(_ context.Context, _ string) (*QueryPaymentResponse, error) {
	return nil, ErrPaymentNotImplemented
}

func (g *LinePayGateway) VerifyNotification(_ context.Context, _ []byte) (*PaymentNotification, error) {
	return nil, ErrPaymentNotImplemented
}

// VisaGateway VISA 信用卡（占位）
type VisaGateway struct{}

func NewVisaGateway() *VisaGateway { return &VisaGateway{} }

func (g *VisaGateway) Name() string { return "VISA" }

func (g *VisaGateway) CreatePayment(_ context.Context, _ *CreatePaymentRequest) (*CreatePaymentResponse, error) {
	// TODO: 对接信用卡支付通道（Stripe / TapPay / 绿界 ECPay 等）
	// https://stripe.com/docs/api/payment_intents
	return nil, ErrPaymentNotImplemented
}

func (g *VisaGateway) QueryPayment(_ context.Context, _ string) (*QueryPaymentResponse, error) {
	return nil, ErrPaymentNotImplemented
}

func (g *VisaGateway) VerifyNotification(_ context.Context, _ []byte) (*PaymentNotification, error) {
	return nil, ErrPaymentNotImplemented
}

// WechatPayGateway 微信支付（占位）
type WechatPayGateway struct{}

func NewWechatPayGateway() *WechatPayGateway { return &WechatPayGateway{} }

func (g *WechatPayGateway) Name() string { return "WechatPay" }

func (g *WechatPayGateway) CreatePayment(_ context.Context, _ *CreatePaymentRequest) (*CreatePaymentResponse, error) {
	// TODO: 对接微信支付 API v3
	// https://pay.weixin.qq.com/docs/merchant/apis/native-payment/direct-jsons/native-prepay.html
	return nil, ErrPaymentNotImplemented
}

func (g *WechatPayGateway) QueryPayment(_ context.Context, _ string) (*QueryPaymentResponse, error) {
	return nil, ErrPaymentNotImplemented
}

func (g *WechatPayGateway) VerifyNotification(_ context.Context, _ []byte) (*PaymentNotification, error) {
	return nil, ErrPaymentNotImplemented
}

// ==================== 支付网关注册中心 ====================

var paymentGateways = map[string]PaymentGateway{
	PayMethodPayPal:  NewPayPalGateway(),
	PayMethodAlipay:  NewAlipayGateway(),
	PayMethodLinePay: NewLinePayGateway(),
	PayMethodVisa:    NewVisaGateway(),
	PayMethodWechat:  NewWechatPayGateway(),
}

// GetPaymentGateway 根据支付方式获取对应网关
func GetPaymentGateway(method string) (PaymentGateway, error) {
	gw, ok := paymentGateways[method]
	if !ok {
		return nil, fmt.Errorf("不支持的支付方式: %s", method)
	}
	return gw, nil
}

// RegisterPaymentGateway 注册自定义支付网关（用于替换占位实现）
func RegisterPaymentGateway(method string, gw PaymentGateway) {
	paymentGateways[method] = gw
}
