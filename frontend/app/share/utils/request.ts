import { getApiBaseUrl } from '../config/api'
import { STORAGE_KEYS, HTTP_STATUS, ROUTE_PATH } from '../constants'

interface RequestOptions {
  url: string
  method?: 'GET' | 'POST' | 'PUT' | 'DELETE'
  data?: any
  header?: Record<string, string>
}

interface ApiResponse<T = any> {
  code: number
  message: string
  data: T
}

// 获取完整URL
function getFullUrl(url: string): string {
  const baseUrl = getApiBaseUrl()
  return baseUrl + url
}

// 统一请求方法
export function request<T = any>(options: RequestOptions): Promise<ApiResponse<T>> {
  return new Promise((resolve, reject) => {
    const token = uni.getStorageSync(STORAGE_KEYS.TOKEN)
    
    const header: Record<string, string> = {
      'Content-Type': 'application/json',
      ...options.header
    }
    
    if (token) {
      header['Authorization'] = `Bearer ${token}`
    }
    
    uni.request({
      url: getFullUrl(options.url),
      method: options.method || 'GET',
      data: options.data,
      header,
      timeout: 10000,
      success: (res) => {
        const statusCode = res.statusCode
        
        // 401 未授权
        if (statusCode === HTTP_STATUS.UNAUTHORIZED) {
          // 如果是登录接口本身的 401，交给调用方处理
          if (options.url.includes('/auth/login')) {
            reject({ statusCode, message: '用户名或密码错误' })
            return
          }
          
          // 其他接口的 401 视为登录失效
          uni.removeStorageSync(STORAGE_KEYS.TOKEN)
          uni.removeStorageSync(STORAGE_KEYS.USER_ID)
          uni.removeStorageSync(STORAGE_KEYS.USER_TYPE)
          uni.reLaunch({ url: ROUTE_PATH.LOGIN })
          reject({ statusCode, message: '登录已过期' })
          return
        }
        
        // 其他HTTP错误
        if (statusCode >= 400) {
          reject({ statusCode, message: (res.data as any)?.message || '请求失败' })
          return
        }
        
        resolve(res.data as ApiResponse<T>)
      },
      fail: (err) => {
        reject({ message: err.errMsg || '网络请求失败' })
      }
    })
  })
}

// GET请求
export function get<T = any>(url: string, data?: any): Promise<ApiResponse<T>> {
  return request<T>({ url, method: 'GET', data })
}

// POST请求
export function post<T = any>(url: string, data?: any): Promise<ApiResponse<T>> {
  return request<T>({ url, method: 'POST', data })
}

// PUT请求
export function put<T = any>(url: string, data?: any): Promise<ApiResponse<T>> {
  return request<T>({ url, method: 'PUT', data })
}

// DELETE请求
export function del<T = any>(url: string, data?: any): Promise<ApiResponse<T>> {
  return request<T>({ url, method: 'DELETE', data })
}

export default {
  request,
  get,
  post,
  put,
  del
}
