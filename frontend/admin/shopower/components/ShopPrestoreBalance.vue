<template>
  <el-card class="balance-card">
    <div class="balance-content">
      <div class="balance-label">
        <span>预存款余额</span>
        <button
          class="eye-toggle"
          type="button"
          :aria-label="showBalance ? '隐藏余额' : '显示余额'"
          @click="showBalance = !showBalance"
        >
          <svg
            v-if="showBalance"
            class="eye-icon"
            viewBox="0 0 24 24"
            fill="none"
            xmlns="http://www.w3.org/2000/svg"
          >
            <path
              d="M2.5 12S6.5 5.5 12 5.5S21.5 12 21.5 12S17.5 18.5 12 18.5S2.5 12 2.5 12Z"
              stroke="currentColor"
              stroke-width="1.8"
              stroke-linejoin="round"
            />
            <circle cx="12" cy="12" r="2.7" stroke="currentColor" stroke-width="1.8" />
          </svg>
          <svg
            v-else
            class="eye-icon"
            viewBox="0 0 24 24"
            fill="none"
            xmlns="http://www.w3.org/2000/svg"
          >
            <path
              d="M3 12S7 5.5 12 5.5S21 12 21 12S17 18.5 12 18.5S3 12 3 12Z"
              stroke="currentColor"
              stroke-width="1.8"
              stroke-linejoin="round"
            />
            <path
              d="M4 4L20 20"
              stroke="currentColor"
              stroke-width="1.8"
              stroke-linecap="round"
            />
          </svg>
        </button>
      </div>
      <div class="balance-value">{{ balanceText }}</div>
      <div class="balance-subtitle">累计提现金额: NT$ 145,456.86</div>
      <div class="balance-actions">
        <el-button size="small" type="default" class="withdraw-btn">提现</el-button>
        <el-button size="small" type="default" class="recharge-btn">充值</el-button>
      </div>
    </div>
  </el-card>
</template>

<script setup lang="ts">
import { computed, ref } from 'vue'

const showBalance = ref(true)
const balanceValue = ref('NT$ 223,560.50')

const balanceText = computed(() => {
  if (showBalance.value) return balanceValue.value
  return 'NT$ ******'
})
</script>

<style scoped lang="scss">
.balance-card {
  background: linear-gradient(135deg, #ff9a3a 0%, #ff7333 55%, #ff6a3a 100%);
  border: none;
  border-radius: 8px; // 设置圆角

  :deep(.el-card__body) {
    padding: 24px;
  }
}

.balance-content {
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.balance-label {
  font-size: 14px;
  color: rgba(255, 255, 255, 0.85);
  display: flex;
  align-items: center;
  gap: 8px;
}

.eye-toggle {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  padding: 0;
  border: none;
  background: transparent;
  color: rgba(255, 255, 255, 0.9);
  cursor: pointer;
}

.eye-icon {
  width: 16px;
  height: 16px;
}

.balance-value {
  font-size: 28px;
  font-weight: bold;
  color: #fff;
}

.balance-subtitle {
  font-size: 12px;
  color: rgba(255, 255, 255, 0.7);
}

.balance-actions {
  display: flex;
  justify-content: flex-end;
  align-items: center;
  gap: 12px;
  margin-top: 12px;
}

.withdraw-btn {
  border: 1px solid rgba(255, 255, 255, 0.9);
  background: transparent;
  color: rgba(255, 255, 255, 0.95);
  border-radius: 6px; // 按钮也设为圆角

  &:hover {
    border-color: rgba(255, 255, 255, 1);
    background-color: rgba(255, 255, 255, 0.08);
  }
}

.recharge-btn {
  border: 1px solid rgba(255, 255, 255, 1);
  background-color: rgba(255, 255, 255, 1);
  color: #ff5a2c;
  border-radius: 6px; // 按钮也设为圆角

  &:hover {
    background-color: rgba(255, 255, 255, 0.92);
  }
}
</style>

