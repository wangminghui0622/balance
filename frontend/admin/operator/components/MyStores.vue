<template>
  <el-card class="stores-card">
    <template #header>
      <div class="card-header">
        <span>我的店铺 ({{ storeList.length }})</span>
        <el-button type="text" size="small">更多</el-button>
      </div>
    </template>
    <div class="stores-list">
      <div
        v-for="(store, index) in storeList"
        :key="index"
        class="store-item"
      >
        <el-avatar :size="40" shape="square" :src="store.avatar" />
        <div class="store-info">
          <div class="store-name">{{ store.name }}</div>
          <div class="store-meta">
            <span class="store-status">{{ store.authorized ? '已授权' : '未授权' }}</span>
            <span class="store-id">店铺ID：{{ store.storeId }}</span>
          </div>
        </div>
      </div>
    </div>
  </el-card>
</template>

<script setup lang="ts">
import { ref } from 'vue'

interface Store {
  avatar: string
  name: string
  storeId: string
  authorized: boolean
}

const storeList = ref<Store[]>([
  {
    avatar: '',
    name: '店铺名称示例文字占位符文字占位符',
    storeId: '1234567890',
    authorized: true
  },
  {
    avatar: '',
    name: '店铺名称示例文字占位符文字占位符',
    storeId: '1234567890',
    authorized: true
  },
  {
    avatar: '',
    name: '店铺名称示例文字占位符文字占位符',
    storeId: '1234567890',
    authorized: true
  },
  {
    avatar: '',
    name: '店铺名称示例文字占位符文字占位符',
    storeId: '1234567890',
    authorized: true
  }
])
</script>

<style scoped lang="scss">
.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  font-weight: 500;
  font-size: 16px;
}

.stores-list {
  display: flex;
  flex-direction: column;
  gap: 16px;
}

.store-item {
  display: flex;
  align-items: flex-start;
  gap: 12px;
  padding: 8px;
  border-radius: 4px;
  transition: background-color 0.3s;

  &:hover {
    background-color: #f5f7fa;
  }
}

.store-info {
  flex: 1;
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.store-name {
  font-size: 14px;
  font-weight: 500;
  color: #303133;
  line-height: 1.4;
}

.store-meta {
  display: flex;
  gap: 12px;
  font-size: 12px;
  color: #909399;
  margin-top: 4px;

  .store-status {
    color: #67c23a;
  }
}
</style>
