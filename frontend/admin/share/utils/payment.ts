/**
 * 第三方支付工具封装
 *
 * 所有第三方支付（PayPal、支付宝、LINE Pay、VISA 等）统一抽象。
 * 当前为预留封装，各实现为占位桩（返回 not implemented），
 * 待正式对接时替换为真实 SDK 调用即可，业务层无需改动。
 */

// ==================== 支付方式定义 ====================

export interface PaymentMethodInfo {
  code: string       // 支付方式编码
  name: string       // 显示名称
  enabled: boolean   // 是否启用
  online: boolean    // 是否为线上支付
  icon: string       // 图标URL
}

/** 支付方式常量 */
export const PayMethod = {
  BANK_TRANSFER: 'bank_transfer',
  CASH: 'cash',
  PAYPAL: 'paypal',
  ALIPAY: 'alipay',
  LINEPAY: 'linepay',
  VISA: 'visa',
  WECHAT: 'wechat',
} as const

export type PayMethodType = typeof PayMethod[keyof typeof PayMethod]

/** 所有支付方式及其元信息 */
export const allPaymentMethods: PaymentMethodInfo[] = [
  {
    code: PayMethod.BANK_TRANSFER,
    name: '银行转账',
    enabled: true,
    online: false,
    icon: '',
  },
  {
    code: PayMethod.CASH,
    name: '现金',
    enabled: true,
    online: false,
    icon: '',
  },
  {
    code: PayMethod.PAYPAL,
    name: 'PayPal',
    enabled: false,
    online: true,
    icon: 'https://www.paypalobjects.com/webstatic/mktg/Logo/pp-logo-100px.png',
  },
  {
    code: PayMethod.ALIPAY,
    name: '支付宝',
    enabled: false,
    online: true,
    icon: 'https://gw.alipayobjects.com/mdn/rms_0c75a8/afts/img/A*V3ICRJ-4bDcAAAAAAAAAAAAAARQnAQ',
  },
  {
    code: PayMethod.LINEPAY,
    name: 'LINE Pay',
    enabled: false,
    online: true,
    icon: 'https://scdn.line-apps.com/linepay/portal/v-240930/portal/img/sp/sp_logo_linepay_white.png',
  },
  {
    code: PayMethod.VISA,
    name: 'VISA',
    enabled: false,
    online: true,
    icon: 'https://upload.wikimedia.org/wikipedia/commons/thumb/5/5e/Visa_Inc._logo.svg/200px-Visa_Inc._logo.svg.png',
  },
  {
    code: PayMethod.WECHAT,
    name: '微信支付',
    enabled: false,
    online: true,
    icon: '',
  },
]

/** 获取当前启用的支付方式 */
export function getEnabledPaymentMethods(): PaymentMethodInfo[] {
  return allPaymentMethods.filter(m => m.enabled)
}

/** 获取当前启用的线下支付方式 */
export function getOfflinePaymentMethods(): PaymentMethodInfo[] {
  return allPaymentMethods.filter(m => m.enabled && !m.online)
}

/** 获取当前启用的线上支付方式 */
export function getOnlinePaymentMethods(): PaymentMethodInfo[] {
  return allPaymentMethods.filter(m => m.enabled && m.online)
}

/** 根据编码获取支付方式名称 */
export function getPaymentMethodName(code: string): string {
  return allPaymentMethods.find(m => m.code === code)?.name || code
}

/** 判断支付方式是否启用 */
export function isPaymentMethodEnabled(code: string): boolean {
  return allPaymentMethods.find(m => m.code === code)?.enabled ?? false
}

/** 判断是否为线上支付 */
export function isOnlinePayment(code: string): boolean {
  return allPaymentMethods.find(m => m.code === code)?.online ?? false
}

// ==================== 支付网关接口 ====================

export interface CreatePaymentParams {
  orderNo: string       // 业务订单号
  amount: number        // 支付金额
  currency: string      // 货币代码
  description: string   // 支付描述
  returnUrl?: string    // 支付成功跳转
  cancelUrl?: string    // 支付取消跳转
}

export interface CreatePaymentResult {
  paymentId: string     // 第三方支付单号
  paymentUrl: string    // 跳转支付URL
  qrCodeUrl?: string    // 二维码URL
  status: string        // created / pending
}

export interface PaymentGateway {
  name: string
  createPayment(params: CreatePaymentParams): Promise<CreatePaymentResult>
}

// ==================== 各渠道占位实现 ====================

class PayPalGateway implements PaymentGateway {
  name = 'PayPal'
  async createPayment(_params: CreatePaymentParams): Promise<CreatePaymentResult> {
    // TODO: 对接 PayPal JS SDK / REST API
    throw new Error('PayPal 支付暂未开通，请使用线下充值')
  }
}

class AlipayGateway implements PaymentGateway {
  name = 'Alipay'
  async createPayment(_params: CreatePaymentParams): Promise<CreatePaymentResult> {
    // TODO: 对接支付宝 SDK（通常由后端创建订单，前端跳转）
    throw new Error('支付宝支付暂未开通，请使用线下充值')
  }
}

class LinePayGateway implements PaymentGateway {
  name = 'LINE Pay'
  async createPayment(_params: CreatePaymentParams): Promise<CreatePaymentResult> {
    // TODO: 对接 LINE Pay API（后端创建，前端跳转）
    throw new Error('LINE Pay 支付暂未开通，请使用线下充值')
  }
}

class VisaGateway implements PaymentGateway {
  name = 'VISA'
  async createPayment(_params: CreatePaymentParams): Promise<CreatePaymentResult> {
    // TODO: 对接 Stripe / TapPay 信用卡支付
    throw new Error('VISA 信用卡支付暂未开通，请使用线下充值')
  }
}

class WechatPayGateway implements PaymentGateway {
  name = 'WechatPay'
  async createPayment(_params: CreatePaymentParams): Promise<CreatePaymentResult> {
    // TODO: 对接微信支付 JSAPI / Native
    throw new Error('微信支付暂未开通，请使用线下充值')
  }
}

// ==================== 支付网关注册中心 ====================

const gatewayRegistry: Record<string, PaymentGateway> = {
  [PayMethod.PAYPAL]: new PayPalGateway(),
  [PayMethod.ALIPAY]: new AlipayGateway(),
  [PayMethod.LINEPAY]: new LinePayGateway(),
  [PayMethod.VISA]: new VisaGateway(),
  [PayMethod.WECHAT]: new WechatPayGateway(),
}

/** 根据支付方式获取对应网关 */
export function getPaymentGateway(method: string): PaymentGateway | null {
  return gatewayRegistry[method] || null
}

/** 注册自定义支付网关（用于替换占位实现） */
export function registerPaymentGateway(method: string, gateway: PaymentGateway): void {
  gatewayRegistry[method] = gateway
}
