<template>
  <div class="orders-page">
    <div class="page-header">
      <h1 class="page-title">订单管理</h1>
    </div>
    <el-card class="summary-section-card">
      <div class="summary-section">
        <div class="section-title-wrapper">
          <span class="title-bar"></span>
          <span class="section-title">我的店铺({{ shopCount }})</span>
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
                <div class="card-title">回款调整(NT$)</div>
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
    <el-card class="filter-card">
      <el-form :model="filterForm" class="filter-form">
        <el-row :gutter="20">
          <el-col :xs="24" :sm="8">
            <el-form-item label="店铺名称/店铺编号">
              <el-select v-model="filterForm.shopId" placeholder="全部" clearable filterable style="width: 100%">
                <el-option label="全部" value="" />
                <el-option v-for="shop in shopOptions" :key="shop.id" :label="shop.name" :value="shop.id" />
              </el-select>
            </el-form-item>
          </el-col>
          <el-col :xs="24" :sm="8">
            <el-form-item label="订单编号">
              <el-input v-model="filterForm.orderNo" placeholder="请输入订单编号" clearable />
            </el-form-item>
          </el-col>
          <el-col :xs="24" :sm="8">
            <el-form-item label="订单状态">
              <el-select v-model="filterForm.status" placeholder="全部" clearable style="width: 100%">
                <el-option label="全部" value="" />
                <el-option label="未结算" value="unsettled" />
                <el-option label="已结算" value="settled" />
              </el-select>
            </el-form-item>
          </el-col>
        </el-row>
        <el-row :gutter="20">
          <el-col :xs="24" :sm="8">
            <el-form-item label="订单日期">
              <el-date-picker
                v-model="filterForm.dateRange"
                type="daterange"
                range-separator="至"
                start-placeholder="开始日期"
                end-placeholder="结束日期"
                style="width: 100%"
              />
            </el-form-item>
          </el-col>
          <el-col :xs="24" :sm="16" class="filter-buttons">
            <el-button type="primary" @click="handleSearch">查询</el-button>
            <el-button @click="handleReset">重置</el-button>
          </el-col>
        </el-row>
      </el-form>
    </el-card>

    <!-- 订单列表区域 -->
    <el-card class="orders-card" v-loading="loading">
      <div class="orders-header">
        <el-tabs v-model="activeTab" @tab-change="handleTabChange" class="order-tabs">
          <el-tab-pane label="全部订单" name="all" />
          <el-tab-pane label="未结算" name="unsettled" />
          <el-tab-pane label="已结算" name="settled" />
          <el-tab-pane label="回款调整" name="adjustments" />
        </el-tabs>
        <div class="action-buttons">
          <el-button :icon="View" @click="toggleProductInfo">
            {{ showProductInfo ? '隐藏商品信息' : '显示商品信息' }}
          </el-button>
          <el-button @click="handleExport">导出报表</el-button>
        </div>
      </div>

      <!-- 订单列表 -->
      <div class="orders-list">
        <div v-for="(order, index) in filteredOrders" :key="index" class="order-item">
          <!-- 订单头部 -->
          <div class="order-header">
            <div class="order-number">
              订单编号: {{ order.orderNo }}
              <el-tag v-if="order.paymentStatus" :type="getPaymentStatusType(order.paymentStatus)" size="small" style="margin-left: 8px">
                {{ getPaymentStatusText(order.paymentStatus) }}
              </el-tag>
            </div>
            <div class="order-amount-info">
              <span v-if="order.unsettledPayment">未结算回款: NT${{ order.unsettledPayment }}</span>
              <span>订单金额: NT${{ formatAmount(order.orderAmount) }}</span>
            </div>
          </div>

          <!-- 订单信息 -->
          <div class="order-info">
            <div class="order-info-line">
              <span class="info-item">下单时间: {{ order.orderTime }}</span>
              <span class="info-item">店铺编号: {{ order.storeId }}</span>
              <span class="info-item">店铺名称: {{ order.storeName }}</span>
              <span v-if="order.shopeeOrderNo" class="info-item">虾皮订单号: {{ order.shopeeOrderNo }}</span>
              <span v-if="order.shopeeStatus" class="info-item info-item-right">虾皮订单状态: {{ order.shopeeStatus }}</span>
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
import { View, Picture } from '@element-plus/icons-vue'
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
const shopCount = ref(3)
const activeTab = ref('all')
const showProductInfo = ref(true)

const summaryData = reactive({
  allOrders: { count: 156, amount: 1256800 },
  unsettledOrders: { count: 23, amount: 186500 },
  settledOrders: { count: 133, amount: 1070300 },
  adjustments: { count: 5, amount: -12500 }
})

const filterForm = reactive({
  shopId: '',
  orderNo: '',
  status: '',
  dateRange: null as [Date, Date] | null
})

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

function handleTabChange() {
  pagination.page = 1
}

function toggleProductInfo() {
  showProductInfo.value = !showProductInfo.value
}

function handleExport() {
  console.log('导出报表')
}

function handleSearch() {
  pagination.page = 1
  fetchOrders()
}

function handleReset() {
  filterForm.shopId = ''
  filterForm.orderNo = ''
  filterForm.status = ''
  filterForm.dateRange = null
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
    if (filterForm.shopId) {
      params.shop_id = filterForm.shopId
    }
    if (filterForm.orderNo) {
      params.order_sn = filterForm.orderNo
    }
    if (filterForm.status) {
      params.status = filterForm.status
    }
    if (filterForm.dateRange && filterForm.dateRange.length === 2) {
      params.start_time = filterForm.dateRange[0].toISOString().split('T')[0]
      params.end_time = filterForm.dateRange[1].toISOString().split('T')[0]
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

onMounted(() => {
  fetchShopOptions()
  fetchOrders()
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
    margin-bottom: 16px;
    
    .title-bar {
      width: 4px;
      height: 16px;
      background: #ff6a3a;
      border-radius: 2px;
      margin-right: 8px;
    }
    
    .section-title {
      font-size: 16px;
      font-weight: 500;
      color: #303133;
    }
  }
  
  .summary-card {
    background: #f8f9fa;
    border-radius: 8px;
    padding: 16px;
    
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
        color: #ff6a3a;
        font-weight: 500;
      }
    }
  }
  
  .filter-card {
    margin-bottom: 20px;
    
    .filter-buttons {
      display: flex;
      align-items: flex-end;
      justify-content: flex-end;
      padding-bottom: 18px;
    }
  }
}

.orders-card {
  .orders-header {
    display: flex;
    justify-content: space-between;
    align-items: center;
    margin-bottom: 20px;
    border-bottom: 1px solid #e4e7ed;

    .order-tabs {
      flex: 1;

      :deep(.el-tabs__header) {
        margin-bottom: 0;
      }

      :deep(.el-tabs__nav-wrap::after) {
        display: none;
      }
    }

    .action-buttons {
      display: flex;
      gap: 10px;
    }
  }

  .orders-list {
    display: flex;
    flex-direction: column;
    gap: 16px;
  }

  .order-item {
    border: 1px solid #e4e7ed;
    border-radius: 8px;
    overflow: hidden;

    .order-header {
      display: flex;
      justify-content: space-between;
      align-items: center;
      padding: 12px 16px;
      background: #f5f7fa;
      border-bottom: 1px solid #e4e7ed;

      .order-number {
        font-weight: 500;
        color: #303133;
      }

      .order-amount-info {
        display: flex;
        gap: 24px;
        color: #606266;
        font-size: 14px;
      }
    }

    .order-info {
      padding: 12px 16px;
      background: #fafafa;
      border-bottom: 1px solid #e4e7ed;

      .order-info-line {
        display: flex;
        flex-wrap: wrap;
        gap: 24px;
        font-size: 13px;
        color: #606266;

        .info-item-right {
          margin-left: auto;
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
