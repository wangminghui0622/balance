<template>
  <el-card class="orders-card">
    <div class="orders-header">
      <el-tabs v-model="activeTab" class="order-tabs">
        <el-tab-pane label="最近订单" name="recent">
          <div class="orders-list">
            <div
              v-for="(order, index) in recentOrders"
              :key="index"
              class="order-item"
            >
            <div class="order-header">
              <div class="order-number">
                订单编号: {{ order.orderNo }}
                <el-tag v-if="order.prepaymentStatus > 0" size="small" :type="PREPAYMENT_STATUS_TAG_TYPE[order.prepaymentStatus]" style="margin-left: 8px;">{{ PREPAYMENT_STATUS_TEXT[order.prepaymentStatus] }}</el-tag>
              </div>
              <div class="order-amount-info">
                <span>未结算佣金: NT${{ order.unsettledCommission }}</span>
                <span>订单金额: NT${{ order.orderAmount }}</span>
              </div>
            </div>
            <div class="order-info">
              <div class="order-info-line">
                <span class="info-item">下单时间: {{ order.orderTime }}</span>
                <span class="info-item">店铺编号: {{ order.storeId }}</span>
                <span class="info-item">店铺名称: {{ order.storeName }}</span>
                <span class="info-item">虾皮订单号: {{ order.orderNo }}</span>
                <span v-if="order.shopeeStatus" class="info-item info-item-right">虾皮订单状态: {{ order.shopeeStatus }}</span>
              </div>
            </div>
            <div v-if="order.products && order.products.length > 0" class="products-section">
              <div v-for="(product, pIndex) in order.products" :key="pIndex" class="product-item">
                <el-avatar :size="80" shape="square" :src="product.image" class="product-image" />
                <div class="product-details">
                  <div class="product-name">{{ product.name }}</div>
                  <div class="product-specs">颜色: {{ product.color }} 尺寸: {{ product.size }}</div>
                  <div class="product-price-info">
                    <span>单价: NT${{ product.unitPrice }}</span>
                    <span>数量: {{ product.quantity }}</span>
                    <span>小计: {{ product.subtotal }}</span>
                  </div>
                </div>
              </div>
              <div class="order-shopee-amount" v-if="order.shopeeAmount">虾皮订单金额: NT${{ order.shopeeAmount }}</div>
              <div class="order-settlement-row">
                <span>已结算佣金: NT${{ order.settledPayment || '0.00' }}</span>
                <span>订单金额: NT${{ order.orderAmount }}</span>
              </div>
            </div>
          </div>
          <el-button
            v-if="showLoadMore()"
            type="primary"
            link
            :loading="loading"
            class="load-more-btn"
            @click="handleLoadMore"
          >
            加载更多
          </el-button>
        </div>
      </el-tab-pane>
      <el-tab-pane label="未结算" name="unsettled">
        <div class="orders-list">
          <div
            v-for="(order, index) in unsettledOrders"
            :key="index"
            class="order-item"
          >
            <div class="order-header">
              <div class="order-number">
                订单编号: {{ order.orderNo }}
                <el-tag size="small" type="warning" style="margin-left: 8px;">待结算</el-tag>
              </div>
              <div class="order-amount-info">
                <span>未结算佣金: NT${{ order.unsettledCommission }}</span>
                <span>未结算回款: NT${{ order.unsettledPayment }}</span>
                <span>订单金额: NT${{ order.orderAmount }}</span>
              </div>
            </div>
            <div class="order-info">
              <div class="order-info-line">
                <span class="info-item">下单时间: {{ order.orderTime }}</span>
                <span class="info-item">店铺编号: {{ order.storeId }}</span>
                <span class="info-item">店铺名称: {{ order.storeName }}</span>
                <span class="info-item">虾皮订单号: {{ order.orderNo }}</span>
                <span v-if="order.shopeeStatus" class="info-item info-item-right">虾皮订单状态: {{ order.shopeeStatus }}</span>
              </div>
            </div>
            <div v-if="order.products && order.products.length > 0" class="products-section">
              <div v-for="(product, pIndex) in order.products" :key="pIndex" class="product-item">
                <el-avatar :size="80" shape="square" :src="product.image" class="product-image" />
                <div class="product-details">
                  <div class="product-name">{{ product.name }}</div>
                  <div class="product-specs">颜色: {{ product.color }} 尺寸: {{ product.size }}</div>
                  <div class="product-price-info">
                    <span>单价: NT${{ product.unitPrice }}</span>
                    <span>数量: {{ product.quantity }}</span>
                    <span>小计: {{ product.subtotal }}</span>
                  </div>
                </div>
              </div>
              <div class="order-shopee-amount" v-if="order.shopeeAmount">虾皮订单金额: NT${{ order.shopeeAmount }}</div>
              <div class="order-settlement-row">
                <span>已结算佣金: NT${{ order.settledPayment || '0.00' }}</span>
                <span>订单金额: NT${{ order.orderAmount }}</span>
              </div>
            </div>
          </div>
          <el-button
            v-if="showLoadMore()"
            type="primary"
            link
            :loading="loading"
            class="load-more-btn"
            @click="handleLoadMore"
          >
            加载更多
          </el-button>
        </div>
      </el-tab-pane>
      <el-tab-pane label="已结算" name="settled">
        <div class="orders-list">
          <div
            v-for="(order, index) in settledOrders"
            :key="index"
            class="order-item"
          >
            <div class="order-header">
              <div class="order-number">
                订单编号: {{ order.orderNo }}
                <el-tag size="small" type="success" style="margin-left: 8px;">已结算</el-tag>
              </div>
              <div class="order-amount-info">
                <span>已结算回款: NT${{ order.settledPayment }}</span>
                <span>订单金额: NT${{ order.orderAmount }}</span>
              </div>
            </div>
            <div class="order-info">
              <div class="order-info-line">
                <span class="info-item">下单时间: {{ order.orderTime }}</span>
                <span class="info-item">店铺编号: {{ order.storeId }}</span>
                <span class="info-item">店铺名称: {{ order.storeName }}</span>
              </div>
            </div>
            <div v-if="order.products && order.products.length > 0" class="products-section">
              <div v-for="(product, pIndex) in order.products" :key="pIndex" class="product-item">
                <el-avatar :size="80" shape="square" :src="product.image" class="product-image" />
                <div class="product-details">
                  <div class="product-name">{{ product.name }}</div>
                  <div class="product-specs">颜色: {{ product.color }} 尺寸: {{ product.size }}</div>
                  <div class="product-price-info">
                    <span>单价: NT${{ product.unitPrice }}</span>
                    <span>数量: {{ product.quantity }}</span>
                    <span>小计: {{ product.subtotal }}</span>
                  </div>
                </div>
              </div>
              <div class="order-settlement-row">
                <span>已结算佣金: NT${{ order.settledPayment || '0.00' }}</span>
                <span>订单金额: NT${{ order.orderAmount }}</span>
              </div>
            </div>
          </div>
          <el-button
            v-if="showLoadMore()"
            type="primary"
            link
            :loading="loading"
            class="load-more-btn"
            @click="handleLoadMore"
          >
            加载更多
          </el-button>
        </div>
      </el-tab-pane>
    </el-tabs>
    <router-link to="/shopowner/orders" class="all-orders-link">
      所有订单
      <svg class="arrow-icon" viewBox="0 0 12 12" fill="none" xmlns="http://www.w3.org/2000/svg">
        <path d="M4.5 2.5L8 6L4.5 9.5" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round"/>
      </svg>
    </router-link>
    </div>
  </el-card>
</template>

<script setup lang="ts">
import { ref, onMounted, computed } from 'vue'
import { ElMessage } from 'element-plus'
import { orderApi, type Order as ApiOrder } from '@share/api/order'
import { HTTP_STATUS, PREPAYMENT_STATUS_TEXT, PREPAYMENT_STATUS_TAG_TYPE } from '@share/constants'

interface Product {
  image: string
  name: string
  color: string
  size: string
  unitPrice: string
  quantity: number
  subtotal: string
}

interface Order {
  orderNo: string
  orderTime: string
  storeName: string
  storeId?: string
  products?: Product[]
  orderAmount: string
  status?: string
  prepaymentStatus: number
  unsettledCommission?: string
  unsettledPayment?: string
  shopeeStatus?: string
  shopeeAmount?: string
  settledPayment?: string
}

const PAGE_SIZE = 10
const activeTab = ref('recent')
const loading = ref(false)
const page = ref(1)
const total = ref(0)

// 格式化时间：将 ISO 8601 格式转换为 YYYY-MM-DD HH:mm:ss
const formatDateTime = (isoString: string | null): string => {
  if (!isoString) return '-'
  const date = new Date(isoString)
  if (isNaN(date.getTime())) return isoString
  const year = date.getFullYear()
  const month = String(date.getMonth() + 1).padStart(2, '0')
  const day = String(date.getDate()).padStart(2, '0')
  const hours = String(date.getHours()).padStart(2, '0')
  const minutes = String(date.getMinutes()).padStart(2, '0')
  const seconds = String(date.getSeconds()).padStart(2, '0')
  return `${year}-${month}-${day} ${hours}:${minutes}:${seconds}`
}

// 原始API数据
const apiOrders = ref<ApiOrder[]>([])

// 将API订单数据转换为显示格式
const transformOrder = (apiOrder: ApiOrder): Order => {
  // 转换所有商品
  const products: Product[] = (apiOrder.items || []).map(item => ({
    image: '',
    name: item.item_name,
    color: item.model_name?.split(',')[0] || '-',
    size: item.model_name?.split(',')[1] || '-',
    unitPrice: item.item_price,
    quantity: item.quantity,
    subtotal: (parseFloat(item.item_price) * item.quantity).toFixed(2)
  }))
  
  // 订单状态映射
  const statusMap: Record<string, string> = {
    'UNPAID': '未付款',
    'READY_TO_SHIP': '待发货',
    'PROCESSED': '已处理',
    'SHIPPED': '已发货',
    'COMPLETED': '已完成',
    'IN_CANCEL': '取消中',
    'CANCELLED': '已取消',
    'INVOICE_PENDING': '待开票'
  }
  
  return {
    orderNo: apiOrder.order_sn,
    orderTime: formatDateTime(apiOrder.create_time || apiOrder.created_at),
    storeId: apiOrder.shop_id.toString(),
    storeName: `店铺 ${apiOrder.shop_id}`,
    products: products.length > 0 ? products : undefined,
    orderAmount: apiOrder.total_amount,
    status: apiOrder.order_status,
    prepaymentStatus: apiOrder.prepayment_status ?? 0,
    shopeeStatus: statusMap[apiOrder.order_status] || apiOrder.order_status,
    shopeeAmount: apiOrder.total_amount,
    unsettledCommission: '0.00',
    unsettledPayment: '0.00',
    settledPayment: '0.00'
  }
}

// 计算属性：最近订单（所有订单）
const recentOrders = computed(() => {
  return apiOrders.value.map(transformOrder)
})

// 计算属性：未结算订单（待发货状态）
const unsettledOrders = computed(() => {
  return apiOrders.value
    .filter(o => ['READY_TO_SHIP', 'PROCESSED', 'SHIPPED'].includes(o.order_status))
    .map(transformOrder)
})

// 计算属性：已结算订单（已完成状态）
const settledOrders = computed(() => {
  return apiOrders.value
    .filter(o => o.order_status === 'COMPLETED')
    .map(transformOrder)
})

// 获取订单列表（首次加载或加载更多）
const fetchOrders = async (isLoadMore = false) => {
  loading.value = true
  try {
    const res = await orderApi.getOrderList({
      page: page.value,
      page_size: PAGE_SIZE
    })
    if (res.code === HTTP_STATUS.OK && res.data) {
      total.value = res.data.total ?? 0
      const list = res.data.list || []
      if (isLoadMore) {
        apiOrders.value = [...apiOrders.value, ...list]
      } else {
        apiOrders.value = list
      }
    }
  } catch (err: any) {
    console.error('获取订单列表失败:', err)
    ElMessage.error(err?.message || '获取订单列表失败')
  } finally {
    loading.value = false
  }
}

// 是否显示加载更多（还有更多订单可加载时显示）
const showLoadMore = () => {
  return apiOrders.value.length < total.value
}

// 加载更多
const handleLoadMore = async () => {
  page.value += 1
  await fetchOrders(true)
}

onMounted(() => {
  fetchOrders(false)
})
</script>

<style scoped lang="scss">
.orders-card {
  height: 100%;
  border-radius: 8px;
  overflow: hidden;
}

.card-header {
  font-weight: 500;
  font-size: 16px;
}

.orders-header {
  position: relative;
}

.order-tabs {
  :deep(.el-tabs__header) {
    margin-bottom: 16px;
  }

  :deep(.el-tabs__item.is-active) {
    color: #303133;
  }

  :deep(.el-tabs__item:not(.is-active):hover) {
    color: #ff6a3a;
  }

  :deep(.el-tabs__active-bar) {
    background-color: #ff6a3a;
  }
}

.all-orders-link {
  position: absolute;
  top: 8px;
  right: 0;
  display: flex;
  align-items: center;
  gap: 5px;
  font-size: 12px;
  font-weight: 400;
  line-height: 1;
  color: #909399;
  text-decoration: none;
  
  &:hover {
    color: #606266;
  }
  
  .arrow-icon {
    width: 12px;
    height: 12px;
  }
}

.orders-list {
  display: flex;
  flex-direction: column;
  gap: 20px;
  max-height: 600px;
  overflow-y: auto;
}

.load-more-btn {
  display: block;
  width: 100%;
  margin-top: 16px;
  text-align: center;
}

.order-item {
  padding: 16px;
  background-color: #f5f7fa;
  border-radius: 8px;
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.order-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.order-header .order-number {
  font-size: 14px;
  font-weight: 500;
  color: #303133;
}

.order-header .order-amount-info {
  display: flex;
  gap: 24px;
  font-size: 13px;
  color: #606266;
}

.order-info {
  margin-bottom: 12px;
  background-color: #ebeef5;
  border-radius: 4px;
  padding: 10px 14px;

  .order-info-line {
    display: flex;
    justify-content: space-between;
    font-size: 12px;
    color: #909399;
    flex-wrap: nowrap;
    gap: 32px;
  }

  .info-item {
    white-space: nowrap;
    min-width: 0;
  }

  .info-item-right {
    margin-left: auto;
  }
}

.products-section {
  padding: 16px 16px 0 16px;
  background-color: #fafafa;
  border-radius: 6px;

  .product-item {
    display: grid;
    grid-template-columns: 80px minmax(0, 1fr);
    column-gap: 16px;
    padding: 12px 0;

    .product-image {
      width: 80px;
      height: 80px;
      border-radius: 6px;
      flex-shrink: 0;
      background-color: #f5f7fa;
    }

    .product-details {
      min-width: 0;
      display: grid;
      grid-template-columns: minmax(0, 1fr) 240px;
      column-gap: 16px;
      row-gap: 6px;

      .product-name {
        font-size: 14px;
        color: #303133;
        font-weight: 500;
        line-height: 1.5;
        grid-column: 1;
        overflow: hidden;
        text-overflow: ellipsis;
        white-space: nowrap;
      }

      .product-specs {
        font-size: 12px;
        color: #909399;
        grid-column: 1;
      }

      .product-price-info {
        display: flex;
        gap: 24px;
        font-size: 13px;
        color: #606266;
        grid-column: 1 / -1;
        justify-content: flex-start;
        align-items: center;
        white-space: nowrap;
        margin-top: 4px;
      }
    }
  }

  .order-shopee-amount {
    font-size: 12px;
    color: #909399;
    font-weight: 400;
    text-align: right;
    margin-top: 0;
    padding-top: 4px;
    border-top: 1px solid #ebeef5;
  }

  .order-settlement-row {
    display: flex;
    justify-content: flex-end;
    gap: 48px;
    font-size: 13px;
    color: #606266;
    margin-top: 12px;
    margin-bottom: -12px;
  }
}
</style>

