<template>
  <view class="container">
    <view class="loading">
      <text>加载中...</text>
    </view>
  </view>
</template>

<script setup lang="ts">
import { onMounted } from 'vue'

onMounted(() => {
  checkLoginStatus()
})

function checkLoginStatus() {
  const token = uni.getStorageSync('token')
  
  if (token) {
    // 已登录，根据用户类型跳转
    const userType = uni.getStorageSync('userType')
    let targetRoute = '/pages/login/login'
    
    if (userType === '1') {
      targetRoute = '/shopower/pages/home/home'
    } else if (userType === '5') {
      targetRoute = '/operator/pages/home/home'
    } else if (userType === '9') {
      targetRoute = '/platform/pages/home/home'
    }
    
    doNavigate(targetRoute)
  } else {
    // 未登录，直接跳转到登录页
    doNavigate('/pages/login/login')
  }
}

function doNavigate(url: string) {
  // 尝试多种跳转方式，确保兼容性
  uni.reLaunch({
    url,
    fail: () => {
      uni.redirectTo({
        url,
        fail: () => {
          // 最后降级：直接修改location
          const base = '/balance/app/'
          window.location.hash = '#' + url
        }
      })
    }
  })
}
</script>

<style lang="scss">
.container {
  display: flex;
  justify-content: center;
  align-items: center;
  height: 100vh;
  background-color: #f5f5f5;
}

.loading {
  text-align: center;
  color: #999;
}
</style>
