import { onMounted, onUnmounted } from 'vue'
import { useRouter } from 'vue-router'
import { useAuthStore } from '@/stores/auth'
import { useAppStore } from '@/stores/app'

const ACTIVITY_KEY = 'senddock:last_activity'
const SESSION_ENDED_KEY = 'senddock:session_ended'

const CLOUD_IDLE_LIMIT_MS = 20 * 60 * 1000
const SELF_HOSTED_FALLBACK_IDLE_LIMIT_MS = 2 * 60 * 60 * 1000

const IDLE_CHECK_INTERVAL_MS = 15_000
const ACTIVITY_WRITE_THROTTLE_MS = 5_000

const USER_ACTIVITY_EVENTS = ['mousedown', 'mousemove', 'keydown', 'scroll', 'touchstart', 'wheel'] as const

let lastActivityFallback = Date.now()

function readLastActivity(): number {
    try {
        const raw = window.localStorage.getItem(ACTIVITY_KEY)
        if (raw !== null) {
            const parsed = Number(raw)
            if (Number.isFinite(parsed)) return parsed
        }
    } catch {}
    return lastActivityFallback
}

function writeLastActivity(at: number) {
    lastActivityFallback = at
    try {
        window.localStorage.setItem(ACTIVITY_KEY, String(at))
    } catch {}
}

export function useSessionWatch() {
    const router = useRouter()
    const auth = useAuthStore()
    const app = useAppStore()

    let idleCheckTimer: ReturnType<typeof setInterval> | null = null
    let lastActivityWriteAt = 0
    let sessionEnding = false

    function idleLimitMs(): number {
        if (app.deploymentMode === 'cloud') return CLOUD_IDLE_LIMIT_MS
        const configured = app.sessionIdleTimeoutMinutes
        if (!Number.isFinite(configured) || configured <= 0) return SELF_HOSTED_FALLBACK_IDLE_LIMIT_MS
        return configured * 60 * 1000
    }

    function recordUserActivity() {
        const now = Date.now()
        if (now - lastActivityWriteAt < ACTIVITY_WRITE_THROTTLE_MS) return
        lastActivityWriteAt = now
        writeLastActivity(now)
    }

    function redirectToLogin() {
        if (router.currentRoute.value.name !== 'login') {
            router.push({ name: 'login', query: { reason: 'idle_timeout' } })
        }
    }

    async function endSession() {
        if (sessionEnding) return
        sessionEnding = true
        stopIdleChecks()

        await auth.logout()

        try {
            window.localStorage.setItem(SESSION_ENDED_KEY, String(Date.now()))
        } catch {}

        redirectToLogin()
    }

    function checkIdle() {
        if (!auth.isAuthenticated) {
            lastActivityWriteAt = Date.now()
            writeLastActivity(lastActivityWriteAt)
            return
        }
        if (Date.now() - readLastActivity() >= idleLimitMs()) {
            void endSession()
        }
    }

    function onVisibilityChange() {
        if (document.visibilityState === 'visible') checkIdle()
    }

    function onStorage(e: StorageEvent) {
        if (e.key === SESSION_ENDED_KEY && e.newValue) {
            sessionEnding = true
            stopIdleChecks()
            auth.isAuthenticated = false
            redirectToLogin()
        }
    }

    function startIdleChecks() {
        if (idleCheckTimer) return
        idleCheckTimer = setInterval(checkIdle, IDLE_CHECK_INTERVAL_MS)
    }

    function stopIdleChecks() {
        if (!idleCheckTimer) return
        clearInterval(idleCheckTimer)
        idleCheckTimer = null
    }

    onMounted(() => {
        writeLastActivity(Date.now())
        for (const event of USER_ACTIVITY_EVENTS) {
            document.addEventListener(event, recordUserActivity, { passive: true, capture: true })
        }
        document.addEventListener('visibilitychange', onVisibilityChange)
        window.addEventListener('storage', onStorage)
        startIdleChecks()
    })

    onUnmounted(() => {
        for (const event of USER_ACTIVITY_EVENTS) {
            document.removeEventListener(event, recordUserActivity, { capture: true })
        }
        document.removeEventListener('visibilitychange', onVisibilityChange)
        window.removeEventListener('storage', onStorage)
        stopIdleChecks()
    })
}
