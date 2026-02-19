// API配置
export const API_CONFIG = {
  // 本地开发环境
  LOCAL: 'http://localhost:19090',
  // 线上生产环境
  PRODUCTION: 'https://kx9y.com'
}

// 获取当前API地址
export function getApiBaseUrl(): string {
  // 开发环境：自动使用本地（通过 vite proxy，所以用空字符串）
  if (import.meta.env.MODE === 'development') {
    return ''  // 开发环境使用 vite proxy，不需要完整URL
  }
  
  // 生产环境：优先从localStorage读取配置（支持手动切换）
  const apiEnv = localStorage.getItem('api_env')
  
  if (apiEnv === 'local') {
    // 生产环境但手动切换到本地（用于测试）
    return API_CONFIG.LOCAL
  }
  
  // 生产环境默认使用线上
  return API_CONFIG.PRODUCTION
}

// 设置API环境
export function setApiEnv(env: 'local' | 'production') {
  localStorage.setItem('api_env', env)
  // 刷新页面以应用新配置
  window.location.reload()
}

// 获取当前API环境
export function getApiEnv(): 'local' | 'production' {
  const apiEnv = localStorage.getItem('api_env')
  return (apiEnv === 'production' ? 'production' : 'local') as 'local' | 'production'
}
