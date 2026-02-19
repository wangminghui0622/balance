<template>
  <div class="orders-page">
    <div class="page-header">
      <h1 class="page-title">我的订单</h1>
      <span class="order-stats-link">订单统计</span>
    </div>

    <!-- 订单概览 -->
    <el-card class="summary-section-card">
      <div class="summary-section">
        <div class="section-title-wrapper">
          <span class="title-bar"></span>
          <span class="section-title">订单概览</span>
        </div>
        <el-row :gutter="20" class="summary-cards">
          <el-col :xs="24" :sm="12" :md="6" :lg="6" :xl="6">
            <div class="summary-card">
              <div class="card-content">
                <div class="card-title">全部订单(NT$)</div>
                <div class="card-value">
                  <span class="count">
                    {{ summaryData.allOrders.count }}<span class="unit">单</span>
                  </span>
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
                  <span class="count">
                    {{ summaryData.unsettledOrders.count }}<span class="unit">单</span>
                  </span>
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
                  <span class="count">
                    {{ summaryData.settledOrders.count }}<span class="unit">单</span>
                  </span>
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
                  <span class="count">
                    {{ summaryData.adjustments.count }}<span class="unit">单</span>
                  </span>
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
              <el-input
                v-model="filterForm.shopKeyword"
                placeholder="请输入店铺名称或编号"
                clearable
                style="width: 100%"
              >
                <template #prefix>
                  <el-icon><Search /></el-icon>
                </template>
              </el-input>
            </el-form-item>
          </el-col>
          <el-col :xs="24" :sm="8">
            <el-form-item label="订单编号">
              <el-input
                v-model="filterForm.orderNo"
                placeholder="请输入订单编号"
                clearable
                style="width: 100%"
              >
                <template #prefix>
                  <el-icon><Search /></el-icon>
                </template>
              </el-input>
            </el-form-item>
          </el-col>
          <el-col :xs="24" :sm="8">
            <el-form-item label="订单结算状态">
              <el-select v-model="filterForm.settlementStatus" placeholder="请选择" style="width: 50%">
                <el-option label="全部" value="all" />
                <el-option label="未结算" value="unsettled" />
                <el-option label="已结算" value="settled" />
              </el-select>
            </el-form-item>
          </el-col>
        </el-row>
        <el-row :gutter="20" style="margin-top: 10px;">
          <el-col :xs="24" :sm="8">
            <el-form-item label="订单付款状态">
              <el-select v-model="filterForm.paymentStatus" placeholder="请选择" style="width: 50%">
                <el-option label="全部" value="all" />
                <el-option label="未付款" value="unpaid" />
                <el-option label="已付款" value="paid" />
                <el-option label="已退款" value="refunded" />
              </el-select>
            </el-form-item>
          </el-col>
          <el-col :xs="24" :sm="8">
            <el-form-item label="日期范围">
              <el-date-picker
                v-model="filterForm.dateRange"
                type="daterange"
                range-separator="至"
                start-placeholder="开始日期"
                end-placeholder="结束日期"
                format="YYYY-MM-DD"
                value-format="YYYY-MM-DD"
                style="width: 100%"
              />
            </el-form-item>
          </el-col>
          <el-col :xs="24" :sm="8">
            <div style="display: flex; justify-content: flex-end; gap: 10px;">
              <el-button type="primary" @click="handleSearch">
                <el-icon><Search /></el-icon>
                查询
              </el-button>
              <el-button @click="handleReset">重置</el-button>
            </div>
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
          <el-tab-pane label="账款调整" name="adjustments" />
        </el-tabs>
        <div class="action-buttons">
          <el-button :icon="View" @click="toggleProductInfo">
            {{ showProductInfo ? '隐藏商品信息' : '显示商品信息' }}
          </el-button>
          <el-button @click="handleExport">
            导出报表
          </el-button>
        </div>
      </div>

      <!-- 订单列表 -->
      <div class="orders-list">
        <div v-for="(order, index) in orderList" :key="index" class="order-item">
          <!-- 订单头部 -->
          <div class="order-header">
            <div class="order-number">
              订单编号: {{ order.orderNo }}
              <el-tag v-if="order.paymentStatus" :type="getPaymentStatusType(order.paymentStatus)" size="small" style="margin-left: 8px">
                {{ getPaymentStatusText(order.paymentStatus) }}
              </el-tag>
            </div>
            <div class="order-amount-info">
              <span v-if="order.adjustmentLabel1">{{ order.adjustmentLabel1 }}</span>
              <span v-if="order.adjustmentLabel2">{{ order.adjustmentLabel2 }}</span>
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
            <!-- 虾皮订单金额/账款调整（使用服务器返回的完整字符串） -->
            <div class="order-shopee-amount" v-if="order.adjustmentLabel3">
              {{ order.adjustmentLabel3 }}
            </div>
          </div>
        </div>

        <!-- 空状态 -->
        <el-empty v-if="orderList.length === 0" description="暂无订单数据" />
      </div>

      <!-- 分页 -->
      <div class="pagination-wrapper" v-if="pagination.total > 0">
        <el-pagination
          v-model:current-page="pagination.page"
          v-model:page-size="pagination.pageSize"
          :total="pagination.total"
          :page-sizes="[10, 20, 50, 100]"
          layout="total, sizes, prev, pager, next, jumper"
          @size-change="handleSizeChange"
          @current-change="handlePageChange"
        />
      </div>
    </el-card>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, computed, onMounted } from 'vue'
import { Search, View, Picture } from '@element-plus/icons-vue'
import { ElMessage } from 'element-plus'

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
  orderAmount: string
  paymentStatus?: string
  shopeeAmount?: string
  products?: Product[]
  // 服务器返回的显示标签
  adjustmentLabel1?: string
  adjustmentLabel2?: string
  adjustmentLabel3?: string
}

const activeTab = ref('all')
const showProductInfo = ref(true)
const loading = ref(false)

const getDefaultDateRange = (): string[] => {
  const today = new Date()
  const tenDaysAgo = new Date()
  tenDaysAgo.setDate(today.getDate() - 10)
  
  const formatDate = (date: Date): string => {
    const year = date.getFullYear()
    const month = String(date.getMonth() + 1).padStart(2, '0')
    const day = String(date.getDate()).padStart(2, '0')
    return `${year}-${month}-${day}`
  }
  
  return [formatDate(tenDaysAgo), formatDate(today)]
}

const filterForm = reactive({
  shopKeyword: '',
  orderNo: '',
  settlementStatus: 'all',
  paymentStatus: 'all',
  dateRange: getDefaultDateRange() as string[] | null,
})

const pagination = reactive({
  page: 1,
  pageSize: 10,
  total: 0,
})

const summaryData = computed(() => ({
  allOrders: { count: 245, amount: 38420.00 },
  unsettledOrders: { count: 12, amount: 456.00 },
  settledOrders: { count: 12, amount: 456.00 },
  adjustments: { count: 1, amount: 34.00 }
}))

const orderList = ref<Order[]>([])

const fetchOrders = async () => {
  loading.value = true
  try {
    // 模拟数据（实际使用时从服务器获取，包含adjustment_label_1/2/3字段）
    orderList.value = [
      // 普通订单（未结算）
      {
        orderNo: 'X250904KQ2P078R',
        orderTime: '2025-12-10 23:59:59',
        storeId: 'S1234567890',
        storeName: '示例文字占位符示例文...',
        shopeeOrderNo: '250904KQ2P078R',
        shopeeStatus: '待发货',
        orderAmount: '36.00',
        paymentStatus: 'paid',
        shopeeAmount: '46.00',
        adjustmentLabel1: '未结算佣金：NT$8.00',
        adjustmentLabel2: '订单金额：NT$36.00',
        adjustmentLabel3: '虾皮订单金额：NT$46.00',
        products: [
          {
            image: '',
            name: '商品名称示例文字占位符替换即可文字占位符替换即可',
            color: 'xxx',
            size: 'xxx',
            unitPrice: '46.00',
            quantity: 1,
            subtotal: '46.00'
          }
        ]
      },
      // 多商品订单（已结算）
      {
        orderNo: 'X250904KQ2P078R',
        orderTime: '2025-12-10 23:59:59',
        storeId: 'S1234567891',
        storeName: '示例文字占位符示例文...',
        shopeeOrderNo: '250904KQ2P078R',
        shopeeStatus: '已完成',
        orderAmount: '198.00',
        paymentStatus: 'paid',
        shopeeAmount: '208.00',
        adjustmentLabel1: '已结算佣金：NT$16.00',
        adjustmentLabel2: '订单金额：NT$198.00',
        adjustmentLabel3: '虾皮订单金额：NT$208.00',
        products: [
          {
            image: '',
            name: '商品名称示例文字占位符替换即可文字占位符替换即可',
            color: 'xxx',
            size: 'xxx',
            unitPrice: '46.00',
            quantity: 2,
            subtotal: '92.00'
          },
          {
            image: '',
            name: '商品名称示例文字占位符替换即可文字占位符替换即可',
            color: 'xxx',
            size: 'xxx',
            unitPrice: '116.00',
            quantity: 1,
            subtotal: '116.00'
          }
        ]
      },
      // 账款调整订单（2个商品）
      {
        orderNo: 'X250904KQ2P078R',
        orderTime: '2025-12-10 23:59:59',
        storeId: 'S1234567890',
        storeName: '示例文字占位符示例文...',
        shopeeOrderNo: '250904KQ2P078R',
        shopeeStatus: '待发货',
        orderAmount: '36.00',
        paymentStatus: 'paid',
        shopeeAmount: '46.00',
        adjustmentLabel1: '账款调整佣金：NT$8.00',
        adjustmentLabel2: '订单账款调整：NT$36.00',
        adjustmentLabel3: '虾皮订单账款调整：NT$46.00',
        products: [
          {
            image: '',
            name: '商品名称示例文字占位符替换即可文字占位符替换即可',
            color: 'xxx',
            size: 'xxx',
            unitPrice: '46.00',
            quantity: 1,
            subtotal: '46.00'
          },
          {
            image: '',
            name: '商品名称示例文字占位符替换即可文字占位符替换即可',
            color: 'xxx',
            size: 'xxx',
            unitPrice: '36.00',
            quantity: 1,
            subtotal: '36.00'
          }
        ]
      }
    ]
    pagination.total = 123
  } catch (err: any) {
    console.error('获取订单列表失败:', err)
    ElMessage.error('获取订单列表失败')
  } finally {
    loading.value = false
  }
}

onMounted(() => {
  fetchOrders()
})

const formatAmount = (amount: number): string => {
  return amount.toLocaleString('zh-CN', {
    minimumFractionDigits: 2,
    maximumFractionDigits: 2
  })
}

const getPaymentStatusType = (status: string): string => {
  const statusMap: Record<string, string> = {
    paid: 'success',
    unpaid: 'warning',
    refunded: 'info'
  }
  return statusMap[status] || 'info'
}

const getPaymentStatusText = (status: string): string => {
  const statusMap: Record<string, string> = {
    paid: '付款状态',
    unpaid: '未付款',
    refunded: '已退款'
  }
  return statusMap[status] || '未知'
}

const handleSearch = () => {
  pagination.page = 1
  fetchOrders()
}

const handleReset = () => {
  filterForm.shopKeyword = ''
  filterForm.orderNo = ''
  filterForm.settlementStatus = 'all'
  filterForm.paymentStatus = 'all'
  filterForm.dateRange = getDefaultDateRange()
  pagination.page = 1
  fetchOrders()
  ElMessage.info('已重置筛选条件')
}

const handleTabChange = () => {
  pagination.page = 1
  fetchOrders()
}

const toggleProductInfo = () => {
  showProductInfo.value = !showProductInfo.value
}

const handleExport = () => {
  ElMessage.info('导出功能开发中...')
}

const handleSizeChange = (size: number) => {
  pagination.pageSize = size
  pagination.page = 1
  fetchOrders()
}

const handlePageChange = (page: number) => {
  pagination.page = page
  fetchOrders()
}
</script>

<style scoped lang="scss">
.orders-page {
  padding: 20px;
  background-color: #f5f7fa;
  min-height: calc(100vh - 60px);
}

.page-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 16px;

  .page-title {
    font-size: 20px;
    font-weight: 500;
    color: #303133;
    margin: 0;
  }

  .order-stats-link {
    font-size: 14px;
    color: #909399;
    cursor: pointer;

    &:hover {
      color: #ff6a3a;
    }
  }
}

.summary-section-card {
  margin-bottom: 20px;
  border-radius: 8px;

  :deep(.el-card__body) {
    padding: 20px;
  }
}

.summary-section {
  .section-title-wrapper {
    display: flex;
    align-items: center;
    gap: 8px;
    margin-bottom: 16px;

    .title-bar {
      width: 3px;
      height: 16px;
      background-color: #ff6a3a;
      border-radius: 2px;
      flex-shrink: 0;
    }

    .section-title {
      font-size: 16px;
      font-weight: 600;
      color: #303133;
    }
  }

  .summary-cards {
    .summary-card {
      padding: 16px;
      background-color: #f5f7fa;
      border-radius: 8px;
      border: 1px solid #ebeef5;
      transition: box-shadow 0.3s;

      &:hover {
        box-shadow: 0 2px 12px 0 rgba(0, 0, 0, 0.1);
      }

      .card-content {
        text-align: center;

        .card-title {
          font-size: 14px;
          color: #909399;
          margin-bottom: 12px;
        }

        .card-value {
          display: flex;
          align-items: center;
          justify-content: center;
          gap: 16px;

          .count {
            font-size: 30px;
            font-weight: 600;
            color: #303133;

            .unit {
              font-size: 12px;
              font-weight: 400;
              margin-left: 2px;
            }
          }

          .equals {
            font-size: 18px;
            color: #606266;
          }

          .amount {
            font-size: 30px;
            font-weight: 600;
            color: #303133;
          }
        }
      }
    }
  }
}

.filter-card {
  margin-bottom: 20px;
  border-radius: 8px;

  :deep(.el-input__wrapper),
  :deep(.el-select .el-select__wrapper) {
    border-radius: 30px !important;
  }
}

.orders-card {
  border-radius: 8px;

  :deep(.el-card__body) {
    padding: 0;
  }

  .orders-header {
    display: flex;
    justify-content: space-between;
    align-items: center;
    padding: 20px;
    border-bottom: 1px solid #ebeef5;

    .order-tabs {
      flex: 1;
    }

    .action-buttons {
      display: flex;
      gap: 12px;
    }
  }

  .orders-list {
    padding: 20px;
  }

  .order-item {
    padding: 20px;
    margin-bottom: 16px;
    background-color: #ffffff;
    border: 1px solid #ebeef5;
    border-radius: 8px;
    transition: box-shadow 0.3s;

    &:hover {
      box-shadow: 0 2px 12px 0 rgba(0, 0, 0, 0.1);
    }

    .order-header {
      display: flex;
      justify-content: space-between;
      align-items: center;
      margin-bottom: 16px;
      padding-bottom: 12px;
      border-bottom: 1px solid #f5f7fa;

      .order-number {
        font-size: 16px;
        font-weight: 600;
        color: #303133;
      }

      .order-amount-info {
        display: flex;
        gap: 16px;
        font-size: 14px;
        color: #606266;
      }
    }

    .order-info {
      margin-bottom: 16px;
      background-color: #ebeef5;
      border-radius: 4px;
      padding: 10px 14px;

      .order-info-line {
        display: flex;
        font-size: 12px;
        color: #909399;
        gap: 32px;
      }

      .info-item-right {
        margin-left: auto;
      }
    }

    .products-section {
      margin-top: 16px;
      padding: 16px;
      background-color: #fafafa;
      border-radius: 6px;

      .product-item {
        display: grid;
        grid-template-columns: 80px 1fr;
        gap: 16px;
        padding: 12px 0;
        border-bottom: 1px solid #ebeef5;

        &:last-child {
          border-bottom: none;
        }

        .product-image {
          width: 80px;
          height: 80px;
          border-radius: 6px;
          background-color: #f5f7fa;

          .image-slot {
            display: flex;
            justify-content: center;
            align-items: center;
            width: 100%;
            height: 100%;
            color: #909399;
            font-size: 24px;
          }
        }

        .product-details {
          min-width: 0;
          display: grid;
          grid-template-columns: minmax(0, 1fr) 280px;
          grid-template-rows: auto auto;
          column-gap: 16px;
          row-gap: 6px;

          .product-name {
            font-size: 14px;
            color: #303133;
            font-weight: 500;
            line-height: 1.5;
            grid-column: 1 / -1;
            overflow: hidden;
            text-overflow: ellipsis;
            white-space: nowrap;
          }

          .product-specs {
            font-size: 12px;
            color: #909399;
            grid-column: 1;
            grid-row: 2;
          }

          .product-price-info {
            display: flex;
            gap: 24px;
            font-size: 13px;
            color: #606266;
            grid-column: 2;
            grid-row: 2;
            justify-content: flex-end;
            align-items: center;
            white-space: nowrap;
          }
        }
      }

      .order-shopee-amount {
        font-size: 12px;
        color: #909399;
        text-align: right;
        padding-top: 12px;
        border-top: 1px solid #ebeef5;
      }
    }
  }
}

.pagination-wrapper {
  padding: 20px;
  display: flex;
  justify-content: center;
}

@media (max-width: 768px) {
  .orders-page {
    padding: 12px;
  }

  .orders-header {
    flex-direction: column;
    align-items: flex-start !important;
    gap: 16px;
  }

  .order-header {
    flex-direction: column;
    align-items: flex-start !important;
    gap: 12px;
  }
}
</style>
