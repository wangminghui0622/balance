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
          <div class="store-status">
            <el-tag size="small" :type="store.isAuthorized ? 'success' : 'warning'">
              {{ store.isAuthorized ? '已授权' : '未授权' }}
            </el-tag>
          </div>
          <div class="store-id">店铺ID: {{ store.storeId }}</div>
          <div class="store-actions">
            <el-button
              type="primary"
              size="small"
              :loading="store.authLoading"
              @click="handleAuth(store)"
            >
              {{ store.isAuthorized ? '重新授权' : '授权' }}
            </el-button>
          </div>
        </div>
      </div>
    </div>
  </el-card>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import { ElMessage } from 'element-plus'
import { shopeeApi } from '@share/api/shopee'

interface Store {
  avatar: string
  name: string
  storeId: string
  shopId: number // Shopee Shop ID
  isAuthorized: boolean
  authLoading?: boolean
}

const storeList = ref<Store[]>([
  {
    avatar: '',
    name: '店铺名称示例文字占位符文字占位符',
    storeId: '1234567890',
    shopId: 226445936, // Shopee Shop ID
    isAuthorized: false,
    authLoading: false
  }
])

const handleAuth = async (store: Store) => {
  if (!store.shopId) {
    ElMessage.warning('店铺 Shop ID 未配置')
    return
  }

  store.authLoading = true
  try {
    const res = await shopeeApi.getAuthURL(store.shopId)
    
    if (res.code === 200 && res.auth_url) {
      // 在新窗口打开授权链接
      window.open(res.auth_url, '_blank')
      ElMessage.success('正在跳转到 Shopee 授权页面...')
    } else {
      ElMessage.error(res.message || '获取授权链接失败')
    }
  } catch (err: any) {
    ElMessage.error(err?.response?.data?.message || err?.message || '获取授权链接失败')
  } finally {
    store.authLoading = false
  }
}
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

.store-status {
  margin: 4px 0;
}

.store-id {
  font-size: 12px;
  color: #909399;
}

.store-actions {
  margin-top: 8px;
}
</style>

