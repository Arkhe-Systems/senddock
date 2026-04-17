import { createRouter, createWebHistory } from 'vue-router'
import { useAuthStore } from '@/stores/auth'
import { useToastStore } from '@/stores/toast'
import { useAppStore } from '@/stores/app'
import { api } from '@/api/client'

interface SetupStatus {
  setup_required: boolean
  deployment_mode: string
}

const router = createRouter({
  history: createWebHistory(import.meta.env.BASE_URL),
  routes: [
    {
      path: '/setup',
      name: 'setup',
      component: () => import('@/views/setup/SetupView.vue'),
    },
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
      meta: { requiresAuth: true },
    },
    {
      path: '/projects/:id',
      component: () => import('@/views/project/ProjectView.vue'),
      meta: { requiresAuth: true },
      children: [
        {
          path: '',
          name: 'project-overview',
          component: () => import('@/views/project/OverviewSection.vue'),
        },
        {
          path: 'subscribers',
          name: 'project-subscribers',
          component: () => import('@/views/project/SubscribersSection.vue'),
        },
        {
          path: 'templates',
          name: 'project-templates',
          component: () => import('@/views/project/TemplatesSection.vue'),
        },
        {
          path: 'logs',
          name: 'project-logs',
          component: () => import('@/views/project/LogsSection.vue'),
        },
        {
          path: 'smtp',
          name: 'project-smtp',
          component: () => import('@/views/project/SmtpSection.vue'),
        },
        {
          path: 'settings',
          name: 'project-settings',
          component: () => import('@/views/project/SettingsSection.vue'),
        },
      ],
    },
    {
      path: '/',
      redirect: '/dashboard',
    },
  ],
})

router.beforeEach(async (to) => {
  const app = useAppStore()

  if (!app.checked) {
    try {
      const status = await api<SetupStatus>('/setup/status')
      app.setupRequired = status.setup_required
      app.deploymentMode = status.deployment_mode
    } catch {
      app.setupRequired = false
    }
    app.checked = true
  }

  if (app.setupRequired && app.deploymentMode === 'self-hosted' && to.name !== 'setup') {
    return { name: 'setup' }
  }

  if (to.name === 'setup' && (!app.setupRequired || app.deploymentMode === 'cloud')) {
    return { name: 'login' }
  }

  if (to.name === 'register' && app.deploymentMode === 'self-hosted') {
    return { name: 'login' }
  }

  if (to.name === 'setup') return

  const auth = useAuthStore()
  await auth.checkAuth()

  if (to.meta.requiresAuth && !auth.isAuthenticated) {
    const toast = useToastStore()

    if (auth.sessionExpired) {
      toast.error('Your session has expired. Please sign in again.')
      return { name: 'login', query: { reason: 'session_expired' } }
    }

    return { name: 'login', query: { reason: 'auth_required' } }
  }

  if ((to.name === 'login' || to.name === 'register') && auth.isAuthenticated) {
    return { name: 'dashboard' }
  }
})

export default router
