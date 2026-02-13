import request from '../utils/request'

// API前缀常量
const SHOPOWER_PREFIX = '/api/v1/balance/admin/shopower'
const OPERATOR_PREFIX = '/api/v1/balance/admin/operator'
const PLATFORM_PREFIX = '/api/v1/balance/admin/platform'

// 结算记录
export interface Settlement {
  id: number
  settlement_no: string
  shop_id: number
  order_sn: string
  order_id: number
  shop_owner_id: number
  operator_id: number
  currency: string
  escrow_amount: string
  goods_cost: string
  shipping_cost: string
  total_cost: string
  profit: string
  platform_share_rate: string
  operator_share_rate: string
  shop_owner_share_rate: string
  platform_share: string
  operator_share: string
  shop_owner_share: string
  operator_income: string
  status: number
  settled_at: string | null
  remark: string
  created_at: string
  updated_at: string
}

// 结算列表响应
export interface SettlementListResponse {
  code: number
  message: string
  data: {
    list: Settlement[]
    total: number
    page: number
    page_size: number
  }
}

// 结算统计
export interface SettlementStats {
  total_settled: string
  total_pending: number
  total_profit: string
}

// 结算统计响应
export interface SettlementStatsResponse {
  code: number
  message: string
  data: SettlementStats
}

// 运营结算API
export const operatorSettlementApi = {
  // 获取结算记录
  getSettlements: (params?: {
    page?: number
    page_size?: number
  }): Promise<SettlementListResponse> => {
    return request.get(`${OPERATOR_PREFIX}/settlements`, { params })
      .then((res: any) => ({
        code: res?.code ?? 500,
        message: res?.message || '',
        data: res?.data || { list: [], total: 0, page: 1, page_size: 10 }
      } as SettlementListResponse))
  },

  // 获取结算统计
  getSettlementStats: (): Promise<SettlementStatsResponse> => {
    return request.get(`${OPERATOR_PREFIX}/settlements/stats`)
      .then((res: any) => ({
        code: res?.code ?? 500,
        message: res?.message || '',
        data: res?.data || { total_settled: '0', total_pending: 0, total_profit: '0' }
      } as SettlementStatsResponse))
  }
}

// 店主结算API
export const shopowerSettlementApi = {
  // 获取结算记录
  getSettlements: (params?: {
    page?: number
    page_size?: number
  }): Promise<SettlementListResponse> => {
    return request.get(`${SHOPOWER_PREFIX}/settlements`, { params })
      .then((res: any) => ({
        code: res?.code ?? 500,
        message: res?.message || '',
        data: res?.data || { list: [], total: 0, page: 1, page_size: 10 }
      } as SettlementListResponse))
  },

  // 获取结算统计
  getSettlementStats: (): Promise<SettlementStatsResponse> => {
    return request.get(`${SHOPOWER_PREFIX}/settlements/stats`)
      .then((res: any) => ({
        code: res?.code ?? 500,
        message: res?.message || '',
        data: res?.data || { total_settled: '0', total_pending: 0, total_profit: '0' }
      } as SettlementStatsResponse))
  }
}

// 平台结算API
export const platformSettlementApi = {
  // 获取结算记录
  getSettlements: (params?: {
    shop_id?: number
    operator_id?: number
    status?: number
    page?: number
    page_size?: number
  }): Promise<SettlementListResponse> => {
    return request.get(`${PLATFORM_PREFIX}/settlements`, { params })
      .then((res: any) => ({
        code: res?.code ?? 500,
        message: res?.message || '',
        data: res?.data || { list: [], total: 0, page: 1, page_size: 10 }
      } as SettlementListResponse))
  },

  // 获取结算统计
  getSettlementStats: (): Promise<SettlementStatsResponse> => {
    return request.get(`${PLATFORM_PREFIX}/settlements/stats`)
      .then((res: any) => ({
        code: res?.code ?? 500,
        message: res?.message || '',
        data: res?.data || { total_settled: '0', total_pending: 0, total_profit: '0' }
      } as SettlementStatsResponse))
  }
}

// 结算状态常量
export const SettlementStatus = {
  PENDING: 0,    // 待结算
  COMPLETED: 1,  // 已结算
  CANCELLED: 2   // 已取消
}

export const SettlementStatusText: Record<number, string> = {
  0: '待结算',
  1: '已结算',
  2: '已取消'
}
