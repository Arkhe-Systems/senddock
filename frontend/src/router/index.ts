import { createRouter, createWebHistory } from 'vue-router'
import { useAuthStore } from '@/stores/auth'

const router = createRouter({
  history: createWebHistory(import.meta.env.BASE_URL),
  routes: [
    {
      path: '/login',
      name: 'login',
      component: () => import('@/views/login/LoginView.vue'),
    },
    {
      path: '/register',
      name: 'register',
      component: () => import('@/views/register/RegisterView.vue'),
    },
    {
      path: '/dashboard',
      name: 'dashboard',
      component: () => import('@/views/dashboard/dashboardView.vue'),
      meta: {requiresAuth: true},
    },
    {
      path: '/',
      redirect: '/dashboard',
    },
  ],
})

router.beforeEach(async (to) => {
  const auth = useAuthStore()
  await auth.checkAuth()

  if (to.meta.requiresAuth && !auth.isAuthenticated) {
    return {name: 'login' , query: {reason: 'auth_required'}}
  }

  if ((to.name === 'login' || to.name === 'register') && auth.isAuthenticated) {
    return {name: 'dashboard'}
  }

})


export default router
