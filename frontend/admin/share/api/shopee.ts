import request from '../utils/request'

export interface ShopeeAuthURLResponse {
  code: number
  message: string
  auth_url?: string
  callback?: string
  is_sandbox?: boolean
}

export interface ShopeeAuthCallbackResponse {
  code: number
  message: string
  data: {
    shop_id: number
    access_token: string
    refresh_token: string
    expire_in: number
    expire_at: string
  }
}

export interface ShopeeAuthConfigResponse {
  code: number
  message: string
  data: {
    partnerID: number
    partnerKey: string
    redirect: string
    isSandbox: boolean
  }
}

export interface ShopeeRebindSendCodeResponse {
  code: number
  message: string
  data: {
    shop_id: number
    email: string
  }
}

export interface ShopeeRebindVerifyResponse {
  code: number
  message: string
}

export interface ShopeeRebindConfirmResponse {
  code: number
  message: string
}

export interface ShopeeRebindCancelResponse {
  code: number
  message: string
}

export interface ShopeeShop {
  id: number
  shopId: number
  shopIdStr: string
  shopName: string
  shopSlug: string | null
  region: string
  partnerId: number
  authStatus: number
  status: number
  suspensionStatus: number
  isCbShop: boolean
  isCodShop: boolean
  isPreferredPlusShop: boolean
  isShopeeVerified: boolean
  ratingStar: number
  ratingBad: number
  ratingGood: number
  ratingNormal: number
  itemCount: number
  followerCount: number
  responseRate: number
  responseTime: number
  cancellationRate: number
  totalSales: number
  totalOrders: number
  totalViews: number
  dailySales: number
  monthlySales: number
  yearlySales: number
  currency: string
  balance: number
  pendingBalance: number
  withdrawnBalance: number
  contactEmail: string | null
  contactPhone: string | null
  country: string | null
  city: string | null
  address: string | null
  zipcode: string | null
  autoSync: boolean
  syncInterval: number
  syncItems: boolean
  syncOrders: boolean
  syncLogistics: boolean
  syncFinance: boolean
  isPrimary: boolean
  lastSyncAt: string | null
  nextSyncAt: string | null
  shopCreatedAt: string | null
  createdAt: string
  updatedAt: string
}

export interface ShopeeShopListResponse {
  code: number
  message: string
  data: {
    list: ShopeeShop[]
    total: number
    page: number
    page_size: number
  }
}

// API前缀常量
const SHOPOWER_PREFIX = '/api/v1/balance/admin/shopower'
const OPERATOR_PREFIX = '/api/v1/balance/admin/operator'
const PLATFORM_PREFIX = '/api/v1/balance/admin/platform'

export const shopeeApi = {
  // 获取配置
  getAuthConfig: (): Promise<ShopeeAuthConfigResponse> => {
    return request.post('/api/v1/balance/admin/shopee/auth/cfg')
  },

  // 获取授权 URL（店主专用）
  getAuthURL: (region?: string): Promise<ShopeeAuthURLResponse> => {
    return request
      .get(`${SHOPOWER_PREFIX}/shops/auth-url`, {
        params: region ? { region } : undefined
      })
      .then((res: any) => {
        return {
          code: res?.code ?? 500,
          message: res?.message || '',
          auth_url: res?.data?.url
        } as ShopeeAuthURLResponse
      })
  },

  // 发送换绑验证码
  sendRebindCode: (shopId: number): Promise<ShopeeRebindSendCodeResponse> => {
    return request.post('/api/v1/balance/admin/shopee/auth/rebind/send-code', {
      shop_id: shopId
    })
  },

  // 验证换绑验证码
  verifyRebindCode: (shopId: number, code: string): Promise<ShopeeRebindVerifyResponse> => {
    return request.post('/api/v1/balance/admin/shopee/auth/rebind/verify', {
      shop_id: shopId,
      code: code
    })
  },

  // 确认换绑
  confirmRebind: (shopId: number, code: string, newAdminId: number): Promise<ShopeeRebindConfirmResponse> => {
    return request.post('/api/v1/balance/admin/shopee/auth/rebind/confirm', {
      shop_id: shopId,
      code: code,
      new_admin_id: newAdminId
    })
  },

  // 取消换绑
  cancelRebind: (shopId: number): Promise<ShopeeRebindCancelResponse> => {
    return request.post('/api/v1/balance/admin/shopee/auth/rebind/cancel', {
      shop_id: shopId
    })
  },

  // 获取店铺列表（店主专用）
  getShopList: (params?: Record<string, any>): Promise<ShopeeShopListResponse> => {
    return request
      .get(`${SHOPOWER_PREFIX}/shops`, {
        params: params || {}
      })
      .then((res: any) => {
        return {
          code: res?.code ?? 500,
          message: res?.message || '',
          data: {
            list: res?.data?.list || [],
            total: res?.data?.total || 0,
            page: res?.data?.page || 1,
            page_size: res?.data?.page_size || 10
          }
        } as ShopeeShopListResponse
      })
  }
}

// 运营专用API
export const operatorShopeeApi = {
  // 获取店铺列表（运营可查看所有店铺）
  getShopList: (params?: Record<string, any>): Promise<ShopeeShopListResponse> => {
    return request
      .get(`${OPERATOR_PREFIX}/shops`, {
        params: params || {}
      })
      .then((res: any) => {
        return {
          code: res?.code ?? 500,
          message: res?.message || '',
          data: {
            list: res?.data?.list || [],
            total: res?.data?.total || 0,
            page: res?.data?.page || 1,
            page_size: res?.data?.page_size || 10
          }
        } as ShopeeShopListResponse
      })
  }
}

// 平台专用API
export const platformShopeeApi = {
  // 获取店铺列表（平台可查看所有店铺）
  getShopList: (params?: Record<string, any>): Promise<ShopeeShopListResponse> => {
    return request
      .get(`${PLATFORM_PREFIX}/shops`, {
        params: params || {}
      })
      .then((res: any) => {
        return {
          code: res?.code ?? 500,
          message: res?.message || '',
          data: {
            list: res?.data?.list || [],
            total: res?.data?.total || 0,
            page: res?.data?.page || 1,
            page_size: res?.data?.page_size || 10
          }
        } as ShopeeShopListResponse
      })
  }
}

// 导出单独的函数方便使用
export const getAuthURL = shopeeApi.getAuthURL
export const getShopList = shopeeApi.getShopList
