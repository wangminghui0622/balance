<template>
  <view class="select-role-page">
    <!-- 顶部橙色区域 -->
    <view class="header">
      <view class="back-btn" @click="goBack">
        <text class="arrow">‹</text>
      </view>
    </view>
    
    <!-- 内容区域 -->
    <view class="content">
      <text class="title">请选择您的身份</text>
      <text class="subtitle">选择适合的身份，以便于我们为您提供最佳服务</text>
      
      <!-- 身份选项 -->
      <view class="role-list">
        <!-- 店主选项 -->
        <view 
          class="role-item" 
          :class="{ active: selectedRole === 'shopowner' }"
          @click="selectRole('shopowner')"
        >
          <view class="role-icon shopowner">
            <text class="icon">🏪</text>
          </view>
          <view class="role-info">
            <text class="role-name">我是店主</text>
            <text class="role-desc">我拥有一家或多家店铺，目前缺乏管理店铺的时间，希望寻找一名合适运营合作，帮助我管理运营店铺。</text>
          </view>
          <view class="check-icon" v-if="selectedRole === 'shopowner'">
            <text class="checked">✓</text>
          </view>
        </view>
        
        <!-- 运营选项 -->
        <view 
          class="role-item" 
          :class="{ active: selectedRole === 'operator' }"
          @click="selectRole('operator')"
        >
          <view class="role-icon operator">
            <text class="icon">💼</text>
          </view>
          <view class="role-info">
            <text class="role-name">我是运营</text>
            <text class="role-desc">我拥有虾皮店铺运营经验，全链路独立负责过店铺的规划、筹措、营销和推广工作，希望寻找到合作店主。</text>
          </view>
          <view class="check-icon" v-if="selectedRole === 'operator'">
            <text class="checked">✓</text>
          </view>
        </view>
      </view>
    </view>
    
    <!-- 底部按钮 -->
    <view class="footer">
      <button class="next-btn" :disabled="!selectedRole" @click="goToRegister">下一步</button>
    </view>
  </view>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import { ROUTE_PATH, USER_TYPE_NUM } from '@share/constants'

type RoleType = 'shopowner' | 'operator' | null

const selectedRole = ref<RoleType>(null)

function selectRole(role: RoleType) {
  selectedRole.value = role
}

function goBack() {
  uni.navigateBack()
}

function goToRegister() {
  if (!selectedRole.value) {
    uni.showToast({ title: '请选择身份', icon: 'none' })
    return
  }
  
  // 传递用户类型到注册页
  const userType = selectedRole.value === 'shopowner' ? USER_TYPE_NUM.SHOPOWNER : USER_TYPE_NUM.OPERATOR
  uni.navigateTo({ 
    url: `${ROUTE_PATH.REGISTER}?userType=${userType}` 
  })
}
</script>

<style lang="scss">
.select-role-page {
  min-height: 100vh;
  background-color: #f5f5f5;
  display: flex;
  flex-direction: column;
}

.header {
  background: linear-gradient(135deg, #ff8c00 0%, #ff6600 100%);
  padding: 60rpx 32rpx 40rpx;
}

.back-btn {
  width: 64rpx;
  height: 64rpx;
  display: flex;
  align-items: center;
  justify-content: center;
  
  .arrow {
    font-size: 48rpx;
    color: #333;
    font-weight: bold;
  }
}

.content {
  flex: 1;
  padding: 48rpx 32rpx;
  background: #f5f5f5;
}

.title {
  display: block;
  font-size: 44rpx;
  font-weight: bold;
  color: #333;
  margin-bottom: 16rpx;
}

.subtitle {
  display: block;
  font-size: 26rpx;
  color: #999;
  margin-bottom: 48rpx;
}

.role-list {
  display: flex;
  flex-direction: column;
  gap: 24rpx;
}

.role-item {
  background: #fff;
  border-radius: 24rpx;
  padding: 32rpx;
  display: flex;
  align-items: flex-start;
  gap: 24rpx;
  border: 3rpx solid transparent;
  transition: all 0.3s;
  
  &.active {
    border-color: #ff6600;
    background: #fff9f5;
  }
}

.role-icon {
  width: 80rpx;
  height: 80rpx;
  border-radius: 16rpx;
  display: flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
  
  &.shopowner {
    background: linear-gradient(135deg, #ff8c00 0%, #ff6600 100%);
  }
  
  &.operator {
    background: #f0f0f0;
  }
  
  .icon {
    font-size: 40rpx;
  }
}

.role-info {
  flex: 1;
  
  .role-name {
    display: block;
    font-size: 32rpx;
    font-weight: bold;
    color: #333;
    margin-bottom: 12rpx;
  }
  
  .role-desc {
    display: block;
    font-size: 24rpx;
    color: #999;
    line-height: 1.6;
  }
}

.check-icon {
  width: 40rpx;
  height: 40rpx;
  border-radius: 50%;
  background: #ff6600;
  display: flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
  
  .checked {
    color: #fff;
    font-size: 24rpx;
  }
}

.footer {
  padding: 32rpx;
  padding-bottom: calc(32rpx + env(safe-area-inset-bottom));
}

.next-btn {
  width: 100%;
  height: 88rpx;
  background: linear-gradient(135deg, #ff8c00 0%, #ff6600 100%);
  border-radius: 44rpx;
  color: #fff;
  font-size: 32rpx;
  font-weight: bold;
  border: none;
  
  &::after {
    border: none;
  }
  
  &[disabled] {
    opacity: 0.5;
  }
}
</style>
