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

export const shopeeApi = {
  // 获取配置
  getAuthConfig: (): Promise<ShopeeAuthConfigResponse> => {
    return request.post('/api/v1/balance/admin/shopee/auth/cfg')
  },

  // 获取授权 URL
  getAuthURL: (shopId?: number): Promise<ShopeeAuthURLResponse> => {
    // Fetch configuration first, then use POST request with JSON data
    return shopeeApi.getAuthConfig().then(configRes => {
      if (configRes.code === 200) {
        const configData = configRes.data;
		console.log(configData.redirect)
		console.log(configData.partnerID)
        const data = {
          partnerID: configData.partnerID,
          partnerKey: configData.partnerKey,
          redirect: configData.redirect,
          isSandbox: configData.isSandbox,
          ...(shopId !== undefined && { shopId: shopId }) // Only include shopId if provided
        };
        return request.post('/api/v1/balance/admin/shopee/auth/url', data);
      } else {
        throw new Error(configRes.message || '获取配置失败');
      }
    });
  }
}
