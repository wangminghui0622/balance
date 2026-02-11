<template>
  <view class="home-page">
    <!-- 顶部用户信息 -->
    <view class="header">
      <view class="user-section">
        <view class="avatar">
          <text class="avatar-text">店</text>
        </view>
        <view class="user-info">
          <view class="greeting">
            <text class="greeting-text">早上好！店主</text>
          </view>
          <view class="notice">
            <text>今天店铺一切正常！</text>
            <text>有</text><text class="notice-highlight">4条</text><text>通知等您查阅...</text>
          </view>
        </view>
      </view>
      <view class="notification-icon">
        <view class="icon-wrapper">
          <text class="bell-icon">💬</text>
          <text class="badge">6</text>
        </view>
      </view>
    </view>
    
    <!-- 预付款余额卡片 -->
    <view class="balance-card">
      <view class="balance-header">
        <text class="balance-title">预付款余额</text>
        <text class="eye-icon">👁</text>
      </view>
      <view class="balance-amount">
        <text class="currency">NT$</text>
        <text class="amount">223,560.50</text>
      </view>
      <view class="balance-footer">
        <text class="total-label">累计充值金额：</text>
        <text class="total-amount">NT$1,445,860.50</text>
      </view>
      <button class="recharge-btn">充值</button>
    </view>
    
    <!-- 订单金额统计 -->
    <view class="stats-card">
      <view class="stats-row">
        <view class="stat-item">
          <text class="stat-label">未结算订单金额(NT$)</text>
          <text class="stat-value">3,246.00</text>
          <text class="stat-sub">托管中订单：6</text>
        </view>
        <view class="stat-item">
          <text class="stat-label">未结算佣金(NT$)</text>
          <text class="stat-value">608.50</text>
          <text class="stat-sub">托管中的订单佣金</text>
        </view>
      </view>
      <view class="stats-row">
        <view class="stat-item">
          <text class="stat-label">已结算订单金额(NT$)</text>
          <text class="stat-value large">1,353,636.00</text>
          <text class="stat-sub">已结算订单：45</text>
        </view>
        <view class="stat-item">
          <text class="stat-label">已结算佣金(NT$)</text>
          <text class="stat-value large">24,543.00</text>
          <text class="stat-sub">已结算的订单佣金</text>
        </view>
      </view>
    </view>
    
    <!-- 近期收益 -->
    <view class="earnings-section">
      <view class="section-header">
        <view class="title-with-bar">
          <view class="title-bar"></view>
          <text class="section-title">近期收益</text>
        </view>
        <text class="section-link">查看详情 ></text>
      </view>
      <view class="earnings-content">
        <view class="earnings-stats">
          <view class="earning-item">
            <text class="earning-label">今日销售(NT$)</text>
            <text class="earning-value highlight">8,420.00</text>
          </view>
          <view class="earning-item">
            <text class="earning-label">订单数量</text>
            <text class="earning-value">45</text>
          </view>
          <view class="earning-item">
            <text class="earning-label">预估佣金(NT$)</text>
            <text class="earning-value">540</text>
          </view>
        </view>
        <!-- 折线图区域 -->
        <view class="chart-area">
          <text class="chart-y-label">销售(NT$)</text>
          <view class="chart-container">
            <svg class="line-chart" viewBox="0 0 300 120" preserveAspectRatio="none">
              <defs>
                <linearGradient id="areaGradient" x1="0%" y1="0%" x2="0%" y2="100%">
                  <stop offset="0%" style="stop-color:rgba(255,106,58,0.35)" />
                  <stop offset="100%" style="stop-color:rgba(255,106,58,0.05)" />
                </linearGradient>
              </defs>
              <path class="area-fill" d="M0,100 L10,85 L30,90 L50,70 L70,75 L90,55 L110,60 L130,45 L150,40 L170,50 L190,35 L210,40 L230,25 L250,30 L270,20 L290,25 L300,15 L300,120 L0,120 Z" fill="url(#areaGradient)" />
              <path class="line-path" d="M0,100 L10,85 L30,90 L50,70 L70,75 L90,55 L110,60 L130,45 L150,40 L170,50 L190,35 L210,40 L230,25 L250,30 L270,20 L290,25 L300,15" fill="none" stroke="#ff6a3a" stroke-width="2" />
            </svg>
            <view class="chart-x-axis">
              <text v-for="d in 10" :key="d">{{ d + 15 }}</text>
            </view>
          </view>
        </view>
      </view>
    </view>
    
    <!-- 最近订单 -->
    <view class="orders-section">
      <view class="orders-tabs">
        <text class="tab active">最近订单</text>
        <text class="tab">未结算</text>
        <text class="tab">已结算</text>
        <text class="tab-link">所有订单 ></text>
      </view>
      
      <view class="order-item" v-for="i in 2" :key="i">
        <view class="order-header">
          <text class="order-no">订单编号：X250904KQ2P078R</text>
          <text class="order-status">付款状态</text>
        </view>
        <view class="order-info">
          <text class="info-text">未结算佣金：NT$8.00</text>
          <text class="info-text">订单金额：NT$36.00</text>
        </view>
        <view class="order-product">
          <view class="product-img"></view>
          <view class="product-info">
            <text class="product-name">商品名称示例文字占位符替换即可文字占...</text>
            <text class="product-spec">颜色：xxx 尺寸：xxx</text>
            <text class="product-price">NT$46.00</text>
          </view>
          <text class="product-qty">x1</text>
        </view>
      </view>
    </view>
    
    <!-- 底部导航 -->
    <view class="tab-bar">
      <view class="tab-item active" @click="switchTab('home')">
        <image class="tab-icon" src="/static/home-active.png" mode="aspectFit" />
        <text class="tab-text">首页</text>
      </view>
      <view class="tab-item" @click="switchTab('orders')">
        <image class="tab-icon" src="/static/order.png" mode="aspectFit" />
        <text class="tab-text">订单</text>
      </view>
      <view class="tab-item" @click="switchTab('mine')">
        <image class="tab-icon" src="/static/me.png" mode="aspectFit" />
        <text class="tab-text">我的</text>
      </view>
    </view>
  </view>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'

const username = ref('店主')

onMounted(() => {
  const userId = uni.getStorageSync('userId')
  if (userId) {
    username.value = userId
  }
})

function switchTab(tab: string) {
  if (tab === 'orders') {
    uni.navigateTo({ url: '/shopower/pages/orders/orders' })
  } else if (tab === 'mine') {
    uni.navigateTo({ url: '/shopower/pages/mine/mine' })
  }
}
</script>

<style lang="scss">
.home-page {
  min-height: 100vh;
  background-color: #f5f5f5;
  padding-bottom: 120rpx;
}

.header {
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
  padding: 60rpx 32rpx 32rpx;
  background: #fff;
}

.user-section {
  display: flex;
  gap: 20rpx;
}

.avatar {
  width: 90rpx;
  height: 90rpx;
  border-radius: 50%;
  background: linear-gradient(135deg, #ff8c00 0%, #ff6600 100%);
  display: flex;
  align-items: center;
  justify-content: center;
  
  .avatar-text {
    color: #fff;
    font-size: 36rpx;
    font-weight: bold;
  }
}

.user-info {
  .greeting {
    display: flex;
    align-items: center;
    gap: 12rpx;
    margin-bottom: 8rpx;
    
    .greeting-text {
      font-size: 32rpx;
      font-weight: bold;
      color: #333;
    }
    
    .role-tag {
      background: linear-gradient(135deg, #ff8c00 0%, #ff6600 100%);
      color: #fff;
      font-size: 22rpx;
      padding: 4rpx 16rpx;
      border-radius: 20rpx;
    }
  }
  
  .notice {
    font-size: 24rpx;
    color: #999;
    
    .notice-highlight {
      color: #ff6600;
    }
  }
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

.balance-card {
  margin: 24rpx 32rpx;
  padding: 32rpx;
  background: linear-gradient(135deg, #ff8c00 0%, #ff6600 100%);
  border-radius: 24rpx;
  position: relative;
  
  .balance-header {
    display: flex;
    align-items: center;
    gap: 12rpx;
    margin-bottom: 16rpx;
    
    .balance-title {
      color: rgba(255,255,255,0.9);
      font-size: 28rpx;
    }
    
    .eye-icon {
      font-size: 28rpx;
    }
  }
  
  .balance-amount {
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
  
  .balance-footer {
    .total-label, .total-amount {
      color: rgba(255,255,255,0.8);
      font-size: 24rpx;
    }
  }
  
  .recharge-btn {
    position: absolute;
    right: 32rpx;
    top: 50%;
    transform: translateY(-50%);
    background: #fff;
    color: #ff6600;
    font-size: 28rpx;
    padding: 16rpx 40rpx;
    border-radius: 32rpx;
    border: none;
  }
}

.stats-card {
  margin: 0 32rpx 24rpx;
  background: #fff;
  border-radius: 24rpx;
  padding: 32rpx;
}

.stats-row {
  display: flex;
  margin-bottom: 32rpx;
  
  &:last-child {
    margin-bottom: 0;
  }
}

.stat-item {
  flex: 1;
  
  .stat-label {
    display: block;
    font-size: 24rpx;
    color: #999;
    margin-bottom: 12rpx;
  }
  
  .stat-value {
    display: block;
    font-size: 36rpx;
    font-weight: bold;
    color: #333;
    margin-bottom: 8rpx;
    
    &.large {
      font-size: 40rpx;
    }
  }
  
  .stat-sub {
    font-size: 22rpx;
    color: #ccc;
  }
}

.earnings-section {
  margin: 0 32rpx 24rpx;
  background: #fff;
  border-radius: 24rpx;
  padding: 32rpx;
}

.section-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 24rpx;
  
  .title-with-bar {
    display: flex;
    align-items: center;
    gap: 12rpx;
    
    .title-bar {
      width: 6rpx;
      height: 32rpx;
      background: #ff6a3a;
      border-radius: 3rpx;
    }
  }
  
  .section-title {
    font-size: 32rpx;
    font-weight: bold;
    color: #333;
  }
  
  .section-link {
    font-size: 26rpx;
    color: #999;
  }
}

.earnings-content {
  display: flex;
  gap: 24rpx;
}

.earnings-stats {
  flex: 0 0 200rpx;
  display: flex;
  flex-direction: column;
  gap: 32rpx;
}

.earning-item {
  .earning-label {
    display: block;
    font-size: 24rpx;
    color: #999;
    margin-bottom: 12rpx;
  }
  
  .earning-value {
    display: block;
    font-size: 36rpx;
    font-weight: bold;
    color: #333;
    
    &.highlight {
      font-size: 44rpx;
      color: #ff6a3a;
    }
  }
}

.chart-area {
  flex: 1;
  position: relative;
  
  .chart-y-label {
    position: absolute;
    top: 0;
    right: 0;
    font-size: 22rpx;
    color: #999;
  }
  
  .chart-container {
    margin-top: 40rpx;
    
    .line-chart {
      width: 100%;
      height: 200rpx;
    }
    
    .chart-x-axis {
      display: flex;
      justify-content: space-between;
      margin-top: 12rpx;
      
      text {
        font-size: 22rpx;
        color: #999;
      }
    }
  }
}

.orders-section {
  margin: 0 32rpx 24rpx;
  background: #fff;
  border-radius: 24rpx;
  padding: 32rpx;
}

.orders-tabs {
  display: flex;
  align-items: center;
  gap: 32rpx;
  margin-bottom: 24rpx;
  border-bottom: 1rpx solid #eee;
  padding-bottom: 24rpx;
  
  .tab {
    font-size: 28rpx;
    color: #999;
    
    &.active {
      color: #333;
      font-weight: bold;
    }
  }
  
  .tab-link {
    margin-left: auto;
    font-size: 26rpx;
    color: #999;
  }
}

.order-item {
  padding: 24rpx 0;
  border-bottom: 1rpx solid #f5f5f5;
  
  &:last-child {
    border-bottom: none;
  }
}

.order-header {
  display: flex;
  justify-content: space-between;
  margin-bottom: 16rpx;
  
  .order-no {
    font-size: 26rpx;
    color: #333;
  }
  
  .order-status {
    font-size: 24rpx;
    color: #ff6600;
  }
}

.order-info {
  display: flex;
  gap: 32rpx;
  margin-bottom: 16rpx;
  
  .info-text {
    font-size: 24rpx;
    color: #999;
  }
}

.order-product {
  display: flex;
  gap: 20rpx;
  
  .product-img {
    width: 120rpx;
    height: 120rpx;
    border-radius: 12rpx;
    background: #f5f5f5;
  }
  
  .product-info {
    flex: 1;
    
    .product-name {
      display: block;
      font-size: 26rpx;
      color: #333;
      margin-bottom: 8rpx;
      overflow: hidden;
      text-overflow: ellipsis;
      white-space: nowrap;
    }
    
    .product-spec {
      display: block;
      font-size: 22rpx;
      color: #999;
      margin-bottom: 8rpx;
    }
    
    .product-price {
      font-size: 28rpx;
      color: #ff6600;
      font-weight: bold;
    }
  }
  
  .product-qty {
    font-size: 24rpx;
    color: #999;
    align-self: flex-end;
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
