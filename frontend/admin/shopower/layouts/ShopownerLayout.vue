<template>
  <BaseLayout
    :menu-items="menuItems"
    :user-name="userName"
    :user-id="userId"
    :message-count="messageCount"
  >
    <router-view />
  </BaseLayout>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import BaseLayout from '@share/components/BaseLayout.vue'
import type { MenuItem } from '@share/types'
import { authApi } from '@share/api/auth'
import { 
  House, 
  Shop, 
  Document, 
  Money 
} from '@element-plus/icons-vue'

const userName = ref('')
const userId = ref('')
const messageCount = ref(0)

onMounted(async () => {
  try {
    const res = await authApi.getCurrentUser()
    if (res?.data && (res?.code === 0 || res?.code === 200)) {
      userName.value = res.data.userName || '店主'
      userId.value = res.data.id != null ? String(res.data.id) : ''
    }
  } catch {
    userName.value = '店主'
    userId.value = ''
  }
})

const menuItems: MenuItem[] = [
  {
    label: '首页',
    path: '/shopowner',
    icon: House
  },
  {
    label: '我的店铺',
    path: '/shopowner/stores',
    icon: Shop
  },
  {
    label: '订单管理',
    path: '/shopowner/orders',
    icon: Document
  },
  {
    label: '财务管理',
    path: '/shopowner/finance',
    icon: Money,
    children: [
      {
        label: '我的账单',
        path: '/shopowner/finance'
      },
      {
        label: '我的佣金',
        path: '/shopowner/finance/commission'
      },
      {
        label: '我的预付款',
        path: '/shopowner/finance/prepayment'
      },
      {
        label: '店主保证金',
        path: '/shopowner/finance/deposit'
      },
      {
        label: '收款账户',
        path: '/shopowner/finance/account'
      }
    ]
  }
]
</script>
