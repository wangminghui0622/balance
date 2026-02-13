import request from '../utils/request'

// API前缀常量
const SHOPOWER_PREFIX = '/api/v1/balance/admin/shopower'
const OPERATOR_PREFIX = '/api/v1/balance/admin/operator'

// 发货请求
export interface ShipOrderRequest {
  shop_id: number
  order_sn: string
  goods_cost: number
  shipping_cost?: number
  pickup_info?: {
    address_id?: number
    pickup_time_id?: string
    tracking_no?: string
  }
}

// 发货记录
export interface ShipmentRecord {
  id: number
  shop_id: number
  order_sn: string
  order_id: number
  shop_owner_id: number
  operator_id: number
  goods_cost: string
  shipping_cost: string
  total_cost: string
  currency: string
  frozen_amount: string
  shipping_carrier: string
  tracking_number: string
  shipped_at: string | null
  status: number
  settlement_id: number
  remark: string
  created_at: string
  updated_at: string
}

// 发货记录列表响应
export interface ShipmentListResponse {
  code: number
  message: string
  data: {
    list: ShipmentRecord[]
    total: number
    page: number
    page_size: number
  }
}

// 发货响应
export interface ShipOrderResponse {
  code: number
  message: string
  data?: ShipmentRecord
}

// 发货参数响应
export interface ShippingParameterResponse {
  code: number
  message: string
  data?: {
    pickup?: {
      address_list: Array<{
        address_id: number
        address: string
        time_slot_list: Array<{
          pickup_time_id: string
          date: string
          time_text: string
        }>
      }>
    }
    dropoff?: {
      branch_list: Array<{
        branch: string
        address: string
      }>
    }
  }
}

// 运营发货API
export const operatorShipmentApi = {
  // 发货
  shipOrder: (data: ShipOrderRequest): Promise<ShipOrderResponse> => {
    return request.post(`${OPERATOR_PREFIX}/shipments/ship`, data)
      .then((res: any) => ({
        code: res?.code ?? 500,
        message: res?.message || '',
        data: res?.data
      } as ShipOrderResponse))
  },

  // 获取待发货订单
  getPendingOrders: (params?: {
    status?: string
    page?: number
    page_size?: number
  }): Promise<any> => {
    return request.get(`${OPERATOR_PREFIX}/orders/pending`, { params })
      .then((res: any) => ({
        code: res?.code ?? 500,
        message: res?.message || '',
        data: res?.data || { list: [], total: 0, page: 1, page_size: 10 }
      }))
  },

  // 获取发货记录
  getShipmentRecords: (params?: {
    status?: number
    page?: number
    page_size?: number
  }): Promise<ShipmentListResponse> => {
    return request.get(`${OPERATOR_PREFIX}/shipments`, { params })
      .then((res: any) => ({
        code: res?.code ?? 500,
        message: res?.message || '',
        data: res?.data || { list: [], total: 0, page: 1, page_size: 10 }
      } as ShipmentListResponse))
  }
}

// 店主发货API
export const shopowerShipmentApi = {
  // 发货
  shipOrder: (shopId: number, orderSn: string, trackingNo?: string): Promise<ShipOrderResponse> => {
    return request.post(`${SHOPOWER_PREFIX}/shipments/ship`, {
      shop_id: shopId,
      order_sn: orderSn,
      tracking_no: trackingNo
    })
      .then((res: any) => ({
        code: res?.code ?? 500,
        message: res?.message || '',
        data: res?.data
      } as ShipOrderResponse))
  },

  // 批量发货
  batchShipOrders: (shopId: number, orderSns: string[]): Promise<any> => {
    return request.post(`${SHOPOWER_PREFIX}/shipments/batch-ship`, {
      shop_id: shopId,
      order_sns: orderSns
    })
      .then((res: any) => ({
        code: res?.code ?? 500,
        message: res?.message || '',
        data: res?.data
      }))
  },

  // 获取发货参数
  getShippingParameter: (shopId: number, orderSn: string): Promise<ShippingParameterResponse> => {
    return request.get(`${SHOPOWER_PREFIX}/shipments/shipping-parameter/${shopId}/${orderSn}`)
      .then((res: any) => ({
        code: res?.code ?? 500,
        message: res?.message || '',
        data: res?.data
      } as ShippingParameterResponse))
  },

  // 获取物流单号
  getTrackingNumber: (shopId: number, orderSn: string): Promise<any> => {
    return request.get(`${SHOPOWER_PREFIX}/shipments/tracking-number/${shopId}/${orderSn}`)
      .then((res: any) => ({
        code: res?.code ?? 500,
        message: res?.message || '',
        data: res?.data
      }))
  },

  // 获取发货记录列表
  getShipmentList: (params?: {
    shop_id?: number
    page?: number
    page_size?: number
  }): Promise<ShipmentListResponse> => {
    return request.get(`${SHOPOWER_PREFIX}/shipments`, { params })
      .then((res: any) => ({
        code: res?.code ?? 500,
        message: res?.message || '',
        data: res?.data || { list: [], total: 0, page: 1, page_size: 10 }
      } as ShipmentListResponse))
  },

  // 同步物流渠道
  syncLogisticsChannels: (shopId: number): Promise<any> => {
    return request.post(`${SHOPOWER_PREFIX}/shipments/sync-logistics/${shopId}`)
      .then((res: any) => ({
        code: res?.code ?? 500,
        message: res?.message || '',
        data: res?.data
      }))
  },

  // 获取物流渠道
  getLogisticsChannels: (shopId: number): Promise<any> => {
    return request.get(`${SHOPOWER_PREFIX}/shipments/logistics/${shopId}`)
      .then((res: any) => ({
        code: res?.code ?? 500,
        message: res?.message || '',
        data: res?.data
      }))
  }
}

// 发货状态常量
export const ShipmentStatus = {
  PENDING: 0,    // 待发货
  SHIPPED: 1,    // 已发货
  COMPLETED: 2,  // 已完成
  CANCELLED: 3,  // 已取消
  FAILED: 4      // 发货失败
}

export const ShipmentStatusText: Record<number, string> = {
  0: '待发货',
  1: '已发货',
  2: '已完成',
  3: '已取消',
  4: '发货失败'
}
