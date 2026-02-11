<template>
  <el-card class="balance-card">
    <div class="balance-content">
      <div class="balance-label">回款余额</div>
      <div class="balance-value">NT$223,560.50</div>
      <el-button type="default" class="withdraw-btn">提现</el-button>
    </div>
  </el-card>
</template>

<script setup lang="ts">
</script>

<style scoped lang="scss">
.balance-card {
  background-color: #606266;
  border: none;

  :deep(.el-card__body) {
    padding: 24px;
  }
}

.balance-content {
  display: flex;
  flex-direction: column;
  gap: 16px;
}

.balance-label {
  font-size: 14px;
  color: rgba(255, 255, 255, 0.8);
}

.balance-value {
  font-size: 32px;
  font-weight: bold;
  color: #fff;
}

.withdraw-btn {
  width: 100%;
  background-color: rgba(255, 255, 255, 0.2);
  color: #fff;
  border: none;

  &:hover {
    background-color: rgba(255, 255, 255, 0.3);
  }
}
</style>
