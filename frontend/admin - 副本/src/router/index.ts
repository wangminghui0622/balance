import { createRouter, createWebHistory } from 'vue-router'
import type { RouteRecordRaw } from 'vue-router'
import { STORAGE_KEYS, USER_TYPE, ROUTE_PATH, getRouteByUserType, type UserType } from '@share/constants'

// 从 localStorage 获取用户类型
function getStoredUserType(): UserType | null {
  const userType = localStorage.getItem(STORAGE_KEYS.USER_TYPE)
  if (userType === USER_TYPE.PLATFORM || userType === USER_TYPE.OPERATOR || userType === USER_TYPE.SHOPOWNER) {
    return userType as UserType
  }
  return null
}

const routes: RouteRecordRaw[] = [
  {
    path: ROUTE_PATH.LOGIN,
    name: 'login',
    component: () => import('../views/Login.vue')
  },
  {
    path: ROUTE_PATH.REGISTER,
    name: 'register',
    component: () => import('../views/Register.vue')
  },
  {
    path: '/register-info',
    name: 'register-info',
    component: () => import('../views/RegisterInfo.vue')
  },
  {
    path: '/shopee/auth',
    name: 'shopee-auth',
    component: () => import('../views/ShopeeAuth.vue'),
    meta: { requiresAuth: true }
  },
  {
    path: '/shopee/auth/callback',
    name: 'shopee-auth-callback',
    component: () => import('../views/ShopeeCallback.vue')
  },
  {
    path: '/shopee/auth/rebind',
    name: 'shopee-auth-rebind',
    component: () => import('../views/ShopeeRebind.vue')
  },
  {
    path: '/',
    redirect: () => {
      // 根据用户类型重定向
      const token = localStorage.getItem(STORAGE_KEYS.TOKEN)
      if (!token) {
        return ROUTE_PATH.LOGIN
      }
      const userType = getStoredUserType()
      return getRouteByUserType(userType)
    }
  },
  {
    path: ROUTE_PATH.PLATFORM,
    component: () => import('@platform/layouts/PlatformLayout.vue'),
    meta: { requiresAuth: true, userType: USER_TYPE.PLATFORM },
    children: [
      {
        path: '',
        name: 'platform-home',
        component: () => import('@platform/views/Home.vue')
      },
      {
        path: 'realtime-dashboard',
        name: 'platform-realtime-dashboard',
        component: () => import('@platform/views/RealtimeDashboard.vue')
      },
      {
        path: 'stores',
        name: 'platform-stores',
        component: () => import('@platform/views/Stores.vue')
      },
      {
        path: 'orders',
        name: 'platform-orders',
        component: () => import('@platform/views/Orders.vue')
      },
      {
        path: 'finance/bills',
        name: 'platform-finance-bills',
        component: () => import('@platform/views/Bills.vue')
      },
      {
        path: 'finance/commission',
        name: 'platform-finance-commission',
        component: () => import('@platform/views/Commission.vue')
      },
      {
        path: 'finance/escrow',
        name: 'platform-finance-escrow',
        component: () => import('@platform/views/Escrow.vue')
      },
      {
        path: 'finance/penalty',
        name: 'platform-finance-penalty',
        component: () => import('@platform/views/Penalty.vue')
      },
      {
        path: 'finance/collection',
        name: 'platform-finance-collection',
        component: () => import('@platform/views/Collection.vue')
      },
      {
        path: 'management/cooperation',
        name: 'platform-management-cooperation',
        component: () => import('@platform/views/Cooperation.vue')
      },
      {
        path: 'management/finance-audit',
        name: 'platform-management-finance-audit',
        component: () => import('@platform/views/FinanceAudit.vue')
      },
      {
        path: 'management/violation',
        name: 'platform-management-violation',
        component: () => import('@platform/views/Violation.vue')
      },
      {
        path: 'management/users',
        name: 'platform-management-users',
        component: () => import('@platform/views/UserManagement.vue')
      },
      {
        path: 'management/settlement',
        name: 'platform-management-settlement',
        component: () => import('@platform/views/Settlement.vue')
      },
      {
        path: 'reports/summary',
        name: 'platform-reports-summary',
        component: () => import('@platform/views/ReportSummary.vue')
      },
      {
        path: 'reports/owner',
        name: 'platform-reports-owner',
        component: () => import('@platform/views/ReportOwner.vue')
      },
      {
        path: 'reports/platform',
        name: 'platform-reports-platform',
        component: () => import('@platform/views/ReportPlatform.vue')
      },
      {
        path: 'reports/operator',
        name: 'platform-reports-operator',
        component: () => import('@platform/views/ReportOperator.vue')
      }
    ]
  },
  {
    path: ROUTE_PATH.OPERATOR,
    component: () => import('@operator/layouts/OperatorLayout.vue'),
    meta: { requiresAuth: true, userType: USER_TYPE.OPERATOR },
    children: [
      {
        path: '',
        name: 'operator-home',
        component: () => import('@operator/views/Home.vue')
      },
      {
        path: 'stores',
        name: 'operator-stores',
        component: () => import('@operator/views/Stores.vue')
      },
      {
        path: 'orders',
        name: 'operator-orders',
        component: () => import('@operator/views/Orders.vue')
      },
      {
        path: 'finance',
        name: 'operator-finance',
        redirect: '/operator/finance/bills'
      },
      {
        path: 'finance/bills',
        name: 'operator-finance-bills',
        component: () => import('@operator/views/FinanceBills.vue')
      },
      {
        path: 'finance/payment',
        name: 'operator-finance-payment',
        component: () => import('@operator/views/FinancePayment.vue')
      },
      {
        path: 'finance/deposit',
        name: 'operator-finance-deposit',
        component: () => import('@operator/views/FinanceDeposit.vue')
      },
      {
        path: 'finance/account',
        name: 'operator-finance-account',
        component: () => import('@operator/views/FinanceAccount.vue')
      }
    ]
  },
  {
    path: ROUTE_PATH.SHOPOWNER,
    component: () => import('@shopowner/layouts/ShopownerLayout.vue'),
    meta: { requiresAuth: true, userType: USER_TYPE.SHOPOWNER },
    children: [
      {
        path: '',
        name: 'shopowner-home',
        component: () => import('@shopowner/views/Home.vue')
      },
      {
        path: 'stores',
        name: 'shopowner-stores',
        component: () => import('@shopowner/views/Stores.vue')
      },
      {
        path: 'orders',
        name: 'shopowner-orders',
        component: () => import('@shopowner/views/Orders.vue')
      },
      {
        path: 'finance',
        name: 'shopowner-finance',
        component: () => import('@shopowner/views/Finance.vue')
      },
      {
        path: 'finance/commission',
        name: 'shopowner-finance-commission',
        component: () => import('@shopowner/views/FinanceCommission.vue')
      },
      {
        path: 'finance/prepayment',
        name: 'shopowner-finance-prepayment',
        component: () => import('@shopowner/views/FinancePrepayment.vue')
      },
      {
        path: 'finance/deposit',
        name: 'shopowner-finance-deposit',
        component: () => import('@shopowner/views/FinanceDeposit.vue')
      },
      {
        path: 'finance/account',
        name: 'shopowner-finance-account',
        component: () => import('@shopowner/views/FinanceAccount.vue')
      }
    ]
  }
]

// 根据环境变量设置路由 base
// 开发环境：/，生产环境：/balance/admin/
const getRouterBase = () => {
  // 开发环境（npm run dev）
  if (import.meta.env.MODE === 'development') {
    return '/'
  }
  // 生产环境（npm run build）
  return '/balance/admin/'
}

const router = createRouter({
  history: createWebHistory(getRouterBase()),
  routes
})

// 路由守卫
router.beforeEach((to, _from, next) => {
  const token = localStorage.getItem(STORAGE_KEYS.TOKEN)
  const userType = getStoredUserType()

  // 如果访问登录页且已登录，根据用户类型重定向
  // 使用 to.name 而不是 to.path，避免生产环境路径前缀问题
  if (to.name === 'login') {
    if (token && userType) {
      next(getRouteByUserType(userType))
    } else {
      next()
    }
    return
  }

  // 需要认证的路由
  if (to.meta.requiresAuth) {
    if (!token) {
      next({ name: 'login' })
      return
    }

    // 检查用户类型是否匹配
    const requiredUserType = to.meta.userType as string
    
    if (!userType || userType !== requiredUserType) {
      // 用户类型不匹配，重定向到对应的页面
      next(getRouteByUserType(userType))
      return
    }
  }

  next()
})

export default router
