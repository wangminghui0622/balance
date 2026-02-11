<template>
  <div class="login-container">
    <el-card class="login-card">
      <template #header>
        <div style="text-align: center">
          <h2>登录</h2>
        </div>
      </template>
      <el-form
        ref="formRef"
        :model="form"
        :rules="rules"
        label-width="0"
        class="centered-form"
        autocomplete="off"
      >
        <el-form-item prop="username">
          <el-input
            v-model="form.username"
            placeholder="请输入用户名"
            autocomplete="off"
            size="large"
          />
        </el-form-item>
        <el-form-item prop="password">
          <el-input
            v-model="form.password"
            :type="showPassword ? 'text' : 'password'"
            placeholder="请输入密码"
            @keyup.enter="handleLogin"
            autocomplete="off"
            size="large"
          >
            <template #suffix>
              <el-icon
                class="password-icon"
                @mousedown="showPassword = true"
                @mouseup="showPassword = false"
                @mouseleave="showPassword = false"
                style="cursor: pointer"
              >
                <View />
              </el-icon>
            </template>
          </el-input>
        </el-form-item>
        <el-form-item>
          <el-button type="primary" @click="handleLogin" :loading="loading" style="width: 100%" size="large">
            登录
          </el-button>
        </el-form-item>
        <el-form-item>
          <div style="text-align: center; width: 100%">
            <el-link type="primary" @click="goToRegister">前往注册</el-link>
          </div>
        </el-form-item>
      </el-form>
    </el-card>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import { View } from '@element-plus/icons-vue'
import { authApi } from '@share/api/auth'
import type { FormInstance, FormRules } from 'element-plus'
import { STORAGE_KEYS, HTTP_STATUS, getRouteByUserType, type UserType } from '@share/constants'

const router = useRouter()
const formRef = ref<FormInstance>()
const loading = ref(false)
const showPassword = ref(false)

const form = reactive({
  username: '',
  password: ''
})

const rules: FormRules = {
  username: [{ required: true, message: '请输入用户名', trigger: 'blur' }],
  password: [
    { required: true, message: '请输入密码', trigger: 'blur' },
    { min: 6, message: '密码至少6个字符', trigger: 'blur' }
  ]
}

const handleLogin = async () => {
  if (!formRef.value) return

  await formRef.value.validate(async (valid) => {
    if (valid) {
      loading.value = true
      try {
        const res = await authApi.login({
          username: form.username,
          password: form.password
        })

        if (res.code === HTTP_STATUS.OK && res.data.token) {
          // 保存token、userId和userType
          localStorage.setItem(STORAGE_KEYS.TOKEN, res.data.token)
          localStorage.setItem(STORAGE_KEYS.USER_ID, res.data.userId.toString())
          localStorage.setItem(STORAGE_KEYS.USER_TYPE, res.data.userType.toString())

          ElMessage.success('登录成功')

          // 根据用户类型路由到不同页面
          const userType = res.data.userType.toString() as UserType
          router.push(getRouteByUserType(userType))
        } else {
          ElMessage.error(res.message || '登录失败')
        }
      } catch (error: any) {
        ElMessage.error(error?.response?.data?.message || error?.message || '登录失败')
      } finally {
        loading.value = false
      }
    }
  })
}

const goToRegister = () => {
  router.push('/register')
}
</script>

<style scoped lang="scss">
.login-container {
  display: flex;
  justify-content: center;
  align-items: center;
  height: 100vh;
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
}

.login-card {
  width: 400px;
}

.centered-form :deep(.el-form-item__label) {
  display: none;
}

.centered-form :deep(.el-form-item__content) {
  width: 100%;
}

.centered-form :deep(.el-form-item) {
  margin-bottom: 20px;
}

.password-icon {
  color: #909399;
  font-size: 16px;
  transition: color 0.3s;

  &:hover {
    color: #409eff;
  }
}
</style>
