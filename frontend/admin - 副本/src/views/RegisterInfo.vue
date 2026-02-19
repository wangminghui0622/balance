<template>
  <div class="register-info-container">
    <el-card class="register-info-card">
      <template #header>
        <div class="header">
          <el-button link @click="goBack" class="back-btn">
            <el-icon><ArrowLeft /></el-icon>
          </el-button>
          <h2>注册{{ roleText }}</h2>
        </div>
      </template>
      
      <!-- 真实姓名 -->
      <div class="section">
        <div class="section-title">真实姓名</div>
        <el-input
          v-model="form.realName"
          placeholder="请输入您的称呼"
          size="large"
        />
      </div>
      
      <!-- 联系方式 -->
      <div class="section">
        <div class="section-title">联系方式</div>
        <div class="section-desc">请至少输入一种联系方式，以方便我们平台工作人员联系您，为您提供更多服务和支持。</div>
        
        <div class="form-item">
          <div class="label">选填：请输入您的电话</div>
          <el-input
            v-model="form.phone"
            placeholder="请输入电话号码"
            size="large"
          />
        </div>
        
        <div class="form-item">
          <div class="label">选填：请输入您的LINE ID</div>
          <el-input
            v-model="form.lineId"
            placeholder="请输入LINE ID"
            size="large"
          />
        </div>
        
        <div class="form-item">
          <div class="label">选填：请输入您的微信账号</div>
          <el-input
            v-model="form.wechat"
            placeholder="请输入微信账号"
            size="large"
          />
        </div>
      </div>
      
      <!-- 提交按钮 -->
      <el-button 
        type="primary" 
        @click="handleRegister" 
        :loading="loading" 
        style="width: 100%; margin-top: 20px" 
        size="large"
      >
        确定并注册
      </el-button>
    </el-card>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, computed, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import { ArrowLeft } from '@element-plus/icons-vue'
import { authApi } from '@share/api/auth'
import { HTTP_STATUS, ROUTE_PATH, USER_TYPE_NUM } from '@share/constants'

const router = useRouter()
const loading = ref(false)

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
    default:
      return ''
  }
})

onMounted(() => {
  // 读取第一步数据
  const data = sessionStorage.getItem('register_step1_data')
  if (data) {
    try {
      step1Data.value = JSON.parse(data)
    } catch (e) {
      ElMessage.error('数据异常，请重新填写')
      router.push(ROUTE_PATH.REGISTER)
    }
  } else {
    ElMessage.warning('请先填写账号信息')
    router.push(ROUTE_PATH.REGISTER)
  }
})

const handleRegister = async () => {
  if (!step1Data.value) {
    ElMessage.error('数据异常，请重新填写')
    return
  }
  
  // 至少填写一种联系方式
  if (!form.phone && !form.lineId && !form.wechat) {
    ElMessage.warning('请至少填写一种联系方式')
    return
  }
  
  loading.value = true
  try {
    const res = await authApi.register({
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
    
    if (res.code === HTTP_STATUS.OK) {
      // 清除临时数据
      sessionStorage.removeItem('register_step1_data')
      
      ElMessage.success('注册成功，请登录')
      router.push(ROUTE_PATH.LOGIN)
    } else {
      ElMessage.error(res.message || '注册失败')
    }
  } catch (error: any) {
    ElMessage.error(error?.response?.data?.message || error?.message || '注册失败')
  } finally {
    loading.value = false
  }
}

const goBack = () => {
  router.back()
}
</script>

<style scoped lang="scss">
.register-info-container {
  display: flex;
  justify-content: center;
  align-items: center;
  min-height: 100vh;
  padding: 20px;
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
}

.register-info-card {
  width: 500px;
  max-width: 100%;
}

.header {
  display: flex;
  align-items: center;
  position: relative;
  
  h2 {
    flex: 1;
    text-align: center;
    margin: 0;
  }
  
  .back-btn {
    position: absolute;
    left: 0;
    font-size: 20px;
  }
}

.section {
  margin-bottom: 24px;
  
  .section-title {
    font-size: 16px;
    font-weight: bold;
    color: #333;
    margin-bottom: 12px;
  }
  
  .section-desc {
    font-size: 13px;
    color: #ff6600;
    margin-bottom: 16px;
    line-height: 1.5;
  }
}

.form-item {
  margin-bottom: 16px;
  
  &:last-child {
    margin-bottom: 0;
  }
  
  .label {
    font-size: 13px;
    color: #666;
    margin-bottom: 8px;
  }
}
</style>
