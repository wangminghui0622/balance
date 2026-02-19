<template>
  <div class="penalty-page">
    <!-- 页面标题 -->
    <div class="page-header">
      <h1 class="page-title">罚补账户</h1>
    </div>

    <!-- 罚补账户余额 -->
    <div class="penalty-summary">
      <div class="summary-title">罚补账户余额</div>
      <div class="summary-card">
        <div class="card-content">
          <div class="card-info">
            <div class="card-label">可提现余额</div>
            <div class="card-value">NT${{ summaryData.balance }}</div>
          </div>
          <div class="card-actions">
            <el-button type="primary">充值</el-button>
            <el-button>提现</el-button>
          </div>
        </div>
      </div>
    </div>

    <!-- 交易列表 -->
    <el-card class="penalty-card" shadow="never">
      <!-- Tab和搜索 -->
      <div class="penalty-header">
        <el-tabs v-model="activeTab" class="penalty-tabs" @tab-change="handleTabChange">
          <el-tab-pane label="全部" name="all" />
          <el-tab-pane label="充值" name="recharge" />
          <el-tab-pane label="提现" name="withdraw" />
          <el-tab-pane label="罚款" name="penalty" />
          <el-tab-pane label="补贴" name="subsidy" />
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
        <el-table-column prop="role" label="角色" min-width="80" />
        <el-table-column prop="name" label="姓名" min-width="100" />
        <el-table-column prop="type" label="交易类型" min-width="80" />
        <el-table-column prop="channel" label="交易渠道" min-width="120" />
        <el-table-column prop="orderNo" label="交易单号" min-width="160" />
        <el-table-column prop="amount" label="交易金额" min-width="120">
          <template #default="{ row }">
            <span class="amount-text">NT${{ row.amount }}</span>
          </template>
        </el-table-column>
        <el-table-column prop="balance" label="余额" min-width="130">
          <template #default="{ row }">
            <span class="balance-text">NT${{ row.balance }}</span>
          </template>
        </el-table-column>
        <el-table-column prop="status" label="状态" min-width="80">
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
import { platformPenaltyApi } from '@share/api/platform'

interface Transaction {
  date: string
  role: string
  name: string
  type: string
  channel: string
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
  balance: '0.00'
})

const pagination = reactive({
  page: 1,
  pageSize: 10,
  total: 0
})

const transactionList = ref<Transaction[]>([])

const fetchPenaltyStats = async () => {
  try {
    const res = await platformPenaltyApi.getPenaltyStats()
    if (res.code === 0 && res.data) {
      summaryData.balance = res.data.balance || '0.00'
    }
  } catch (err) {
    console.error('获取罚补统计失败:', err)
  }
}

const fetchTransactions = async () => {
  loading.value = true
  try {
    const res = await platformPenaltyApi.getPenaltyList({
      type: activeTab.value === 'all' ? undefined : activeTab.value,
      page: pagination.page,
      page_size: pagination.pageSize
    })
    if (res.code === 0 && res.data) {
      transactionList.value = res.data.list.map((item: any) => ({
        date: item.date,
        role: item.role,
        name: item.name,
        type: item.type,
        channel: item.channel,
        orderNo: item.order_no,
        amount: item.amount,
        balance: item.balance,
        status: item.status
      }))
      pagination.total = res.data.total
    }
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
  fetchPenaltyStats()
  fetchTransactions()
})
</script>

<style lang="scss" scoped>
.penalty-page {
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

.penalty-summary {
  margin-bottom: 20px;

  .summary-title {
    font-size: 14px;
    font-weight: 500;
    color: #606266;
    margin-bottom: 12px;
  }

  .summary-card {
    background: #fff;
    border: 1px solid #ebeef5;
    border-radius: 8px;
    padding: 20px;

    .card-content {
      display: flex;
      justify-content: space-between;
      align-items: center;
    }

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

    .card-actions {
      display: flex;
      gap: 12px;
    }
  }
}

.penalty-card {
  border-radius: 8px;

  :deep(.el-card__body) {
    padding: 0;
  }
}

.penalty-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 0 20px;
  border-bottom: 1px solid #ebeef5;

  .penalty-tabs {
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

@media (max-width: 768px) {
  .penalty-page {
    padding: 12px;
  }

  .penalty-summary .summary-card .card-content {
    flex-direction: column;
    align-items: flex-start;
    gap: 16px;
  }

  .penalty-header {
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
