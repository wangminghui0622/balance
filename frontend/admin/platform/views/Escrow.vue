<template>
  <div class="escrow-page">
    <!-- 页面标题 -->
    <div class="page-header">
      <h1 class="page-title">托管账户</h1>
    </div>

    <!-- 托管账户余额 -->
    <div class="escrow-summary">
      <div class="summary-title">托管账户余额</div>
      <div class="summary-cards">
        <div class="summary-card">
          <div class="card-label">余额</div>
          <div class="card-value">¥{{ summaryData.balance }}</div>
        </div>
        <div class="card-divider">=</div>
        <div class="summary-card">
          <div class="card-label">订单付款(代收)</div>
          <div class="card-value">¥{{ summaryData.orderPayment }}</div>
        </div>
        <div class="card-divider">-</div>
        <div class="summary-card">
          <div class="card-label">订单结算(代付)</div>
          <div class="card-value">¥{{ summaryData.orderSettlement }}</div>
        </div>
        <div class="card-divider">+</div>
        <div class="summary-card">
          <div class="card-label">账款调整</div>
          <div class="card-value">¥{{ summaryData.adjustment }}</div>
        </div>
      </div>
    </div>

    <!-- 交易列表 -->
    <el-card class="escrow-card" shadow="never">
      <!-- Tab和搜索 -->
      <div class="escrow-header">
        <el-tabs v-model="activeTab" class="escrow-tabs" @tab-change="handleTabChange">
          <el-tab-pane label="全部" name="all" />
          <el-tab-pane label="订单付款" name="payment" />
          <el-tab-pane label="订单结算" name="settlement" />
          <el-tab-pane label="账款调整" name="adjustment" />
        </el-tabs>
        <div class="header-actions">
          <el-input
            v-model="searchKeyword"
            placeholder="快速搜索"
            :prefix-icon="Search"
            clearable
            style="width: 160px"
            @input="handleSearch"
          />
          <el-date-picker
            v-model="dateRange"
            type="daterange"
            range-separator="-"
            start-placeholder="开始日期"
            end-placeholder="结束日期"
            value-format="YYYY-MM-DD"
            style="width: 240px"
            @change="handleDateChange"
          />
        </div>
      </div>

      <!-- 交易表格 -->
      <el-table :data="transactionList" style="width: 100%" v-loading="loading">
        <el-table-column prop="date" label="日期" min-width="160" />
        <el-table-column prop="type" label="交易类型" min-width="100" />
        <el-table-column prop="storeId" label="店铺编号" min-width="120" />
        <el-table-column prop="orderNo" label="订单编号" min-width="160" />
        <el-table-column prop="amount" label="交易金额" min-width="120">
          <template #default="{ row }">
            <span class="amount-text">NT${{ row.amount }}</span>
          </template>
        </el-table-column>
        <el-table-column prop="balance" label="余额" min-width="140">
          <template #default="{ row }">
            <span class="balance-text">NT${{ row.balance }}</span>
          </template>
        </el-table-column>
        <el-table-column prop="status" label="状态" min-width="100">
          <template #default="{ row }">
            <span class="status-text">{{ row.status }}</span>
          </template>
        </el-table-column>
      </el-table>

      <!-- 分页 -->
      <div class="pagination-wrapper">
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
import { ref, reactive, onMounted } from 'vue'
import { Search } from '@element-plus/icons-vue'

interface Transaction {
  date: string
  type: string
  storeId: string
  orderNo: string
  amount: string
  balance: string
  status: string
}

const activeTab = ref('all')
const searchKeyword = ref('')
const dateRange = ref<string[]>(['2025-09-01', '2025-09-10'])
const loading = ref(false)

const summaryData = reactive({
  balance: '5,000.00',
  orderPayment: '5,000.00',
  orderSettlement: '5,000.00',
  adjustment: '5,000.00'
})

const pagination = reactive({
  page: 1,
  pageSize: 10,
  total: 123
})

const transactionList = ref<Transaction[]>([])

const fetchTransactions = async () => {
  loading.value = true
  try {
    // 模拟数据
    transactionList.value = Array.from({ length: 10 }, (_, i) => ({
      date: '2026-12-12 23:59:59',
      type: i === 1 ? '订单结算' : i === 2 ? '账款调整' : '订单付款',
      storeId: 'S1234567890',
      orderNo: 'X250904KQ2P078R',
      amount: '1,000.00',
      balance: '223,560.50',
      status: '已完成'
    }))
    pagination.total = 123
  } catch (err) {
    console.error('获取交易列表失败:', err)
  } finally {
    loading.value = false
  }
}

const handleTabChange = (_tab: string) => {
  pagination.page = 1
  fetchTransactions()
}

const handleSearch = () => {
  pagination.page = 1
  fetchTransactions()
}

const handleDateChange = () => {
  pagination.page = 1
  fetchTransactions()
}

const handleSizeChange = (size: number) => {
  pagination.pageSize = size
  pagination.page = 1
  fetchTransactions()
}

const handlePageChange = (page: number) => {
  pagination.page = page
  fetchTransactions()
}

onMounted(() => {
  fetchTransactions()
})
</script>

<style lang="scss" scoped>
.escrow-page {
  padding: 20px;
  background-color: #f5f7fa;
  min-height: 100vh;
}

.page-header {
  margin-bottom: 20px;

  .page-title {
    font-size: 20px;
    font-weight: 600;
    color: #303133;
    margin: 0;
  }
}

.escrow-summary {
  margin-bottom: 20px;

  .summary-title {
    font-size: 14px;
    font-weight: 500;
    color: #606266;
    margin-bottom: 12px;
  }

  .summary-cards {
    display: flex;
    align-items: center;
    gap: 0;
    background: #fff;
    border: 1px solid #ebeef5;
    border-radius: 8px;
    padding: 20px;
  }

  .summary-card {
    flex: 1;
    padding: 0 20px;

    .card-label {
      font-size: 12px;
      color: #909399;
      margin-bottom: 8px;
    }

    .card-value {
      font-size: 24px;
      font-weight: 600;
      color: #303133;
    }
  }

  .card-divider {
    font-size: 30px;
    font-weight: 500;
    color: #909399;
    padding: 0 20px;
  }
}

.escrow-card {
  border-radius: 8px;

  :deep(.el-card__body) {
    padding: 0;
  }
}

.escrow-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 0 20px;
  border-bottom: 1px solid #ebeef5;

  .escrow-tabs {
    :deep(.el-tabs__header) {
      margin: 0;
    }

    :deep(.el-tabs__nav-wrap::after) {
      display: none;
    }
  }

  .header-actions {
    display: flex;
    gap: 12px;
    align-items: center;
  }
}

.pagination-wrapper {
  padding: 20px;
  display: flex;
  justify-content: center;
}

.amount-text {
  color: #303133;
}

.balance-text {
  color: #303133;
  font-weight: 500;
}

.status-text {
  color: #909399;
}

@media (max-width: 1200px) {
  .escrow-summary .summary-cards {
    flex-wrap: wrap;
    gap: 16px;
  }

  .escrow-summary .card-divider {
    display: none;
  }

  .escrow-summary .summary-card {
    flex: 0 0 calc(50% - 8px);
    padding: 12px;
    border: 1px solid #ebeef5;
    border-radius: 6px;
  }
}

@media (max-width: 768px) {
  .escrow-page {
    padding: 12px;
  }

  .escrow-summary .summary-card {
    flex: 0 0 100%;
  }

  .escrow-header {
    flex-direction: column;
    align-items: flex-start;
    gap: 12px;
    padding: 12px 16px;
  }

  .header-actions {
    flex-wrap: wrap;
  }
}
</style>
