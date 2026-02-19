<template>
  <div class="finance-bills-page">
    <div class="page-header">
      <h1 class="page-title">我的账单</h1>
    </div>

    <!-- 账单总览 -->
    <div class="summary-section">
      <div class="summary-title">账单总览</div>
      <el-row :gutter="20" class="summary-cards">
        <el-col :span="6">
          <div class="stat-card">
            <div class="stat-label">预估未结算回款</div>
            <div class="stat-value">NT${{ formatAmount(summaryData.unsettled) }}</div>
          </div>
        </el-col>
        <el-col :span="6">
          <div class="stat-card">
            <div class="stat-label">已结算回款</div>
            <div class="stat-value">NT${{ formatAmount(summaryData.settled) }}</div>
          </div>
        </el-col>
        <el-col :span="6">
          <div class="stat-card">
            <div class="stat-label">账款调整</div>
            <div class="stat-value">NT${{ formatAmount(summaryData.adjustments) }}</div>
          </div>
        </el-col>
        <el-col :span="6">
          <div class="stat-card highlight">
            <div class="stat-label">预估回款总额（未结算+已结算+账款调整）</div>
            <div class="stat-value">NT${{ formatAmount(summaryData.total) }}</div>
          </div>
        </el-col>
      </el-row>
    </div>

    <!-- 账单列表 -->
    <el-card class="bills-card" v-loading="loading">
      <div class="bills-header">
        <div class="tab-buttons">
          <span 
            :class="['tab-btn', activeTab === 'unsettled' ? 'active' : '']" 
            @click="activeTab = 'unsettled'"
          >未结算</span>
          <span 
            :class="['tab-btn', activeTab === 'settled' ? 'active' : '']" 
            @click="activeTab = 'settled'"
          >已结算</span>
          <span 
            :class="['tab-btn', activeTab === 'adjustments' ? 'active' : '']" 
            @click="activeTab = 'adjustments'"
          >账款调整</span>
        </div>
        <div class="action-buttons">
          <el-input v-model="filterForm.keyword" placeholder="快速搜索" clearable class="search-input">
            <template #suffix>
              <el-icon><Search /></el-icon>
            </template>
          </el-input>
          <el-checkbox v-model="exportReport">导出报表</el-checkbox>
        </div>
      </div>

      <el-table :data="billList" style="width: 100%">
        <el-table-column prop="date" label="日期" min-width="160" />
        <el-table-column prop="storeId" label="店铺编号" min-width="140" />
        <el-table-column prop="orderNo" label="订单编号" min-width="180" />
        <el-table-column prop="orderStatus" label="订单状态" min-width="100" />
        <el-table-column prop="orderTimer" label="订单计时" min-width="120" />
        <el-table-column prop="orderAmount" label="订单金额" min-width="120">
          <template #default="{ row }">
            NT${{ formatAmount(row.orderAmount) }}
          </template>
        </el-table-column>
        <el-table-column prop="unsettledPayment" label="未结算回款" min-width="120">
          <template #default="{ row }">
            NT${{ formatAmount(row.unsettledPayment) }}
          </template>
        </el-table-column>
      </el-table>

      <div class="pagination-wrapper">
        <el-pagination
          v-model:current-page="pagination.page"
          v-model:page-size="pagination.pageSize"
          :total="pagination.total"
          :page-sizes="[10, 20, 50, 100]"
          layout="total, sizes, prev, pager, next, jumper"
        />
      </div>
    </el-card>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted } from 'vue'
import { Search } from '@element-plus/icons-vue'

const loading = ref(false)
const activeTab = ref('unsettled')
const exportReport = ref(false)

const summaryData = reactive({
  unsettled: 245,
  settled: 12300,
  adjustments: 12300,
  total: 123456
})

const filterForm = reactive({
  keyword: ''
})

const billList = ref([
  { date: '2026-12-12 23:59:59', storeId: 'S1234567890', orderNo: 'X250904KQ2P078R', orderStatus: '待发货', orderTimer: '1天 03:23:42', orderAmount: 68, unsettledPayment: 8 },
  { date: '2026-12-12 23:59:59', storeId: 'S1234567890', orderNo: 'X250904KQ2P078R', orderStatus: '待发货', orderTimer: '1天 03:23:42', orderAmount: 68, unsettledPayment: 8 },
  { date: '2026-12-12 23:59:59', storeId: 'S1234567890', orderNo: 'X250904KQ2P078R', orderStatus: '待发货', orderTimer: '1天 03:23:42', orderAmount: 68, unsettledPayment: 8 },
  { date: '2026-12-12 23:59:59', storeId: 'S1234567890', orderNo: 'X250904KQ2P078R', orderStatus: '待发货', orderTimer: '1天 03:23:42', orderAmount: 68, unsettledPayment: 8 },
  { date: '2026-12-12 23:59:59', storeId: 'S1234567890', orderNo: 'X250904KQ2P078R', orderStatus: '待发货', orderTimer: '1天 03:23:42', orderAmount: 68, unsettledPayment: 8 },
  { date: '2026-12-12 23:59:59', storeId: 'S1234567890', orderNo: 'X250904KQ2P078R', orderStatus: '待发货', orderTimer: '1天 03:23:42', orderAmount: 68, unsettledPayment: 8 },
  { date: '2026-12-12 23:59:59', storeId: 'S1234567890', orderNo: 'X250904KQ2P078R', orderStatus: '待发货', orderTimer: '1天 03:23:42', orderAmount: 68, unsettledPayment: 8 },
  { date: '2026-12-12 23:59:59', storeId: 'S1234567890', orderNo: 'X250904KQ2P078R', orderStatus: '待发货', orderTimer: '1天 03:23:42', orderAmount: 68, unsettledPayment: 8 },
  { date: '2026-12-12 23:59:59', storeId: 'S1234567890', orderNo: 'X250904KQ2P078R', orderStatus: '待发货', orderTimer: '1天 03:23:42', orderAmount: 68, unsettledPayment: 8 },
  { date: '2026-12-12 23:59:59', storeId: 'S1234567890', orderNo: 'X250904KQ2P078R', orderStatus: '待发货', orderTimer: '1天 03:23:42', orderAmount: 68, unsettledPayment: 8 }
])

const pagination = reactive({
  page: 1,
  pageSize: 10,
  total: 45
})

function formatAmount(value: number): string {
  return value.toLocaleString('en-US', { minimumFractionDigits: 2, maximumFractionDigits: 2 })
}

onMounted(() => {
  // 加载数据
})
</script>

<style scoped lang="scss">
.finance-bills-page {
  .page-header {
    margin-bottom: 20px;

    .page-title {
      font-size: 20px;
      font-weight: 500;
      color: #303133;
      margin: 0;
    }
  }

  .summary-section {
    margin-bottom: 20px;

    .summary-title {
      font-size: 16px;
      font-weight: 500;
      color: #303133;
      margin-bottom: 16px;
    }

    .stat-card {
      background: #fff;
      border-radius: 8px;
      padding: 20px;
      border: 1px solid #ebeef5;
      height: 100%;

      &.highlight {
        border-color: #ff6a3a;
      }

      .stat-label {
        font-size: 12px;
        color: #909399;
        margin-bottom: 12px;
      }

      .stat-value {
        font-size: 28px;
        font-weight: 600;
        color: #303133;
      }
    }
  }
}

.bills-card {
  .bills-header {
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
      align-items: center;

      .search-input {
        width: 150px;
      }
    }
  }

  .pagination-wrapper {
    display: flex;
    justify-content: center;
    margin-top: 20px;
  }
}
</style>
