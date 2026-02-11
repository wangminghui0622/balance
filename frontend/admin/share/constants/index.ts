/**
 * 应用常量定义
 */

// ==================== LocalStorage Keys ====================
export const STORAGE_KEYS = {
  TOKEN: 'token',
  USER_ID: 'userId',
  USER_TYPE: 'userType'
} as const

// ==================== 用户类型 ====================
// 字符串类型（用于用户ID前缀匹配）
export const USER_TYPE = {
  PLATFORM: '9',    // 平台管理员
  OPERATOR: '5',    // 运营人员
  SHOPOWNER: '1'    // 店主
} as const

// 数字类型（用于API请求）
export const USER_TYPE_NUM = {
  PLATFORM: 9,    // 平台管理员
  OPERATOR: 5,    // 运营人员
  SHOPOWNER: 1    // 店主
} as const

export type UserType = typeof USER_TYPE[keyof typeof USER_TYPE]
export type UserTypeNum = typeof USER_TYPE_NUM[keyof typeof USER_TYPE_NUM]

// ==================== 路由路径 ====================
export const ROUTE_PATH = {
  LOGIN: '/login',
  REGISTER: '/register',
  PLATFORM: '/platform',
  OPERATOR: '/operator',
  SHOPOWNER: '/shopowner'
} as const

// ==================== HTTP 状态码 ====================
export const HTTP_STATUS = {
  OK: 0,             // API 成功响应 code
  HTTP_OK: 200,      // HTTP 200 状态码
  UNAUTHORIZED: 401
} as const

// ==================== 用户类型到路由的映射 ====================
export const USER_TYPE_TO_ROUTE: Record<UserType, string> = {
  [USER_TYPE.PLATFORM]: ROUTE_PATH.PLATFORM,
  [USER_TYPE.OPERATOR]: ROUTE_PATH.OPERATOR,
  [USER_TYPE.SHOPOWNER]: ROUTE_PATH.SHOPOWNER
} as const

// ==================== 工具函数 ====================
/**
 * 根据用户ID获取用户类型
 */
export function getUserType(userId: string | null): UserType | null {
  if (!userId) return null
  
  if (userId.startsWith(USER_TYPE.PLATFORM)) {
    return USER_TYPE.PLATFORM
  } else if (userId.startsWith(USER_TYPE.OPERATOR)) {
    return USER_TYPE.OPERATOR
  } else if (userId.startsWith(USER_TYPE.SHOPOWNER)) {
    return USER_TYPE.SHOPOWNER
  }
  
  return null
}

/**
 * 根据用户类型获取路由路径
 */
export function getRouteByUserType(userType: UserType | null): string {
  if (!userType) return ROUTE_PATH.LOGIN
  return USER_TYPE_TO_ROUTE[userType] || ROUTE_PATH.LOGIN
}
