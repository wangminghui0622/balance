<template>
  <el-card class="alert-card">
    <template #header>
      <div class="card-header">
        <span>店主充值预警({{ alertList.length }})</span>
        <el-button type="text" size="small">更多</el-button>
      </div>
    </template>
    <div class="alert-list">
      <div v-for="(item, index) in alertList" :key="index" class="alert-item">
        <el-avatar :size="32" :src="item.avatar" />
        <div class="alert-info">
          <div class="owner-name">{{ item.ownerName }}</div>
          <div class="financial-info">
            <span>可用: {{ item.available }}</span>
            <span>冻结: {{ item.frozen }}</span>
            <span>欠费: {{ item.owed }}</span>
          </div>
        </div>
        <el-tag :type="getTagType(item.status)" size="small">
          {{ item.status }}
        </el-tag>
      </div>
    </div>
  </el-card>
</template>

<script setup lang="ts">
import { ref } from 'vue'

interface RechargeAlert {
  avatar: string
  ownerName: string
  available: string
  frozen: string
  owed: string
  status: string
}

const alertList = ref<RechargeAlert[]>([
  {
    avatar: '',
    ownerName: '店主名称占位符',
    available: '0.00',
    frozen: '123.45',
    owed: '123.45',
    status: '预付款不足'
  },
  {
    avatar: '',
    ownerName: '店主名称占位符',
    available: '0.00',
    frozen: '123.45',
    owed: '123.45',
    status: '预付款欠费'
  },
  {
    avatar: '',
    ownerName: '店主名称占位符',
    available: '0.00',
    frozen: '123.45',
    owed: '123.45',
    status: '预付款逾期'
  },
  {
    avatar: '',
    ownerName: '店主名称占位符',
    available: '0.00',
    frozen: '123.45',
    owed: '123.45',
    status: '预付款不足'
  }
])

const getTagType = (status: string) => {
  if (status.includes('欠费') || status.includes('逾期')) {
    return 'danger'
  }
  return 'warning'
}
</script>

<style scoped lang="scss">
.alert-card {
  height: 100%;
  display: flex;
  flex-direction: column;

  :deep(.el-card__body) {
    padding: 12px;
  }
}

.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  font-weight: 500;
  font-size: 16px;
}

.alert-list {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.alert-item {
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 8px;
  background-color: #f5f7fa;
  border-radius: 4px;
}

.alert-info {
  flex: 1;
  display: flex;
  flex-direction: column;
  gap: 2px;
}

.owner-name {
  font-size: 14px;
  font-weight: 500;
  color: #303133;
}

.financial-info {
  font-size: 12px;
  color: #909399;
  display: flex;
  gap: 12px;
}
</style>
