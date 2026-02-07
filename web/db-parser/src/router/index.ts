import { createRouter, createWebHistory } from 'vue-router'
import HomeView from '@/views/HomeView.vue'
import { useAuthStore } from '@/stores/auth'

const router = createRouter({
  history: createWebHistory(import.meta.env.BASE_URL),
  routes: [
    {
      path: '/login',
      name: 'login',
      component: () => import('@/views/LoginView.vue'),
      meta: { requiresGuest: true },
    },
    {
      path: '/register',
      name: 'register',
      component: () => import('@/views/RegisterView.vue'),
      meta: { requiresGuest: true },
    },
    {
      path: '/',
      name: 'home',
      component: HomeView,
      meta: { requiresAuth: true },
    },
    {
      path: '/history',
      name: 'history',
      component: () => import('@/views/HistoryView.vue'),
      meta: { requiresAuth: true },
    },
    {
      path: '/tenants',
      name: 'tenants',
      component: () => import('@/views/TenantsView.vue'),
      meta: { requiresAuth: true, roles: ['SUPER_ADMIN', 'REALM_ADMIN'] },
    },
    {
      path: '/:pathMatch(.*)*',
      redirect: '/',
    },
  ],
})

router.beforeEach(async (to, from, next) => {
  const auth = useAuthStore()
  
  if (!auth.isAuthenticated && to.meta.requiresAuth !== false) {
    await auth.checkAuth()
  }

  if (to.meta.requiresAuth && !auth.isAuthenticated) {
    next('/login')
  } else if (to.meta.requiresGuest && auth.isAuthenticated) {
    next('/')
  } else if (to.meta.roles && !to.meta.roles.includes(auth.userRole)) {
    next('/')
  } else {
    next()
  }
})

export default router
