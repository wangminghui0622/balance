import request from '../utils/request'

// API前缀常量
const SHOPOWER_PREFIX = '/api/v1/balance/admin/shopower'
const OPERATOR_PREFIX = '/api/v1/balance/admin/operator'

// 预付款账户
export interface PrepaymentAccount {
  id: number
  admin_id: number
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

// 账户响应
export interface AccountResponse<T> {
  code: number
  message: string
  data: T
}

// 流水列表响应
export interface TransactionListResponse {
  code: number
  message: string
  data: {
    list: AccountTransaction[]
    total: number
    page: number
    page_size: number
  }
}

// 店主账户API
export const shopowerAccountApi = {
  // 获取预付款账户
  getPrepaymentAccount: (): Promise<AccountResponse<PrepaymentAccount>> => {
    return request.get(`${SHOPOWER_PREFIX}/account/prepayment`)
      .then((res: any) => ({
        code: res?.code ?? 500,
        message: res?.message || '',
        data: res?.data
      } as AccountResponse<PrepaymentAccount>))
  },

  // 获取保证金账户
  getDepositAccount: (): Promise<AccountResponse<DepositAccount>> => {
    return request.get(`${SHOPOWER_PREFIX}/account/deposit`)
      .then((res: any) => ({
        code: res?.code ?? 500,
        message: res?.message || '',
        data: res?.data
      } as AccountResponse<DepositAccount>))
  },

  // 获取预付款流水
  getPrepaymentTransactions: (params?: {
    page?: number
    page_size?: number
    transaction_type?: string
  }): Promise<TransactionListResponse> => {
    return request.get(`${SHOPOWER_PREFIX}/account/prepayment/transactions`, { params })
      .then((res: any) => ({
        code: res?.code ?? 500,
        message: res?.message || '',
        data: res?.data || { list: [], total: 0, page: 1, page_size: 10 }
      } as TransactionListResponse))
  },

  // 获取保证金流水
  getDepositTransactions: (params?: {
    page?: number
    page_size?: number
  }): Promise<TransactionListResponse> => {
    return request.get(`${SHOPOWER_PREFIX}/account/deposit/transactions`, { params })
      .then((res: any) => ({
        code: res?.code ?? 500,
        message: res?.message || '',
        data: res?.data || { list: [], total: 0, page: 1, page_size: 10 }
      } as TransactionListResponse))
  }
}

// 充值申请
export interface RechargeRecord {
  id: number
  application_no: string
  admin_id: number
  account_type: string
  amount: string
  currency: string
  payment_method: string
  payment_proof: string
  status: number
  audit_remark: string
  audit_by: number
  audit_at: string | null
  remark: string
  created_at: string
  updated_at: string
}

// 充值申请列表响应
export interface RechargeListResponse {
  code: number
  message: string
  data: {
    list: RechargeRecord[]
    total: number
    page: number
    page_size: number
  }
}

// 店主充值/提现API
export const shopowerFinanceApi = {
  // 充值（直接到账，无需审核）
  recharge: (data: {
    account_type: string
    amount: number
    payment_method: string
    payment_proof?: string
    remark?: string
  }): Promise<AccountResponse<RechargeRecord>> => {
    return request.post(`${SHOPOWER_PREFIX}/recharge`, data)
      .then((res: any) => ({
        code: res?.code ?? 500,
        message: res?.message || '',
        data: res?.data
      } as AccountResponse<RechargeRecord>))
  },

  // 获取充值申请列表
  getRechargeRecords: (params?: {
    status?: number
    page?: number
    page_size?: number
  }): Promise<RechargeListResponse> => {
    return request.get(`${SHOPOWER_PREFIX}/recharge/list`, { params })
      .then((res: any) => ({
        code: res?.code ?? 500,
        message: res?.message || '',
        data: res?.data || { list: [], total: 0, page: 1, page_size: 10 }
      } as RechargeListResponse))
  }
}

// 充值申请状态常量
export const RechargeStatus = {
  PENDING: 0,   // 待审核
  APPROVED: 1,  // 已通过
  REJECTED: 2   // 已拒绝
}

export const RechargeStatusText: Record<number, string> = {
  0: '待审核',
  1: '已通过',
  2: '已拒绝'
}

export const RechargeStatusColor: Record<number, string> = {
  0: '#E6A23C',
  1: '#67C23A',
  2: '#F56C6C'
}

// 运营账户API
export const operatorAccountApi = {
  // 获取运营账户
  getAccount: (): Promise<AccountResponse<OperatorAccount>> => {
    return request.get(`${OPERATOR_PREFIX}/account`)
      .then((res: any) => ({
        code: res?.code ?? 500,
        message: res?.message || '',
        data: res?.data
      } as AccountResponse<OperatorAccount>))
  },

  // 获取账户流水
  getTransactions: (params?: {
    page?: number
    page_size?: number
  }): Promise<TransactionListResponse> => {
    return request.get(`${OPERATOR_PREFIX}/account/transactions`, { params })
      .then((res: any) => ({
        code: res?.code ?? 500,
        message: res?.message || '',
        data: res?.data || { list: [], total: 0, page: 1, page_size: 10 }
      } as TransactionListResponse))
  }
}

// 账户状态常量
export const AccountStatus = {
  NORMAL: 1,   // 正常
  FROZEN: 2    // 冻结
}

export const AccountStatusText: Record<number, string> = {
  1: '正常',
  2: '冻结'
}

// 交易类型常量
export const TransactionType = {
  RECHARGE: 'recharge',           // 充值
  WITHDRAW: 'withdraw',           // 提现
  FREEZE: 'freeze',               // 冻结
  UNFREEZE: 'unfreeze',           // 解冻
  ORDER_PAY: 'order_pay',         // 订单支付
  ORDER_REFUND: 'order_refund',   // 订单退款
  PROFIT_SHARE: 'profit_share',   // 利润分成
  COST_SETTLE: 'cost_settle',     // 成本结算
  PLATFORM_FEE: 'platform_fee',   // 平台服务费
  DEPOSIT_PAY: 'deposit_pay',     // 保证金缴纳
  DEPOSIT_REFUND: 'deposit_refund' // 保证金退还
}

export const TransactionTypeText: Record<string, string> = {
  'recharge': '充值',
  'withdraw': '提现',
  'freeze': '冻结',
  'unfreeze': '解冻',
  'order_pay': '订单支付',
  'order_refund': '订单退款',
  'profit_share': '利润分成',
  'cost_settle': '成本结算',
  'platform_fee': '平台服务费',
  'deposit_pay': '保证金缴纳',
  'deposit_refund': '保证金退还'
}
