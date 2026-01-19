<template>
  <el-card class="kpi-card">
    <div class="kpi-content">
      <div class="kpi-title">{{ title }}</div>
      <div class="kpi-value">{{ formatValue(value) }}</div>
      <div class="kpi-subtitle">{{ subtitle }}</div>
    </div>
  </el-card>
</template>

<script setup lang="ts">
interface Props {
  title: string
  value: number
  subtitle: string
}

defineProps<Props>()

const formatValue = (value: number) => {
  return value.toLocaleString('zh-CN', {
    minimumFractionDigits: 2,
    maximumFractionDigits: 2
  })
}
</script>

<style scoped lang="scss">
.kpi-card {
  height: 100%;
  border-radius: 8px;
}

.kpi-content {
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.kpi-title {
  font-size: 14px;
  color: #909399;
  line-height: 1.5;
}

.kpi-value {
  font-size: 28px;
  font-weight: bold;
  color: #303133;
  line-height: 1.2;
}

.kpi-subtitle {
  font-size: 12px;
  color: #909399;
}
</style>
