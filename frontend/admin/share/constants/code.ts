/**
 * 业务错误码定义
 * 与后端 utils/code.go 保持一致
 */

// ==================== 通用状态码 ====================
export const CODE = {
  // 成功
  SUCCESS: 0,

  // 通用错误 (400-499)
  BAD_REQUEST: 400,
  UNAUTHORIZED: 401,
  FORBIDDEN: 403,
  NOT_FOUND: 404,

  // 服务器错误 (500-599)
  INTERNAL_ERROR: 500,

  // 认证相关错误码 (1000-1099)
  USERNAME_EXISTS: 1001,
  EMAIL_EXISTS: 1002,
  INVALID_USER_TYPE: 1003,
  INVALID_CREDENTIAL: 1004,
  ACCOUNT_DISABLED: 1005,
  USER_NOT_FOUND: 1006,
  EMAIL_NOT_REGISTER: 1007,
  INVALID_TOKEN: 1008,

  // 邮箱验证码相关错误码 (1100-1199)
  EMAIL_CODE_EXPIRED: 1101,
  EMAIL_CODE_INVALID: 1102,

  // 店铺相关错误码 (1200-1299)
  SHOP_NOT_FOUND: 1201,
  SHOP_UNAUTHORIZED: 1202,
  SHOP_ALREADY_BOUND: 1203,
  SHOP_NO_PERMISSION: 1204,
  SHOP_TOKEN_EXPIRED: 1205,
  SHOP_SYNCING: 1206,

  // 订单相关错误码 (1300-1399)
  ORDER_NOT_FOUND: 1301,
  ORDER_CANNOT_SHIP: 1302,
  ORDER_SHIPPING: 1303,

  // 发货相关错误码 (1400-1499)
  SHIPMENT_NOT_FOUND: 1401,
} as const

export type CodeType = typeof CODE[keyof typeof CODE]

// ==================== 错误码消息映射 ====================
export const CODE_MESSAGES: Record<number, string> = {
  [CODE.SUCCESS]: '成功',
  [CODE.BAD_REQUEST]: '参数错误',
  [CODE.UNAUTHORIZED]: '未授权',
  [CODE.FORBIDDEN]: '禁止访问',
  [CODE.NOT_FOUND]: '资源不存在',
  [CODE.INTERNAL_ERROR]: '内部错误',

  // 认证相关
  [CODE.USERNAME_EXISTS]: '用户名已存在',
  [CODE.EMAIL_EXISTS]: '邮箱已被注册',
  [CODE.INVALID_USER_TYPE]: '无效的用户类型',
  [CODE.INVALID_CREDENTIAL]: '用户名或密码错误',
  [CODE.ACCOUNT_DISABLED]: '账户已被禁用',
  [CODE.USER_NOT_FOUND]: '用户不存在',
  [CODE.EMAIL_NOT_REGISTER]: '该邮箱未注册',
  [CODE.INVALID_TOKEN]: '无效的令牌',

  // 邮箱验证码相关
  [CODE.EMAIL_CODE_EXPIRED]: '验证码已过期或不存在',
  [CODE.EMAIL_CODE_INVALID]: '验证码错误',

  // 店铺相关
  [CODE.SHOP_NOT_FOUND]: '店铺不存在',
  [CODE.SHOP_UNAUTHORIZED]: '店铺未授权',
  [CODE.SHOP_ALREADY_BOUND]: '该店铺已被其他用户绑定',
  [CODE.SHOP_NO_PERMISSION]: '店铺不存在或无权限访问',
  [CODE.SHOP_TOKEN_EXPIRED]: '刷新Token已过期，请重新授权',
  [CODE.SHOP_SYNCING]: '正在同步中，请稍后再试',

  // 订单相关
  [CODE.ORDER_NOT_FOUND]: '订单不存在',
  [CODE.ORDER_CANNOT_SHIP]: '订单状态不允许发货',
  [CODE.ORDER_SHIPPING]: '订单正在发货中，请勿重复操作',

  // 发货相关
  [CODE.SHIPMENT_NOT_FOUND]: '发货记录不存在',
}

// ==================== 工具函数 ====================
/**
 * 获取错误码对应的消息
 */
export function getCodeMessage(code: number): string {
  return CODE_MESSAGES[code] || '未知错误'
}

/**
 * 判断是否成功
 */
export function isSuccess(code: number): boolean {
  return code === CODE.SUCCESS
}

/**
 * 判断是否需要重新登录
 */
export function needReLogin(code: number): boolean {
  return code === CODE.UNAUTHORIZED || code === CODE.INVALID_TOKEN
}
