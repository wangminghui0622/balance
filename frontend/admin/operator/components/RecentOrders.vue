<template>
  <el-card class="orders-card">
    <template #header>
      <div class="card-header">
        <span>最近订单</span>
      </div>
    </template>
    <el-tabs v-model="activeTab" class="order-tabs">
      <el-tab-pane label="未结算" name="unsettled">
        <div class="orders-list">
          <div
            v-for="(order, index) in unsettledOrders"
            :key="index"
            class="order-item"
          >
            <div class="order-header">
              <div class="order-number">
                订单编号: {{ order.orderNo }} 付款状态
              </div>
            </div>
            <div class="order-info">
              <div class="info-row">
                <span>下单时间: {{ order.orderTime }}</span>
              </div>
              <div class="info-row">
                <span>店铺编号: {{ order.storeId }}</span>
              </div>
              <div class="info-row">
                <span>店铺名称: {{ order.storeName }}</span>
              </div>
              <div class="info-row">
                <span>虾皮订单号: {{ order.shopeeOrderNo }}</span>
              </div>
            </div>
            <div class="product-info" v-if="order.product">
              <el-avatar :size="60" shape="square" :src="order.product.image" />
              <div class="product-details">
                <div class="product-name">{{ order.product.name }}</div>
                <div class="product-spec">
                  颜色: {{ order.product.color }} 尺寸: {{ order.product.size }}
                </div>
                <div class="product-price">
                  <span>单价: NT${{ order.product.unitPrice }}</span>
                  <span>数量: {{ order.product.quantity }}</span>
                  <span>小计: {{ order.product.subtotal }}</span>
                </div>
              </div>
            </div>
            <div class="order-summary">
              <div class="summary-row">
                <span>未结算回款: NT${{ order.unsettledPayment }}</span>
                <span>订单金额: NT${{ order.orderAmount }}</span>
              </div>
              <div class="summary-row">
                <span>虾皮订单状态: {{ order.shopeeStatus }}</span>
                <span>虾皮订单金额: NT${{ order.shopeeAmount }}</span>
              </div>
              <div class="summary-row">
                <span>已结算回款: NT${{ order.settledPayment }}</span>
                <span>订单金额: NT${{ order.settledOrderAmount }}</span>
              </div>
            </div>
          </div>
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
                订单编号: {{ order.orderNo }} 付款状态
              </div>
            </div>
            <div class="order-info">
              <div class="info-row">
                <span>下单时间: {{ order.orderTime }}</span>
              </div>
              <div class="info-row">
                <span>店铺编号: {{ order.storeId }}</span>
              </div>
              <div class="info-row">
                <span>店铺名称: {{ order.storeName }}</span>
              </div>
            </div>
            <div class="order-summary">
              <div class="summary-row">
                <span>已结算回款: NT${{ order.settledPayment }}</span>
                <span>订单金额: NT${{ order.orderAmount }}</span>
              </div>
            </div>
          </div>
        </div>
      </el-tab-pane>
    </el-tabs>
  </el-card>
</template>

<script setup lang="ts">
import { ref } from 'vue'

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
  storeId: string
  storeName: string
  shopeeOrderNo?: string
  product?: Product
  unsettledPayment?: string
  orderAmount: string
  shopeeStatus?: string
  shopeeAmount?: string
  settledPayment: string
  settledOrderAmount?: string
}

const activeTab = ref('unsettled')

const unsettledOrders = ref<Order[]>([
  {
    orderNo: 'X250904KQ2P078R',
    orderTime: '2025-12-10 23:59:59',
    storeId: '51234567890',
    storeName: '示例文字占位符示例文',
    shopeeOrderNo: '250904KQ2P07BR',
    product: {
      image: '',
      name: '商品名称示例文字占位符替换即可文字占位符替换即可',
      color: 'xxx',
      size: 'xxx',
      unitPrice: '46789.00',
      quantity: 1378,
      subtotal: '4634567.00'
    },
    unsettledPayment: '8.00',
    orderAmount: '36897.00',
    shopeeStatus: '待发货',
    shopeeAmount: '46.00',
    settledPayment: '16.00',
    settledOrderAmount: '198.00'
  },
  {
    orderNo: 'X250904KQ2P078S',
    orderTime: '2025-12-10 22:30:00',
    storeId: '51234567891',
    storeName: '示例文字占位符示例文',
    shopeeOrderNo: '250904KQ2P07BS',
    product: {
      image: '',
      name: '商品名称示例文字占位符',
      color: '红色',
      size: 'M',
      unitPrice: '120.00',
      quantity: 2,
      subtotal: '240.00'
    },
    unsettledPayment: '20.00',
    orderAmount: '220.00',
    shopeeStatus: '待发货',
    shopeeAmount: '240.00',
    settledPayment: '0.00',
    settledOrderAmount: '0.00'
  }
])

const settledOrders = ref<Order[]>([
  {
    orderNo: 'X250904KQ2P078T',
    orderTime: '2025-12-09 15:20:00',
    storeId: '51234567892',
    storeName: '示例文字占位符示例文',
    settledPayment: '50.00',
    orderAmount: '450.00'
  }
])
</script>

<style scoped lang="scss">
.orders-card {
  height: 100%;
}

.card-header {
  font-weight: 500;
  font-size: 16px;
}

.order-tabs {
  :deep(.el-tabs__header) {
    margin-bottom: 16px;
  }
}

.orders-list {
  display: flex;
  flex-direction: column;
  gap: 20px;
  max-height: 600px;
  overflow-y: auto;
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
  .order-number {
    font-size: 14px;
    font-weight: 500;
    color: #303133;
  }
}

.order-info {
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.info-row {
  font-size: 12px;
  color: #606266;
}

.product-info {
  display: flex;
  gap: 12px;
  padding: 12px;
  background-color: #fff;
  border-radius: 4px;
}

.product-details {
  flex: 1;
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.product-name {
  font-size: 14px;
  color: #303133;
  line-height: 1.4;
}

.product-spec {
  font-size: 12px;
  color: #909399;
}

.product-price {
  display: flex;
  gap: 16px;
  font-size: 12px;
  color: #606266;
}

.order-summary {
  display: flex;
  flex-direction: column;
  gap: 8px;
  padding-top: 12px;
  border-top: 1px solid #e4e7ed;
}

.summary-row {
  display: flex;
  justify-content: space-between;
  font-size: 12px;
  color: #606266;
}
</style>
