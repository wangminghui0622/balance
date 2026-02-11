<template>
  <view class="home-page">
    <!-- 顶部用户信息 -->
    <view class="header">
      <view class="user-info">
        <view class="avatar">
          <text class="avatar-text">平</text>
        </view>
        <view class="info">
          <text class="name">{{ username }}</text>
          <text class="role">平台管理员</text>
        </view>
      </view>
      <view class="logout-btn" @click="handleLogout">
        <text>退出</text>
      </view>
    </view>
    
    <!-- 功能菜单 -->
    <view class="menu-grid">
      <view class="menu-item">
        <view class="icon users">👥</view>
        <text class="label">用户管理</text>
      </view>
      <view class="menu-item">
        <view class="icon stores">🏪</view>
        <text class="label">店铺管理</text>
      </view>
      <view class="menu-item">
        <view class="icon finance">💰</view>
        <text class="label">财务管理</text>
      </view>
      <view class="menu-item">
        <view class="icon settings">⚙️</view>
        <text class="label">系统设置</text>
      </view>
    </view>
    
    <!-- 平台数据概览 -->
    <view class="stats-card">
      <text class="card-title">平台数据</text>
      <view class="stats-grid">
        <view class="stat-item">
          <text class="value">0</text>
          <text class="label">总用户数</text>
        </view>
        <view class="stat-item">
          <text class="value">0</text>
          <text class="label">总店铺数</text>
        </view>
        <view class="stat-item">
          <text class="value">0</text>
          <text class="label">总订单数</text>
        </view>
        <view class="stat-item">
          <text class="value">0</text>
          <text class="label">总交易额</text>
        </view>
      </view>
    </view>
  </view>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { STORAGE_KEYS, ROUTE_PATH } from '@share/constants'

const username = ref('')

onMounted(() => {
  const userId = uni.getStorageSync(STORAGE_KEYS.USER_ID)
  username.value = userId || '管理员'
})

function handleLogout() {
  uni.showModal({
    title: '提示',
    content: '确定要退出登录吗？',
    success: (res) => {
      if (res.confirm) {
        uni.removeStorageSync(STORAGE_KEYS.TOKEN)
        uni.removeStorageSync(STORAGE_KEYS.USER_ID)
        uni.removeStorageSync(STORAGE_KEYS.USER_TYPE)
        uni.reLaunch({ url: ROUTE_PATH.LOGIN })
      }
    }
  })
}
</script>

<style lang="scss">
.home-page {
  min-height: 100vh;
  background-color: #f5f5f5;
  padding: 32rpx;
}

.header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 32rpx;
}

.user-info {
  display: flex;
  align-items: center;
  gap: 20rpx;
}

.avatar {
  width: 80rpx;
  height: 80rpx;
  border-radius: 50%;
  background: linear-gradient(135deg, #9c27b0 0%, #7b1fa2 100%);
  display: flex;
  align-items: center;
  justify-content: center;
  
  .avatar-text {
    color: #fff;
    font-size: 32rpx;
    font-weight: bold;
  }
}

.info {
  .name {
    display: block;
    font-size: 32rpx;
    font-weight: bold;
    color: #333;
  }
  
  .role {
    display: block;
    font-size: 24rpx;
    color: #999;
  }
}

.logout-btn {
  padding: 16rpx 32rpx;
  background: #fff;
  border-radius: 32rpx;
  
  text {
    font-size: 26rpx;
    color: #666;
  }
}

.menu-grid {
  display: grid;
  grid-template-columns: repeat(4, 1fr);
  gap: 24rpx;
  background: #fff;
  border-radius: 24rpx;
  padding: 32rpx;
  margin-bottom: 32rpx;
}

.menu-item {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 12rpx;
  
  .icon {
    width: 80rpx;
    height: 80rpx;
    border-radius: 16rpx;
    display: flex;
    align-items: center;
    justify-content: center;
    font-size: 40rpx;
    
    &.users {
      background: #f3e5f5;
    }
    
    &.stores {
      background: #e3f2fd;
    }
    
    &.finance {
      background: #fff3e0;
    }
    
    &.settings {
      background: #e8f5e9;
    }
  }
  
  .label {
    font-size: 24rpx;
    color: #666;
  }
}

.stats-card {
  background: #fff;
  border-radius: 24rpx;
  padding: 32rpx;
}

.card-title {
  display: block;
  font-size: 32rpx;
  font-weight: bold;
  color: #333;
  margin-bottom: 24rpx;
}

.stats-grid {
  display: grid;
  grid-template-columns: repeat(4, 1fr);
  gap: 16rpx;
}

.stat-item {
  text-align: center;
  
  .value {
    display: block;
    font-size: 40rpx;
    font-weight: bold;
    color: #9c27b0;
    margin-bottom: 8rpx;
  }
  
  .label {
    font-size: 24rpx;
    color: #999;
  }
}
</style>
