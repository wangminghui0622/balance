import axios from 'axios'
import { getApiBaseUrl } from '../config/api'

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
    const token = localStorage.getItem('token')
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
    // 401错误，清除token并跳转到登录页
    if (error.response?.status === 401) {
      localStorage.removeItem('token')
      localStorage.removeItem('userId')
      window.location.href = '/login'
    }
    return Promise.reject(error)
  }
)

export default request
