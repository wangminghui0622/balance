<template>
  <div class="report-page">
    <div class="page-header">
      <h2>平台报表</h2>
    </div>

    <!-- Tab切换 -->
    <div class="tab-section">
      <el-tabs v-model="activeTab" @tab-change="handleTabChange">
        <el-tab-pane label="平台罚补账户余额" name="penalty" />
        <el-tab-pane label="平台佣金账户余额" name="commission" />
        <el-tab-pane label="平台托管账户余额" name="escrow" />
      </el-tabs>
    </div>

    <!-- 汇总卡片 - 罚补账户余额 -->
    <el-card v-if="activeTab === 'penalty'" class="summary-card multi" shadow="never">
      <div class="summary-row with-operators">
        <div class="summary-item">
          <div class="item-label">平台罚补账户余额</div>
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
        <span class="operator">+</span>
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

    <!-- 汇总卡片 - 佣金账户余额 -->
    <el-card v-else-if="activeTab === 'commission'" class="summary-card multi" shadow="never">
      <div class="summary-row with-operators">
        <div class="summary-item">
          <div class="item-label">平台佣金账户余额</div>
          <div class="item-value">NT$5,450.00</div>
        </div>
        <span class="operator">=</span>
        <div class="summary-item">
          <div class="item-label">佣金</div>
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

    <!-- 汇总卡片 - 托管账户余额 -->
    <el-card v-else class="summary-card multi" shadow="never">
      <div class="summary-row with-operators">
        <div class="summary-item">
          <div class="item-label">托管账户余额</div>
          <div class="item-value">NT$5,450.00</div>
        </div>
        <span class="operator">=</span>
        <div class="summary-item">
          <div class="item-label">订单付款</div>
          <div class="item-value">¥5,000.00</div>
        </div>
        <span class="operator">-</span>
        <div class="summary-item">
          <div class="item-label">订单结算</div>
          <div class="item-value">¥5,000.00</div>
        </div>
        <span class="operator">+</span>
        <div class="summary-item">
          <div class="item-label">账款调整</div>
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

      <!-- 罚补账户表格（有角色和姓名） -->
      <el-table v-if="activeTab === 'penalty'" :data="tableData" style="width: 100%" v-loading="loading">
        <el-table-column prop="date" label="日期" min-width="150" />
        <el-table-column prop="role" label="角色" min-width="80" />
        <el-table-column prop="name" label="姓名" min-width="80" />
        <el-table-column prop="transType" label="交易类型" min-width="80" />
        <el-table-column prop="transChannel" label="交易渠道" min-width="120" />
        <el-table-column prop="transNo" label="交易单号" min-width="150" />
        <el-table-column prop="transAmount" label="交易金额" min-width="100" />
        <el-table-column prop="balance" label="余额" min-width="100" />
        <el-table-column prop="status" label="状态" min-width="80">
          <template #default="{ row }">
            <span :class="['status-tag', row.status === '已完成' ? 'completed' : 'pending']">
              {{ row.status }}
            </span>
          </template>
        </el-table-column>
      </el-table>

      <!-- 佣金账户表格 -->
      <el-table v-else-if="activeTab === 'commission'" :data="commissionTableData" style="width: 100%" v-loading="loading">
        <el-table-column prop="date" label="日期" min-width="150" />
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

      <!-- 托管账户表格 -->
      <el-table v-else :data="escrowTableData" style="width: 100%" v-loading="loading">
        <el-table-column prop="date" label="日期" min-width="150" />
        <el-table-column prop="transType" label="交易类型" min-width="100" />
        <el-table-column prop="storeNo" label="店铺编号" min-width="120" />
        <el-table-column prop="orderNo" label="订单编号" min-width="150" />
        <el-table-column prop="transAmount" label="交易金额" min-width="120" />
        <el-table-column prop="balance" label="余额" min-width="120" />
        <el-table-column prop="status" label="状态" min-width="80">
          <template #default="{ row }">
            <span :class="['status-tag', row.status === '已完成' ? 'completed' : 'pending']">
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
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'

const activeTab = ref('penalty')
const loading = ref(false)

const pagination = ref({
  page: 1,
  pageSize: 10,
  total: 123
})

interface TableRecord {
  date: string
  role: string
  name: string
  transType: string
  transChannel: string
  transNo: string
  transAmount: string
  balance: string
  status: string
}

interface CommissionTableRecord {
  date: string
  transType: string
  transChannel: string
  transNo: string
  transAmount: string
  balance: string
  status: string
}

interface EscrowTableRecord {
  date: string
  transType: string
  storeNo: string
  orderNo: string
  transAmount: string
  balance: string
  status: string
}

const tableData = ref<TableRecord[]>([])
const commissionTableData = ref<CommissionTableRecord[]>([])
const escrowTableData = ref<EscrowTableRecord[]>([])

const roles = ['店主', '运营']
const names = ['张小明', '李玉豪', '张启发', '周晓琳', '马清润']
const penaltyTransTypes = ['充值', '提现', '罚款', '补贴']
const commissionTransTypes = ['佣金', '提现', '账款调整']

const fetchData = () => {
  loading.value = true
  setTimeout(() => {
    if (activeTab.value === 'penalty') {
      tableData.value = Array.from({ length: 10 }, (_, i) => ({
        date: '2026-12-12 23:59:59',
        role: roles[i % roles.length],
        name: names[i % names.length],
        transType: penaltyTransTypes[i % penaltyTransTypes.length],
        transChannel: '文字占位符占位符',
        transNo: 'X250904KQ2P078R',
        transAmount: 'NT$1,000.00',
        balance: 'NT$223,560.50',
        status: '已完成'
      }))
    } else if (activeTab.value === 'commission') {
      commissionTableData.value = Array.from({ length: 10 }, (_, i) => ({
        date: '2026-12-12 23:59:59',
        transType: commissionTransTypes[i % commissionTransTypes.length],
        transChannel: '文字占位符占位符',
        transNo: 'X250904KQ2P078R',
        transAmount: 'NT$1,000.00',
        balance: 'NT$223,560.50',
        status: '已结算'
      }))
    } else {
      escrowTableData.value = Array.from({ length: 10 }, () => ({
        date: '2026-12-12 23:59:59',
        transType: '佣金收入',
        storeNo: '文字占位符占位符',
        orderNo: 'X250904KQ2P078R',
        transAmount: 'NT$1,000.00',
        balance: 'NT$223,560.50',
        status: '已完成'
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
</style>
