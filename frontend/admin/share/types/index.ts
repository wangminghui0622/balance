import type { Component } from 'vue'

// 用户类型
export type UserType = 'operator' | 'platform' | 'shopowner'

// 菜单项类型
export interface MenuItem {
  label: string
  path: string
  icon?: string | Component
  children?: MenuItem[]
}

// 统计数据
export interface StatData {
  label: string
  value: number
  compareValue: number
  percentage: number
  trend: 'up' | 'down'
}
