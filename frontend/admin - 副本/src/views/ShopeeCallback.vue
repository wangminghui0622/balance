<template>
  <div class="callback-container">
    <el-card class="callback-card">
      <template #header>
        <div class="card-header">
          <span>Shopee 授权结果</span>
        </div>
      </template>

      <div v-if="loading" class="loading-container">
        <el-icon class="is-loading"><Loading /></el-icon>
        <p>正在处理授权结果...</p>
      </div>

      <div v-else-if="success" class="success-container">
        <el-result icon="success" title="授权成功！">
          <template #sub-title>
            <div class="result-content">
              <p><strong>Shop ID:</strong> {{ resultData?.shop_id }}</p>
              <p><strong>Access Token:</strong> {{ resultData?.access_token }}</p>
              <p><strong>Refresh Token:</strong> {{ resultData?.refresh_token }}</p>
              <p><strong>过期时间:</strong> {{ resultData?.expire_at }}</p>
              <p style="margin-top: 20px; color: #909399;">
                Token 已自动保存到数据库，现在可以使用 Shopee API 了。
              </p>
            </div>
          </template>
          <template #extra>
            <el-button type="primary" @click="goToHome" @mouseover="logButtonHover">返回首页</el-button>
          </template>
        </el-result>
      </div>

      <div v-else-if="error" class="error-container">
        <el-result icon="error" title="授权失败">
          <template #sub-title>
            <p>{{ error }}</p>
          </template>
          <template #extra>
            <el-button type="primary" @click="goToAuth">重新授权</el-button>
            <el-button @click="goToHome" @mouseover="logButtonHover">返回首页</el-button>
          </template>
        </el-result>
      </div>
    </el-card>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import { ElMessage } from 'element-plus'
import { Loading } from '@element-plus/icons-vue'
import request from '../../share/utils/request'
import { STORAGE_KEYS } from '../../share/constants'

console.log('ShopeeCallback 组件开始加载');

const router = useRouter()
const route = useRoute()

const loading = ref(true)
const success = ref(false)
const error = ref('')
const resultData = ref<{
  shop_id: number
  access_token: string
  refresh_token: string
  expire_in: number
  expire_at: string
} | null>(null)

onMounted(async () => {
  console.log('onMounted 执行，路由参数:', route.query);
  const successParam = route.query.success as string
  const shopId = route.query.shop_id as string
  const errorParam = route.query.error as string

  loading.value = false
  console.log('loading 设置为 false');

  if (successParam === 'true' && shopId) {
    console.log('授权成功，设置 success 为 true');
    // 授权成功（后端重定向过来的）
    success.value = true
    resultData.value = {
      shop_id: parseInt(shopId),
      access_token: '已保存到数据库',
      refresh_token: '已保存到数据库',
      expire_in: 0,
      expire_at: '请查看数据库或后端日志'
    }
    ElMessage.success('授权成功，Token 已保存到数据库')
    
    // 自动触发绑定操作
    await autoBindShop();
  } else {
    console.log('授权失败或参数错误，设置 error:', errorParam);
    // 授权失败或参数错误
    error.value = errorParam || '授权失败或参数错误'
    if (error.value) {
      ElMessage.error(error.value)
    }
  }
  console.log('onMounted 执行完成，当前状态 - success:', success.value, 'error:', error.value);
})

const goToAuth = () => {
  console.log('goToAuth 被调用');
  router.push('/shopee/auth')
}

const logButtonHover = () => {
  console.log('鼠标悬停在返回首页按钮上');
}

// 自动绑定店铺
const autoBindShop = async () => {
  console.log('autoBindShop 函数开始执行');
  // 获取当前用户的token
  const token = localStorage.getItem(STORAGE_KEYS.TOKEN) || sessionStorage.getItem(STORAGE_KEYS.TOKEN);
  console.log('token:**************', token);
  console.log('当前 resultData:', resultData.value);
  console.log('当前 route.query:', route.query);
  
  if (token) {
    console.log('检测到token，准备调用后端bind接口');
    try {
      // 调用后端的bind接口，将shop_id发送到服务器（token通过Authorization header传递）
      console.log('发送请求到后端: /api/v1/balance/admin/shopower/shops/bind');
      const response = await request.post('/api/v1/balance/admin/shopower/shops/bind', {
        shop_id: resultData.value?.shop_id || parseInt(route.query.shop_id as string || '0')
      });
      console.log('后端响应:', response);
      
      // 由于request拦截器返回response.data，我们需要断言类型
      const responseData: any = response;
      
      // 检查响应状态码（后端成功返回 code=0）
      if (responseData.code === 0) {
        console.log('自动绑定成功:', responseData);
        ElMessage.success('店铺绑定成功！');
        
        // 绑定成功后跳转到首页，并传递刷新参数
        console.log('跳转到首页');
        setTimeout(() => {
          router.push({
            path: '/',
            query: { refresh: 'true' }
          });
        }, 1500); // 稍微延迟一下，让用户看到成功消息
      } else {
        console.error('自动绑定失败，响应代码:', responseData.code, '消息:', responseData.message);
        ElMessage.error(responseData.message || '店铺绑定失败，请稍后重试');
        // 即使绑定失败也跳转到首页
        setTimeout(() => {
          router.push({
            path: '/',
            query: { refresh: 'true' }
          });
        }, 1500);
      }
    } catch (bindError) {
      console.error('自动绑定失败:', bindError);
      ElMessage.error('店铺绑定请求失败，请稍后重试');
      // 即使绑定失败也跳转到首页，并传递刷新参数
      setTimeout(() => {
        router.push({
          path: '/',
          query: { refresh: 'true' }
        });
      }, 1500);
    }
  } else {
    console.warn('未找到用户token，无法自动绑定');
    // 如果没有token，则直接跳转到首页
    console.log('无token，直接跳转到首页');
    setTimeout(() => {
      router.push({
        path: '/',
        query: { refresh: 'true' }
      });
    }, 1500);
  }
}

const goToHome = async () => {
  console.log('goToHome 函数开始执行（通过按钮点击）');
  // 调用相同的绑定逻辑
  await autoBindShop();
}
</script>

<style scoped>
.callback-container {
  padding: 20px;
  max-width: 800px;
  margin: 0 auto;
}

.callback-card {
  width: 100%;
}

.card-header {
  font-size: 18px;
  font-weight: bold;
}

.loading-container {
  text-align: center;
  padding: 40px;
}

.loading-container .el-icon {
  font-size: 48px;
  color: #409eff;
}

.success-container,
.error-container {
  padding: 20px;
}

.result-content {
  text-align: left;
  max-width: 500px;
  margin: 0 auto;
}

.result-content p {
  margin: 10px 0;
  word-break: break-all;
}
</style>
