import { post, get } from '../utils/request'

// API路径前缀
const API_PREFIX = '/api/v1/balance/app'

// 登录请求
export interface LoginRequest {
  username: string
  password: string
}

// 登录响应
export interface LoginResponse {
  token: string
  user_id: string
  username: string
  user_type: number
}

// 注册请求
export interface RegisterRequest {
  username: string
  email: string
  emailCode: string
  password: string
  userType: number
  realName?: string
  phone?: string
  lineId?: string
  wechat?: string
}

// 发送验证码请求
export interface SendCodeRequest {
  email: string
  type?: string
}

// 重置密码请求
export interface ResetPasswordRequest {
  email: string
  emailCode: string
  newPassword: string
}

// 用户信息
export interface UserInfo {
  user_id: string
  username: string
  user_type: number
}

// 登录
export function login(data: LoginRequest) {
  return post<LoginResponse>(`${API_PREFIX}/auth/login`, data)
}

// 注册
export function register(data: RegisterRequest) {
  return post<LoginResponse>(`${API_PREFIX}/auth/register`, data)
}

// 发送邮箱验证码
export function sendEmailCode(data: SendCodeRequest) {
  return post<{ message: string }>(`${API_PREFIX}/auth/send-code`, data)
}

// 获取当前用户信息
export function getCurrentUser() {
  return get<UserInfo>(`${API_PREFIX}/auth/me`)
}

// 重置密码
export function resetPassword(data: ResetPasswordRequest) {
  return post<{ message: string }>(`${API_PREFIX}/auth/reset-password`, data)
}
