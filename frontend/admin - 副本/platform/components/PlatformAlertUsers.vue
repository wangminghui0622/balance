<template>
  <el-card class="alert-card">
    <template #header>
      <div class="card-header">
        <span>平台警示用户</span>
      </div>
    </template>
    <div class="alert-list">
      <div v-for="item in alertData" :key="item.type" class="alert-item">
        <div class="alert-label">{{ item.label }}</div>
        <div class="alert-value">{{ item.value }}</div>
      </div>
    </div>
  </el-card>
</template>

<script setup lang="ts">
import { ref } from 'vue'

interface AlertItem {
  label: string
  value: number
  type: string
}

const alertData = ref<AlertItem[]>([
  {
    label: '无效店主',
    value: 5,
    type: 'invalid_owner'
  },
  {
    label: '无效运营',
    value: 2,
    type: 'invalid_operator'
  },
  {
    label: '闲置店铺',
    value: 12,
    type: 'idle_store'
  }
])
</script>

<style scoped lang="scss">
.alert-card {
  height: 100%;
}

.card-header {
  font-weight: 500;
  font-size: 16px;
}

.alert-list {
  display: flex;
  flex-direction: row;
  gap: 20px;
  justify-content: space-around;
}

.alert-item {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 8px;
  flex: 1;
}

.alert-label {
  font-size: 14px;
  color: #606266;
}

.alert-value {
  font-size: 20px;
  font-weight: bold;
  color: #f56c6c;
}
</style>
