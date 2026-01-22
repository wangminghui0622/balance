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

export const shopeeApi = {
  // 获取授权 URL
  getAuthURL: (shopId?: number): Promise<ShopeeAuthURLResponse> => {
    const params = shopId ? { shop_id: shopId } : {}
    return request.get('/api/v1/balance/admin/shopee/auth/url', { params })
  }
}
