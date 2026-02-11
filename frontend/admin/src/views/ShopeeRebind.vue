<template>
  <div class="rebind-container">
    <el-card class="rebind-card">
      <template #header>
        <div class="card-header">
          <span>店铺换绑确认</span>
        </div>
      </template>

      <div class="rebind-content">
        <!-- 账号对比显示 -->
        <div class="account-comparison">
          <div class="account-item old-account">
            <div class="account-label">已绑定账号</div>
            <div class="account-value">{{ boundAdminID || boundUserName }}</div>
            <div class="account-email" v-if="boundEmail">{{ boundEmail }}</div>
            <div class="account-phone" v-if="boundPhone">{{ boundPhone }}</div>
          </div>
          
          <div class="arrow">
            <el-icon :size="32"><ArrowRight /></el-icon>
          </div>
          
          <div class="account-item new-account">
            <div class="account-label">新账号</div>
            <div class="account-value">{{ currentUserID || currentUserName }}</div>
            <div class="account-email">{{ currentUserEmail || '--' }}</div>
            <div class="account-phone">{{ currentUserPhone || '--' }}</div>
          </div>
        </div>

        <!-- 验证码输入区域 -->
        <div class="verification-section">
          <div class="verification-content">
            <div class="code-label">验证码</div>
            <div class="code-input-group">
              <el-input
                v-model="form.code"
                placeholder="请输入验证码"
                :maxlength="6"
                style="width: 200px; margin-right: 12px;"
                :disabled="!codeSent"
              />
              <el-button
                type="primary"
                :loading="sendingCode"
                :disabled="codeSent && countdown > 0"
                @click="handleSendCode"
              >
                {{ codeSent && countdown > 0 ? `重新发送(${countdown}s)` : '发送验证码' }}
              </el-button>
            </div>
            <div class="code-tip" v-if="codeSent">
              <el-text type="info" size="small">验证码已发送到邮箱：{{ boundEmail }}</el-text>
            </div>
          </div>
        </div>

        <!-- 操作按钮 -->
        <div class="action-buttons">
          <el-button
            type="primary"
            :loading="confirming"
            :disabled="!form.code || form.code.length !== 6"
            @click="handleConfirm"
          >
            确认
          </el-button>
          <el-button @click="handleCancel">取消</el-button>
        </div>
      </div>
    </el-card>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted, onUnmounted } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import { ElMessage } from 'element-plus'
import { ArrowRight } from '@element-plus/icons-vue'
import { shopeeApi } from '../../share/api/shopee'
import { authApi } from '../../share/api/auth'
import { STORAGE_KEYS, HTTP_STATUS } from '../../share/constants'

const router = useRouter()
const route = useRoute()

// URL 参数
const shopID = ref<number>(0)
const boundAdminID = ref<number>(0)
const boundUserName = ref<string>('')
const boundEmail = ref<string>('')

// 当前登录用户信息
const currentUserID = ref<string>('')
const currentUserName = ref<string>('')
const currentUserEmail = ref<string>('')
const currentUserPhone = ref<string>('')

// 已绑定账号信息
const boundPhone = ref<string>('')

// 表单数据
const form = reactive({
  code: ''
})

// 状态
const codeSent = ref(false)
const sendingCode = ref(false)
const confirming = ref(false)
const countdown = ref(0)

// 倒计时定时器
let countdownTimer: ReturnType<typeof setInterval> | null = null

onMounted(async () => {
  // 从 URL 参数获取信息
  shopID.value = parseInt(route.query.shop_id as string || '0')
  boundAdminID.value = parseInt(route.query.bound_admin_id as string || '0')
  boundUserName.value = decodeURIComponent(route.query.bound_user_name as string || '')
  boundEmail.value = decodeURIComponent(route.query.bound_email as string || '')

  // 获取当前登录用户信息
  const userId = localStorage.getItem(STORAGE_KEYS.USER_ID)
  if (userId) {
    currentUserID.value = userId
    // 获取用户详细信息（包括邮箱和手机号）
    try {
      const userRes = await authApi.getCurrentUser()
      if (userRes.code === HTTP_STATUS.OK && userRes.data) {
        currentUserName.value = userRes.data.userName
        currentUserEmail.value = userRes.data.email
        currentUserPhone.value = userRes.data.phone
      }
    } catch (error: any) {
      console.error('获取用户信息失败:', error)
      // 如果获取失败，只显示用户ID
    }
    
    // 获取已绑定账号的详细信息（包括手机号）
    if (boundAdminID.value > 0) {
      try {
        // 这里可以通过后端API获取已绑定账号的详细信息
        // 暂时先不实现，因为需要额外的API接口
      } catch (error: any) {
        console.error('获取已绑定账号信息失败:', error)
      }
    }
  } else {
    // 如果没有登录，提示用户需要登录
    ElMessage.warning('请先登录后再进行换绑操作')
    setTimeout(() => {
      router.push('/login')
    }, 2000)
    return
  }

  // 检查必要参数
  if (!shopID.value || !boundAdminID.value) {
    ElMessage.error('缺少必要参数')
    router.push('/')
    return
  }
})

// 发送验证码
const handleSendCode = async () => {
  if (!shopID.value) {
    ElMessage.error('店铺ID无效')
    return
  }

  sendingCode.value = true
  try {
    const res = await shopeeApi.sendRebindCode(shopID.value)

    if (res.code === HTTP_STATUS.OK) {
      codeSent.value = true
      ElMessage.success('验证码已发送到邮箱')
      
      // 开始倒计时
      countdown.value = 60
      countdownTimer = setInterval(() => {
        countdown.value--
        if (countdown.value <= 0) {
          if (countdownTimer) {
            clearInterval(countdownTimer)
            countdownTimer = null
          }
        }
      }, 1000)
    } else {
      ElMessage.error(res.message || '发送验证码失败')
    }
  } catch (error: any) {
    ElMessage.error(error?.response?.data?.message || error?.message || '发送验证码失败')
  } finally {
    sendingCode.value = false
  }
}

// 确认换绑
const handleConfirm = async () => {
  if (!form.code || form.code.length !== 6) {
    ElMessage.warning('请输入6位验证码')
    return
  }

  if (!currentUserID.value) {
    ElMessage.error('未获取到当前用户信息')
    return
  }

  confirming.value = true
  try {
    // 先验证验证码
    const verifyRes = await shopeeApi.verifyRebindCode(shopID.value, form.code)

    if (verifyRes.code !== HTTP_STATUS.OK) {
      ElMessage.error(verifyRes.message || '验证码验证失败')
      return
    }

    // 验证通过，执行换绑
    const confirmRes = await shopeeApi.confirmRebind(shopID.value, form.code, parseInt(currentUserID.value))

    if (confirmRes.code === HTTP_STATUS.OK) {
      ElMessage.success('换绑成功')
      setTimeout(() => {
        router.push('/')
      }, 1500)
    } else {
      ElMessage.error(confirmRes.message || '换绑失败')
    }
  } catch (error: any) {
    ElMessage.error(error?.response?.data?.message || error?.message || '换绑失败')
  } finally {
    confirming.value = false
  }
}

// 取消换绑
const handleCancel = async () => {
  if (!shopID.value) {
    router.push('/')
    return
  }

  try {
    const res = await shopeeApi.cancelRebind(shopID.value)

    if (res.code === HTTP_STATUS.OK) {
      ElMessage.success('已取消换绑，token 已更新')
    } else {
      ElMessage.warning(res.message || '取消换绑失败')
    }
  } catch (error: any) {
    ElMessage.warning(error?.response?.data?.message || error?.message || '取消换绑失败')
  } finally {
    setTimeout(() => {
      router.push('/')
    }, 1500)
  }
}

// 清理定时器
onUnmounted(() => {
  if (countdownTimer) {
    clearInterval(countdownTimer)
    countdownTimer = null
  }
})
</script>

<style scoped lang="scss">
.rebind-container {
  padding: 40px 20px;
  max-width: 900px;
  margin: 0 auto;
  min-height: calc(100vh - 200px);
  display: flex;
  align-items: center;
  justify-content: center;
}

.rebind-card {
  width: 100%;
}

.card-header {
  font-size: 20px;
  font-weight: bold;
  text-align: center;
}

.rebind-content {
  padding: 40px 20px;
}

.account-comparison {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 40px;
  margin-bottom: 40px;
  padding: 30px;
  background: #f5f7fa;
  border-radius: 8px;
}

.account-item {
  flex: 1;
  text-align: center;
  padding: 20px;
  background: white;
  border-radius: 8px;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.1);
}

.account-label {
  font-size: 14px;
  color: #909399;
  margin-bottom: 12px;
}

.account-value {
  font-size: 24px;
  font-weight: bold;
  color: #303133;
  margin-bottom: 8px;
}

.account-email {
  font-size: 12px;
  color: #909399;
  margin-bottom: 4px;
}

.account-phone {
  font-size: 12px;
  color: #909399;
}

.old-account {
  .account-value {
    color: #f56c6c;
  }
}

.new-account {
  .account-value {
    color: #67c23a;
  }
}

.arrow {
  color: #409eff;
  flex-shrink: 0;
}

.verification-section {
  margin-bottom: 30px;
  padding: 20px;
  background: #fafafa;
  border-radius: 8px;
}

.verification-content {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 12px;
}

.code-label {
  font-size: 14px;
  color: #606266;
  margin-bottom: 4px;
}

.code-input-group {
  display: flex;
  align-items: center;
  justify-content: center;
}

.code-tip {
  margin-top: 4px;
  text-align: center;
}

.action-buttons {
  display: flex;
  justify-content: center;
  gap: 16px;
  margin-top: 30px;
}
</style>
