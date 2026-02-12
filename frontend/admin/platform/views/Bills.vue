<template>
  <div class="bills-page">
    <!-- 页面标题 -->
    <div class="page-header">
      <h1 class="page-title">我的账单</h1>
    </div>

    <!-- 账单总览 -->
    <div class="bills-summary">
      <div class="summary-title">账单总览</div>
      <div class="summary-cards">
        <div class="summary-card">
          <div class="card-label">预估未结算佣金</div>
          <div class="card-value">NT${{ summaryData.unsettled }}</div>
        </div>
        <div class="summary-card">
          <div class="card-label">已结算佣金</div>
          <div class="card-value">NT${{ summaryData.settled }}</div>
        </div>
        <div class="summary-card">
          <div class="card-label">账款调整</div>
          <div class="card-value">NT${{ summaryData.adjustment }}</div>
        </div>
        <div class="summary-card highlight">
          <div class="card-label">预估佣金总额（未结算+已结算+账款调整）</div>
          <div class="card-value">NT${{ summaryData.total }}</div>
        </div>
      </div>
    </div>

    <!-- 账单列表 -->
    <el-card class="bills-card" shadow="never">
      <!-- Tab和搜索 -->
      <div class="bills-header">
        <el-tabs v-model="activeTab" class="bills-tabs" @tab-change="handleTabChange">
          <el-tab-pane label="未结算" name="unsettled" />
          <el-tab-pane label="已结算" name="settled" />
          <el-tab-pane label="账款调整" name="adjustment" />
        </el-tabs>
        <div class="header-actions">
          <el-input
            v-model="searchKeyword"
            placeholder="快速搜索"
            :prefix-icon="Search"
            clearable
            style="width: 200px"
            @input="handleSearch"
          />
          <el-button :icon="Download">导出报表</el-button>
        </div>
      </div>

      <!-- 账单表格 -->
      <el-table :data="billList" style="width: 100%" v-loading="loading">
        <el-table-column prop="date" label="日期" min-width="180" />
        <el-table-column prop="storeId" label="店铺编号" min-width="140" />
        <el-table-column prop="orderNo" label="订单编号" min-width="180" />
        <el-table-column prop="orderStatus" label="订单状态" min-width="100">
          <template #default="{ row }">
            <span>{{ row.orderStatus }}</span>
          </template>
        </el-table-column>
        <el-table-column prop="countdown" label="订单计时" min-width="120" />
        <el-table-column prop="orderAmount" label="订单金额" min-width="120">
          <template #default="{ row }">
            <span class="amount-text">NT${{ row.orderAmount }}</span>
          </template>
        </el-table-column>
        <el-table-column prop="commission" label="未结算佣金" min-width="120">
          <template #default="{ row }">
            <span class="commission-text">NT${{ row.commission }}</span>
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
import { Search, Download } from '@element-plus/icons-vue'

interface Bill {
  date: string
  storeId: string
  orderNo: string
  orderStatus: string
  countdown: string
  orderAmount: string
  commission: string
}

const activeTab = ref('unsettled')
const searchKeyword = ref('')
const loading = ref(false)

const summaryData = reactive({
  unsettled: '245',
  settled: '123,00',
  adjustment: '123,00',
  total: '123,456.00'
})

const pagination = reactive({
  page: 1,
  pageSize: 10,
  total: 123
})

const billList = ref<Bill[]>([])

const fetchBills = async () => {
  loading.value = true
  try {
    // 模拟数据
    billList.value = Array.from({ length: 10 }, (_, i) => ({
      date: '2026-12-12 23:59:59',
      storeId: 'S1234567890',
      orderNo: 'X250904KQ2P078R',
      orderStatus: '待发货',
      countdown: '1天 03:23:42',
      orderAmount: '68.00',
      commission: '8.00'
    }))
    pagination.total = 123
  } catch (err) {
    console.error('获取账单列表失败:', err)
  } finally {
    loading.value = false
  }
}

const handleTabChange = (tab: string) => {
  pagination.page = 1
  fetchBills()
}

const handleSearch = () => {
  pagination.page = 1
  fetchBills()
}

const handleSizeChange = (size: number) => {
  pagination.pageSize = size
  pagination.page = 1
  fetchBills()
}

const handlePageChange = (page: number) => {
  pagination.page = page
  fetchBills()
}

onMounted(() => {
  fetchBills()
})
</script>

<style lang="scss" scoped>
.bills-page {
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

.bills-summary {
  margin-bottom: 20px;

  .summary-title {
    font-size: 14px;
    font-weight: 500;
    color: #606266;
    margin-bottom: 12px;
  }

  .summary-cards {
    display: grid;
    grid-template-columns: repeat(4, 1fr);
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

    &.highlight {
      .card-value {
        color: #ff6a3a;
      }
    }
  }
}

.bills-card {
  border-radius: 8px;

  :deep(.el-card__body) {
    padding: 0;
  }
}

.bills-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 0 20px;
  border-bottom: 1px solid #ebeef5;

  .bills-tabs {
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

.commission-text {
  color: #ff6a3a;
  font-weight: 500;
}

@media (max-width: 1200px) {
  .bills-summary .summary-cards {
    grid-template-columns: repeat(2, 1fr);
  }
}

@media (max-width: 768px) {
  .bills-page {
    padding: 12px;
  }

  .bills-summary .summary-cards {
    grid-template-columns: 1fr;
  }

  .bills-header {
    flex-direction: column;
    align-items: flex-start;
    gap: 12px;
    padding: 12px 16px;
  }
}
</style>
