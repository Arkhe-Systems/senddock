import { createApp } from 'vue'
import { createPinia } from 'pinia'

import App from './App.vue'
import router from './router'
import { setSessionExpiredHandler, setRateLimitedHandler } from './api/client'
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

let lastRateLimitToastAt = 0
setRateLimitedHandler(() => {
    const now = Date.now()
    if (now - lastRateLimitToastAt < 5000) return
    lastRateLimitToastAt = now
    const toast = useToastStore()
    toast.error('Too many requests. Slow down and try again in a minute.')
})

app.mount('#app')
