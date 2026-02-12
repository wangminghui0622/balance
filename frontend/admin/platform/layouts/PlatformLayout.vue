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
import { ref, markRaw } from 'vue'
import BaseLayout from '@share/components/BaseLayout.vue'
import type { MenuItem } from '@share/types'
import { 
  House, 
  User, 
  Shop, 
  Document, 
  Wallet, 
  Tools, 
  DataAnalysis 
} from '@element-plus/icons-vue'

// 使用 markRaw 避免 Vue 将组件转为响应式
const HouseIcon = markRaw(House)
const UserIcon = markRaw(User)
const ShopIcon = markRaw(Shop)
const DocumentIcon = markRaw(Document)
const WalletIcon = markRaw(Wallet)
const ToolsIcon = markRaw(Tools)
const DataAnalysisIcon = markRaw(DataAnalysis)

const userName = ref('Hector')
const userAccount = ref('1234567890')
const messageCount = ref(4)

const menuItems: MenuItem[] = [
  {
    label: '首页',
    path: '/platform',
    icon: HouseIcon
  },
  {
    label: '用户',
    path: '/platform/users',
    icon: UserIcon
  },
  {
    label: '店铺',
    path: '/platform/stores',
    icon: ShopIcon
  },
  {
    label: '订单',
    path: '/platform/orders',
    icon: DocumentIcon
  },
  {
    label: '财务',
    path: '/platform/finance',
    icon: WalletIcon,
    children: [
      { label: '我的账单', path: '/platform/finance/bills' },
      { label: '我的佣金', path: '/platform/finance/commission' },
      { label: '托管账户', path: '/platform/finance/escrow' },
      { label: '罚补账户', path: '/platform/finance/penalty' },
      { label: '收款账户', path: '/platform/finance/collection' }
    ]
  },
  {
    label: '管理',
    path: '/platform/management',
    icon: ToolsIcon,
    children: [
      { label: '合作管理', path: '/platform/management/cooperation' },
      { label: '财务审核', path: '/platform/management/finance-audit' },
      { label: '违规处理', path: '/platform/management/violation' },
      { label: '用户管理', path: '/platform/management/users' },
      { label: '结算管理', path: '/platform/management/settlement' }
    ]
  },
  {
    label: '报表',
    path: '/platform/reports',
    icon: DataAnalysisIcon,
    children: [
      { label: '汇总报表', path: '/platform/reports/summary' },
      { label: '店主报表', path: '/platform/reports/owner' },
      { label: '平台报表', path: '/platform/reports/platform' },
      { label: '运营报表', path: '/platform/reports/operator' }
    ]
  }
]
</script>
