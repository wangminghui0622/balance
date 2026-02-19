<template>
  <div class="report-page">
    <div class="page-header">
      <h2>运营报表</h2>
      <el-button @click="searchOperatorDialogVisible = true">查询运营</el-button>
    </div>

    <!-- Tab切换 -->
    <div class="tab-section">
      <el-tabs v-model="activeTab" @tab-change="handleTabChange">
        <el-tab-pane label="运营回款账户余额" name="payment" />
        <el-tab-pane label="运营保证金账户余额" name="deposit" />
      </el-tabs>
    </div>

    <!-- 汇总卡片 - 回款账户余额 -->
    <el-card v-if="activeTab === 'payment'" class="summary-card multi" shadow="never">
      <div class="summary-row with-operators">
        <div class="summary-item">
          <div class="item-label">运营回款账户余额</div>
          <div class="item-value">NT$5,450.00</div>
        </div>
        <span class="operator">=</span>
        <div class="summary-item">
          <div class="item-label">回款</div>
          <div class="item-value">¥5,000.00</div>
        </div>
        <span class="operator">-</span>
        <div class="summary-item">
          <div class="item-label">提现</div>
          <div class="item-value">¥5,000.00</div>
        </div>
        <span class="operator">+</span>
        <div class="summary-item">
          <div class="item-label">账款调整</div>
          <div class="item-value">¥5,000.00</div>
        </div>
      </div>
    </el-card>

    <!-- 汇总卡片 - 保证金账户余额 -->
    <el-card v-else class="summary-card multi" shadow="never">
      <div class="summary-row with-operators">
        <div class="summary-item">
          <div class="item-label">运营保证金账户余额</div>
          <div class="item-value">NT$5,450.00</div>
        </div>
        <span class="operator">=</span>
        <div class="summary-item">
          <div class="item-label">充值</div>
          <div class="item-value">¥5,000.00</div>
        </div>
        <span class="operator">-</span>
        <div class="summary-item">
          <div class="item-label">提现</div>
          <div class="item-value">¥5,000.00</div>
        </div>
        <span class="operator">-</span>
        <div class="summary-item">
          <div class="item-label">罚款</div>
          <div class="item-value">¥5,000.00</div>
        </div>
        <span class="operator">+</span>
        <div class="summary-item">
          <div class="item-label">补贴</div>
          <div class="item-value">¥5,000.00</div>
        </div>
      </div>
    </el-card>

    <!-- 全部明细 -->
    <el-card class="detail-card" shadow="never">
      <div class="detail-header">
        <span class="detail-title">全部明细</span>
        <el-checkbox>导出报表</el-checkbox>
      </div>

      <el-table :data="tableData" style="width: 100%" v-loading="loading">
        <el-table-column prop="date" label="日期" min-width="150" />
        <el-table-column prop="operator" label="运营" min-width="100" />
        <el-table-column prop="transType" label="交易类型" min-width="100" />
        <el-table-column prop="transChannel" label="交易渠道" min-width="120" />
        <el-table-column prop="transNo" label="交易单号" min-width="150" />
        <el-table-column prop="transAmount" label="交易金额" min-width="120" />
        <el-table-column prop="balance" label="余额" min-width="120" />
        <el-table-column prop="status" label="状态" min-width="80">
          <template #default="{ row }">
            <span :class="['status-tag', row.status === '已结算' ? 'completed' : 'pending']">
              {{ row.status }}
            </span>
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

    <!-- 查询运营弹框 -->
    <el-dialog
      v-model="searchOperatorDialogVisible"
      title="查询运营"
      width="900px"
      :close-on-click-modal="false"
    >
      <div class="search-header">
        <el-input v-model="searchOperatorKeyword" placeholder="输入运营名称、账号等搜索" style="width: 300px" />
        <el-button type="primary" @click="handleSearchOperator">搜索</el-button>
      </div>

      <el-table :data="operatorSearchList" style="width: 100%; margin-top: 16px" class="nowrap-table">
        <el-table-column prop="operatorId" label="运营ID" min-width="100" />
        <el-table-column prop="operatorName" label="运营" min-width="100" />
        <el-table-column prop="paymentAmount" label="回款余额" min-width="120" />
        <el-table-column prop="depositAmount" label="保证金余额" min-width="120" />
        <el-table-column prop="totalAmount" label="累计可用余额" min-width="120" />
      </el-table>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'

const activeTab = ref('payment')
const loading = ref(false)
const searchOperatorDialogVisible = ref(false)
const searchOperatorKeyword = ref('')

const pagination = ref({
  page: 1,
  pageSize: 10,
  total: 123
})

interface TableRecord {
  date: string
  operator: string
  transType: string
  transChannel: string
  transNo: string
  transAmount: string
  balance: string
  status: string
}

interface OperatorSearchRecord {
  operatorId: string
  operatorName: string
  paymentAmount: string
  depositAmount: string
  totalAmount: string
}

const tableData = ref<TableRecord[]>([])
const operatorSearchList = ref<OperatorSearchRecord[]>([])

const paymentTransTypes = ['回款', '提现', '账款调整']
const depositTransTypes = ['充值', '提现', '罚款', '补贴']

const fetchData = () => {
  loading.value = true
  setTimeout(() => {
    const types = activeTab.value === 'payment' ? paymentTransTypes : depositTransTypes
    const status = activeTab.value === 'payment' ? '已结算' : '已完成'
    tableData.value = Array.from({ length: 10 }, (_, i) => ({
      date: '2026-12-12 23:59:59',
      operator: '文字占位符',
      transType: types[i % types.length],
      transChannel: '文字占位符占位符',
      transNo: 'X250904KQ2P078R',
      transAmount: 'NT$1,000.00',
      balance: 'NT$223,560.50',
      status
    }))
    loading.value = false
  }, 300)
}

const handleTabChange = () => {
  pagination.value.page = 1
  fetchData()
}

const handleSizeChange = (size: number) => {
  pagination.value.pageSize = size
  pagination.value.page = 1
  fetchData()
}

const handlePageChange = (page: number) => {
  pagination.value.page = page
  fetchData()
}

const handleSearchOperator = () => {
  operatorSearchList.value = [
    { operatorId: '1234567890', operatorName: '文字占位符', paymentAmount: 'NT$5,450.00', depositAmount: 'NT$5,450.00', totalAmount: 'NT$5,450.00' },
    { operatorId: '1234567890', operatorName: '文字占位符', paymentAmount: 'NT$5,450.00', depositAmount: 'NT$5,450.00', totalAmount: 'NT$5,450.00' },
    { operatorId: '1234567890', operatorName: '文字占位符', paymentAmount: 'NT$5,450.00', depositAmount: 'NT$5,450.00', totalAmount: 'NT$5,450.00' },
    { operatorId: '1234567890', operatorName: '文字占位符', paymentAmount: 'NT$5,450.00', depositAmount: 'NT$5,450.00', totalAmount: 'NT$5,450.00' }
  ]
}

onMounted(() => {
  fetchData()
})
</script>

<style scoped lang="scss">
.report-page {
  padding: 20px;
  background: #f5f7fa;
  min-height: 100%;
}

.page-header {
  display: flex;
  align-items: center;
  gap: 16px;
  margin-bottom: 20px;
  
  h2 {
    margin: 0;
    font-size: 18px;
    font-weight: 500;
    color: #303133;
  }
}

.tab-section {
  margin-bottom: 20px;
  
  :deep(.el-tabs__nav-wrap::after) {
    display: none;
  }
  
  :deep(.el-tabs__item) {
    color: #606266;
    
    &.is-active {
      color: #303133;
      font-weight: 500;
    }
  }
  
  :deep(.el-tabs__active-bar) {
    background-color: #f90;
  }
}

.summary-card {
  margin-bottom: 20px;
  
  .summary-label {
    font-size: 12px;
    color: #909399;
    margin-bottom: 8px;
  }
  
  .summary-value {
    font-size: 24px;
    font-weight: 500;
    color: #303133;
  }
  
  &.multi {
    .summary-row {
      display: flex;
      align-items: center;
      gap: 10px;
      
      &.with-operators {
        .operator {
          font-size: 18px;
          color: #909399;
          font-weight: 300;
        }
      }
    }
    
    .summary-item {
      background: #fff;
      border: 1px solid #ebeef5;
      border-radius: 4px;
      padding: 12px 16px;
      flex: 1;
      
      .item-label {
        font-size: 12px;
        color: #909399;
        margin-bottom: 6px;
      }
      
      .item-value {
        font-size: 18px;
        font-weight: 500;
        color: #303133;
      }
    }
  }
}

.detail-card {
  .detail-header {
    display: flex;
    justify-content: space-between;
    align-items: center;
    margin-bottom: 16px;
    
    .detail-title {
      font-size: 14px;
      font-weight: 500;
      color: #303133;
    }
  }
}

.status-tag {
  &.completed {
    color: #67c23a;
  }
  
  &.pending {
    color: #909399;
  }
}

.pagination-wrapper {
  padding: 20px 0;
  display: flex;
  justify-content: center;
}

.search-header {
  display: flex;
  gap: 12px;
  align-items: center;
}

.nowrap-table {
  :deep(.el-table__header th .cell),
  :deep(.el-table__body td .cell) {
    white-space: nowrap;
  }
}
</style>
