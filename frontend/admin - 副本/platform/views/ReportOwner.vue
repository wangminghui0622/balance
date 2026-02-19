<template>
  <div class="report-page">
    <div class="page-header">
      <h2>店主报表</h2>
      <el-button @click="searchOwnerDialogVisible = true">查询店主</el-button>
    </div>

    <!-- Tab切换 -->
    <div class="tab-section">
      <el-tabs v-model="activeTab" @tab-change="handleTabChange">
        <el-tab-pane label="店主订单未付款金额" name="unpaid" />
        <el-tab-pane label="店主预付款账户余额" name="prepay" />
        <el-tab-pane label="店主佣金账户余额" name="commission" />
        <el-tab-pane label="店主保证金账户余额" name="deposit" />
      </el-tabs>
    </div>

    <!-- 汇总卡片 - 简单模式 -->
    <el-card v-if="activeTab === 'unpaid'" class="summary-card" shadow="never">
      <div class="summary-label">{{ summaryLabel }}</div>
      <div class="summary-value">NT$5,450.00</div>
    </el-card>

    <!-- 汇总卡片 - 预付款账户余额（多项） -->
    <el-card v-else-if="activeTab === 'prepay'" class="summary-card multi" shadow="never">
      <div class="summary-row with-operators">
        <div class="summary-item">
          <div class="item-label">店主预付款账户余额</div>
          <div class="item-value">NT$5,450.00</div>
        </div>
        <span class="operator">=</span>
        <div class="summary-item">
          <div class="item-label">充值金额</div>
          <div class="item-value">¥5,000.00</div>
        </div>
        <span class="operator">+</span>
        <div class="summary-item">
          <div class="item-label">转存金额</div>
          <div class="item-value">¥5,000.00</div>
        </div>
        <span class="operator">-</span>
        <div class="summary-item">
          <div class="item-label">提现金额</div>
          <div class="item-value">¥5,000.00</div>
        </div>
        <span class="operator">-</span>
        <div class="summary-item">
          <div class="item-label">订单付款</div>
          <div class="item-value">¥5,000.00</div>
        </div>
        <span class="operator">+</span>
        <div class="summary-item">
          <div class="item-label">账款调整</div>
          <div class="item-value">¥5,000.00</div>
        </div>
      </div>
    </el-card>

    <!-- 汇总卡片 - 佣金账户余额（多项） -->
    <el-card v-else-if="activeTab === 'commission'" class="summary-card multi" shadow="never">
      <div class="summary-row with-operators">
        <div class="summary-item">
          <div class="item-label">店主佣金账户余额</div>
          <div class="item-value">NT$5,450.00</div>
        </div>
        <span class="operator">=</span>
        <div class="summary-item">
          <div class="item-label">佣金</div>
          <div class="item-value">¥5,000.00</div>
        </div>
        <span class="operator">+</span>
        <div class="summary-item">
          <div class="item-label">转存</div>
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

    <!-- 汇总卡片 - 保证金账户余额（多项） -->
    <el-card v-else-if="activeTab === 'deposit'" class="summary-card multi" shadow="never">
      <div class="summary-row with-operators">
        <div class="summary-item">
          <div class="item-label">店主保证金账户余额</div>
          <div class="item-value">NT$5,450.00</div>
        </div>
        <span class="operator">=</span>
        <div class="summary-item">
          <div class="item-label">充值</div>
          <div class="item-value">¥5,000.00</div>
        </div>
        <span class="operator">-</span>
        <div class="summary-item">
          <div class="item-label">退保</div>
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

    <!-- 汇总卡片 - 其他Tab -->
    <el-card v-else class="summary-card" shadow="never">
      <div class="summary-label">{{ summaryLabel }}</div>
      <div class="summary-value">NT$5,450.00</div>
    </el-card>

    <!-- 全部明细 -->
    <el-card class="detail-card" shadow="never">
      <div class="detail-header">
        <span class="detail-title">全部明细</span>
        <el-checkbox>导出报表</el-checkbox>
      </div>

      <!-- 订单未付款表格 -->
      <el-table v-if="activeTab === 'unpaid'" :data="tableData" style="width: 100%" v-loading="loading">
        <el-table-column prop="date" label="日期" min-width="150" />
        <el-table-column prop="owner" label="店主" min-width="100" />
        <el-table-column prop="storeNo" label="店铺编号" min-width="100" />
        <el-table-column prop="storeName" label="店铺名称" min-width="120" />
        <el-table-column prop="orderNo" label="订单编号" min-width="150" />
        <el-table-column prop="orderAmount" label="订单金额" min-width="120" />
        <el-table-column prop="balance" label="余额" min-width="120" />
        <el-table-column prop="status" label="订单状态" min-width="100">
          <template #default="{ row }">
            <span :class="['status-tag', row.status === '未付款' ? 'unpaid' : 'paid']">
              {{ row.status }}
            </span>
          </template>
        </el-table-column>
      </el-table>

      <!-- 预付款/佣金/保证金表格 -->
      <el-table v-else :data="transactionData" style="width: 100%" v-loading="loading">
        <el-table-column prop="date" label="日期" min-width="150" />
        <el-table-column prop="owner" label="店主" min-width="100" />
        <el-table-column prop="transType" label="交易类型" min-width="100" />
        <el-table-column prop="transChannel" label="交易渠道" min-width="120" />
        <el-table-column prop="transNo" label="交易单号" min-width="150" />
        <el-table-column prop="transAmount" label="交易金额" min-width="120" />
        <el-table-column prop="balance" label="余额" min-width="120" />
        <el-table-column prop="transStatus" :label="activeTab === 'commission' || activeTab === 'deposit' ? '状态' : '交易状态'" min-width="100">
          <template #default="{ row }">
            <span :class="['status-tag', row.transStatus === '已完成' ? 'completed' : 'pending']">
              {{ row.transStatus }}
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

    <!-- 查询店主弹框 -->
    <el-dialog
      v-model="searchOwnerDialogVisible"
      title="查询店主"
      width="1100px"
      :close-on-click-modal="false"
    >
      <div class="search-owner-header">
        <el-input v-model="searchOwnerKeyword" placeholder="请输入店主ID或名称" style="width: 300px" />
        <el-button type="primary" @click="handleSearchOwner">搜索</el-button>
      </div>

      <el-table :data="ownerSearchList" style="width: 100%; margin-top: 16px" class="nowrap-table">
        <el-table-column prop="ownerId" label="店主ID" min-width="100" />
        <el-table-column prop="ownerName" label="店主" min-width="80" />
        <el-table-column prop="unpaidAmount" label="店主订单未付款金额" min-width="130" />
        <el-table-column prop="prepayAmount" label="店主预付款账户余额" min-width="130" />
        <el-table-column prop="commissionAmount" label="店主佣金账户余额" min-width="120" />
        <el-table-column prop="depositAmount" label="店主保证金账户余额" min-width="130" />
        <el-table-column prop="totalAmount" label="累计可用余额" min-width="100" />
      </el-table>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'

const activeTab = ref('unpaid')
const loading = ref(false)
const searchOwnerDialogVisible = ref(false)
const searchOwnerKeyword = ref('1234567890')

interface OwnerSearchRecord {
  ownerId: string
  ownerName: string
  unpaidAmount: string
  prepayAmount: string
  commissionAmount: string
  depositAmount: string
  totalAmount: string
}

const ownerSearchList = ref<OwnerSearchRecord[]>([
  { ownerId: '1234567890', ownerName: '文字占位符', unpaidAmount: 'NT$5,450.00', prepayAmount: 'NT$5,450.00', commissionAmount: 'NT$5,450.00', depositAmount: 'NT$5,450.00', totalAmount: 'NT$5,450.00' },
  { ownerId: '1234567890', ownerName: '文字占位符', unpaidAmount: 'NT$5,450.00', prepayAmount: 'NT$5,450.00', commissionAmount: 'NT$5,450.00', depositAmount: 'NT$5,450.00', totalAmount: 'NT$5,450.00' },
  { ownerId: '1234567890', ownerName: '文字占位符', unpaidAmount: 'NT$5,450.00', prepayAmount: 'NT$5,450.00', commissionAmount: 'NT$5,450.00', depositAmount: 'NT$5,450.00', totalAmount: 'NT$5,450.00' },
  { ownerId: '1234567890', ownerName: '文字占位符', unpaidAmount: 'NT$5,450.00', prepayAmount: 'NT$5,450.00', commissionAmount: 'NT$5,450.00', depositAmount: 'NT$5,450.00', totalAmount: 'NT$5,450.00' }
])

const handleSearchOwner = () => {
  console.log('搜索店主:', searchOwnerKeyword.value)
}

const summaryLabel = computed(() => {
  const labels: Record<string, string> = {
    unpaid: '店主订单未付款金额(=订单欠费金额)',
    prepay: '店主预付款账户余额',
    commission: '店主佣金账户余额',
    deposit: '店主保证金账户余额'
  }
  return labels[activeTab.value]
})

const pagination = ref({
  page: 1,
  pageSize: 10,
  total: 123
})

interface TableRecord {
  date: string
  owner: string
  storeNo: string
  storeName: string
  orderNo: string
  orderAmount: string
  balance: string
  status: string
}

interface TransactionRecord {
  date: string
  owner: string
  transType: string
  transChannel: string
  transNo: string
  transAmount: string
  balance: string
  transStatus: string
}

const tableData = ref<TableRecord[]>([])
const transactionData = ref<TransactionRecord[]>([])

const prepayTransTypes = ['充值', '转存', '提现', '订单付款', '账款调整']
const commissionTransTypes = ['佣金', '转存', '提现', '账款调整']
const depositTransTypes = ['充值', '提现', '罚款', '补贴']

const getTransTypes = () => {
  if (activeTab.value === 'commission') return commissionTransTypes
  if (activeTab.value === 'deposit') return depositTransTypes
  return prepayTransTypes
}

const fetchData = () => {
  loading.value = true
  setTimeout(() => {
    if (activeTab.value === 'unpaid') {
      tableData.value = Array.from({ length: 10 }, () => ({
        date: '2026-12-12 23:59:59',
        owner: '文字占位符',
        storeNo: '12536582',
        storeName: '文字占位符占位符',
        orderNo: 'X250904KQ2P078R',
        orderAmount: 'NT$1,000.00',
        balance: 'NT$223,560.50',
        status: '未付款'
      }))
    } else {
      const types = getTransTypes()
      transactionData.value = Array.from({ length: 10 }, (_, i) => ({
        date: '2026-12-12 23:59:59',
        owner: '文字占位符',
        transType: types[i % types.length],
        transChannel: '文字占位符占位符',
        transNo: 'X250904KQ2P078R',
        transAmount: 'NT$1,000.00',
        balance: 'NT$223,560.50',
        transStatus: '已完成'
      }))
    }
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
  &.unpaid {
    color: #f90;
  }
  
  &.paid,
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

.search-owner-header {
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
