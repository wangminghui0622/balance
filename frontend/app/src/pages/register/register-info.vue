<template>
  <view class="register-info-page">
    <!-- 顶部区域 -->
    <view class="header">
      <view class="back-btn" @click="goBack">
        <text class="arrow">‹</text>
      </view>
      <text class="header-title">注册{{ roleText }}</text>
    </view>
    
    <!-- 内容区域 -->
    <view class="content">
      <!-- 真实姓名 -->
      <view class="section">
        <text class="section-title">真实姓名</text>
        <view class="form-item">
          <input 
            class="input" 
            type="text" 
            v-model="form.realName" 
            placeholder="请输入您的称呼"
            placeholder-class="placeholder"
          />
        </view>
      </view>
      
      <!-- 联系方式 -->
      <view class="section">
        <text class="section-title">联系方式</text>
        <text class="section-desc">请至少输入一种联系方式，以方便我们平台工作人员联系您，为您提供更多服务和支持。</text>
        
        <view class="form-item">
          <text class="label">选填：请输入您的电话</text>
          <input 
            class="input" 
            type="text" 
            v-model="form.phone" 
            placeholder="请输入电话号码"
            placeholder-class="placeholder"
          />
        </view>
        
        <view class="form-item">
          <text class="label">选填：请输入您的LINE ID</text>
          <input 
            class="input" 
            type="text" 
            v-model="form.lineId" 
            placeholder="请输入LINE ID"
            placeholder-class="placeholder"
          />
        </view>
        
        <view class="form-item">
          <text class="label">选填：请输入您的微信账号</text>
          <input 
            class="input" 
            type="text" 
            v-model="form.wechat" 
            placeholder="请输入微信账号"
            placeholder-class="placeholder"
          />
        </view>
      </view>
    </view>
    
    <!-- 底部按钮 -->
    <view class="footer">
      <button class="submit-btn" :loading="loading" @click="handleRegister">确定并注册</button>
    </view>
    
    <!-- 注册成功弹窗 -->
    <view class="success-modal" v-if="showSuccessModal">
      <view class="modal-mask"></view>
      <view class="modal-content">
        <view class="success-icon">
          <view class="check-circle">
            <text class="check-mark">✓</text>
          </view>
        </view>
        <text class="success-title">注册成功</text>
        <text class="success-desc">恭喜您注册成功，欢迎加入我们!</text>
      </view>
    </view>
  </view>
</template>

<script setup lang="ts">
import { ref, reactive, computed, onMounted } from 'vue'
import { register } from '@share/api/auth'
import { USER_TYPE_NUM } from '@share/constants'

const loading = ref(false)
const showSuccessModal = ref(false)

interface Step1Data {
  username: string
  email: string
  emailCode: string
  password: string
  userType: number
}

const step1Data = ref<Step1Data | null>(null)

const form = reactive({
  realName: '',
  phone: '',
  lineId: '',
  wechat: ''
})

const roleText = computed(() => {
  if (!step1Data.value) return ''
  switch (step1Data.value.userType) {
    case USER_TYPE_NUM.SHOPOWNER:
      return '店主'
    case USER_TYPE_NUM.OPERATOR:
      return '运营'
    case USER_TYPE_NUM.PLATFORM:
      return '平台'
    default:
      return ''
  }
})

onMounted(() => {
  // 读取第一步数据
  const data = uni.getStorageSync('register_step1_data')
  if (data) {
    try {
      step1Data.value = JSON.parse(data)
    } catch (e) {
      uni.showToast({ title: '数据异常，请重新填写', icon: 'none' })
      uni.navigateBack()
    }
  } else {
    uni.showToast({ title: '请先填写账号信息', icon: 'none' })
    uni.navigateBack()
  }
})

async function handleRegister() {
  if (!step1Data.value) {
    uni.showToast({ title: '数据异常，请重新填写', icon: 'none' })
    return
  }
  
  // 至少填写一种联系方式
  if (!form.phone && !form.lineId && !form.wechat) {
    uni.showToast({ title: '请至少填写一种联系方式', icon: 'none' })
    return
  }
  
  loading.value = true
  try {
    const res = await register({
      username: step1Data.value.username,
      email: step1Data.value.email,
      emailCode: step1Data.value.emailCode,
      password: step1Data.value.password,
      userType: step1Data.value.userType,
      realName: form.realName,
      phone: form.phone,
      lineId: form.lineId,
      wechat: form.wechat
    })
    
    if (res.code === 0) {
      // 清除临时数据
      uni.removeStorageSync('register_step1_data')
      
      // 显示成功弹窗
      showSuccessModal.value = true
      
      // 2秒后跳转到登录页
      setTimeout(() => {
        showSuccessModal.value = false
        uni.reLaunch({ url: '/pages/login/login' })
      }, 2000)
    } else {
      uni.showToast({ title: res.message || '注册失败', icon: 'none' })
    }
  } catch (err: any) {
    uni.showToast({ title: err.message || '注册失败', icon: 'none' })
  } finally {
    loading.value = false
  }
}

function goBack() {
  uni.navigateBack()
}
</script>

<style lang="scss">
.register-info-page {
  min-height: 100vh;
  background-color: #f5f5f5;
  display: flex;
  flex-direction: column;
}

.header {
  background: #fff;
  padding: 60rpx 32rpx 30rpx;
  display: flex;
  align-items: center;
  position: relative;
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

.header-title {
  position: absolute;
  left: 50%;
  transform: translateX(-50%);
  font-size: 34rpx;
  font-weight: bold;
  color: #333;
}

.content {
  flex: 1;
  padding: 24rpx 32rpx;
}

.section {
  background: #fff;
  border-radius: 24rpx;
  padding: 32rpx;
  margin-bottom: 24rpx;
}

.section-title {
  display: block;
  font-size: 32rpx;
  font-weight: bold;
  color: #333;
  margin-bottom: 16rpx;
}

.section-desc {
  display: block;
  font-size: 24rpx;
  color: #999;
  line-height: 1.6;
  margin-bottom: 24rpx;
}

.form-item {
  margin-bottom: 24rpx;
  
  &:last-child {
    margin-bottom: 0;
  }
  
  .label {
    display: block;
    font-size: 26rpx;
    color: #666;
    margin-bottom: 12rpx;
  }
  
  .input {
    width: 100%;
    height: 80rpx;
    border-bottom: 1rpx solid #eee;
    font-size: 28rpx;
  }
  
  .placeholder {
    color: #ccc;
  }
}

.footer {
  padding: 32rpx;
  padding-bottom: calc(32rpx + env(safe-area-inset-bottom));
}

.submit-btn {
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

.success-modal {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  z-index: 999;
  display: flex;
  align-items: center;
  justify-content: center;
  
  .modal-mask {
    position: absolute;
    top: 0;
    left: 0;
    right: 0;
    bottom: 0;
    background: rgba(0, 0, 0, 0.5);
  }
  
  .modal-content {
    position: relative;
    width: 500rpx;
    background: #fff;
    border-radius: 24rpx;
    padding: 60rpx 40rpx;
    display: flex;
    flex-direction: column;
    align-items: center;
  }
  
  .success-icon {
    margin-bottom: 32rpx;
    
    .check-circle {
      width: 120rpx;
      height: 120rpx;
      background: linear-gradient(135deg, #52c41a 0%, #73d13d 100%);
      border-radius: 50%;
      display: flex;
      align-items: center;
      justify-content: center;
      box-shadow: 0 8rpx 24rpx rgba(82, 196, 26, 0.3);
      
      .check-mark {
        font-size: 60rpx;
        color: #fff;
        font-weight: bold;
      }
    }
  }
  
  .success-title {
    font-size: 36rpx;
    font-weight: bold;
    color: #333;
    margin-bottom: 16rpx;
  }
  
  .success-desc {
    font-size: 28rpx;
    color: #999;
    text-align: center;
  }
}
</style>
