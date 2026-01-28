import request from '../utils/request'

export interface ShopeeAuthURLResponse {
  code: number
  message: string
  auth_url: string
  callback: string
  is_sandbox: boolean
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
  authTime: string | null
  tokenExpireAt: string | null
  lastSyncAt: string | null
  nextSyncAt: string | null
  shopCreatedAt: string | null
  createdAt: string
  updatedAt: string
}

export interface ShopeeShopListResponse {
  code: number
  message: string
  data: ShopeeShop[]
}

export const shopeeApi = {
  // 获取配置
  getAuthConfig: (): Promise<ShopeeAuthConfigResponse> => {
    return request.post('/api/v1/balance/admin/shopee/auth/cfg')
  },

  // 获取授权 URL
  getAuthURL: (): Promise<ShopeeAuthURLResponse> => {
    // 使用固定参数
    const data = {
      partnerID: 1203446,
      partnerKey: "shpk724b6a656d626b696b756345464e6b614d524664716c61525a4e4e4f466c",
      redirect: "https://kx9y.com",
      isSandbox: true
    };
    return request.post('/api/v1/balance/admin/shopee/auth/url', data);
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

  // 获取店铺列表
  getShopList: (): Promise<ShopeeShopListResponse> => {
    return request.post('/api/v1/balance/admin/shopee/shop/list')
  }
}
