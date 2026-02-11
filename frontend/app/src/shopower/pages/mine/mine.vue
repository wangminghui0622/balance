<template>
  <view class="mine-page">
    <!-- 顶部用户信息 -->
    <view class="header">
      <view class="user-section">
        <view class="avatar">
          <text class="avatar-text">店</text>
        </view>
        <view class="user-info">
          <text class="username">Hector</text>
          <text class="email">530542673@qq.com</text>
        </view>
      </view>
      <text class="edit-btn">编辑信息</text>
      <view class="notification-icon">
        <view class="icon-wrapper">
          <text class="bell-icon">💬</text>
          <text class="badge">6</text>
        </view>
      </view>
    </view>
    
    <!-- 我的佣金卡片 -->
    <view class="commission-card">
      <view class="commission-header">
        <text class="commission-title">我的佣金</text>
        <text class="commission-icon">💰</text>
      </view>
      <view class="commission-amount">
        <text class="currency">NT$</text>
        <text class="amount">4,450.00</text>
      </view>
      <view class="commission-footer">
        <text class="total-label">累计佣金：</text>
        <text class="total-amount">NT$ 33,260.00</text>
      </view>
      <view class="commission-btns">
        <button class="btn primary">提现</button>
        <button class="btn outline">提现</button>
      </view>
    </view>
    
    <!-- 店铺和保证金 -->
    <view class="info-cards">
      <view class="info-card">
        <text class="info-label">我的店铺</text>
        <text class="info-icon">🏪</text>
        <text class="info-value">4</text>
      </view>
      <view class="info-card">
        <text class="info-label">我的保证金</text>
        <text class="info-icon">💎</text>
        <text class="info-value">¥3000</text>
      </view>
    </view>
    
    <!-- 邀请好友 -->
    <view class="invite-section">
      <view class="invite-item">
        <text class="invite-icon">👥</text>
        <text class="invite-text">邀请好友</text>
        <text class="invite-link">完成任务返佣金 ></text>
      </view>
    </view>
    
    <!-- 菜单列表 -->
    <view class="menu-list">
      <view class="menu-item">
        <text class="menu-icon">🔒</text>
        <text class="menu-text">账户安全</text>
        <text class="menu-arrow">></text>
      </view>
      <view class="menu-item">
        <text class="menu-icon">💬</text>
        <text class="menu-text">官方客服</text>
        <text class="menu-arrow">></text>
      </view>
      <view class="menu-item">
        <text class="menu-icon">ℹ️</text>
        <text class="menu-text">关于我们</text>
        <text class="menu-arrow">></text>
      </view>
      <view class="menu-item">
        <text class="menu-icon">📝</text>
        <text class="menu-text">意见反馈</text>
        <text class="menu-arrow">></text>
      </view>
      <view class="menu-item">
        <text class="menu-icon">⚙️</text>
        <text class="menu-text">设置</text>
        <text class="menu-arrow">></text>
      </view>
      <view class="menu-item logout-item" @click="handleLogout">
        <text class="menu-icon">🚪</text>
        <text class="menu-text">退出登录</text>
        <text class="menu-arrow">></text>
      </view>
    </view>
    
    <!-- 底部导航 -->
    <view class="tab-bar">
      <view class="tab-item" @click="switchTab('home')">
        <image class="tab-icon" src="/static/home.png" mode="aspectFit" />
        <text class="tab-text">首页</text>
      </view>
      <view class="tab-item" @click="switchTab('orders')">
        <image class="tab-icon" src="/static/order.png" mode="aspectFit" />
        <text class="tab-text">订单</text>
      </view>
      <view class="tab-item active" @click="switchTab('mine')">
        <image class="tab-icon" src="/static/me-active.png" mode="aspectFit" />
        <text class="tab-text">我的</text>
      </view>
    </view>
  </view>
</template>

<script setup lang="ts">
function switchTab(tab: string) {
  if (tab === 'home') {
    uni.navigateTo({ url: '/shopower/pages/home/home' })
  } else if (tab === 'orders') {
    uni.navigateTo({ url: '/shopower/pages/orders/orders' })
  }
}

function handleLogout() {
  uni.showModal({
    title: '提示',
    content: '确定要退出登录吗？',
    success: (res) => {
      if (res.confirm) {
        // 清除localStorage
        uni.removeStorageSync('token')
        uni.removeStorageSync('userId')
        uni.removeStorageSync('userType')
        uni.removeStorageSync('rememberLogin')
        // #ifdef H5
        // 清除sessionStorage
        sessionStorage.removeItem('token')
        sessionStorage.removeItem('userId')
        sessionStorage.removeItem('userType')
        // #endif
        uni.reLaunch({ url: '/pages/login/login' })
      }
    }
  })
}
</script>

<style lang="scss">
.mine-page {
  min-height: 100vh;
  background-color: #f5f5f5;
  padding-bottom: 120rpx;
}

.header {
  display: flex;
  align-items: center;
  padding: 60rpx 32rpx 32rpx;
  background: #fff;
  position: relative;
}

.user-section {
  display: flex;
  align-items: center;
  gap: 20rpx;
  flex: 1;
}

.avatar {
  width: 100rpx;
  height: 100rpx;
  border-radius: 50%;
  background: linear-gradient(135deg, #ff8c00 0%, #ff6600 100%);
  display: flex;
  align-items: center;
  justify-content: center;
  
  .avatar-text {
    color: #fff;
    font-size: 40rpx;
    font-weight: bold;
  }
}

.user-info {
  .username {
    display: block;
    font-size: 36rpx;
    font-weight: bold;
    color: #333;
    margin-bottom: 8rpx;
  }
  
  .email {
    font-size: 24rpx;
    color: #999;
  }
}

.edit-btn {
  font-size: 26rpx;
  color: #ff6600;
  margin-right: 32rpx;
}

.notification-icon {
  .icon-wrapper {
    position: relative;
    width: 60rpx;
    height: 60rpx;
    display: flex;
    align-items: center;
    justify-content: center;
    
    .bell-icon {
      font-size: 40rpx;
      color: #666;
    }
    
    .badge {
      position: absolute;
      top: -8rpx;
      right: -8rpx;
      background: #ff6600;
      color: #fff;
      font-size: 20rpx;
      min-width: 32rpx;
      height: 32rpx;
      border-radius: 16rpx;
      display: flex;
      align-items: center;
      justify-content: center;
    }
  }
}

.commission-card {
  margin: 24rpx 32rpx;
  padding: 32rpx;
  background: linear-gradient(135deg, #4a90d9 0%, #357abd 100%);
  border-radius: 24rpx;
  position: relative;
  
  .commission-header {
    display: flex;
    align-items: center;
    gap: 12rpx;
    margin-bottom: 16rpx;
    
    .commission-title {
      color: rgba(255,255,255,0.9);
      font-size: 28rpx;
    }
    
    .commission-icon {
      width: 40rpx;
      height: 40rpx;
    }
  }
  
  .commission-amount {
    margin-bottom: 16rpx;
    
    .currency {
      color: #fff;
      font-size: 32rpx;
      margin-right: 8rpx;
    }
    
    .amount {
      color: #fff;
      font-size: 56rpx;
      font-weight: bold;
    }
  }
  
  .commission-footer {
    margin-bottom: 24rpx;
    
    .total-label, .total-amount {
      color: rgba(255,255,255,0.8);
      font-size: 24rpx;
    }
  }
  
  .commission-btns {
    display: flex;
    gap: 16rpx;
    
    .btn {
      flex: 1;
      height: 64rpx;
      border-radius: 32rpx;
      font-size: 28rpx;
      border: none;
      
      &.primary {
        background: #ff6600;
        color: #fff;
      }
      
      &.outline {
        background: transparent;
        border: 2rpx solid #fff;
        color: #fff;
      }
    }
  }
}

.info-cards {
  display: flex;
  gap: 16rpx;
  margin: 0 32rpx 24rpx;
}

.info-card {
  flex: 1;
  background: #fff;
  border-radius: 16rpx;
  padding: 24rpx;
  display: flex;
  flex-direction: column;
  
  .info-label {
    font-size: 24rpx;
    color: #666;
    margin-bottom: 12rpx;
  }
  
  .info-icon {
    width: 60rpx;
    height: 60rpx;
    margin-bottom: 12rpx;
  }
  
  .info-value {
    font-size: 32rpx;
    font-weight: bold;
    color: #333;
  }
}

.invite-section {
  margin: 0 32rpx 24rpx;
  background: #fff;
  border-radius: 16rpx;
  padding: 24rpx 32rpx;
}

.invite-item {
  display: flex;
  align-items: center;
  
  .invite-icon {
    font-size: 36rpx;
    margin-right: 16rpx;
  }
  
  .invite-text {
    font-size: 28rpx;
    color: #333;
    flex: 1;
  }
  
  .invite-link {
    font-size: 24rpx;
    color: #ff6600;
  }
}

.menu-list {
  margin: 0 32rpx;
  background: #fff;
  border-radius: 16rpx;
  overflow: hidden;
}

.menu-item {
  display: flex;
  align-items: center;
  padding: 32rpx;
  border-bottom: 1rpx solid #f5f5f5;
  
  &:last-child {
    border-bottom: none;
  }
  
  .menu-icon {
    font-size: 36rpx;
    margin-right: 20rpx;
  }
  
  .menu-text {
    flex: 1;
    font-size: 28rpx;
    color: #333;
  }
  
  .menu-arrow {
    font-size: 28rpx;
    color: #ccc;
  }
}

.tab-bar {
  position: fixed;
  bottom: 0;
  left: 0;
  right: 0;
  height: 100rpx;
  background: #fff;
  display: flex;
  box-shadow: 0 -2rpx 10rpx rgba(0,0,0,0.05);
  padding-bottom: env(safe-area-inset-bottom);
}

.tab-item {
  flex: 1;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  gap: 4rpx;
  
  .tab-icon {
    width: 48rpx;
    height: 48rpx;
  }
  
  .tab-text {
    font-size: 22rpx;
    color: #999;
  }
  
  &.active {
    .tab-text {
      color: #ff6600;
    }
  }
}
</style>
