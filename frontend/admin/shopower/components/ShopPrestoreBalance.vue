<template>
  <el-card class="balance-card">
    <div class="balance-content">
      <div class="balance-label">预存款余额</div>
      <div class="balance-value">NT$223,560.50</div>
      <div class="balance-subtitle">累计提现金额: NT$145,456.86</div>
      <div class="balance-actions">
        <el-button size="small" type="default" class="recharge-btn">充值</el-button>
        <el-button size="small" type="default" class="withdraw-btn">提现</el-button>
      </div>
    </div>
  </el-card>
</template>

<script setup lang="ts">
</script>

<style scoped lang="scss">
.balance-card {
  background-color: #303133;
  border: none;

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
  gap: 12px;
  margin-top: 8px;
}

.recharge-btn,
.withdraw-btn {
  flex: 1;
  border: none;
  background-color: rgba(255, 255, 255, 0.15);
  color: #fff;

  &:hover {
    background-color: rgba(255, 255, 255, 0.25);
  }
}
</style>

