// API配置
export const API_CONFIG = {
  // 本地开发环境
  LOCAL: 'http://localhost:19090',
  // 线上生产环境
  PRODUCTION: 'https://kx9y.com'
}

// 获取当前API地址
export function getApiBaseUrl(): string {
  // #ifdef H5
  // H5环境：开发环境使用proxy，生产环境使用线上地址
  if (import.meta.env.MODE === 'development') {
    return ''  // 开发环境使用 vite proxy
  }
  
  // 生产环境：优先从storage读取配置（支持手动切换）
  const apiEnv = uni.getStorageSync('api_env')
  
  if (apiEnv === 'local') {
    return API_CONFIG.LOCAL
  }
  
  return API_CONFIG.PRODUCTION
  // #endif
  
  // #ifndef H5
  // 非H5环境（小程序等）：直接使用线上地址
  const apiEnv = uni.getStorageSync('api_env')
  if (apiEnv === 'local') {
    return API_CONFIG.LOCAL
  }
  return API_CONFIG.PRODUCTION
  // #endif
}

// 设置API环境
export function setApiEnv(env: 'local' | 'production') {
  uni.setStorageSync('api_env', env)
}

// 获取当前API环境
export function getApiEnv(): 'local' | 'production' {
  const apiEnv = uni.getStorageSync('api_env')
  return (apiEnv === 'production' ? 'production' : 'local') as 'local' | 'production'
}
