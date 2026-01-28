import { createRouter, createWebHistory } from 'vue-router'
import type { RouteRecordRaw } from 'vue-router'
import { STORAGE_KEYS, USER_TYPE, ROUTE_PATH, getUserType, getRouteByUserType } from '@share/constants'

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
      // 根据用户ID前缀重定向
      const userId = localStorage.getItem(STORAGE_KEYS.USER_ID)
      if (!userId) {
        return ROUTE_PATH.LOGIN
      }
      const userType = getUserType(userId)
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
  const userId = localStorage.getItem(STORAGE_KEYS.USER_ID)

  // 如果访问登录页且已登录，根据用户类型重定向
  if (to.path === ROUTE_PATH.LOGIN) {
    if (token && userId) {
      const userType = getUserType(userId)
      next(getRouteByUserType(userType))
    } else {
      next()
    }
    return
  }

  // 需要认证的路由
  if (to.meta.requiresAuth) {
    if (!token || !userId) {
      next(ROUTE_PATH.LOGIN)
      return
    }

    // 检查用户类型是否匹配
    const requiredUserType = to.meta.userType as string
    const currentUserType = getUserType(userId)
    
    if (!currentUserType || currentUserType !== requiredUserType) {
      // 用户类型不匹配，重定向到对应的页面
      next(getRouteByUserType(currentUserType))
      return
    }
  }

  next()
})

export default router
