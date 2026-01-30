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
                <el-tag size="small" type="success" style="margin-left: 8px;">付款状态</el-tag>
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
            <div v-if="order.product" class="products-section">
              <div class="product-item">
                <el-avatar :size="80" shape="square" :src="order.product.image" class="product-image" />
                <div class="product-details">
                  <div class="product-name">{{ order.product.name }}</div>
                  <div class="product-specs">颜色: {{ order.product.color }} 尺寸: {{ order.product.size }}</div>
                  <div class="product-price-info">
                    <span>单价: NT${{ order.product.unitPrice }}</span>
                    <span>数量: {{ order.product.quantity }}</span>
                    <span>小计: {{ order.product.subtotal }}</span>
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
            <div v-if="order.product" class="products-section">
              <div class="product-item">
                <el-avatar :size="80" shape="square" :src="order.product.image" class="product-image" />
                <div class="product-details">
                  <div class="product-name">{{ order.product.name }}</div>
                  <div class="product-specs">颜色: {{ order.product.color }} 尺寸: {{ order.product.size }}</div>
                  <div class="product-price-info">
                    <span>单价: NT${{ order.product.unitPrice }}</span>
                    <span>数量: {{ order.product.quantity }}</span>
                    <span>小计: {{ order.product.subtotal }}</span>
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
            <div v-if="order.product" class="products-section">
              <div class="product-item">
                <el-avatar :size="80" shape="square" :src="order.product.image" class="product-image" />
                <div class="product-details">
                  <div class="product-name">{{ order.product.name }}</div>
                  <div class="product-specs">颜色: {{ order.product.color }} 尺寸: {{ order.product.size }}</div>
                  <div class="product-price-info">
                    <span>单价: NT${{ order.product.unitPrice }}</span>
                    <span>数量: {{ order.product.quantity }}</span>
                    <span>小计: {{ order.product.subtotal }}</span>
                  </div>
                </div>
              </div>
              <div class="order-settlement-row">
                <span>已结算佣金: NT${{ order.settledPayment || '0.00' }}</span>
                <span>订单金额: NT${{ order.orderAmount }}</span>
              </div>
            </div>
          </div>
        </div>
      </el-tab-pane>
    </el-tabs>
    <a href="#" class="all-orders-link">
      所有订单
      <svg class="arrow-icon" viewBox="0 0 12 12" fill="none" xmlns="http://www.w3.org/2000/svg">
        <path d="M4.5 2.5L8 6L4.5 9.5" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round"/>
      </svg>
    </a>
    </div>
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
  storeName: string
  storeId?: string
  product?: Product
  orderAmount: string
  status?: string
  unsettledCommission?: string
  unsettledPayment?: string
  shopeeStatus?: string
  shopeeAmount?: string
  settledPayment?: string
}

const activeTab = ref('recent')

const recentOrders = ref<Order[]>([
  {
    orderNo: 'X250904KQ2P078A',
    orderTime: '2025-12-11 10:30:00',
    storeId: 'S1234567890',
    storeName: '店铺名称示例文字占位符文字占位符',
    product: {
      image: '',
      name: '商品名称示例文字占位符替换即可文字占位符替换即可',
      color: '蓝色',
      size: 'L',
      unitPrice: '2388.00',
      quantity: 1231,
      subtotal: '345688.00'
    },
    unsettledCommission: '2335.00',
    orderAmount: '88444.00',
    status: '待处理',
    unsettledPayment: '0.00',
    shopeeStatus: '待发货',
    shopeeAmount: '88.00',
    settledPayment: '0.00'
  },
  {
    orderNo: 'X250904KQ2P078B',
    orderTime: '2025-12-11 09:15:00',
    storeId: 'S1234567891',
    storeName: '店铺名称示例文字占位符文字占位符',
    product: {
      image: '',
      name: '商品名称示例文字占位符',
      color: '红色',
      size: 'M',
      unitPrice: '120.00',
      quantity: 2,
      subtotal: '240.00'
    },
    unsettledCommission: '12.00',
    orderAmount: '240.00',
    status: '处理中',
    unsettledPayment: '0.00',
    shopeeStatus: '已发货',
    shopeeAmount: '240.00',
    settledPayment: '0.00'
  }
])

const unsettledOrders = ref<Order[]>([
  {
    orderNo: 'X250904KQ2P078R',
    orderTime: '2025-12-10 23:59:59',
    storeId: 'S1234567890',
    storeName: '店铺名称示例文字占位符文字占位符',
    product: {
      image: '',
      name: '商品名称示例文字占位符替换即可文字占位符替换即可',
      color: 'xxx',
      size: 'xxx',
      unitPrice: '46.00',
      quantity: 1,
      subtotal: '46.00'
    },
    unsettledCommission: '5.00',
    unsettledPayment: '8.00',
    orderAmount: '36.00',
    shopeeStatus: '待发货',
    shopeeAmount: '46.00',
    settledPayment: '0.00'
  }
])

const settledOrders = ref<Order[]>([
  {
    orderNo: 'X250904KQ2P078T',
    orderTime: '2025-12-09 15:20:00',
    storeId: 'S1234567892',
    storeName: '店铺名称示例文字占位符文字占位符',
    product: {
      image: '',
      name: '商品名称示例文字占位符',
      color: '红色',
      size: 'M',
      unitPrice: '120.00',
      quantity: 2,
      subtotal: '240.00'
    },
    orderAmount: '450.00',
    settledPayment: '50.00'
  }
])
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

