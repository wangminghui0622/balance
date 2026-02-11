import request from '../utils/request'

export interface LoginRequest {
  username: string
  password: string
}

export interface RegisterRequest {
  username: string
  password: string
  email: string
  emailCode: string
  userType: number // 1=店铺, 5=运营
  realName?: string
  phone?: string
  lineId?: string
  wechat?: string
}

export interface SendCodeRequest {
  email: string
}

export interface SendCodeResponse {
  code: number
  message: string
}

export interface LoginResponse {
  code: number
  message: string
  data: {
    token: string
    userId: number
    userType: number // 1=店主, 5=运营, 9=平台
  }
}

export interface RegisterResponse {
  code: number
  message: string
}

export interface CurrentUserResponse {
  code: number
  message: string
  data: {
    id: number
    userNo: string
    userType: number
    userName: string
    email: string
    phone: string
    avatar: string
  }
}

export const authApi = {
  login: (data: LoginRequest): Promise<LoginResponse> => {
    return request.post('/api/v1/balance/admin/auth/login', data)
  },
  register: (data: RegisterRequest): Promise<RegisterResponse> => {
    return request.post('/api/v1/balance/admin/auth/register', data)
  },
  sendEmailCode: (data: SendCodeRequest): Promise<SendCodeResponse> => {
    return request.post('/api/v1/balance/admin/auth/send-code', data)
  },
  getCurrentUser: (): Promise<CurrentUserResponse> => {
    return request.get('/api/v1/balance/admin/auth/me')
  }
}
