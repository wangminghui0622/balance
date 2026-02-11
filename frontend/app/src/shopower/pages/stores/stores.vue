<template>
  <view class="stores-page">
    <view class="header">
      <text class="title">店铺管理</text>
    </view>
    
    <view class="content">
      <view class="empty" v-if="stores.length === 0">
        <text class="icon">🏪</text>
        <text class="text">暂无店铺</text>
        <button class="add-btn" @click="addStore">添加店铺</button>
      </view>
      
      <view class="store-list" v-else>
        <view class="store-item" v-for="store in stores" :key="store.id">
          <view class="store-info">
            <text class="name">{{ store.name }}</text>
            <text class="status" :class="{ active: store.status === 1 }">
              {{ store.status === 1 ? '正常' : '禁用' }}
            </text>
          </view>
        </view>
      </view>
    </view>
  </view>
</template>

<script setup lang="ts">
import { ref } from 'vue'

interface Store {
  id: number
  name: string
  status: number
}

const stores = ref<Store[]>([])

function addStore() {
  uni.showToast({ title: '功能开发中', icon: 'none' })
}
</script>

<style lang="scss">
.stores-page {
  min-height: 100vh;
  background-color: #f5f5f5;
}

.header {
  background: #fff;
  padding: 32rpx;
  
  .title {
    font-size: 36rpx;
    font-weight: bold;
    color: #333;
  }
}

.content {
  padding: 32rpx;
}

.empty {
  display: flex;
  flex-direction: column;
  align-items: center;
  padding: 120rpx 0;
  
  .icon {
    font-size: 120rpx;
    margin-bottom: 32rpx;
  }
  
  .text {
    font-size: 28rpx;
    color: #999;
    margin-bottom: 48rpx;
  }
  
  .add-btn {
    padding: 24rpx 64rpx;
    background: linear-gradient(135deg, #ff8c00 0%, #ff6600 100%);
    border-radius: 44rpx;
    color: #fff;
    font-size: 28rpx;
    border: none;
    
    &::after {
      border: none;
    }
  }
}

.store-list {
  display: flex;
  flex-direction: column;
  gap: 24rpx;
}

.store-item {
  background: #fff;
  border-radius: 16rpx;
  padding: 32rpx;
}

.store-info {
  display: flex;
  justify-content: space-between;
  align-items: center;
  
  .name {
    font-size: 30rpx;
    color: #333;
  }
  
  .status {
    font-size: 24rpx;
    color: #999;
    
    &.active {
      color: #52c41a;
    }
  }
}
</style>
