<template>
  <div class="shopee-auth-container">
    <el-card class="auth-card">
      <template #header>
        <div class="card-header">
          <span>Shopee 授权</span>
        </div>
      </template>

      <el-form label-width="120px">
        <el-form-item>
          <el-button
            type="primary"
            :loading="loading"
            @click="handleAuth"
          >
            {{ loading ? '获取授权链接中...' : '开始授权' }}
          </el-button>
        </el-form-item>
      </el-form>

      <el-alert
        v-if="authURL"
        type="success"
        :closable="false"
        style="margin-top: 20px"
      >
        <template #title>
          <div>
            <p>授权链接已生成，请点击下方按钮跳转到 Shopee 进行授权：</p>
            <el-button
              type="primary"
              @click="openAuthURL"
              style="margin-top: 10px"
            >
              跳转到 Shopee 授权页面
            </el-button>
          </div>
        </template>
      </el-alert>

      <el-alert
        v-if="error"
        type="error"
        :closable="false"
        style="margin-top: 20px"
        :title="error"
      />
    </el-card>

    <el-card class="info-card" style="margin-top: 20px">
      <template #header>
        <div class="card-header">
          <span>使用说明</span>
        </div>
      </template>
      <ol>
        <li>固定参数已配置（partnerID, partnerKey, redirect, isSandbox）</li>
        <li>点击"开始授权"按钮获取授权链接</li>
        <li>跳转到 Shopee 授权页面完成授权</li>
        <li>授权成功后，access_token 和 refresh_token 会自动保存到数据库</li>
      </ol>
    </el-card>
  </div>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import { ElMessage } from 'element-plus'
import { shopeeApi } from '@share/api/shopee'
import { HTTP_STATUS } from '@share/constants'

const loading = ref(false)
const authURL = ref('')
const error = ref('')

const handleAuth = async () => {
  loading.value = true
  error.value = ''
  authURL.value = ''

  try {
    const res = await shopeeApi.getAuthURL()
    
    if (res.code === HTTP_STATUS.OK && res.auth_url) {
      authURL.value = res.auth_url
      ElMessage.success('授权链接生成成功')
    } else {
      error.value = res.message || '获取授权链接失败'
      ElMessage.error(error.value)
    }
  } catch (err: any) {
    error.value = err?.response?.data?.message || err?.message || '获取授权链接失败'
    ElMessage.error(error.value)
  } finally {
    loading.value = false
  }
}

const openAuthURL = () => {
  if (authURL.value) {
    window.open(authURL.value, '_blank')
  }
}
</script>

<style scoped>
.shopee-auth-container {
  padding: 20px;
  max-width: 800px;
  margin: 0 auto;
}

.auth-card,
.info-card {
  width: 100%;
}

.card-header {
  font-size: 18px;
  font-weight: bold;
}

.form-tip {
  font-size: 12px;
  color: #909399;
  margin-top: 5px;
}

.info-card ol {
  line-height: 2;
  padding-left: 20px;
}

.info-card li {
  margin-bottom: 10px;
}
</style>
