<template>
  <view class="orders-page">
    <!-- 顶部标题栏 -->
    <view class="header">
      <text class="title">订单</text>
      <view class="header-icons">
        <view class="icon-btn search-btn">
          <view class="search-circle"></view>
          <view class="search-handle"></view>
        </view>
        <view class="icon-btn filter-btn">
          <view class="filter-shape"></view>
          <view class="filter-lines">
            <view class="filter-line"></view>
            <view class="filter-line"></view>
            <view class="filter-line"></view>
          </view>
        </view>
      </view>
    </view>
    
    <!-- 本月销售统计 -->
    <view class="sales-summary">
      <view class="sales-header">
        <text class="sales-label">本月销售</text>
        <text class="arrow">></text>
      </view>
      <view class="sales-content">
        <view class="sales-count">
          <text class="count-value">245</text>
          <text class="count-unit">单</text>
        </view>
        <text class="equals">=</text>
        <view class="sales-amount">
          <text class="currency">NT$</text>
          <text class="amount-value">38,420.00</text>
        </view>
      </view>
    </view>
    
    <!-- 订单状态统计 -->
    <view class="status-cards">
      <view class="status-card">
        <text class="status-label">待结算</text>
        <text class="arrow">></text>
        <text class="status-value">608.50</text>
        <text class="status-sub">订单：6</text>
      </view>
      <view class="status-card">
        <text class="status-label">已结算</text>
        <text class="arrow">></text>
        <text class="status-value">21,608.50</text>
        <text class="status-sub">订单：242</text>
      </view>
      <view class="status-card">
        <text class="status-label">账款调整</text>
        <text class="arrow">></text>
        <text class="status-value">608.50</text>
        <text class="status-sub">订单：1</text>
      </view>
    </view>
    
    <!-- 近期订单 -->
    <view class="orders-section">
      <view class="section-header">
        <text class="section-title">近期订单(45)</text>
      </view>
      
      <view class="order-item" v-for="i in 3" :key="i">
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
      <view class="tab-item" @click="switchTab('home')">
        <image class="tab-icon" src="/static/home.png" mode="aspectFit" />
        <text class="tab-text">首页</text>
      </view>
      <view class="tab-item active" @click="switchTab('orders')">
        <image class="tab-icon" src="/static/order-active.png" mode="aspectFit" />
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
function switchTab(tab: string) {
  if (tab === 'home') {
    uni.navigateTo({ url: '/shopower/pages/home/home' })
  } else if (tab === 'mine') {
    uni.navigateTo({ url: '/shopower/pages/mine/mine' })
  }
}
</script>

<style lang="scss">
.orders-page {
  min-height: 100vh;
  background-color: #f5f5f5;
  padding-bottom: 120rpx;
}

.header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  background: #fff;
  padding: 60rpx 32rpx 32rpx;
  
  .title {
    font-size: 36rpx;
    font-weight: bold;
    color: #333;
  }
  
  .header-icons {
    display: flex;
    gap: 32rpx;
    
    .icon-btn {
      width: 48rpx;
      height: 48rpx;
      position: relative;
    }
    
    .search-btn {
      .search-circle {
        width: 28rpx;
        height: 28rpx;
        border: 4rpx solid #333;
        border-radius: 50%;
        position: absolute;
        top: 4rpx;
        left: 4rpx;
      }
      
      .search-handle {
        width: 14rpx;
        height: 4rpx;
        background: #333;
        position: absolute;
        bottom: 8rpx;
        right: 4rpx;
        transform: rotate(45deg);
        border-radius: 2rpx;
      }
    }
    
    .filter-btn {
      display: flex;
      flex-direction: row;
      
      .filter-shape {
        width: 0;
        height: 0;
        border-left: 16rpx solid transparent;
        border-right: 16rpx solid transparent;
        border-top: 20rpx solid #333;
        position: absolute;
        top: 8rpx;
        left: 8rpx;
      }
      
      .filter-lines {
        position: absolute;
        right: 0;
        top: 8rpx;
        display: flex;
        flex-direction: column;
        gap: 6rpx;
        
        .filter-line {
          height: 4rpx;
          background: #333;
          border-radius: 2rpx;
          
          &:nth-child(1) { width: 20rpx; }
          &:nth-child(2) { width: 16rpx; }
          &:nth-child(3) { width: 12rpx; }
        }
      }
    }
  }
}

.sales-summary {
  margin: 24rpx 32rpx;
  background: #fff;
  border-radius: 24rpx;
  padding: 32rpx;
}

.sales-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 16rpx;
  
  .sales-label {
    font-size: 28rpx;
    color: #666;
  }
  
  .arrow {
    color: #999;
  }
}

.sales-content {
  display: flex;
  align-items: center;
  gap: 24rpx;
}

.sales-count {
  .count-value {
    font-size: 48rpx;
    font-weight: bold;
    color: #ff6600;
  }
  
  .count-unit {
    font-size: 28rpx;
    color: #ff6600;
    margin-left: 8rpx;
  }
}

.equals {
  font-size: 32rpx;
  color: #999;
}

.sales-amount {
  .currency {
    font-size: 28rpx;
    color: #333;
  }
  
  .amount-value {
    font-size: 40rpx;
    font-weight: bold;
    color: #333;
  }
}

.status-cards {
  display: flex;
  gap: 16rpx;
  margin: 0 32rpx 24rpx;
}

.status-card {
  flex: 1;
  background: #fff;
  border-radius: 16rpx;
  padding: 24rpx;
  position: relative;
  
  .status-label {
    display: block;
    font-size: 24rpx;
    color: #666;
    margin-bottom: 12rpx;
  }
  
  .arrow {
    position: absolute;
    top: 24rpx;
    right: 24rpx;
    font-size: 24rpx;
    color: #ccc;
  }
  
  .status-value {
    display: block;
    font-size: 32rpx;
    font-weight: bold;
    color: #333;
    margin-bottom: 8rpx;
  }
  
  .status-sub {
    font-size: 22rpx;
    color: #999;
  }
}

.orders-section {
  margin: 0 32rpx;
  background: #fff;
  border-radius: 24rpx;
  padding: 32rpx;
}

.section-header {
  margin-bottom: 24rpx;
  
  .section-title {
    font-size: 32rpx;
    font-weight: bold;
    color: #333;
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
