<template>
  <view class="set-password-page">
    <!-- 顶部区域 -->
    <view class="header">
      <view class="back-btn" @click="goBack">
        <text class="arrow">‹</text>
      </view>
      <text class="header-title">设置新密码</text>
    </view>
    
    <!-- 表单区域 -->
    <view class="form-container">
      <!-- 设定密码 -->
      <view class="form-section">
        <text class="section-title">设定密码</text>
        <text class="section-desc">请选择密码，密码位数应为8-16位，且至少包含数字、字母、特殊符号中的两种</text>
        
        <view class="form-item">
          <input 
            class="input" 
            :type="showPassword ? 'text' : 'password'" 
            v-model="form.password" 
            placeholder="请输入密码"
            placeholder-class="placeholder"
          />
          <view class="eye-icon" @click="showPassword = !showPassword">
            <text>{{ showPassword ? '👁' : '👁‍🗨' }}</text>
          </view>
        </view>
        
        <view class="form-item">
          <input 
            class="input" 
            :type="showConfirmPassword ? 'text' : 'password'" 
            v-model="form.confirmPassword" 
            placeholder="请输入确认密码"
            placeholder-class="placeholder"
          />
          <view class="eye-icon" @click="showConfirmPassword = !showConfirmPassword">
            <text>{{ showConfirmPassword ? '👁' : '👁‍🗨' }}</text>
          </view>
        </view>
      </view>
    </view>
    
    <!-- 底部区域 -->
    <view class="footer">
      <view class="help-link">
        <text class="text">没有收到验证码？</text>
        <text class="link" @click="goBack">修改邮箱</text>
      </view>
      <button class="submit-btn" :loading="loading" @click="handleSubmit">确定</button>
    </view>
    
    <!-- 成功弹窗 -->
    <view class="success-modal" v-if="showSuccessModal">
      <view class="modal-mask"></view>
      <view class="modal-content">
        <view class="success-icon">
          <view class="check-circle">
            <text class="check-mark">✓</text>
          </view>
        </view>
        <text class="success-title">密码重置成功</text>
        <text class="success-desc">请使用新密码登录</text>
      </view>
    </view>
  </view>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted } from 'vue'
import { resetPassword } from '@share/api/auth'

const loading = ref(false)
const showPassword = ref(false)
const showConfirmPassword = ref(false)
const showSuccessModal = ref(false)

interface ResetData {
  email: string
  emailCode: string
}

const resetData = ref<ResetData | null>(null)

const form = reactive({
  password: '',
  confirmPassword: ''
})

onMounted(() => {
  const data = uni.getStorageSync('reset_password_data')
  if (data) {
    try {
      resetData.value = JSON.parse(data)
    } catch (e) {
      uni.showToast({ title: '数据异常，请重新操作', icon: 'none' })
      uni.navigateBack()
    }
  } else {
    uni.showToast({ title: '请先验证邮箱', icon: 'none' })
    uni.navigateBack()
  }
})

function validatePassword(password: string): boolean {
  if (password.length < 8 || password.length > 16) {
    return false
  }
  
  let typeCount = 0
  if (/[0-9]/.test(password)) typeCount++
  if (/[a-zA-Z]/.test(password)) typeCount++
  if (/[^0-9a-zA-Z]/.test(password)) typeCount++
  
  return typeCount >= 2
}

async function handleSubmit() {
  if (!resetData.value) {
    uni.showToast({ title: '数据异常，请重新操作', icon: 'none' })
    return
  }
  
  if (!form.password) {
    uni.showToast({ title: '请输入密码', icon: 'none' })
    return
  }
  
  if (!validatePassword(form.password)) {
    uni.showToast({ title: '密码需8-16位，包含数字、字母、符号中至少两种', icon: 'none' })
    return
  }
  
  if (form.password !== form.confirmPassword) {
    uni.showToast({ title: '两次密码不一致', icon: 'none' })
    return
  }
  
  loading.value = true
  try {
    const res = await resetPassword({
      email: resetData.value.email,
      emailCode: resetData.value.emailCode,
      newPassword: form.password
    })
    
    if (res.code === 0) {
      // 清除临时数据
      uni.removeStorageSync('reset_password_data')
      
      // 显示成功弹窗
      showSuccessModal.value = true
      
      // 2秒后跳转到登录页
      setTimeout(() => {
        showSuccessModal.value = false
        uni.reLaunch({ url: '/pages/login/login' })
      }, 2000)
    } else {
      uni.showToast({ title: res.message || '重置失败', icon: 'none' })
    }
  } catch (err: any) {
    uni.showToast({ title: err.message || '重置失败', icon: 'none' })
  } finally {
    loading.value = false
  }
}

function goBack() {
  uni.navigateBack()
}
</script>

<style lang="scss">
.set-password-page {
  min-height: 100vh;
  background-color: #fff;
  display: flex;
  flex-direction: column;
}

.header {
  display: flex;
  align-items: center;
  padding: 24rpx 32rpx;
  padding-top: calc(24rpx + env(safe-area-inset-top));
  border-bottom: 1rpx solid #f5f5f5;
}

.back-btn {
  width: 60rpx;
  height: 60rpx;
  display: flex;
  align-items: center;
  justify-content: center;
  
  .arrow {
    font-size: 48rpx;
    color: #333;
    font-weight: 300;
  }
}

.header-title {
  flex: 1;
  text-align: center;
  font-size: 34rpx;
  font-weight: bold;
  color: #333;
  margin-right: 60rpx;
}

.form-container {
  flex: 1;
  padding: 32rpx;
}

.form-section {
  .section-title {
    display: block;
    font-size: 28rpx;
    color: #333;
    margin-bottom: 16rpx;
    padding-left: 16rpx;
    border-left: 4rpx solid #ff6600;
  }
  
  .section-desc {
    display: block;
    font-size: 24rpx;
    color: #ff6600;
    margin-bottom: 24rpx;
    line-height: 1.5;
  }
}

.form-item {
  display: flex;
  align-items: center;
  border-bottom: 1rpx solid #eee;
  padding: 16rpx 0;
  margin-bottom: 16rpx;
  
  .input {
    flex: 1;
    height: 60rpx;
    font-size: 28rpx;
    color: #333;
  }
  
  .placeholder {
    color: #ccc;
  }
  
  .eye-icon {
    padding: 16rpx;
    font-size: 32rpx;
  }
}

.footer {
  padding: 32rpx;
  padding-bottom: calc(32rpx + env(safe-area-inset-bottom));
}

.help-link {
  display: flex;
  justify-content: center;
  align-items: center;
  margin-bottom: 32rpx;
  
  .text {
    font-size: 26rpx;
    color: #999;
  }
  
  .link {
    font-size: 26rpx;
    color: #ff6600;
    margin-left: 8rpx;
  }
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
