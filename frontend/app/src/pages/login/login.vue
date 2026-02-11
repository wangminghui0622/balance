<template>
  <view class="login-page">
    <!-- 顶部橙色区域 -->
    <view class="header">
      <view class="header-content">
        <text class="greeting">你好!</text>
        <text class="title">欢迎使用天平系统</text>
      </view>
    </view>
    
    <!-- 白色表单区域 -->
    <view class="form-container">
      <view class="form-card">
        <!-- 账号输入 -->
        <view class="form-item">
          <text class="label">账号</text>
          <input 
            class="input" 
            type="text" 
            v-model="form.username" 
            placeholder="请输入手机号/邮箱/账号名"
            placeholder-class="placeholder"
          />
        </view>
        
        <!-- 密码输入 -->
        <view class="form-item">
          <text class="label">密码</text>
          <view class="password-wrapper">
            <input 
              class="input" 
              :type="showPassword ? 'text' : 'password'" 
              v-model="form.password" 
              placeholder="请输入密码"
              placeholder-class="placeholder"
            />
            <view class="eye-icon" @click="showPassword = !showPassword">
              <text class="iconfont">{{ showPassword ? '👁' : '👁‍🗨' }}</text>
            </view>
          </view>
        </view>
        
        <!-- 记住登录 & 忘记密码 -->
        <view class="options">
          <view class="remember" @click="rememberLogin = !rememberLogin">
            <view class="checkbox" :class="{ checked: rememberLogin }"></view>
            <text class="text">保持登录状态</text>
          </view>
          <text class="forgot" @click="goToForgotPassword">忘记密码?</text>
        </view>
        
        <!-- 登录按钮 -->
        <button class="login-btn" :loading="loading" @click="handleLogin">登录</button>
        
        <!-- 第三方登录 -->
        <view class="divider">
          <view class="line"></view>
          <text class="text">或者</text>
          <view class="line"></view>
        </view>
        
        <view class="social-login">
          <view class="social-icon google">G</view>
          <view class="social-icon line-icon">L</view>
          <view class="social-icon facebook">f</view>
        </view>
        
        <!-- 注册链接 -->
        <view class="register-link">
          <text class="text">还没有账号?</text>
          <text class="link" @click="goToRegister">注册></text>
        </view>
      </view>
    </view>
  </view>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted } from 'vue'
import { login } from '@share/api/auth'

const loading = ref(false)

// 获取token，优先从localStorage，其次从sessionStorage
function getToken(): string {
  let token = uni.getStorageSync('token')
  // #ifdef H5
  if (!token) {
    token = sessionStorage.getItem('token') || ''
  }
  // #endif
  return token
}

function getUserType(): string {
  let userType = uni.getStorageSync('userType')
  // #ifdef H5
  if (!userType) {
    userType = sessionStorage.getItem('userType') || ''
  }
  // #endif
  return userType
}

onMounted(() => {
  // 检查是否已登录，已登录则跳转
  const token = getToken()
  if (token) {
    const userType = getUserType()
    let targetRoute = '/pages/login/login'
    if (userType === '1') {
      targetRoute = '/shopower/pages/home/home'
    } else if (userType === '5') {
      targetRoute = '/operator/pages/home/home'
    } else if (userType === '9') {
      targetRoute = '/platform/pages/home/home'
    }
    uni.reLaunch({ url: targetRoute })
  }
})
const showPassword = ref(false)
const rememberLogin = ref(false)

const form = reactive({
  username: '',
  password: ''
})

async function handleLogin() {
  if (!form.username) {
    uni.showToast({ title: '请输入账号', icon: 'none' })
    return
  }
  if (!form.password) {
    uni.showToast({ title: '请输入密码', icon: 'none' })
    return
  }
  
  loading.value = true
  try {
    const res = await login({
      username: form.username,
      password: form.password
    })
    
    if (res.code === 0 && res.data) {
      // 根据user_id前缀判断用户类型
      const userId = String(res.data.user_id || res.data.userId || '')
      let userType = '1'
      if (userId && userId.startsWith('9')) {
        userType = '9'
      } else if (userId && userId.startsWith('5')) {
        userType = '5'
      }
      
      // 根据"保持登录"选项决定存储方式
      if (rememberLogin.value) {
        // 勾选：使用localStorage永久保存
        uni.setStorageSync('token', res.data.token)
        uni.setStorageSync('userId', userId)
        uni.setStorageSync('userType', userType)
        uni.setStorageSync('rememberLogin', 'true')
      } else {
        // 不勾选：使用sessionStorage，关闭浏览器后清除
        // #ifdef H5
        sessionStorage.setItem('token', res.data.token)
        sessionStorage.setItem('userId', userId)
        sessionStorage.setItem('userType', userType)
        // #endif
        // #ifndef H5
        uni.setStorageSync('token', res.data.token)
        uni.setStorageSync('userId', userId)
        uni.setStorageSync('userType', userType)
        // #endif
      }
      
      uni.showToast({ title: '登录成功', icon: 'success' })
      
      // 跳转到对应页面
      setTimeout(() => {
        let targetRoute = '/shopower/pages/home/home'
        if (userType === '5') {
          targetRoute = '/operator/pages/home/home'
        } else if (userType === '9') {
          targetRoute = '/platform/pages/home/home'
        }
        uni.reLaunch({ url: targetRoute })
      }, 1000)
    } else {
      uni.showToast({ title: res.message || '登录失败', icon: 'none' })
    }
  } catch (err: any) {
    uni.showToast({ title: err.message || '登录失败', icon: 'none' })
  } finally {
    loading.value = false
  }
}

function goToRegister() {
  uni.navigateTo({ url: '/pages/register/select-role' })
}

function goToForgotPassword() {
  uni.navigateTo({ url: '/pages/forgot-password/forgot-password' })
}
</script>

<style lang="scss">
.login-page {
  min-height: 100vh;
  background-color: #f5f5f5;
}

.header {
  background: linear-gradient(135deg, #ff8c00 0%, #ff6600 100%);
  padding: 120rpx 48rpx 160rpx;
  border-radius: 0 0 0 80rpx;
}

.header-content {
  .greeting {
    display: block;
    font-size: 48rpx;
    font-weight: bold;
    color: #fff;
    margin-bottom: 16rpx;
  }
  
  .title {
    display: block;
    font-size: 40rpx;
    color: #fff;
  }
}

.form-container {
  padding: 0 32rpx;
  margin-top: -80rpx;
}

.form-card {
  background: #fff;
  border-radius: 32rpx;
  padding: 48rpx 40rpx;
  box-shadow: 0 8rpx 32rpx rgba(0, 0, 0, 0.08);
}

.form-item {
  margin-bottom: 40rpx;
  
  .label {
    display: block;
    font-size: 28rpx;
    color: #333;
    margin-bottom: 16rpx;
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

.password-wrapper {
  display: flex;
  align-items: center;
  
  .input {
    flex: 1;
  }
  
  .eye-icon {
    padding: 16rpx;
    font-size: 32rpx;
    color: #999;
  }
}

.options {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 48rpx;
  
  .remember {
    display: flex;
    align-items: center;
    
    .checkbox {
      width: 32rpx;
      height: 32rpx;
      border: 2rpx solid #ddd;
      border-radius: 50%;
      margin-right: 12rpx;
      
      &.checked {
        background: #ff6600;
        border-color: #ff6600;
      }
    }
    
    .text {
      font-size: 24rpx;
      color: #999;
    }
  }
  
  .forgot {
    font-size: 24rpx;
    color: #ff6600;
  }
}

.login-btn {
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

.divider {
  display: flex;
  align-items: center;
  margin: 48rpx 0;
  
  .line {
    flex: 1;
    height: 1rpx;
    background: #eee;
  }
  
  .text {
    padding: 0 24rpx;
    font-size: 24rpx;
    color: #999;
  }
}

.social-login {
  display: flex;
  justify-content: center;
  gap: 40rpx;
  margin-bottom: 48rpx;
  
  .social-icon {
    width: 80rpx;
    height: 80rpx;
    border-radius: 50%;
    display: flex;
    align-items: center;
    justify-content: center;
    font-size: 36rpx;
    font-weight: bold;
    
    &.google {
      background: #f5f5f5;
      color: #ea4335;
    }
    
    &.line-icon {
      background: #00c300;
      color: #fff;
    }
    
    &.facebook {
      background: #1877f2;
      color: #fff;
    }
  }
}

.register-link {
  text-align: center;
  
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
</style>
