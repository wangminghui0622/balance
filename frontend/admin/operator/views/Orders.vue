<template>
  <div class="orders-page">
    <div class="page-header">
      <h1 class="page-title">我的订单</h1>
    </div>
    <el-card class="summary-section-card">
      <div class="summary-section">
        <div class="section-title-wrapper">
          <span class="section-title">订单概览</span>
          <el-button type="primary" link size="small" class="stat-link">订单统计</el-button>
        </div>
        <el-row :gutter="20" class="summary-cards">
          <el-col :xs="24" :sm="12" :md="6" :lg="6" :xl="6">
            <div class="summary-card">
              <div class="card-content">
                <div class="card-title">全部订单(NT$)</div>
                <div class="card-value">
                  <span class="count">{{ summaryData.allOrders.count }}<span class="unit">单</span></span>
                  <span class="equals">=</span>
                  <span class="amount">{{ formatAmount(summaryData.allOrders.amount) }}</span>
                </div>
              </div>
            </div>
          </el-col>
          <el-col :xs="24" :sm="12" :md="6" :lg="6" :xl="6">
            <div class="summary-card">
              <div class="card-content">
                <div class="card-title">未结算订单(NT$)</div>
                <div class="card-value">
                  <span class="count">{{ summaryData.unsettledOrders.count }}<span class="unit">单</span></span>
                  <span class="equals">=</span>
                  <span class="amount">{{ formatAmount(summaryData.unsettledOrders.amount) }}</span>
                </div>
              </div>
            </div>
          </el-col>
          <el-col :xs="24" :sm="12" :md="6" :lg="6" :xl="6">
            <div class="summary-card">
              <div class="card-content">
                <div class="card-title">已结算订单(NT$)</div>
                <div class="card-value">
                  <span class="count">{{ summaryData.settledOrders.count }}<span class="unit">单</span></span>
                  <span class="equals">=</span>
                  <span class="amount">{{ formatAmount(summaryData.settledOrders.amount) }}</span>
                </div>
              </div>
            </div>
          </el-col>
          <el-col :xs="24" :sm="12" :md="6" :lg="6" :xl="6">
            <div class="summary-card">
              <div class="card-content">
                <div class="card-title">账款调整(NT$)</div>
                <div class="card-value">
                  <span class="count">{{ summaryData.adjustments.count }}<span class="unit">单</span></span>
                  <span class="equals">=</span>
                  <span class="amount">{{ formatAmount(summaryData.adjustments.amount) }}</span>
                </div>
              </div>
            </div>
          </el-col>
        </el-row>
      </div>
    </el-card>

    <!-- 搜索筛选栏 -->
    <div class="filter-section">
      <div class="filter-left">
        <el-input v-model="filterForm.shopKeyword" placeholder="店铺名称/店铺编号" clearable class="filter-input">
          <template #prefix>
            <el-icon><Search /></el-icon>
          </template>
        </el-input>
        <el-input v-model="filterForm.orderNo" placeholder="订单编号" clearable class="filter-input">
          <template #prefix>
            <el-icon><Search /></el-icon>
          </template>
        </el-input>
        <el-select v-model="filterForm.settlementStatus" placeholder="订单结算状态" clearable class="filter-select">
          <el-option label="全部" value="" />
          <el-option label="未结算" value="unsettled" />
          <el-option label="已结算" value="settled" />
        </el-select>
        <el-select v-model="filterForm.paymentStatus" placeholder="选择付款状态" clearable class="filter-select">
          <el-option label="全部" value="" />
          <el-option label="已付款" value="paid" />
          <el-option label="未付款" value="unpaid" />
        </el-select>
        <el-date-picker
          v-model="filterForm.startDate"
          type="date"
          placeholder="开始日期"
          class="filter-date"
        />
        <span class="date-separator">-</span>
        <el-date-picker
          v-model="filterForm.endDate"
          type="date"
          placeholder="结束日期"
          class="filter-date"
        />
      </div>
      <div class="filter-right">
        <el-button type="primary" @click="handleSearch">查询</el-button>
        <el-button @click="handleReset">重置</el-button>
      </div>
    </div>

    <!-- 订单列表区域 -->
    <el-card class="orders-card" v-loading="loading">
      <div class="orders-header">
        <div class="tab-buttons">
          <span 
            :class="['tab-btn', activeTab === 'all' ? 'active' : '']" 
            @click="activeTab = 'all'"
          >全部订单</span>
          <span 
            :class="['tab-btn', activeTab === 'unsettled' ? 'active' : '']" 
            @click="activeTab = 'unsettled'"
          >未结算</span>
          <span 
            :class="['tab-btn', activeTab === 'settled' ? 'active' : '']" 
            @click="activeTab = 'settled'"
          >已结算</span>
        </div>
        <div class="action-buttons">
          <el-checkbox v-model="showProductInfo">商品信息</el-checkbox>
          <el-checkbox v-model="exportReport">导出报表</el-checkbox>
        </div>
      </div>

      <!-- 订单列表 -->
      <div class="orders-list">
        <div v-for="(order, index) in filteredOrders" :key="index" class="order-item">
          <!-- 订单头部 -->
          <div class="order-row-header">
            <div class="order-left">
              <span class="order-number">订单编号：{{ order.orderNo }}</span>
              <el-tag v-if="order.paymentStatus" :type="getPaymentStatusType(order.paymentStatus)" size="small">
                {{ getPaymentStatusText(order.paymentStatus) }}
              </el-tag>
            </div>
            <div class="order-right">
              <span class="payment-info" v-if="order.unsettledPayment">未结算回款：<span class="highlight">NT${{ order.unsettledPayment }}</span></span>
              <span class="payment-info" v-if="order.settledPayment">已结算回款：<span class="highlight">NT${{ order.settledPayment }}</span></span>
              <span class="amount-info">订单金额：<span class="amount">NT${{ formatAmount(order.orderAmount) }}</span></span>
            </div>
          </div>

          <!-- 订单信息 -->
          <div class="order-row-info">
            <div class="info-left">
              <span class="info-item">下单时间：{{ order.orderTime }}</span>
              <span class="info-item">店铺编号：{{ order.storeId }}</span>
              <span class="info-item">店铺名称：{{ order.storeName }}</span>
              <span class="info-item">虾皮订单号：{{ order.shopeeOrderNo || '-' }}</span>
            </div>
            <div class="info-right">
              <span class="shopee-status">虾皮订单状态：{{ order.shopeeStatus || '待发货' }}</span>
            </div>
          </div>

          <!-- 商品信息 -->
          <div v-if="showProductInfo && order.products && order.products.length > 0" class="products-section">
            <div v-for="(product, pIndex) in order.products" :key="pIndex" class="product-item">
              <el-image :src="product.image || '/placeholder.png'" :alt="product.name" class="product-image" fit="cover">
                <template #error>
                  <div class="image-slot">
                    <el-icon><Picture /></el-icon>
                  </div>
                </template>
              </el-image>
              <div class="product-details">
                <div class="product-name">{{ product.name }}</div>
                <div class="product-specs">
                  颜色: {{ product.color || 'xxx' }}&nbsp;&nbsp;&nbsp;&nbsp;尺寸: {{ product.size || 'xxx' }}
                </div>
                <div class="product-price-info">
                  <span>单价: NT${{ product.unitPrice }}</span>
                  <span>数量: {{ product.quantity }}</span>
                  <span>小计: {{ product.subtotal }}</span>
                </div>
              </div>
            </div>
            <!-- 已结算回款和订单金额 -->
            <div class="order-settlement-row">
              <span>已结算回款: NT${{ order.settledPayment || '0.00' }}</span>
              <span>订单金额: NT${{ formatAmount(order.orderAmount) }}</span>
            </div>
          </div>
        </div>

        <!-- 空状态 -->
        <el-empty v-if="filteredOrders.length === 0" description="暂无订单数据" />
      </div>

      <!-- 分页 -->
      <div class="pagination-wrapper" v-if="filteredOrders.length > 0">
        <el-pagination
          v-model:current-page="pagination.page"
          v-model:page-size="pagination.pageSize"
          :total="pagination.total"
          :page-sizes="[10, 20, 50, 100]"
          layout="total, sizes, prev, pager, next, jumper"
          @size-change="handleSizeChange"
          @current-change="handleCurrentChange"
        />
      </div>
    </el-card>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, computed, onMounted } from 'vue'
import { Search, Picture } from '@element-plus/icons-vue'
import { operatorOrderApi } from '@share/api/order'
import { operatorShopeeApi } from '@share/api/shopee'

interface Product {
  image: string
  name: string
  color?: string
  size?: string
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
  shopeeStatus?: string
  orderAmount: number
  paymentStatus?: string
  unsettledPayment?: string
  settledPayment?: string
  products?: Product[]
}

const loading = ref(false)
const activeTab = ref('all')
const showProductInfo = ref(true)

const summaryData = reactive({
  allOrders: { count: 156, amount: 1256800 },
  unsettledOrders: { count: 23, amount: 186500 },
  settledOrders: { count: 133, amount: 1070300 },
  adjustments: { count: 5, amount: -12500 }
})

const filterForm = reactive({
  shopKeyword: '',
  orderNo: '',
  settlementStatus: '',
  paymentStatus: '',
  startDate: null as Date | null,
  endDate: null as Date | null
})

const exportReport = ref(false)

const shopOptions = ref([
  { id: '1001', name: '示例店铺1' },
  { id: '1002', name: '示例店铺2' },
  { id: '1003', name: '示例店铺3' }
])

const orders = ref<Order[]>([])

const filteredOrders = computed(() => {
  let result = [...orders.value]
  if (activeTab.value === 'unsettled') {
    result = result.filter(order => order.unsettledPayment)
  } else if (activeTab.value === 'settled') {
    result = result.filter(order => order.settledPayment)
  } else if (activeTab.value === 'adjustments') {
    result = []
  }
  return result
})

const pagination = reactive({
  page: 1,
  pageSize: 10,
  total: 156
})

function formatAmount(value: number): string {
  return value.toLocaleString('en-US', { minimumFractionDigits: 2, maximumFractionDigits: 2 })
}

function getPaymentStatusType(status: string): string {
  const map: Record<string, string> = {
    paid: 'success',
    unpaid: 'warning',
    refunded: 'info'
  }
  return map[status] || 'info'
}

function getPaymentStatusText(status: string): string {
  const map: Record<string, string> = {
    paid: '已付款',
    unpaid: '未付款',
    refunded: '已退款'
  }
  return map[status] || '未知'
}

function handleSearch() {
  pagination.page = 1
  fetchOrders()
}

function handleReset() {
  filterForm.shopKeyword = ''
  filterForm.orderNo = ''
  filterForm.settlementStatus = ''
  filterForm.paymentStatus = ''
  filterForm.startDate = null
  filterForm.endDate = null
  handleSearch()
}

function handleSizeChange() {
  pagination.page = 1
  fetchOrders()
}

function handleCurrentChange() {
  fetchOrders()
}

async function fetchOrders() {
  loading.value = true
  try {
    const params: Record<string, any> = {
      page: pagination.page,
      page_size: pagination.pageSize
    }
    if (filterForm.shopKeyword) {
      params.shop_keyword = filterForm.shopKeyword
    }
    if (filterForm.orderNo) {
      params.order_sn = filterForm.orderNo
    }
    if (filterForm.settlementStatus) {
      params.settlement_status = filterForm.settlementStatus
    }
    if (filterForm.paymentStatus) {
      params.payment_status = filterForm.paymentStatus
    }
    if (filterForm.startDate) {
      params.start_time = filterForm.startDate.toISOString().split('T')[0]
    }
    if (filterForm.endDate) {
      params.end_time = filterForm.endDate.toISOString().split('T')[0]
    }

    const res = await operatorOrderApi.getOrderList(params)
    if (res.code === 0 && res.data) {
      orders.value = res.data.list.map((item: any) => ({
        orderNo: item.order_sn,
        orderTime: item.create_time,
        storeId: String(item.shop_id),
        storeName: item.shop_name || '-',
        shopeeOrderNo: item.order_sn,
        shopeeStatus: item.order_status,
        orderAmount: parseFloat(item.total_amount) || 0,
        paymentStatus: item.order_status === 'COMPLETED' ? 'paid' : 'unpaid',
        products: item.items || []
      }))
      pagination.total = res.data.total
    }
  } catch (error) {
    console.error('获取订单列表失败', error)
  } finally {
    loading.value = false
  }
}

async function fetchShopOptions() {
  try {
    const res = await operatorShopeeApi.getShopList({ page: 1, page_size: 100 })
    if (res.code === 0 && res.data) {
      shopOptions.value = res.data.list.map((item: any) => ({
        id: String(item.shopId),
        name: item.shopName
      }))
    }
  } catch (error) {
    console.error('获取店铺列表失败', error)
  }
}

function loadMockOrders() {
  orders.value = Array.from({ length: 6 }, (_, i) => ({
    orderNo: `X250904KQ2P078R`,
    orderTime: '2025-12-10 23:59:59',
    storeId: `S123456789${i}`,
    storeName: '示例文字占位符示例文...',
    shopeeOrderNo: `250904KQ2P0y8R`,
    shopeeStatus: i % 2 === 0 ? '待发货' : '已完成',
    orderAmount: i % 2 === 0 ? 36 : 198,
    paymentStatus: 'paid',
    unsettledPayment: i % 3 === 0 ? '8.00' : undefined,
    settledPayment: i % 3 !== 0 ? '16.00' : undefined,
    products: []
  }))
  pagination.total = 123
  summaryData.allOrders = { count: 245, amount: 38420 }
  summaryData.unsettledOrders = { count: 12, amount: 456 }
  summaryData.settledOrders = { count: 12, amount: 456 }
  summaryData.adjustments = { count: 12, amount: 456 }
}

onMounted(() => {
  fetchShopOptions()
  loadMockOrders()
})
</script>

<style scoped lang="scss">
.orders-page {
  .page-header {
    margin-bottom: 20px;
    
    .page-title {
      font-size: 20px;
      font-weight: 500;
      color: #303133;
      margin: 0;
    }
  }
  
  .summary-section-card {
    margin-bottom: 20px;
  }
  
  .section-title-wrapper {
    display: flex;
    align-items: center;
    justify-content: space-between;
    margin-bottom: 16px;
    
    .section-title {
      font-size: 16px;
      font-weight: 500;
      color: #303133;
    }

    .stat-link {
      font-size: 14px;
    }
  }
  
  .summary-card {
    background: #f8f9fa;
    border-radius: 8px;
    padding: 16px;
    border: 1px solid #ebeef5;
    
    .card-title {
      font-size: 14px;
      color: #909399;
      margin-bottom: 8px;
    }
    
    .card-value {
      display: flex;
      align-items: baseline;
      gap: 8px;
      
      .count {
        font-size: 24px;
        font-weight: 600;
        color: #303133;
        
        .unit {
          font-size: 14px;
          font-weight: normal;
          margin-left: 2px;
        }
      }
      
      .equals {
        color: #909399;
      }
      
      .amount {
        font-size: 16px;
        color: #303133;
        font-weight: 500;
      }
    }
  }

  .filter-section {
    display: flex;
    justify-content: space-between;
    align-items: center;
    margin-bottom: 20px;
    padding: 16px 20px;
    background: #fff;
    border-radius: 8px;
    border: 1px solid #ebeef5;

    .filter-left {
      display: flex;
      gap: 12px;
      align-items: center;
    }

    .filter-right {
      display: flex;
      gap: 8px;
    }

    .filter-input {
      width: 160px;
    }

    .filter-select {
      width: 140px;
    }

    .filter-date {
      width: 130px;
    }

    .date-separator {
      color: #909399;
    }
  }

  :deep(.el-input__wrapper),
  :deep(.el-select .el-select__wrapper) {
    border-radius: 4px;
  }
}

.orders-card {
  .orders-header {
    display: flex;
    justify-content: space-between;
    align-items: center;
    padding-bottom: 12px;
    border-bottom: 1px solid #e4e7ed;
    margin-bottom: 20px;

    .tab-buttons {
      display: flex;
      gap: 24px;

      .tab-btn {
        font-size: 14px;
        color: #909399;
        cursor: pointer;
        padding-bottom: 12px;
        border-bottom: 2px solid transparent;
        margin-bottom: -13px;

        &:hover {
          color: #303133;
        }

        &.active {
          color: #303133;
          font-weight: 500;
          border-bottom-color: #303133;
        }
      }
    }

    .action-buttons {
      display: flex;
      gap: 16px;
    }
  }

  .orders-list {
    display: flex;
    flex-direction: column;
    gap: 16px;
  }

  .order-item {
    border-bottom: 1px solid #ebeef5;
    padding: 16px 0;

    &:last-child {
      border-bottom: none;
    }

    .order-row-header {
      display: flex;
      justify-content: space-between;
      align-items: center;
      margin-bottom: 12px;

      .order-left {
        display: flex;
        align-items: center;
        gap: 12px;

        .order-number {
          font-weight: 500;
          color: #303133;
          font-size: 14px;
        }
      }

      .order-right {
        display: flex;
        gap: 24px;
        font-size: 14px;

        .payment-info {
          color: #909399;

          .highlight {
            color: #ff6a3a;
            font-weight: 500;
          }
        }

        .amount-info {
          color: #909399;

          .amount {
            color: #303133;
            font-weight: 500;
          }
        }
      }
    }

    .order-row-info {
      display: flex;
      justify-content: space-between;
      align-items: center;
      padding: 8px 16px;
      background: #fafafa;
      border-radius: 4px;

      .info-left {
        display: flex;
        gap: 32px;

        .info-item {
          font-size: 13px;
          color: #909399;
        }
      }

      .info-right {
        .shopee-status {
          font-size: 13px;
          color: #ff6a3a;
        }
      }
    }

    .products-section {
      padding: 16px;

      .product-item {
        display: flex;
        gap: 16px;
        padding: 12px 0;
        border-bottom: 1px solid #f0f0f0;

        &:last-child {
          border-bottom: none;
        }

        .product-image {
          width: 80px;
          height: 80px;
          border-radius: 4px;
          background: #f5f7fa;

          .image-slot {
            display: flex;
            align-items: center;
            justify-content: center;
            width: 100%;
            height: 100%;
            color: #909399;
            font-size: 24px;
          }
        }

        .product-details {
          flex: 1;

          .product-name {
            font-size: 14px;
            font-weight: 500;
            color: #303133;
            margin-bottom: 8px;
          }

          .product-specs {
            font-size: 12px;
            color: #909399;
            margin-bottom: 8px;
          }

          .product-price-info {
            display: flex;
            gap: 24px;
            font-size: 13px;
            color: #606266;
          }
        }
      }

      .order-settlement-row {
        display: flex;
        justify-content: flex-end;
        gap: 24px;
        padding-top: 12px;
        font-size: 14px;
        color: #303133;
      }
    }
  }

  .pagination-wrapper {
    display: flex;
    justify-content: flex-end;
    margin-top: 20px;
  }
}
</style>
