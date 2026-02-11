<template>
  <view class="forgot-page">
    <!-- 顶部区域 -->
    <view class="header">
      <view class="back-btn" @click="goBack">
        <text class="arrow">‹</text>
      </view>
      <text class="header-title">忘记密码</text>
    </view>
    
    <!-- 表单区域 -->
    <view class="form-container">
      <!-- 登录账号 -->
      <view class="form-section">
        <text class="section-title">登录账号</text>
        <view class="form-item">
          <input 
            class="input" 
            type="text" 
            v-model="form.email" 
            placeholder="12345678@qq.com"
            placeholder-class="placeholder"
          />
        </view>
        <text class="hint-text" :class="{ error: emailError }">{{ emailError || '请输入正确邮箱' }}</text>
      </view>
      
      <!-- 验证码 -->
      <view class="form-section">
        <text class="section-title">验证码</text>
        <view class="form-item code-item">
          <input 
            class="input" 
            type="text" 
            v-model="form.emailCode" 
            placeholder="请输入验证码"
            placeholder-class="placeholder"
          />
          <text 
            class="send-code-btn" 
            :class="{ disabled: countdown > 0 }"
            @click="sendCode"
          >
            {{ countdown > 0 ? `${countdown}s` : '发送验证码' }}
          </text>
        </view>
      </view>
    </view>
    
    <!-- 底部区域 -->
    <view class="footer">
      <view class="help-link">
        <text class="text">没有收到验证码？</text>
        <text class="link">修改邮箱</text>
      </view>
      <button class="next-btn" @click="handleNext">下一步</button>
    </view>
  </view>
</template>

<script setup lang="ts">
import { ref, reactive } from 'vue'
import { sendEmailCode } from '@share/api/auth'

const form = reactive({
  email: '',
  emailCode: ''
})

const emailError = ref('')
const countdown = ref(0)
let timer: any = null

function validateEmail(email: string): boolean {
  return /^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(email)
}

async function sendCode() {
  if (countdown.value > 0) return
  
  if (!form.email) {
    emailError.value = '请输入邮箱'
    return
  }
  
  if (!validateEmail(form.email)) {
    emailError.value = '请输入正确邮箱'
    return
  }
  
  emailError.value = ''
  
  try {
    const res = await sendEmailCode({ email: form.email, type: 'reset' })
    if (res.code === 0) {
      uni.showToast({ title: '验证码已发送', icon: 'success' })
      countdown.value = 60
      timer = setInterval(() => {
        countdown.value--
        if (countdown.value <= 0) {
          clearInterval(timer)
        }
      }, 1000)
    } else {
      uni.showToast({ title: res.message || '发送失败', icon: 'none' })
    }
  } catch (err: any) {
    uni.showToast({ title: err.message || '发送失败', icon: 'none' })
  }
}

function handleNext() {
  if (!form.email) {
    emailError.value = '请输入邮箱'
    return
  }
  
  if (!validateEmail(form.email)) {
    emailError.value = '请输入正确邮箱'
    return
  }
  
  if (!form.emailCode) {
    uni.showToast({ title: '请输入验证码', icon: 'none' })
    return
  }
  
  if (form.emailCode.length !== 6) {
    uni.showToast({ title: '验证码为6位数字', icon: 'none' })
    return
  }
  
  // 保存数据到临时存储，跳转到设置新密码页面
  uni.setStorageSync('reset_password_data', JSON.stringify({
    email: form.email,
    emailCode: form.emailCode
  }))
  
  uni.navigateTo({ url: '/pages/forgot-password/set-password' })
}

function goBack() {
  uni.navigateBack()
}
</script>

<style lang="scss">
.forgot-page {
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
  margin-bottom: 40rpx;
  
  .section-title {
    display: block;
    font-size: 28rpx;
    color: #333;
    margin-bottom: 16rpx;
    padding-left: 16rpx;
    border-left: 4rpx solid #ff6600;
  }
}

.form-item {
  border-bottom: 1rpx solid #eee;
  padding: 16rpx 0;
  
  .input {
    width: 100%;
    height: 60rpx;
    font-size: 28rpx;
    color: #333;
  }
  
  .placeholder {
    color: #ccc;
  }
}

.code-item {
  display: flex;
  align-items: center;
  
  .input {
    flex: 1;
  }
  
  .send-code-btn {
    font-size: 26rpx;
    color: #ff6600;
    white-space: nowrap;
    padding-left: 24rpx;
    
    &.disabled {
      color: #999;
    }
  }
}

.hint-text {
  display: block;
  font-size: 24rpx;
  color: #ff6600;
  margin-top: 12rpx;
  
  &.error {
    color: #ff4d4f;
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
}
</style>
