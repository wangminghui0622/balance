<template>
  <div class="finance-page">
    <div class="page-header">
      <h1 class="page-title">财务管理</h1>
    </div>
    
    <!-- 财务概览卡片 -->
    <el-row :gutter="20" class="overview-row">
      <el-col :xs="24" :sm="12" :md="6">
        <el-card class="overview-card payment">
          <div class="card-icon">💰</div>
          <div class="card-info">
            <div class="card-label">回款余额(NT$)</div>
            <div class="card-value">{{ formatAmount(overview.paymentBalance) }}</div>
          </div>
        </el-card>
      </el-col>
      <el-col :xs="24" :sm="12" :md="6">
        <el-card class="overview-card unsettled">
          <div class="card-icon">📋</div>
          <div class="card-info">
            <div class="card-label">未结算回款(NT$)</div>
            <div class="card-value">{{ formatAmount(overview.unsettledPayment) }}</div>
          </div>
        </el-card>
      </el-col>
      <el-col :xs="24" :sm="12" :md="6">
        <el-card class="overview-card settled">
          <div class="card-icon">✅</div>
          <div class="card-info">
            <div class="card-label">已结算回款(NT$)</div>
            <div class="card-value">{{ formatAmount(overview.settledPayment) }}</div>
          </div>
        </el-card>
      </el-col>
      <el-col :xs="24" :sm="12" :md="6">
        <el-card class="overview-card withdrawn">
          <div class="card-icon">🏦</div>
          <div class="card-info">
            <div class="card-label">已提现金额(NT$)</div>
            <div class="card-value">{{ formatAmount(overview.withdrawnAmount) }}</div>
          </div>
        </el-card>
      </el-col>
    </el-row>

    <!-- 功能入口 -->
    <el-row :gutter="20" class="menu-row">
      <el-col :xs="24" :sm="12" :md="6">
        <el-card class="menu-card" @click="goToPayment">
          <div class="menu-icon">💵</div>
          <div class="menu-title">回款管理</div>
          <div class="menu-desc">查看回款明细和结算记录</div>
        </el-card>
      </el-col>
      <el-col :xs="24" :sm="12" :md="6">
        <el-card class="menu-card" @click="goToWithdraw">
          <div class="menu-icon">🏧</div>
          <div class="menu-title">提现管理</div>
          <div class="menu-desc">申请提现和查看提现记录</div>
        </el-card>
      </el-col>
      <el-col :xs="24" :sm="12" :md="6">
        <el-card class="menu-card" @click="goToAccount">
          <div class="menu-icon">🏦</div>
          <div class="menu-title">账户管理</div>
          <div class="menu-desc">管理收款账户信息</div>
        </el-card>
      </el-col>
      <el-col :xs="24" :sm="12" :md="6">
        <el-card class="menu-card" @click="goToStatement">
          <div class="menu-icon">📊</div>
          <div class="menu-title">对账单</div>
          <div class="menu-desc">查看和下载对账单</div>
        </el-card>
      </el-col>
    </el-row>

    <!-- 最近交易记录 -->
    <el-card class="recent-card">
      <template #header>
        <div class="card-header">
          <span class="title">最近交易记录</span>
          <el-button type="primary" link @click="goToPayment">查看全部</el-button>
        </div>
      </template>
      <el-table :data="recentTransactions" style="width: 100%">
        <el-table-column prop="time" label="时间" width="160" />
        <el-table-column prop="type" label="类型" width="100">
          <template #default="{ row }">
            <el-tag :type="row.type === 'income' ? 'success' : 'danger'" size="small">
              {{ row.type === 'income' ? '收入' : '支出' }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="description" label="描述" min-width="200" />
        <el-table-column prop="amount" label="金额(NT$)" width="140" align="right">
          <template #default="{ row }">
            <span :class="row.type === 'income' ? 'text-success' : 'text-danger'">
              {{ row.type === 'income' ? '+' : '-' }}{{ formatAmount(row.amount) }}
            </span>
          </template>
        </el-table-column>
        <el-table-column prop="balance" label="余额(NT$)" width="140" align="right">
          <template #default="{ row }">
            {{ formatAmount(row.balance) }}
          </template>
        </el-table-column>
      </el-table>
    </el-card>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive } from 'vue'
import { useRouter } from 'vue-router'

const router = useRouter()

const overview = reactive({
  paymentBalance: 125680.50,
  unsettledPayment: 18650.00,
  settledPayment: 1070300.00,
  withdrawnAmount: 950000.00
})

const recentTransactions = ref([
  { time: '2025-01-15 14:30:00', type: 'income', description: '订单回款 - X250904KQ2P078R', amount: 288, balance: 125680.50 },
  { time: '2025-01-14 10:20:00', type: 'income', description: '订单回款 - X250904KQ2P079S', amount: 384, balance: 125392.50 },
  { time: '2025-01-13 16:45:00', type: 'expense', description: '提现到银行账户', amount: 50000, balance: 125008.50 },
  { time: '2025-01-12 09:00:00', type: 'income', description: '订单回款 - X250904KQ2P080T', amount: 512, balance: 175008.50 }
])

function formatAmount(value: number): string {
  return value.toLocaleString('en-US', { minimumFractionDigits: 2, maximumFractionDigits: 2 })
}

function goToPayment() {
  router.push('/operator/finance/payment')
}

function goToWithdraw() {
  router.push('/operator/finance/withdraw')
}

function goToAccount() {
  router.push('/operator/finance/account')
}

function goToStatement() {
  router.push('/operator/finance/statement')
}
</script>

<style scoped lang="scss">
.finance-page {
  .page-header {
    margin-bottom: 20px;
    
    .page-title {
      font-size: 20px;
      font-weight: 500;
      color: #303133;
      margin: 0;
    }
  }
  
  .overview-row {
    margin-bottom: 20px;
  }
  
  .overview-card {
    display: flex;
    align-items: center;
    padding: 20px;
    
    .card-icon {
      font-size: 40px;
      margin-right: 16px;
    }
    
    .card-info {
      .card-label {
        font-size: 14px;
        color: #909399;
        margin-bottom: 4px;
      }
      
      .card-value {
        font-size: 24px;
        font-weight: 600;
        color: #303133;
      }
    }
  }
  
  .menu-row {
    margin-bottom: 20px;
  }
  
  .menu-card {
    cursor: pointer;
    text-align: center;
    padding: 24px;
    transition: all 0.3s;
    
    &:hover {
      transform: translateY(-4px);
      box-shadow: 0 4px 12px rgba(0, 0, 0, 0.1);
    }
    
    .menu-icon {
      font-size: 48px;
      margin-bottom: 12px;
    }
    
    .menu-title {
      font-size: 16px;
      font-weight: 500;
      color: #303133;
      margin-bottom: 8px;
    }
    
    .menu-desc {
      font-size: 12px;
      color: #909399;
    }
  }
  
  .recent-card {
    .card-header {
      display: flex;
      justify-content: space-between;
      align-items: center;
      
      .title {
        font-size: 16px;
        font-weight: 500;
      }
    }
  }
  
  .text-success {
    color: #67c23a;
  }
  
  .text-danger {
    color: #f56c6c;
  }
}
</style>
