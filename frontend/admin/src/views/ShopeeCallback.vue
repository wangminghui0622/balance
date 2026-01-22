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
            <el-button type="primary" @click="goToHome">返回首页</el-button>
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
            <el-button @click="goToHome">返回首页</el-button>
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
  const successParam = route.query.success as string
  const shopId = route.query.shop_id as string
  const errorParam = route.query.error as string

  loading.value = false

  if (successParam === 'true' && shopId) {
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
  } else {
    // 授权失败或参数错误
    error.value = errorParam || '授权失败或参数错误'
    if (error.value) {
      ElMessage.error(error.value)
    }
  }
})

const goToAuth = () => {
  router.push('/shopee/auth')
}

const goToHome = () => {
  router.push('/')
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
