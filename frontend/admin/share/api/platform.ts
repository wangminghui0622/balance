import request from '../utils/request'

// API前缀常量
const PLATFORM_PREFIX = '/api/v1/balance/admin/platform'

// ==================== 结算管理 ====================

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

// 结算统计响应
export interface SettlementStatsResponse {
  code: number
  message: string
  data: {
    total_settled: string
    total_pending: number
    total_profit: string
  }
}

// 平台结算API
export const platformSettlementApi = {
  // 获取结算记录
  getSettlements: (params?: {
    shop_owner_id?: number
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
  },

  // 获取待结算订单
  getPendingSettlements: (params?: {
    page?: number
    page_size?: number
  }): Promise<any> => {
    return request.get(`${PLATFORM_PREFIX}/settlements/pending`, { params })
      .then((res: any) => ({
        code: res?.code ?? 500,
        message: res?.message || '',
        data: res?.data || { list: [], total: 0, page: 1, page_size: 10 }
      }))
  },

  // 手动触发结算处理
  processSettlement: (): Promise<any> => {
    return request.post(`${PLATFORM_PREFIX}/settlements/process`)
      .then((res: any) => ({
        code: res?.code ?? 500,
        message: res?.message || '',
        data: res?.data
      }))
  }
}

// ==================== 合作管理 ====================

// 合作关系
export interface Cooperation {
  id: number
  shop_id: number
  shop_name: string
  shop_owner_id: number
  shop_owner_name: string
  operator_id: number
  operator_name: string
  status: number
  status_text: string
  platform_share_rate: string
  operator_share_rate: string
  shop_owner_share_rate: string
  assigned_at: string
  created_at: string
}

// 合作列表响应
export interface CooperationListResponse {
  code: number
  message: string
  data: {
    list: Cooperation[]
    total: number
    page: number
    page_size: number
  }
}

// 合作统计响应
export interface CooperationStatsResponse {
  code: number
  message: string
  data: {
    total: number
    active: number
    cancelled: number
  }
}

// 创建合作请求
export interface CreateCooperationRequest {
  shop_id: number
  operator_id: number
  platform_share_rate?: number
  operator_share_rate?: number
  shop_owner_share_rate?: number
}

// 更新合作请求
export interface UpdateCooperationRequest {
  status?: number
  platform_share_rate?: number
  operator_share_rate?: number
  shop_owner_share_rate?: number
}

// 平台合作管理API
export const platformCooperationApi = {
  // 获取合作列表
  getCooperations: (params?: {
    status?: number
    keyword?: string
    page?: number
    page_size?: number
  }): Promise<CooperationListResponse> => {
    return request.get(`${PLATFORM_PREFIX}/cooperations`, { params })
      .then((res: any) => ({
        code: res?.code ?? 500,
        message: res?.message || '',
        data: res?.data || { list: [], total: 0, page: 1, page_size: 10 }
      } as CooperationListResponse))
  },

  // 创建合作关系
  createCooperation: (data: CreateCooperationRequest): Promise<any> => {
    return request.post(`${PLATFORM_PREFIX}/cooperations`, data)
      .then((res: any) => ({
        code: res?.code ?? 500,
        message: res?.message || '',
        data: res?.data
      }))
  },

  // 更新合作关系
  updateCooperation: (id: number, data: UpdateCooperationRequest): Promise<any> => {
    return request.put(`${PLATFORM_PREFIX}/cooperations/${id}`, data)
      .then((res: any) => ({
        code: res?.code ?? 500,
        message: res?.message || '',
        data: res?.data
      }))
  },

  // 取消合作关系
  cancelCooperation: (id: number): Promise<any> => {
    return request.delete(`${PLATFORM_PREFIX}/cooperations/${id}`)
      .then((res: any) => ({
        code: res?.code ?? 500,
        message: res?.message || '',
        data: res?.data
      }))
  },

  // 获取合作统计
  getCooperationStats: (): Promise<CooperationStatsResponse> => {
    return request.get(`${PLATFORM_PREFIX}/cooperations/stats`)
      .then((res: any) => ({
        code: res?.code ?? 500,
        message: res?.message || '',
        data: res?.data || { total: 0, active: 0, cancelled: 0 }
      } as CooperationStatsResponse))
  },

  // 获取运营列表（下拉选择）
  getOperatorList: (): Promise<any> => {
    return request.get(`${PLATFORM_PREFIX}/operators`)
      .then((res: any) => ({
        code: res?.code ?? 500,
        message: res?.message || '',
        data: res?.data || []
      }))
  },

  // 获取店主列表（下拉选择）
  getShopOwnerList: (): Promise<any> => {
    return request.get(`${PLATFORM_PREFIX}/shop-owners`)
      .then((res: any) => ({
        code: res?.code ?? 500,
        message: res?.message || '',
        data: res?.data || []
      }))
  }
}

// ==================== 账户管理 ====================

// 预付款账户
export interface PrepaymentAccount {
  id: number
  admin_id: number
  username: string
  email: string
  balance: string
  frozen_amount: string
  total_recharge: string
  total_consume: string
  currency: string
  status: number
  created_at: string
  updated_at: string
}

// 保证金账户
export interface DepositAccount {
  id: number
  admin_id: number
  username: string
  email: string
  balance: string
  required_amount: string
  currency: string
  status: number
  created_at: string
  updated_at: string
}

// 运营账户
export interface OperatorAccount {
  id: number
  admin_id: number
  username: string
  email: string
  balance: string
  frozen_amount: string
  total_earnings: string
  total_withdrawn: string
  currency: string
  status: number
  created_at: string
  updated_at: string
}

// 账户流水
export interface AccountTransaction {
  id: number
  transaction_no: string
  account_type: string
  admin_id: number
  transaction_type: string
  amount: string
  balance_before: string
  balance_after: string
  related_order_sn: string
  related_id: number
  remark: string
  operator_id: number
  created_at: string
}

// 账户列表响应
export interface AccountListResponse<T> {
  code: number
  message: string
  data: {
    list: T[]
    total: number
    page: number
    page_size: number
  }
}

// 账户统计响应
export interface AccountStatsResponse {
  code: number
  message: string
  data: {
    prepayment: {
      total_balance: string
      account_count: number
    }
    deposit: {
      total_balance: string
      account_count: number
    }
    operator: {
      total_balance: string
      account_count: number
    }
  }
}

// 充值请求
export interface RechargeRequest {
  admin_id: number
  amount: number
  remark?: string
}

// 平台账户管理API
export const platformAccountApi = {
  // 获取预付款账户列表
  getPrepaymentAccounts: (params?: {
    page?: number
    page_size?: number
  }): Promise<AccountListResponse<PrepaymentAccount>> => {
    return request.get(`${PLATFORM_PREFIX}/accounts/prepayment`, { params })
      .then((res: any) => ({
        code: res?.code ?? 500,
        message: res?.message || '',
        data: res?.data || { list: [], total: 0, page: 1, page_size: 10 }
      } as AccountListResponse<PrepaymentAccount>))
  },

  // 获取保证金账户列表
  getDepositAccounts: (params?: {
    page?: number
    page_size?: number
  }): Promise<AccountListResponse<DepositAccount>> => {
    return request.get(`${PLATFORM_PREFIX}/accounts/deposit`, { params })
      .then((res: any) => ({
        code: res?.code ?? 500,
        message: res?.message || '',
        data: res?.data || { list: [], total: 0, page: 1, page_size: 10 }
      } as AccountListResponse<DepositAccount>))
  },

  // 获取运营账户列表
  getOperatorAccounts: (params?: {
    page?: number
    page_size?: number
  }): Promise<AccountListResponse<OperatorAccount>> => {
    return request.get(`${PLATFORM_PREFIX}/accounts/operator`, { params })
      .then((res: any) => ({
        code: res?.code ?? 500,
        message: res?.message || '',
        data: res?.data || { list: [], total: 0, page: 1, page_size: 10 }
      } as AccountListResponse<OperatorAccount>))
  },

  // 预付款充值
  rechargePrepayment: (data: RechargeRequest): Promise<any> => {
    return request.post(`${PLATFORM_PREFIX}/accounts/prepayment/recharge`, data)
      .then((res: any) => ({
        code: res?.code ?? 500,
        message: res?.message || '',
        data: res?.data
      }))
  },

  // 缴纳保证金
  payDeposit: (data: RechargeRequest): Promise<any> => {
    return request.post(`${PLATFORM_PREFIX}/accounts/deposit/pay`, data)
      .then((res: any) => ({
        code: res?.code ?? 500,
        message: res?.message || '',
        data: res?.data
      }))
  },

  // 获取账户流水
  getAccountTransactions: (params?: {
    account_type?: string
    admin_id?: number
    page?: number
    page_size?: number
  }): Promise<AccountListResponse<AccountTransaction>> => {
    return request.get(`${PLATFORM_PREFIX}/accounts/transactions`, { params })
      .then((res: any) => ({
        code: res?.code ?? 500,
        message: res?.message || '',
        data: res?.data || { list: [], total: 0, page: 1, page_size: 10 }
      } as AccountListResponse<AccountTransaction>))
  },

  // 获取账户统计
  getAccountStats: (): Promise<AccountStatsResponse> => {
    return request.get(`${PLATFORM_PREFIX}/accounts/stats`)
      .then((res: any) => ({
        code: res?.code ?? 500,
        message: res?.message || '',
        data: res?.data
      } as AccountStatsResponse))
  }
}

// 合作状态常量
export const CooperationStatus = {
  ACTIVE: 1,    // 合作中
  PAUSED: 2,    // 暂停
  RELEASED: 3   // 已解除
}

export const CooperationStatusText: Record<number, string> = {
  1: '合作中',
  2: '暂停',
  3: '已解除'
}

// ==================== 佣金管理 ====================

// 平台佣金API
export const platformCommissionApi = {
  // 获取佣金统计
  getCommissionStats: (): Promise<any> => {
    return request.get(`${PLATFORM_PREFIX}/commission/stats`)
      .then((res: any) => ({
        code: res?.code ?? 500,
        message: res?.message || '',
        data: res?.data
      }))
  },

  // 获取佣金列表
  getCommissionList: (params?: {
    type?: string
    page?: number
    page_size?: number
  }): Promise<any> => {
    return request.get(`${PLATFORM_PREFIX}/commission/list`, { params })
      .then((res: any) => ({
        code: res?.code ?? 500,
        message: res?.message || '',
        data: res?.data || { list: [], total: 0, page: 1, page_size: 10 }
      }))
  }
}

// ==================== 财务审核 ====================

// 平台财务审核API
export const platformFinanceAuditApi = {
  // 获取审核统计
  getAuditStats: (type?: string): Promise<any> => {
    return request.get(`${PLATFORM_PREFIX}/finance/audit/stats`, { params: { type } })
      .then((res: any) => ({
        code: res?.code ?? 500,
        message: res?.message || '',
        data: res?.data
      }))
  },

  // 获取提现审核列表
  getWithdrawAuditList: (params?: {
    status?: string
    keyword?: string
    withdraw_type?: string
    page?: number
    page_size?: number
  }): Promise<any> => {
    return request.get(`${PLATFORM_PREFIX}/finance/audit/withdraw`, { params })
      .then((res: any) => ({
        code: res?.code ?? 500,
        message: res?.message || '',
        data: res?.data || { list: [], total: 0, page: 1, page_size: 10 }
      }))
  },

  // 获取充值审核列表
  getRechargeAuditList: (params?: {
    status?: string
    page?: number
    page_size?: number
  }): Promise<any> => {
    return request.get(`${PLATFORM_PREFIX}/finance/audit/recharge`, { params })
      .then((res: any) => ({
        code: res?.code ?? 500,
        message: res?.message || '',
        data: res?.data || { list: [], total: 0, page: 1, page_size: 10 }
      }))
  },

  // 审批操作
  approveAudit: (data: {
    transaction_id: number
    action: 'approve' | 'reject'
    remark?: string
  }): Promise<any> => {
    return request.post(`${PLATFORM_PREFIX}/finance/audit/approve`, data)
      .then((res: any) => ({
        code: res?.code ?? 500,
        message: res?.message || '',
        data: res?.data
      }))
  },

  // 创建提现申请
  createWithdrawApplication: (data: {
    account_type: string
    amount: number
    remark?: string
  }): Promise<any> => {
    return request.post(`${PLATFORM_PREFIX}/finance/withdraw/apply`, data)
      .then((res: any) => ({
        code: res?.code ?? 500,
        message: res?.message || '',
        data: res?.data
      }))
  }
}

// ==================== 罚补账户 ====================

// 平台罚补账户API
export const platformPenaltyApi = {
  // 获取罚补账户统计
  getPenaltyStats: (): Promise<any> => {
    return request.get(`${PLATFORM_PREFIX}/penalty/stats`)
      .then((res: any) => ({
        code: res?.code ?? 500,
        message: res?.message || '',
        data: res?.data
      }))
  },

  // 获取罚补交易列表
  getPenaltyList: (params?: {
    type?: string
    page?: number
    page_size?: number
  }): Promise<any> => {
    return request.get(`${PLATFORM_PREFIX}/penalty/list`, { params })
      .then((res: any) => ({
        code: res?.code ?? 500,
        message: res?.message || '',
        data: res?.data || { list: [], total: 0, page: 1, page_size: 10 }
      }))
  },

  // 创建罚款/补贴
  createPenalty: (data: {
    admin_id: number
    type: 'penalty' | 'subsidy'
    amount: number
    remark?: string
  }): Promise<any> => {
    return request.post(`${PLATFORM_PREFIX}/penalty/create`, data)
      .then((res: any) => ({
        code: res?.code ?? 500,
        message: res?.message || '',
        data: res?.data
      }))
  }
}

// ==================== 收款账户 ====================

// 收款账户
export interface CollectionAccount {
  id: number
  name: string
  account: string
  payee: string
  status: string
  is_default: boolean
  bank_branch?: string
}

// 平台收款账户API
export const platformCollectionApi = {
  // 获取收款账户列表
  getCollectionAccounts: (adminId?: number): Promise<any> => {
    return request.get(`${PLATFORM_PREFIX}/collection/accounts`, { params: { admin_id: adminId } })
      .then((res: any) => ({
        code: res?.code ?? 500,
        message: res?.message || '',
        data: res?.data || { wallets: [], banks: [] }
      }))
  },

  // 创建收款账户
  createCollectionAccount: (data: {
    account_type: 'wallet' | 'bank'
    account_name: string
    account_no: string
    payee: string
    bank_name?: string
    bank_branch?: string
    is_default?: boolean
  }): Promise<any> => {
    return request.post(`${PLATFORM_PREFIX}/collection/accounts`, data)
      .then((res: any) => ({
        code: res?.code ?? 500,
        message: res?.message || '',
        data: res?.data
      }))
  },

  // 更新收款账户
  updateCollectionAccount: (id: number, data: {
    account_name: string
    account_no: string
    payee: string
    bank_name?: string
    bank_branch?: string
    is_default?: boolean
  }): Promise<any> => {
    return request.put(`${PLATFORM_PREFIX}/collection/accounts/${id}`, data)
      .then((res: any) => ({
        code: res?.code ?? 500,
        message: res?.message || '',
        data: res?.data
      }))
  },

  // 删除收款账户
  deleteCollectionAccount: (id: number): Promise<any> => {
    return request.delete(`${PLATFORM_PREFIX}/collection/accounts/${id}`)
      .then((res: any) => ({
        code: res?.code ?? 500,
        message: res?.message || '',
        data: res?.data
      }))
  },

  // 设置默认账户
  setDefaultAccount: (id: number): Promise<any> => {
    return request.post(`${PLATFORM_PREFIX}/collection/accounts/${id}/default`)
      .then((res: any) => ({
        code: res?.code ?? 500,
        message: res?.message || '',
        data: res?.data
      }))
  }
}
