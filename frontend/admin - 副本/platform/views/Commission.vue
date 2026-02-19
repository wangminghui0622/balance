<template>
  <div class="commission-page">
    <!-- 页面标题 -->
    <div class="page-header">
      <h1 class="page-title">我的佣金</h1>
    </div>

    <!-- 佣金总览 -->
    <div class="commission-summary">
      <div class="summary-header">
        <span class="summary-title">佣金总览</span>
        <el-link type="primary" class="stats-link">佣金统计</el-link>
      </div>
      <div class="summary-cards">
        <div class="summary-card main-card">
          <div class="card-content">
            <div class="card-info">
              <div class="card-label">可提现金额</div>
              <div class="card-value">NT${{ summaryData.withdrawable }}</div>
              <div class="card-sub">累计佣金：NT${{ summaryData.totalCommission }}</div>
            </div>
            <el-button type="primary" class="withdraw-btn">提现</el-button>
          </div>
        </div>
        <div class="summary-card">
          <div class="card-label">即将结算佣金</div>
          <div class="card-value">NT${{ summaryData.pending }}</div>
          <div class="card-sub">根据托管中的订单预估佣金收益</div>
        </div>
      </div>
    </div>

    <!-- 佣金列表 -->
    <el-card class="commission-card" shadow="never">
      <!-- Tab和搜索 -->
      <div class="commission-header">
        <el-tabs v-model="activeTab" class="commission-tabs" @tab-change="handleTabChange">
          <el-tab-pane label="全部" name="all" />
          <el-tab-pane label="佣金" name="commission" />
          <el-tab-pane label="提现" name="withdraw" />
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

      <!-- 佣金表格 -->
      <el-table :data="commissionList" style="width: 100%" v-loading="loading">
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
import { platformCommissionApi } from '@share/api/platform'

interface CommissionRecord {
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
  withdrawable: '0.00',
  totalCommission: '0.00',
  pending: '0.00'
})

const pagination = reactive({
  page: 1,
  pageSize: 10,
  total: 0
})

const commissionList = ref<CommissionRecord[]>([])

const fetchCommissionStats = async () => {
  try {
    const res = await platformCommissionApi.getCommissionStats()
    if (res.code === 0 && res.data) {
      summaryData.withdrawable = res.data.withdrawable || '0.00'
      summaryData.totalCommission = res.data.total_commission || '0.00'
      summaryData.pending = res.data.pending || '0.00'
    }
  } catch (err) {
    console.error('获取佣金统计失败:', err)
  }
}

const fetchCommissions = async () => {
  loading.value = true
  try {
    const res = await platformCommissionApi.getCommissionList({
      type: activeTab.value === 'all' ? undefined : activeTab.value,
      page: pagination.page,
      page_size: pagination.pageSize
    })
    if (res.code === 0 && res.data) {
      commissionList.value = res.data.list.map((item: any) => ({
        date: item.date,
        type: item.type,
        storeId: String(item.store_id),
        orderNo: item.order_no,
        amount: item.amount,
        balance: item.balance,
        status: item.status
      }))
      pagination.total = res.data.total
    }
  } catch (err) {
    console.error('获取佣金列表失败:', err)
  } finally {
    loading.value = false
  }
}

const handleTabChange = (_tab: string) => {
  pagination.page = 1
  fetchCommissions()
}

const handleSearch = () => {
  pagination.page = 1
  fetchCommissions()
}

const handleDateChange = () => {
  pagination.page = 1
  fetchCommissions()
}

const handleSizeChange = (size: number) => {
  pagination.pageSize = size
  pagination.page = 1
  fetchCommissions()
}

const handlePageChange = (page: number) => {
  pagination.page = page
  fetchCommissions()
}

onMounted(() => {
  fetchCommissionStats()
  fetchCommissions()
})
</script>

<style lang="scss" scoped>
.commission-page {
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

.commission-summary {
  margin-bottom: 20px;

  .summary-header {
    display: flex;
    justify-content: space-between;
    align-items: center;
    margin-bottom: 12px;

    .summary-title {
      font-size: 14px;
      font-weight: 500;
      color: #606266;
    }

    .stats-link {
      font-size: 13px;
    }
  }

  .summary-cards {
    display: grid;
    grid-template-columns: 2fr 1fr;
    gap: 16px;
  }

  .summary-card {
    background: #fff;
    border: 1px solid #ebeef5;
    border-radius: 8px;
    padding: 20px;

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

    .card-sub {
      font-size: 12px;
      color: #909399;
      margin-top: 8px;
    }

    &.main-card {
      .card-content {
        display: flex;
        justify-content: space-between;
        align-items: center;
      }

      .withdraw-btn {
        height: 40px;
        padding: 0 24px;
      }
    }
  }
}

.commission-card {
  border-radius: 8px;

  :deep(.el-card__body) {
    padding: 0;
  }
}

.commission-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 0 20px;
  border-bottom: 1px solid #ebeef5;

  .commission-tabs {
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
  .commission-summary .summary-cards {
    grid-template-columns: 1fr;
  }
}

@media (max-width: 768px) {
  .commission-page {
    padding: 12px;
  }

  .commission-header {
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
