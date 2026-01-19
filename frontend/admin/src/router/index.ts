import { createRouter, createWebHistory } from 'vue-router'
import type { RouteRecordRaw } from 'vue-router'

const routes: RouteRecordRaw[] = [
  {
    path: '/login',
    name: 'login',
    component: () => import('../views/Login.vue')
  },
  {
    path: '/register',
    name: 'register',
    component: () => import('../views/Register.vue')
  },
  {
    path: '/',
    redirect: () => {
      // 根据用户ID前缀重定向
      const userId = localStorage.getItem('userId')
      if (!userId) {
        return '/login'
      }
      if (userId.startsWith('9')) {
        return '/platform'
      } else if (userId.startsWith('5')) {
        return '/operator'
      } else if (userId.startsWith('1')) {
        return '/shopowner'
      }
      return '/login'
    }
  },
  {
    path: '/platform',
    component: () => import('@platform/layouts/PlatformLayout.vue'),
    meta: { requiresAuth: true, userType: '9' },
    children: [
      {
        path: '',
        name: 'platform-home',
        component: () => import('@platform/views/Home.vue')
      }
    ]
  },
  {
    path: '/operator',
    component: () => import('@operator/layouts/OperatorLayout.vue'),
    meta: { requiresAuth: true, userType: '5' },
    children: [
      {
        path: '',
        name: 'operator-home',
        component: () => import('@operator/views/Home.vue')
      }
    ]
  },
  {
    path: '/shopowner',
    component: () => import('@shopowner/layouts/ShopownerLayout.vue'),
    meta: { requiresAuth: true, userType: '1' },
    children: [
      {
        path: '',
        name: 'shopowner-home',
        component: () => import('@shopowner/views/Home.vue')
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
  const token = localStorage.getItem('token')
  const userId = localStorage.getItem('userId')

  // 如果访问登录页且已登录，根据用户类型重定向
  if (to.path === '/login') {
    if (token && userId) {
      if (userId.startsWith('9')) {
        next('/platform')
      } else if (userId.startsWith('5')) {
        next('/operator')
      } else if (userId.startsWith('1')) {
        next('/shopowner')
      } else {
        next()
      }
    } else {
      next()
    }
    return
  }

  // 需要认证的路由
  if (to.meta.requiresAuth) {
    if (!token || !userId) {
      next('/login')
      return
    }

    // 检查用户类型是否匹配
    const requiredUserType = to.meta.userType as string
    if (!userId.startsWith(requiredUserType)) {
      // 用户类型不匹配，重定向到对应的页面
      if (userId.startsWith('9')) {
        next('/platform')
      } else if (userId.startsWith('5')) {
        next('/operator')
      } else if (userId.startsWith('1')) {
        next('/shopowner')
      } else {
        next('/login')
      }
      return
    }
  }

  next()
})

export default router
