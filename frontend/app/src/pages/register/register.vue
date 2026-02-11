<template>
  <view class="register-page">
    <!-- 顶部橙色区域 -->
    <view class="header">
      <view class="back-btn" @click="goBack">
        <text class="arrow">‹</text>
      </view>
    </view>
    
    <!-- 内容区域 -->
    <view class="content">
      <text class="title">创建账号</text>
      <text class="subtitle">{{ roleText }}注册</text>
      
      <!-- 表单 -->
      <view class="form">
        <view class="form-item">
          <text class="label">账号</text>
          <input 
            class="input" 
            type="text" 
            v-model="form.username" 
            placeholder="请输入账号名（6-16个字符）"
            placeholder-class="placeholder"
          />
        </view>
        
        <view class="form-item">
          <text class="label">邮箱</text>
          <input 
            class="input" 
            type="text" 
            v-model="form.email" 
            placeholder="请输入邮箱"
            placeholder-class="placeholder"
          />
        </view>
        
        <view class="form-item">
          <text class="label">验证码</text>
          <view class="code-wrapper">
            <input 
              class="input code-input" 
              type="number" 
              v-model="form.emailCode" 
              placeholder="请输入验证码"
              placeholder-class="placeholder"
              maxlength="6"
            />
            <view 
              class="send-btn" 
              :class="{ disabled: codeCooldown > 0 || !form.email || sendingCode }"
              @click="handleSendCode"
            >
              <text class="send-text">{{ codeCooldown > 0 ? `${codeCooldown}s` : '发送验证码' }}</text>
            </view>
          </view>
        </view>
        
        <view class="form-item">
          <text class="label">密码</text>
          <view class="password-wrapper">
            <input 
              class="input" 
              :type="showPassword ? 'text' : 'password'" 
              v-model="form.password" 
              placeholder="请输入密码（至少6个字符）"
              placeholder-class="placeholder"
            />
            <view class="eye-icon" @click="showPassword = !showPassword">
              <text>{{ showPassword ? '👁' : '👁‍🗨' }}</text>
            </view>
          </view>
        </view>
        
        <view class="form-item">
          <text class="label">确认密码</text>
          <view class="password-wrapper">
            <input 
              class="input" 
              :type="showConfirmPassword ? 'text' : 'password'" 
              v-model="form.confirmPassword" 
              placeholder="请再次输入密码"
              placeholder-class="placeholder"
            />
            <view class="eye-icon" @click="showConfirmPassword = !showConfirmPassword">
              <text>{{ showConfirmPassword ? '👁' : '👁‍🗨' }}</text>
            </view>
          </view>
        </view>
      </view>
      
      <!-- 下一步按钮 -->
      <button class="register-btn" :loading="loading" @click="handleNext">下一步</button>
      
      <!-- 登录链接 -->
      <view class="login-link">
        <text class="text">已有账号?</text>
        <text class="link" @click="goToLogin">去登录></text>
      </view>
    </view>
  </view>
</template>

<script setup lang="ts">
import { ref, reactive, computed } from 'vue'
import { onLoad } from '@dcloudio/uni-app'
import { sendEmailCode } from '@share/api/auth'
import { ROUTE_PATH, USER_TYPE_NUM } from '@share/constants'

const loading = ref(false)
const showPassword = ref(false)
const showConfirmPassword = ref(false)
const userType = ref(USER_TYPE_NUM.SHOPOWNER)

const sendingCode = ref(false)
const codeCooldown = ref(0)
let cooldownTimer: ReturnType<typeof setInterval> | null = null

const form = reactive({
  username: '',
  email: '',
  emailCode: '',
  password: '',
  confirmPassword: ''
})

const roleText = computed(() => {
  switch (userType.value) {
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

onLoad((options: any) => {
  if (options?.userType) {
    userType.value = parseInt(options.userType)
  }
})

async function handleSendCode() {
  if (!form.email) {
    uni.showToast({ title: '请先输入邮箱', icon: 'none' })
    return
  }
  if (!/^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(form.email)) {
    uni.showToast({ title: '邮箱格式不正确', icon: 'none' })
    return
  }
  if (codeCooldown.value > 0 || sendingCode.value) {
    return
  }

  sendingCode.value = true
  try {
    const res = await sendEmailCode({ email: form.email })
    if (res.code === 0) {
      uni.showToast({ title: '验证码已发送', icon: 'success' })
      codeCooldown.value = 60
      cooldownTimer = setInterval(() => {
        codeCooldown.value--
        if (codeCooldown.value <= 0) {
          if (cooldownTimer) {
            clearInterval(cooldownTimer)
            cooldownTimer = null
          }
        }
      }, 1000)
    } else {
      uni.showToast({ title: res.message || '发送失败', icon: 'none' })
    }
  } catch (err: any) {
    uni.showToast({ title: err.message || '发送失败', icon: 'none' })
  } finally {
    sendingCode.value = false
  }
}

function handleNext() {
  if (!form.username) {
    uni.showToast({ title: '请输入账号', icon: 'none' })
    return
  }
  if (form.username.length < 6 || form.username.length > 16) {
    uni.showToast({ title: '账号长度为6-16个字符', icon: 'none' })
    return
  }
  if (/\s/.test(form.username)) {
    uni.showToast({ title: '账号不能包含空格', icon: 'none' })
    return
  }
  if (!form.email) {
    uni.showToast({ title: '请输入邮箱', icon: 'none' })
    return
  }
  if (!/^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(form.email)) {
    uni.showToast({ title: '邮箱格式不正确', icon: 'none' })
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
  if (!form.password) {
    uni.showToast({ title: '请输入密码', icon: 'none' })
    return
  }
  if (form.password.length < 6) {
    uni.showToast({ title: '密码至少6位', icon: 'none' })
    return
  }
  if (form.password !== form.confirmPassword) {
    uni.showToast({ title: '两次密码不一致', icon: 'none' })
    return
  }
  
  // 保存第一步数据到本地存储，跳转到下一步
  const registerData = {
    username: form.username,
    email: form.email,
    emailCode: form.emailCode,
    password: form.password,
    userType: userType.value
  }
  uni.setStorageSync('register_step1_data', JSON.stringify(registerData))
  
  // 跳转到填写个人信息页面
  uni.navigateTo({ url: '/pages/register/register-info' })
}

function goBack() {
  uni.navigateBack()
}

function goToLogin() {
  uni.reLaunch({ url: ROUTE_PATH.LOGIN })
}
</script>

<style lang="scss">
.register-page {
  min-height: 100vh;
  background-color: #f5f5f5;
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
  padding: 48rpx 32rpx;
}

.title {
  display: block;
  font-size: 44rpx;
  font-weight: bold;
  color: #333;
  margin-bottom: 8rpx;
}

.subtitle {
  display: block;
  font-size: 28rpx;
  color: #ff6600;
  margin-bottom: 48rpx;
}

.form {
  background: #fff;
  border-radius: 24rpx;
  padding: 32rpx;
  margin-bottom: 48rpx;
}

.form-item {
  margin-bottom: 32rpx;
  
  &:last-child {
    margin-bottom: 0;
  }
  
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

.code-wrapper {
  display: flex;
  align-items: center;
  gap: 16rpx;
  
  .code-input {
    flex: 1;
  }
  
  .send-btn {
    padding: 0 24rpx;
    height: 64rpx;
    background: linear-gradient(135deg, #ff8c00 0%, #ff6600 100%);
    border-radius: 32rpx;
    display: flex;
    align-items: center;
    justify-content: center;
    flex-shrink: 0;
    
    &.disabled {
      opacity: 0.5;
    }
    
    .send-text {
      font-size: 24rpx;
      color: #fff;
      white-space: nowrap;
    }
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

.register-btn {
  width: 100%;
  height: 88rpx;
  background: linear-gradient(135deg, #ff8c00 0%, #ff6600 100%);
  border-radius: 44rpx;
  color: #fff;
  font-size: 32rpx;
  font-weight: bold;
  border: none;
  margin-bottom: 32rpx;
  
  &::after {
    border: none;
  }
}

.login-link {
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
