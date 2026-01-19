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

const userName = ref('ShopOwner')
const userAccount = ref('SO001')
const messageCount = ref(0)

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
    icon: Money
  }
]
</script>
