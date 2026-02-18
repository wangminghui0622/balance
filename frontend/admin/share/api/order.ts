import request from '../utils/request'

// API前缀常量
const SHOPOWER_PREFIX = '/api/v1/balance/admin/shopower'
const OPERATOR_PREFIX = '/api/v1/balance/admin/operator'
const PLATFORM_PREFIX = '/api/v1/balance/admin/platform'

// 订单商品
export interface OrderItem {
  id: number
  order_id: number
  shop_id: number
  order_sn: string
  item_id: number
  item_name: string
  item_sku: string
  model_id: number
  model_name: string
  model_sku: string
  quantity: number
  item_price: string
  created_at: string
  updated_at: string
}

// 订单地址
export interface OrderAddress {
  id: number
  order_id: number
  shop_id: number
  order_sn: string
  name: string
  phone: string
  town: string
  district: string
  city: string
  state: string
  region: string
  zipcode: string
  full_address: string
  created_at: string
  updated_at: string
}

// 订单
export interface Order {
  id: number
  shop_id: number
  order_sn: string
  region: string
  order_status: string
  status_locked: boolean
  status_remark: string
  buyer_user_id: number
  buyer_username: string
  total_amount: string
  currency: string
  shipping_carrier: string
  tracking_number: string
  ship_by_date: string | null
  pay_time: string | null
  create_time: string | null
  update_time: string | null
  created_at: string
  updated_at: string
  items?: OrderItem[]
  address?: OrderAddress
  // 预付款状态 (0=未检查, 1=充足/已付款, 2=不足/未付款)
  prepayment_status: number
  prepayment_snapshot?: string
  prepayment_checked_at?: string | null
  // 账款调整相关显示字段（服务器返回完整字符串）
  adjustment_label_1?: string  // 例如: "账款调整佣金：NT$8.00" 或 "未结算佣金: NT$8.00"
  adjustment_label_2?: string  // 例如: "订单账款调整：NT$36.00" 或 "订单金额: NT$36.00"
  adjustment_label_3?: string  // 例如: "虾皮订单账款调整：NT$46.00" 或 "虾皮订单金额: NT$46.00"
}

// 订单列表响应（sheepx格式）
export interface OrderListResponse {
  code: number
  message: string
  data: {
    list: Order[]
    total: number
    page: number
    page_size: number
  }
}

// 订单详情响应
export interface OrderDetailResponse {
  code: number
  message: string
  data: Order
}

// 订单同步请求
export interface SyncOrdersRequest {
  shop_id: number
  time_from?: number
  time_to?: number
  order_status?: string
}

// 订单同步响应
export interface SyncOrdersResponse {
  code: number
  message: string
  data?: {
    message: string
  }
}

// 订单统计数据
export interface OrderSummary {
  allOrders: { count: number; amount: number }
  unsettledOrders: { count: number; amount: number }
  settledOrders: { count: number; amount: number }
  adjustments: { count: number; amount: number }
}

// 订单列表查询参数
export interface OrderListParams {
  shop_id?: number
  status?: string
  order_status?: string
  order_sn?: string
  start_time?: string
  end_time?: string
  page?: number
  page_size?: number
}

// 店主专用订单API
export const orderApi = {
  // 获取订单列表
  getOrderList: (params?: OrderListParams): Promise<OrderListResponse> => {
    return request.get(`${SHOPOWER_PREFIX}/orders`, { params })
      .then((res: any) => ({
        code: res?.code ?? 500,
        message: res?.message || '',
        data: res?.data || { list: [], total: 0, page: 1, page_size: 10 }
      } as OrderListResponse))
  },

  // 获取订单详情
  getOrderDetail: (shopId: number, orderSn: string): Promise<OrderDetailResponse> => {
    return request.get(`${SHOPOWER_PREFIX}/orders/${shopId}/${orderSn}`)
      .then((res: any) => ({
        code: res?.code ?? 500,
        message: res?.message || '',
        data: res?.data
      } as OrderDetailResponse))
  },

  // 获取待发货订单
  getReadyToShipOrders: (params?: { shop_id?: number; page?: number; page_size?: number }): Promise<OrderListResponse> => {
    return request.get(`${SHOPOWER_PREFIX}/orders/ready-to-ship`, { params })
      .then((res: any) => ({
        code: res?.code ?? 500,
        message: res?.message || '',
        data: res?.data || { list: [], total: 0, page: 1, page_size: 10 }
      } as OrderListResponse))
  },

  // 同步订单
  syncOrders: (data: SyncOrdersRequest): Promise<SyncOrdersResponse> => {
    return request.post(`${SHOPOWER_PREFIX}/orders/sync`, data)
      .then((res: any) => ({
        code: res?.code ?? 500,
        message: res?.message || '',
        data: res?.data
      } as SyncOrdersResponse))
  },

  // 刷新单个订单
  refreshOrder: (shopId: number, orderSn: string): Promise<OrderDetailResponse> => {
    return request.post(`${SHOPOWER_PREFIX}/orders/${shopId}/${orderSn}/refresh`)
      .then((res: any) => ({
        code: res?.code ?? 500,
        message: res?.message || '',
        data: res?.data
      } as OrderDetailResponse))
  }
}

// 运营专用订单API
export const operatorOrderApi = {
  // 获取订单列表（运营可查看所有订单）
  getOrderList: (params?: OrderListParams): Promise<OrderListResponse> => {
    return request.get(`${OPERATOR_PREFIX}/orders`, { params })
      .then((res: any) => ({
        code: res?.code ?? 500,
        message: res?.message || '',
        data: res?.data || { list: [], total: 0, page: 1, page_size: 10 }
      } as OrderListResponse))
  },

  // 获取订单详情
  getOrderDetail: (shopId: number, orderSn: string): Promise<OrderDetailResponse> => {
    return request.get(`${OPERATOR_PREFIX}/orders/${shopId}/${orderSn}`)
      .then((res: any) => ({
        code: res?.code ?? 500,
        message: res?.message || '',
        data: res?.data
      } as OrderDetailResponse))
  }
}

// 平台专用订单API
export const platformOrderApi = {
  // 获取订单列表（平台可查看所有订单）
  getOrderList: (params?: OrderListParams): Promise<OrderListResponse> => {
    return request.get(`${PLATFORM_PREFIX}/orders`, { params })
      .then((res: any) => ({
        code: res?.code ?? 500,
        message: res?.message || '',
        data: res?.data || { list: [], total: 0, page: 1, page_size: 10 }
      } as OrderListResponse))
  },

  // 获取订单详情
  getOrderDetail: (shopId: number, orderSn: string): Promise<OrderDetailResponse> => {
    return request.get(`${PLATFORM_PREFIX}/orders/${shopId}/${orderSn}`)
      .then((res: any) => ({
        code: res?.code ?? 500,
        message: res?.message || '',
        data: res?.data
      } as OrderDetailResponse))
  }
}

// 导出单独的函数方便使用
export const getOrderList = orderApi.getOrderList
export const getOrderDetail = orderApi.getOrderDetail
export const getReadyToShipOrders = orderApi.getReadyToShipOrders
export const syncOrders = orderApi.syncOrders
export const refreshOrder = orderApi.refreshOrder
