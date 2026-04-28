import { createApp } from 'vue'
import { createPinia } from 'pinia'

import App from './App.vue'
import router from './router'
import { setSessionExpiredHandler } from './api/client'
import { useAuthStore } from './stores/auth'
import { useToastStore } from './stores/toast'

import './assets/main.css'

const app = createApp(App)

app.use(createPinia())
app.use(router)

setSessionExpiredHandler(() => {
    const auth = useAuthStore()
    if (auth.sessionExpired) return
    auth.isAuthenticated = false
    auth.sessionExpired = true
    const toast = useToastStore()
    toast.error('Your session has expired. Please sign in again.')
    if (router.currentRoute.value.name !== 'login') {
        router.push({ name: 'login', query: { reason: 'session_expired' } })
    }
})

app.mount('#app')
