<template>
  <el-card class="alert-card">
    <template #header>
      <div class="card-header">
        <span>店主异常预警</span>
        <el-button type="text" size="small">更多</el-button>
      </div>
    </template>
    <el-tabs v-model="activeTab" class="alert-tabs">
      <el-tab-pane label="授权异常(8)" name="auth">
        <div class="alert-list">
          <div v-for="(item, index) in alertList" :key="index" class="alert-item">
            <el-avatar :size="40" shape="square" :src="item.avatar" />
            <div class="alert-info">
              <div class="store-name">{{ item.storeName }}</div>
              <div class="owner-name">店主: {{ item.ownerName }}</div>
              <div class="time">时间: {{ item.time }}</div>
            </div>
            <el-button size="small" type="primary">检查</el-button>
          </div>
        </div>
      </el-tab-pane>
      <el-tab-pane label="通知异常(2)" name="notification">
        <div class="empty-tab">暂无数据</div>
      </el-tab-pane>
      <el-tab-pane label="申请提现(5)" name="withdraw">
        <div class="empty-tab">暂无数据</div>
      </el-tab-pane>
    </el-tabs>
  </el-card>
</template>

<script setup lang="ts">
import { ref } from 'vue'

interface AnomalyAlert {
  avatar: string
  storeName: string
  ownerName: string
  time: string
}

const activeTab = ref('auth')

const alertList = ref<AnomalyAlert[]>([
  {
    avatar: '',
    storeName: '店铺名称示例文字占位符',
    ownerName: '名称示例',
    time: '2026-12-12 23:59:59'
  },
  {
    avatar: '',
    storeName: '店铺名称示例文字占位符',
    ownerName: '名称示例',
    time: '2026-12-12 23:59:59'
  },
  {
    avatar: '',
    storeName: '店铺名称示例文字占位符',
    ownerName: '名称示例',
    time: '2026-12-12 23:59:59'
  }
])
</script>

<style scoped lang="scss">
.alert-card {
  height: 100%;
  display: flex;
  flex-direction: column;
}

.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  font-weight: 500;
  font-size: 16px;
}

.alert-tabs {
  :deep(.el-tabs__header) {
    margin-bottom: 16px;
  }
}

.alert-list {
  display: flex;
  flex-direction: column;
  gap: 16px;
}

.alert-item {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 12px;
  background-color: #f5f7fa;
  border-radius: 4px;
}

.alert-info {
  flex: 1;
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.store-name {
  font-size: 14px;
  font-weight: 500;
  color: #303133;
}

.owner-name,
.time {
  font-size: 12px;
  color: #909399;
}

.empty-tab {
  text-align: center;
  padding: 40px;
  color: #909399;
}
</style>
