<template>
  <BaseLayout
    :menu-items="menuItems"
    :user-name="userName"
    :user-account="userAccount"
    :message-count="messageCount"
  >
    <router-view />
  </BaseLayout>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import BaseLayout from '@share/components/BaseLayout.vue'
import type { MenuItem } from '@share/types'
import { 
  House, 
  User, 
  Shop, 
  Document, 
  Money, 
  Setting, 
  DataAnalysis 
} from '@element-plus/icons-vue'

const userName = ref('Hector')
const userAccount = ref('1234567890')
const messageCount = ref(4)

const menuItems: MenuItem[] = [
  {
    label: '首页',
    path: '/platform',
    icon: House
  },
  {
    label: '用户',
    path: '/platform/users',
    icon: User
  },
  {
    label: '店铺',
    path: '/platform/stores',
    icon: Shop
  },
  {
    label: '订单',
    path: '/platform/orders',
    icon: Document
  },
  {
    label: '财务',
    path: '/platform/finance',
    icon: Money,
    children: [
      { label: '财务概览', path: '/platform/finance/overview' },
      { label: '充值记录', path: '/platform/finance/recharge' }
    ]
  },
  {
    label: '管理',
    path: '/platform/management',
    icon: Setting,
    children: [
      { label: '系统设置', path: '/platform/management/settings' }
    ]
  },
  {
    label: '报表',
    path: '/platform/reports',
    icon: DataAnalysis,
    children: [
      { label: '数据报表', path: '/platform/reports/data' }
    ]
  }
]
</script>
