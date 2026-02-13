<template>
  <div class="finance-payment-page">
    <div class="page-header">
      <h1 class="page-title">我的回款</h1>
    </div>

    <!-- 回款总览 -->
    <div class="summary-section">
      <div class="summary-header">
        <span class="summary-title">回款总览</span>
        <el-button type="primary" link size="small">回款统计</el-button>
      </div>
      <el-row :gutter="20" class="summary-cards">
        <el-col :span="16">
          <div class="stat-card main-card">
            <div class="card-content">
              <div class="card-info">
                <div class="stat-label">可提现金额</div>
                <div class="stat-value large">NT${{ formatAmount(summaryData.withdrawable) }}</div>
                <div class="stat-sub">累计回款：NT${{ formatAmount(summaryData.totalPayment) }}</div>
              </div>
              <el-button type="primary" @click="handleWithdraw">提现</el-button>
            </div>
          </div>
        </el-col>
        <el-col :span="8">
          <div class="stat-card">
            <div class="stat-label">即将结算回款</div>
            <div class="stat-value">NT${{ formatAmount(summaryData.upcoming) }}</div>
            <div class="stat-sub">根据未结算订单预估即将结算金额</div>
          </div>
        </el-col>
      </el-row>
    </div>

    <!-- 回款列表 -->
    <el-card class="payment-card" v-loading="loading">
      <div class="payment-header">
        <div class="tab-buttons">
          <span 
            :class="['tab-btn', activeTab === 'all' ? 'active' : '']" 
            @click="activeTab = 'all'"
          >全部</span>
          <span 
            :class="['tab-btn', activeTab === 'payment' ? 'active' : '']" 
            @click="activeTab = 'payment'"
          >回款</span>
          <span 
            :class="['tab-btn', activeTab === 'withdraw' ? 'active' : '']" 
            @click="activeTab = 'withdraw'"
          >提现</span>
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
      </div>

      <el-table :data="paymentList" style="width: 100%">
        <el-table-column prop="date" label="日期" min-width="160" />
        <el-table-column prop="transType" label="交易类型" min-width="100" />
        <el-table-column prop="storeId" label="店铺编号" min-width="140" />
        <el-table-column prop="orderNo" label="订单编号" min-width="180" />
        <el-table-column prop="amount" label="交易金额" min-width="120">
          <template #default="{ row }">
            NT${{ formatAmount(row.amount) }}
          </template>
        </el-table-column>
        <el-table-column prop="balance" label="余额" min-width="140">
          <template #default="{ row }">
            NT${{ formatAmount(row.balance) }}
          </template>
        </el-table-column>
        <el-table-column prop="status" label="状态" min-width="100">
          <template #default="{ row }">
            {{ row.status }}
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
import { ElMessage } from 'element-plus'

const loading = ref(false)
const activeTab = ref('all')

const summaryData = reactive({
  withdrawable: 5450,
  totalPayment: 44543,
  upcoming: 123
})

const filterForm = reactive({
  keyword: '',
  startDate: null as Date | null,
  endDate: null as Date | null
})

const paymentList = ref([
  { date: '2026-12-12 23:59:59', transType: '回款收入', storeId: 'S1234567890', orderNo: 'X250904KQ2P078R', amount: 1000, balance: 223560.50, status: '已结算' },
  { date: '2026-12-12 23:59:59', transType: '回款收入', storeId: 'S1234567890', orderNo: 'X250904KQ2P078R', amount: 1000, balance: 223560.50, status: '已结算' },
  { date: '2026-12-12 23:59:59', transType: '回款收入', storeId: 'S1234567890', orderNo: 'X250904KQ2P078R', amount: 1000, balance: 223560.50, status: '已结算' },
  { date: '2026-12-12 23:59:59', transType: '回款收入', storeId: 'S1234567890', orderNo: 'X250904KQ2P078R', amount: 1000, balance: 223560.50, status: '已结算' },
  { date: '2026-12-12 23:59:59', transType: '回款收入', storeId: 'S1234567890', orderNo: 'X250904KQ2P078R', amount: 1000, balance: 223560.50, status: '已结算' },
  { date: '2026-12-12 23:59:59', transType: '回款收入', storeId: 'S1234567890', orderNo: 'X250904KQ2P078R', amount: 1000, balance: 223560.50, status: '已结算' },
  { date: '2026-12-12 23:59:59', transType: '回款收入', storeId: 'S1234567890', orderNo: 'X250904KQ2P078R', amount: 1000, balance: 223560.50, status: '已结算' },
  { date: '2026-12-12 23:59:59', transType: '回款收入', storeId: 'S1234567890', orderNo: 'X250904KQ2P078R', amount: 1000, balance: 223560.50, status: '已结算' },
  { date: '2026-12-12 23:59:59', transType: '回款收入', storeId: 'S1234567890', orderNo: 'X250904KQ2P078R', amount: 1000, balance: 223560.50, status: '已结算' },
  { date: '2026-12-12 23:59:59', transType: '回款收入', storeId: 'S1234567890', orderNo: 'X250904KQ2P078R', amount: 1000, balance: 223560.50, status: '已结算' }
])

function handleWithdraw() {
  ElMessage.info('提现功能开发中...')
}

const pagination = reactive({
  page: 1,
  pageSize: 10,
  total: 68
})

function formatAmount(value: number): string {
  return value.toLocaleString('en-US', { minimumFractionDigits: 2, maximumFractionDigits: 2 })
}

onMounted(() => {
  // 加载数据
})
</script>

<style scoped lang="scss">
.finance-payment-page {
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

    .summary-header {
      display: flex;
      justify-content: space-between;
      align-items: center;
      margin-bottom: 16px;

      .summary-title {
        font-size: 16px;
        font-weight: 500;
        color: #303133;
      }
    }

    .stat-card {
      background: #fff;
      border-radius: 8px;
      padding: 20px;
      border: 1px solid #ebeef5;
      height: 100%;

      &.main-card {
        border-color: #ff6a3a;

        .card-content {
          display: flex;
          justify-content: space-between;
          align-items: center;
        }

        .card-info {
          .stat-value.large {
            font-size: 32px;
          }
        }
      }

      .stat-label {
        font-size: 12px;
        color: #909399;
        margin-bottom: 8px;
      }

      .stat-value {
        font-size: 28px;
        font-weight: 600;
        color: #303133;
        margin-bottom: 8px;
      }

      .stat-sub {
        font-size: 12px;
        color: #909399;
      }
    }
  }
}

.payment-card {
  .payment-header {
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
      gap: 12px;
      align-items: center;

      .search-input {
        width: 150px;
      }

      .filter-date {
        width: 130px;
      }

      .date-separator {
        color: #909399;
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
