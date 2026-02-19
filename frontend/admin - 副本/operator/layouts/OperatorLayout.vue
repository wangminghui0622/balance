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
  Shop, 
  Document, 
  Money
} from '@element-plus/icons-vue'

const userName = ref('Hector')
const userAccount = ref('1234567890')
const messageCount = ref(0)

const menuItems: MenuItem[] = [
  {
    label: '首页',
    path: '/operator',
    icon: House
  },
  {
    label: '店铺',
    path: '/operator/stores',
    icon: Shop
  },
  {
    label: '订单',
    path: '/operator/orders',
    icon: Document
  },
  {
    label: '财务',
    path: '/operator/finance',
    icon: Money,
    children: [
      { label: '我的账单', path: '/operator/finance/bills' },
      { label: '我的回款', path: '/operator/finance/payment' },
      { label: '运营保证金', path: '/operator/finance/deposit' },
      { label: '收款账户', path: '/operator/finance/account' }
    ]
  }
]
</script>
