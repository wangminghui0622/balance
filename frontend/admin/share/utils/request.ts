import axios from 'axios'
import { getApiBaseUrl } from '../config/api'
import { STORAGE_KEYS, HTTP_STATUS, ROUTE_PATH } from '../constants'

// 从配置获取API地址（支持本地和线上切换）
// 优先使用环境变量，否则从配置获取（支持动态切换）
const getBaseURL = () => {
  return import.meta.env.VITE_API_BASE_URL || getApiBaseUrl()
}

const request = axios.create({
  baseURL: getBaseURL(),
  timeout: 10000
})

// 动态更新baseURL（切换环境时使用）
export function updateBaseURL() {
  request.defaults.baseURL = getBaseURL()
}

// 请求拦截器
request.interceptors.request.use(
  (config) => {
    // 从localStorage获取token并添加到请求头
    const token = localStorage.getItem(STORAGE_KEYS.TOKEN)
    if (token) {
      config.headers.Authorization = `Bearer ${token}`
    }
    return config
  },
  (error) => {
    return Promise.reject(error)
  }
)

// 响应拦截器
request.interceptors.response.use(
  (response) => {
    return response.data
  },
  (error) => {
    // 401 错误统一处理（注意：登录接口本身的 401 不应重定向）
    if (error.response?.status === HTTP_STATUS.UNAUTHORIZED) {
      const url = error.config?.url || ''

      // 如果是登录接口本身的 401，交给调用方（如 Login.vue）自行处理并提示“用户名或密码错误”
      if (url.includes('/api/v1/balance/admin/auth/login')) {
        return Promise.reject(error)
      }

      // 其他接口的 401 视为登录失效，清除本地信息并跳转到登录页
      localStorage.removeItem(STORAGE_KEYS.TOKEN)
      localStorage.removeItem(STORAGE_KEYS.USER_ID)
      window.location.href = ROUTE_PATH.LOGIN
    }
    return Promise.reject(error)
  }
)

export default request
