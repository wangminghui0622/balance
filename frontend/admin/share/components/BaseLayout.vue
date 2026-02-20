<template>
  <el-container class="base-layout">
    <!-- 侧边栏 -->
    <el-aside :width="isCollapse ? '64px' : '160px'" class="sidebar">
      <div class="logo">
        <span v-if="!isCollapse">XShopee V2.1</span>
        <span v-else>XS</span>
      </div>
      <el-menu
        :default-active="activeMenu"
        :collapse="isCollapse"
        :unique-opened="true"
        router
        class="sidebar-menu"
      >
        <template v-for="item in menuItems" :key="item.path">
          <el-menu-item v-if="!item.children" :index="item.path">
            <el-icon v-if="item.icon">
              <component :is="item.icon" />
            </el-icon>
            <span>{{ item.label }}</span>
          </el-menu-item>
          <el-sub-menu v-else :index="item.path">
            <template #title>
              <el-icon v-if="item.icon">
                <component :is="item.icon" />
              </el-icon>
              <span>{{ item.label }}</span>
            </template>
            <el-menu-item
              v-for="child in item.children"
              :key="child.path"
              :index="child.path"
            >
              {{ child.label }}
            </el-menu-item>
          </el-sub-menu>
        </template>
      </el-menu>
      <div class="user-info">
        <div class="user-info-content">
          <div class="user-info-left">
            <template v-if="(userId ?? '') !== ''">
              <div class="user-name">账号: {{ userName }}</div>
              <div class="user-account user-id" :title="String(userId)">id: {{ userId }}</div>
            </template>
            <template v-else>
              <div class="user-name">{{ userName }}</div>
              <div class="user-account">账户: {{ userAccount }}</div>
            </template>
          </div>
          <div class="user-info-right">
            <el-popover
              placement="top"
              :width="110"
              trigger="click"
              popper-class="logout-popover"
            >
              <template #reference>
                <div class="logout-arrow">
                  <el-icon>
                    <ArrowUp />
                  </el-icon>
                </div>
              </template>
              <div class="logout-menu" style="width: 110px;">
                <div class="logout-item" @click="handleLogout">
                  <el-icon><SwitchButton /></el-icon>
                  <span>退出登录</span>
                </div>
              </div>
            </el-popover>
          </div>
        </div>
      </div>
    </el-aside>

    <!-- 主内容区 -->
    <el-container>
      <!-- 顶部导航栏 -->
      <el-header class="header">
        <div class="header-left">
          <el-icon class="collapse-icon" @click="toggleCollapse">
            <component :is="isCollapse ? 'Expand' : 'Fold'" />
          </el-icon>
        </div>
        <div class="header-right">
          <!-- 只在生产环境显示API环境切换开关 -->
          <div v-if="isProduction" class="api-env-switch">
            <span class="api-env-label">本地</span>
            <el-switch
              v-model="apiEnv"
              @change="handleApiEnvChange"
            />
            <span class="api-env-label">线上</span>
          </div>
          <el-icon class="header-icon">
            <Search />
          </el-icon>
          <el-badge :value="messageCount" class="header-icon chat-badge">
            <el-icon>
              <ChatDotRound />
            </el-icon>
          </el-badge>
          <el-icon class="header-icon">
            <Setting />
          </el-icon>
        </div>
      </el-header>

      <!-- 内容区域 -->
      <el-main class="main-content">
        <slot />
      </el-main>
    </el-container>
  </el-container>
</template>

<script setup lang="ts">
import { ref, computed } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { ElMessageBox, ElSwitch } from 'element-plus'
import { Search, ArrowUp, SwitchButton, ChatDotRound, Setting } from '@element-plus/icons-vue'
import type { MenuItem } from '@share/types'
import { getApiEnv, setApiEnv } from '@share/config/api'
import { STORAGE_KEYS, ROUTE_PATH } from '@share/constants'

interface Props {
  menuItems: MenuItem[]
  userName?: string
  userAccount?: string
  userId?: string | number
  messageCount?: number
}

withDefaults(defineProps<Props>(), {
  userName: 'Hector',
  userAccount: '1234567890',
  userId: undefined,
  messageCount: 4
})

const route = useRoute()
const router = useRouter()
const isCollapse = ref(false)

// 只在生产环境显示API环境切换开关
const isProduction = import.meta.env.MODE === 'production'
const apiEnv = ref<boolean>(getApiEnv() === 'production')

const activeMenu = computed(() => route.path)

const toggleCollapse = () => {
  isCollapse.value = !isCollapse.value
}

const handleApiEnvChange = (value: string | number | boolean) => {
  const boolValue = value === true || value === 'true' || value === 1
  const env = boolValue ? 'production' : 'local'
  ElMessageBox.confirm(
    `确定要切换到${boolValue ? '线上' : '本地'}环境吗？页面将自动刷新。`,
    '切换API环境',
    {
      confirmButtonText: '确定',
      cancelButtonText: '取消',
      type: 'warning'
    }
  ).then(() => {
    setApiEnv(env)
  }).catch(() => {
    // 取消时恢复原状态
    apiEnv.value = getApiEnv() === 'production'
  })
}

const handleLogout = async () => {
  try {
    await ElMessageBox.confirm('确定要退出登录吗？', '提示', {
      confirmButtonText: '确定',
      cancelButtonText: '取消',
      type: 'warning'
    })
    
    // 清除本地存储
    localStorage.removeItem(STORAGE_KEYS.TOKEN)
    localStorage.removeItem(STORAGE_KEYS.USER_ID)
    
    // 跳转到登录页
    router.push(ROUTE_PATH.LOGIN)
  } catch {
    // 用户取消，不做任何操作
  }
}
</script>

<style scoped lang="scss">
.base-layout {
  height: 100vh;
}

.sidebar {
  background-color: #fff;
  border-right: 1px solid #e4e7ed;
  transition: width 0.3s;
  display: flex;
  flex-direction: column;
}

.logo {
  height: 60px;
  line-height: 60px;
  text-align: center;
  font-size: 18px;
  font-weight: bold;
  color: #ff6600;
  border-bottom: 1px solid #e4e7ed;
}

.sidebar-menu {
  flex: 1;
  border: none;
}

:deep(.sidebar-menu .el-menu-item.is-active) {
  color: #ff6a3a;
  background-color: rgba(255, 106, 58, 0.08);
}

:deep(.sidebar-menu .el-menu-item:not(.is-active):hover) {
  background-color: transparent !important;
}

:deep(.sidebar-menu .el-menu-item.is-active .el-icon) {
  color: #ff6a3a;
}

:deep(.sidebar-menu .el-sub-menu.is-active > .el-sub-menu__title) {
  color: #ff6a3a;
}

:deep(.sidebar-menu .el-sub-menu.is-active > .el-sub-menu__title .el-icon) {
  color: #ff6a3a;
}

 :deep(.sidebar-menu .el-sub-menu__title:hover) {
   background-color: transparent !important;
 }

.user-info {
  padding: 16px;
  border-top: 1px solid #e4e7ed;
  background-color: #f5f7fa;
}

.user-info-content {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.user-info-left {
  flex: 1;
}

.user-name {
  font-weight: 500;
  margin-bottom: 4px;
}

.user-account {
  font-size: 12px;
  color: #909399;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.user-info-right {
  display: flex;
  align-items: center;
}

.logout-arrow {
  width: 32px;
  height: 32px;
  display: flex;
  justify-content: center;
  align-items: center;
  border-radius: 50%;
  background-color: #f5f7fa;
  cursor: pointer;
  transition: all 0.3s;
  color: #606266;

  &:hover {
    background-color: #ff6600;
    color: #fff;
    transform: translateY(-2px);
  }

  .el-icon {
    font-size: 16px;
  }
}

.logout-menu {
  padding: 0;
  border-radius: 18px;
  overflow: hidden;
  background-color: #fff;
  display: flex;
  align-items: center;
}

.logout-item {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 6px;
  padding: 8px 16px;
  cursor: pointer;
  transition: background-color 0.3s;
  font-size: 13px;
  color: #606266;
  border-radius: 18px;

  &:hover {
    background-color: #f5f7fa;
    color: #ff6600;
  }

  .el-icon {
    font-size: 14px;
  }
}

// 弹出菜单椭圆形样式 - 横向椭圆形
:deep(.logout-popover) {
  border-radius: 20px !important;
  padding: 0 !important;
  width: 110px !important;
  min-width: 110px !important;
  max-width: 110px !important;
  
  .el-popover__inner {
    border-radius: 20px !important;
    padding: 0 !important;
    background-color: #fff !important;
    width: 110px !important;
    min-width: 110px !important;
    max-width: 110px !important;
  }
  
  .el-popover__title {
    border-radius: 20px !important;
  }
}

.header {
  background-color: #fff;
  border-bottom: 1px solid #e4e7ed;
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 0 20px;
}

.header-left {
  display: flex;
  align-items: center;
}

.collapse-icon {
  font-size: 20px;
  cursor: pointer;
  color: #606266;
  
  &:hover {
    color: #ff6600;
  }
}

.header-right {
  display: flex;
  align-items: center;
  gap: 20px;
}

.api-env-switch {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 0 12px;
  border-right: 1px solid #e4e7ed;
  margin-right: 8px;
}

.api-env-label {
  font-size: 12px;
  color: #909399;
}

.header-icon {
  font-size: 20px;
  cursor: pointer;
  color: #606266;
  
  &:hover {
    color: #ff6600;
  }
}

.chat-badge {
  :deep(.el-badge__content) {
    background-color: #ff6a3a;
    border: none;
    font-size: 9px;
    height: 14px;
    line-height: 14px;
    padding: 0 4px;
  }
}

.main-content {
  background-color: #f5f7fa;
  padding: 20px;
  overflow-y: auto;
}
</style>

<style>
/* 全局样式 - 确保弹出框横向椭圆形 */
.logout-popover {
  border-radius: 20px !important;
  padding: 0 !important;
  width: 110px !important;
  min-width: 110px !important;
  max-width: 110px !important;
}

.logout-popover .el-popover__inner {
  border-radius: 20px !important;
  padding: 0 !important;
  background-color: #fff !important;
  width: 110px !important;
  min-width: 110px !important;
  max-width: 110px !important;
}

.logout-popover .el-popover__title {
  border-radius: 20px !important;
}
</style>
