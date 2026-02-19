<template>
  <el-card class="data-card">
    <template #header>
      <div class="card-header">
        <span>平台用户数据</span>
      </div>
    </template>
    <div class="stat-list">
      <div v-for="item in statData" :key="item.label" class="stat-item">
        <div class="stat-label">{{ item.label }}</div>
        <div class="stat-value">{{ item.value }}</div>
        <div class="stat-compare">
          <span>vs.上周 {{ item.compareValue }}</span>
          <span :class="['trend', item.trend]">
            {{ item.trend === 'up' ? '↑' : '↓' }} {{ item.percentage }}%
          </span>
        </div>
      </div>
    </div>
  </el-card>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import type { StatData } from '@share/types'

const statData = ref<StatData[]>([
  {
    label: '店主数量',
    value: 45,
    compareValue: 10,
    percentage: 36.8,
    trend: 'up'
  },
  {
    label: '运营数量',
    value: 2,
    compareValue: 1,
    percentage: 36.8,
    trend: 'up'
  },
  {
    label: '店铺数量',
    value: 10,
    compareValue: 1,
    percentage: 36.8,
    trend: 'up'
  }
])
</script>

<style scoped lang="scss">
.data-card {
  height: 100%;
}

.card-header {
  font-weight: 500;
  font-size: 16px;
}

.stat-list {
  display: flex;
  flex-direction: row;
  gap: 20px;
  justify-content: space-around;
}

.stat-item {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.stat-label {
  font-size: 14px;
  color: #909399;
}

.stat-value {
  font-size: 24px;
  font-weight: bold;
  color: #303133;
}

.stat-compare {
  font-size: 12px;
  color: #909399;
  display: flex;
  gap: 8px;
}

.trend {
  &.up {
    color: #67c23a;
  }
  &.down {
    color: #f56c6c;
  }
}
</style>
