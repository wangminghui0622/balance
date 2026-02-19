<template>
  <el-card class="alert-card">
    <template #header>
      <div class="card-header">
        <span>运营业绩预警</span>
        <el-button type="text" size="small">更多</el-button>
      </div>
    </template>
    <el-tabs v-model="activeTab" class="alert-tabs">
      <el-tab-pane label="1日单量预警(3)" name="1day">
        <div class="alert-list">
          <div v-for="(item, index) in alertList" :key="index" class="alert-item">
            <el-avatar :size="40" shape="square" :src="item.avatar" />
            <div class="alert-info">
              <div class="store-name">{{ item.storeName }}</div>
              <div class="store-id">店铺ID: {{ item.storeId }}</div>
              <div class="order-count">今日单量: {{ item.todayOrders }}</div>
            </div>
            <el-button size="small" type="primary">检查</el-button>
          </div>
        </div>
      </el-tab-pane>
      <el-tab-pane label="7日业绩预警(8)" name="7day">
        <div class="empty-tab">暂无数据</div>
      </el-tab-pane>
      <el-tab-pane label="30日业绩预警(2)" name="30day">
        <div class="empty-tab">暂无数据</div>
      </el-tab-pane>
    </el-tabs>
  </el-card>
</template>

<script setup lang="ts">
import { ref } from 'vue'

interface PerformanceAlert {
  avatar: string
  storeName: string
  storeId: string
  todayOrders: number
}

const activeTab = ref('1day')

const alertList = ref<PerformanceAlert[]>([
  {
    avatar: '',
    storeName: '店铺名称示例文字占位符',
    storeId: '1234567890',
    todayOrders: 0
  },
  {
    avatar: '',
    storeName: '店铺名称示例文字占位符',
    storeId: '1234567890',
    todayOrders: 0
  },
  {
    avatar: '',
    storeName: '店铺名称示例文字占位符',
    storeId: '1234567890',
    todayOrders: 0
  }
])
</script>

<style scoped lang="scss">
.alert-card {
  // 移除 height: 100%，让卡片根据内容自适应高度
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

.store-id,
.order-count {
  font-size: 12px;
  color: #909399;
}

.empty-tab {
  text-align: center;
  padding: 40px;
  color: #909399;
}
</style>
