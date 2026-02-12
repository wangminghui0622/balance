<template>
  <div class="collection-page">
    <!-- 页面标题 -->
    <div class="page-header">
      <h1 class="page-title">收款账户</h1>
    </div>

    <!-- 电子钱包 -->
    <div class="account-section">
      <div class="section-title">
        <span class="title-indicator"></span>
        电子钱包
      </div>
      <div class="account-cards">
        <div 
          v-for="wallet in wallets" 
          :key="wallet.id" 
          class="account-card wallet-card"
          :class="{ 'is-default': wallet.isDefault }"
        >
          <div class="card-header">
            <div class="card-logo">
              <div class="logo-icon paypal">P</div>
            </div>
            <div class="card-info">
              <div class="card-name">{{ wallet.name }}</div>
              <el-tag size="small" type="primary">{{ wallet.status }}</el-tag>
            </div>
            <div class="card-default" v-if="wallet.isDefault">
              <el-icon><CircleCheckFilled /></el-icon>
              <span>默认</span>
            </div>
            <div class="card-default not-default" v-else>
              <span>○</span>
              <span>默认</span>
            </div>
          </div>
          <div class="card-details">
            <div class="detail-row">
              <div class="detail-item">
                <span class="detail-label">账号</span>
                <span class="detail-value">{{ wallet.account }}</span>
              </div>
              <div class="detail-item">
                <span class="detail-label">收款人</span>
                <span class="detail-value">{{ wallet.payee }}</span>
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>

    <!-- 银行账号 -->
    <div class="account-section">
      <div class="section-title">
        <span class="title-indicator"></span>
        银行账号
      </div>
      <div class="account-cards">
        <div 
          v-for="bank in banks" 
          :key="bank.id" 
          class="account-card bank-card"
          :class="{ 'is-default': bank.isDefault }"
        >
          <div class="card-header">
            <div class="card-logo">
              <div class="logo-icon bank">
                <el-icon><CreditCard /></el-icon>
              </div>
            </div>
            <div class="card-info">
              <div class="card-name">{{ bank.name }}</div>
              <el-tag size="small" :type="bank.status === '未激活' ? 'danger' : 'primary'">{{ bank.status }}</el-tag>
            </div>
            <div class="card-default" v-if="bank.isDefault">
              <el-icon><CircleCheckFilled /></el-icon>
              <span>默认</span>
            </div>
            <div class="card-default not-default" v-else>
              <span>○</span>
              <span>默认</span>
            </div>
          </div>
          <div class="card-details">
            <div class="detail-row">
              <div class="detail-item">
                <span class="detail-label">账号</span>
                <span class="detail-value">{{ bank.account }}</span>
              </div>
              <div class="detail-item">
                <span class="detail-label">币种</span>
                <span class="detail-value">{{ bank.currency }}</span>
              </div>
            </div>
          </div>
        </div>

        <!-- 新增银行账号卡片 -->
        <div class="account-card add-card" @click="handleAddBank">
          <div class="add-content">
            <el-icon class="add-icon"><Plus /></el-icon>
            <span>新增银行账号</span>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import { CircleCheckFilled, Plus, CreditCard } from '@element-plus/icons-vue'
import { ElMessage } from 'element-plus'

interface Wallet {
  id: number
  name: string
  status: string
  account: string
  payee: string
  isDefault: boolean
}

interface Bank {
  id: number
  name: string
  status: string
  account: string
  currency: string
  isDefault: boolean
}

const wallets = ref<Wallet[]>([
  {
    id: 1,
    name: 'PayPal支付>',
    status: '连接',
    account: '12345678910121314​15',
    payee: '123456789101',
    isDefault: true
  }
])

const banks = ref<Bank[]>([
  {
    id: 1,
    name: '汇丰银行>',
    status: '未激活',
    account: '12345678910121314​15',
    currency: 'CNY',
    isDefault: false
  }
])

const handleAddBank = () => {
  ElMessage.info('新增银行账号功能开发中')
}
</script>

<style lang="scss" scoped>
.collection-page {
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

.account-section {
  background: #fff;
  border: 1px solid #ebeef5;
  border-radius: 8px;
  padding: 20px;
  margin-bottom: 20px;

  .section-title {
    display: flex;
    align-items: center;
    gap: 8px;
    font-size: 14px;
    font-weight: 500;
    color: #303133;
    margin-bottom: 16px;

    .title-indicator {
      width: 3px;
      height: 14px;
      background-color: #ff6a3a;
      border-radius: 2px;
    }
  }

  .account-cards {
    display: flex;
    flex-wrap: wrap;
    gap: 16px;
  }
}

.account-card {
  width: 260px;
  border-radius: 8px;
  overflow: hidden;

  &.wallet-card {
    .card-header {
      background: linear-gradient(135deg, #0070ba 0%, #003087 100%);
    }
  }

  &.bank-card {
    .card-header {
      background: linear-gradient(135deg, #e74c3c 0%, #c0392b 100%);
    }
  }

  .card-header {
    padding: 16px;
    color: #fff;
    display: flex;
    align-items: flex-start;
    gap: 12px;
    position: relative;

    .card-logo {
      .logo-icon {
        width: 40px;
        height: 40px;
        border-radius: 50%;
        display: flex;
        align-items: center;
        justify-content: center;
        font-size: 20px;
        font-weight: bold;

        &.paypal {
          background: #fff;
          color: #0070ba;
        }

        &.bank {
          background: #fff;
          color: #e74c3c;
        }
      }
    }

    .card-info {
      flex: 1;

      .card-name {
        font-size: 14px;
        font-weight: 500;
        margin-bottom: 6px;
      }

      :deep(.el-tag) {
        background: rgba(255, 255, 255, 0.2);
        border: none;
        color: #fff;
      }
    }

    .card-default {
      display: flex;
      align-items: center;
      gap: 4px;
      font-size: 12px;
      color: #fff;

      &.not-default {
        opacity: 0.6;
      }
    }
  }

  .card-details {
    background: #fff;
    padding: 12px 16px;
    border: 1px solid #ebeef5;
    border-top: none;
    border-radius: 0 0 8px 8px;

    .detail-row {
      display: flex;
      gap: 24px;
    }

    .detail-item {
      display: flex;
      flex-direction: column;
      gap: 4px;

      .detail-label {
        font-size: 12px;
        color: #909399;
        white-space: nowrap;
      }

      .detail-value {
        font-size: 12px;
        color: #303133;
        white-space: nowrap;
      }
    }
  }

  &.add-card {
    border: 1px dashed #dcdfe6;
    background: #fff;
    display: flex;
    align-items: center;
    justify-content: center;
    min-height: 160px;
    cursor: pointer;
    transition: all 0.3s;

    &:hover {
      border-color: #ff6a3a;
      
      .add-content {
        color: #ff6a3a;
      }
    }

    .add-content {
      display: flex;
      flex-direction: column;
      align-items: center;
      gap: 8px;
      color: #909399;

      .add-icon {
        font-size: 24px;
      }

      span {
        font-size: 14px;
      }
    }
  }
}

@media (max-width: 768px) {
  .collection-page {
    padding: 12px;
  }

  .account-card {
    width: 100%;
  }
}
</style>
