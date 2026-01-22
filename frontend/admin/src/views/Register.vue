<template>
  <div class="register-container">
    <el-card class="register-card">
      <template #header>
        <div style="text-align: center">
          <h2>注册账号</h2>
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
        <el-form-item prop="userType">
          <el-select v-model="form.userType" placeholder="请选择用户类型" style="width: 100%" size="large">
            <el-option label="店铺" :value="USER_TYPE_NUM.SHOPOWNER" />
            <el-option label="运营" :value="USER_TYPE_NUM.OPERATOR" />
          </el-select>
        </el-form-item>
        <el-form-item prop="username">
          <el-input
            v-model="form.username"
            placeholder="请输入用户名（6-16个字符，不能包含空格）"
            autocomplete="off"
            size="large"
            @keydown.space.prevent
          />
        </el-form-item>
        <el-form-item prop="email">
          <el-input
            v-model="form.email"
            placeholder="请输入邮箱"
            autocomplete="off"
            size="large"
          />
        </el-form-item>
        <el-form-item prop="password">
          <el-input
            v-model="form.password"
            :type="showPassword ? 'text' : 'password'"
            placeholder="请输入密码（至少6个字符）"
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
        <div class="password-hint">密码至少6个字符</div>
        <el-form-item prop="confirmPassword">
          <el-input
            v-model="form.confirmPassword"
            :type="showConfirmPassword ? 'text' : 'password'"
            placeholder="请再次输入密码"
            autocomplete="off"
            size="large"
          >
            <template #suffix>
              <el-icon
                class="password-icon"
                @mousedown="showConfirmPassword = true"
                @mouseup="showConfirmPassword = false"
                @mouseleave="showConfirmPassword = false"
                style="cursor: pointer"
              >
                <View />
              </el-icon>
            </template>
          </el-input>
        </el-form-item>
        <el-form-item>
          <el-button type="primary" @click="handleRegister" :loading="loading" style="width: 100%" size="large">
            注册
          </el-button>
        </el-form-item>
        <el-form-item>
          <div style="text-align: center; width: 100%">
            <el-link type="primary" @click="goToLogin">已有账号？前往登录</el-link>
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
import { HTTP_STATUS, ROUTE_PATH, USER_TYPE_NUM } from '@share/constants'

const router = useRouter()
const formRef = ref<FormInstance>()
const loading = ref(false)
const showPassword = ref(false)
const showConfirmPassword = ref(false)

const form = reactive({
  userType: USER_TYPE_NUM.SHOPOWNER, // 默认选择店铺
  username: '',
  email: '',
  password: '',
  confirmPassword: ''
})

const validateConfirmPassword = (_rule: any, value: string, callback: any) => {
  if (value !== form.password) {
    callback(new Error('两次输入的密码不一致'))
  } else {
    callback()
  }
}

const validateEmail = (_rule: any, value: string, callback: any) => {
  if (value && !/^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(value)) {
    callback(new Error('邮箱格式不正确'))
  } else {
    callback()
  }
}

const validateUsername = (_rule: any, value: string, callback: any) => {
  if (!value) {
    callback()
    return
  }
  // 检查是否包含空格、空字符或tab
  if (/\s/.test(value)) {
    callback(new Error('用户名不能包含空格、空字符或Tab'))
  } else {
    callback()
  }
}

const rules: FormRules = {
  userType: [{ required: true, message: '请选择用户类型', trigger: 'change' }],
  username: [
    { required: true, message: '请输入用户名', trigger: 'blur' },
    { min: 6, max: 16, message: '用户名长度为6-16个字符', trigger: 'blur' },
    { validator: validateUsername, trigger: 'blur' }
  ],
  email: [
    { required: true, message: '请输入邮箱', trigger: 'blur' },
    { validator: validateEmail, trigger: 'blur' }
  ],
  password: [
    { required: true, message: '请输入密码', trigger: 'blur' },
    { min: 6, message: '密码至少6个字符', trigger: 'blur' }
  ],
  confirmPassword: [
    { required: true, message: '请确认密码', trigger: 'blur' },
    { validator: validateConfirmPassword, trigger: 'blur' }
  ]
}

const handleRegister = async () => {
  if (!formRef.value) return

  await formRef.value.validate(async (valid) => {
    if (valid) {
      loading.value = true
      try {
        const res = await authApi.register({
          username: form.username,
          email: form.email,
          password: form.password,
          userType: form.userType
        })

        if (res.code === HTTP_STATUS.OK) {
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
  })
}

const goToLogin = () => {
  router.push(ROUTE_PATH.LOGIN)
}
</script>

<style scoped lang="scss">
.register-container {
  display: flex;
  justify-content: center;
  align-items: center;
  height: 100vh;
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
}

.register-card {
  width: 450px;
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

.password-hint {
  margin: -10px 0 10px;
  font-size: 12px;
  color: #909399;
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
