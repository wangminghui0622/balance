<template>
  <div class="finance-deposit-page">
    <div class="page-header">
      <h1 class="page-title">运营保证金</h1>
    </div>

    <!-- 保证金概览 -->
    <div class="summary-section">
      <div class="summary-title">保证金概览</div>
      <div class="summary-card">
        <div class="card-content">
          <div class="card-info">
            <div class="stat-label">保证金余额</div>
            <div class="stat-value">¥{{ formatAmount(summaryData.balance) }}</div>
            <div class="stat-sub">保证金门槛：¥{{ formatAmount(summaryData.threshold) }}或以上</div>
          </div>
          <div class="card-actions">
            <el-button type="primary" @click="handleDeposit">充值</el-button>
            <el-button @click="handleWithdraw">提现</el-button>
          </div>
        </div>
      </div>
    </div>

    <!-- 交易记录 -->
    <el-card class="records-card" v-loading="loading">
      <div class="records-header">
        <div class="tab-buttons">
          <span 
            :class="['tab-btn', activeTab === 'all' ? 'active' : '']" 
            @click="activeTab = 'all'"
          >全部</span>
          <span 
            :class="['tab-btn', activeTab === 'deposit' ? 'active' : '']" 
            @click="activeTab = 'deposit'"
          >充值</span>
          <span 
            :class="['tab-btn', activeTab === 'withdraw' ? 'active' : '']" 
            @click="activeTab = 'withdraw'"
          >提现</span>
          <span 
            :class="['tab-btn', activeTab === 'penalty' ? 'active' : '']" 
            @click="activeTab = 'penalty'"
          >罚扣</span>
          <span 
            :class="['tab-btn', activeTab === 'subsidy' ? 'active' : '']" 
            @click="activeTab = 'subsidy'"
          >补贴</span>
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

      <el-table :data="recordList" style="width: 100%">
        <el-table-column prop="date" label="日期" min-width="160" />
        <el-table-column prop="transType" label="交易类型" min-width="100" />
        <el-table-column prop="channel" label="交易渠道" min-width="160" />
        <el-table-column prop="transNo" label="交易单号" min-width="180" />
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
        <el-table-column prop="status" label="状态" min-width="100" />
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

    <!-- 充值弹框 -->
    <el-dialog v-model="depositVisible" title="充值保证金" width="400px">
      <el-form :model="depositForm" label-width="100px">
        <el-form-item label="充值金额">
          <el-input-number v-model="depositForm.amount" :min="100" :step="100" style="width: 100%" />
        </el-form-item>
        <el-form-item label="支付方式">
          <el-select v-model="depositForm.payMethod" style="width: 100%">
            <el-option label="银行转账" value="bank" />
            <el-option label="支付宝" value="alipay" />
            <el-option label="微信支付" value="wechat" />
          </el-select>
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="depositVisible = false">取消</el-button>
        <el-button type="primary" @click="confirmDeposit">确认充值</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted } from 'vue'
import { ElMessage } from 'element-plus'
import { Search } from '@element-plus/icons-vue'

const loading = ref(false)
const depositVisible = ref(false)
const activeTab = ref('all')

const summaryData = reactive({
  balance: 5000,
  threshold: 3000
})

const filterForm = reactive({
  keyword: '',
  startDate: null as Date | null,
  endDate: null as Date | null
})

const depositForm = reactive({
  amount: 1000,
  payMethod: 'bank'
})

const recordList = ref([
  { date: '2026-12-12 23:59:59', transType: '充值', channel: '文字占位符文字占位符', transNo: 'X250904KQ2P078R', amount: 1000, balance: 223560.50, status: '已完成' },
  { date: '2026-12-12 23:59:59', transType: '充值', channel: '文字占位符文字占位符', transNo: 'X250904KQ2P078R', amount: 1000, balance: 223560.50, status: '已完成' },
  { date: '2026-12-12 23:59:59', transType: '充值', channel: '文字占位符文字占位符', transNo: 'X250904KQ2P078R', amount: 1000, balance: 223560.50, status: '已完成' },
  { date: '2026-12-12 23:59:59', transType: '充值', channel: '文字占位符文字占位符', transNo: 'X250904KQ2P078R', amount: 1000, balance: 223560.50, status: '已完成' },
  { date: '2026-12-12 23:59:59', transType: '充值', channel: '文字占位符文字占位符', transNo: 'X250904KQ2P078R', amount: 1000, balance: 223560.50, status: '已完成' },
  { date: '2026-12-12 23:59:59', transType: '充值', channel: '文字占位符文字占位符', transNo: 'X250904KQ2P078R', amount: 1000, balance: 223560.50, status: '已完成' },
  { date: '2026-12-12 23:59:59', transType: '充值', channel: '文字占位符文字占位符', transNo: 'X250904KQ2P078R', amount: 1000, balance: 223560.50, status: '已完成' },
  { date: '2026-12-12 23:59:59', transType: '充值', channel: '文字占位符文字占位符', transNo: 'X250904KQ2P078R', amount: 1000, balance: 223560.50, status: '已完成' },
  { date: '2026-12-12 23:59:59', transType: '充值', channel: '文字占位符文字占位符', transNo: 'X250904KQ2P078R', amount: 1000, balance: 223560.50, status: '已完成' },
  { date: '2026-12-12 23:59:59', transType: '充值', channel: '文字占位符文字占位符', transNo: 'X250904KQ2P078R', amount: 1000, balance: 223560.50, status: '已完成' }
])

const pagination = reactive({
  page: 1,
  pageSize: 10,
  total: 25
})

function formatAmount(value: number): string {
  return Math.abs(value).toLocaleString('en-US', { minimumFractionDigits: 2, maximumFractionDigits: 2 })
}

function handleDeposit() {
  depositVisible.value = true
}

function handleWithdraw() {
  ElMessage.info('提现功能开发中...')
}

function confirmDeposit() {
  ElMessage.success('充值申请已提交')
  depositVisible.value = false
}

onMounted(() => {
  // 加载数据
})
</script>

<style scoped lang="scss">
.finance-deposit-page {
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

    .summary-card {
      background: #fff;
      border-radius: 8px;
      padding: 20px;
      border: 1px solid #ebeef5;

      .card-content {
        display: flex;
        justify-content: space-between;
        align-items: center;
      }

      .card-info {
        .stat-label {
          font-size: 12px;
          color: #909399;
          margin-bottom: 8px;
        }

        .stat-value {
          font-size: 32px;
          font-weight: 600;
          color: #303133;
          margin-bottom: 8px;
        }

        .stat-sub {
          font-size: 12px;
          color: #909399;
        }
      }

      .card-actions {
        display: flex;
        gap: 12px;
      }
    }
  }
}

.records-card {
  .records-header {
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
