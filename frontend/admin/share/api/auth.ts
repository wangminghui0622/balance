import request from '../utils/request'

export interface LoginRequest {
  username: string
  password: string
}

export interface RegisterRequest {
  username: string
  password: string
  email: string
  userType: number // 1=店铺, 5=运营
}

export interface LoginResponse {
  code: number
  message: string
  data: {
    token: string
    userId: number
  }
}

export interface RegisterResponse {
  code: number
  message: string
}

export const authApi = {
  login: (data: LoginRequest): Promise<LoginResponse> => {
    return request.post('/api/v1/balance/admin/auth/login', data)
  },
  register: (data: RegisterRequest): Promise<RegisterResponse> => {
    return request.post('/api/v1/balance/admin/auth/register', data)
  }
}
