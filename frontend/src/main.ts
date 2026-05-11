import { createApp } from 'vue'
import { createPinia } from 'pinia'
import router from '@/app/router'
import { setupRouterGuards } from '@/app/router/guards'
import { setUnauthorizedHandler } from '@/shared/api/client'
import { useAuthStore } from '@/app/stores/auth.store'
import { useThemeStore } from '@/app/stores/theme.store'
import { loadLoggerConfig, logger } from '@/shared/logger'
import '@/styles/index.css'
import App from './App.vue'

const app = createApp(App)
const pinia = createPinia()

app.use(pinia)
app.use(router)

useThemeStore(pinia).initTheme()

setupRouterGuards(router)

void loadLoggerConfig().then(() => {
  window.addEventListener('error', (event) => {
    const message = event.message || 'Uncaught error'
    logger.error(message, {
      filename: event.filename,
      lineno: event.lineno,
      colno: event.colno,
      stack: event.error instanceof Error ? event.error.stack : undefined,
    })
  })

  window.addEventListener('unhandledrejection', (event) => {
    const reason = event.reason
    const message = reason instanceof Error ? reason.message : 'Unhandled promise rejection'
    logger.error(message, {
      stack: reason instanceof Error ? reason.stack : undefined,
      reason: typeof reason === 'string' ? reason : undefined,
    })
  })
})

setUnauthorizedHandler(() => {
  const authStore = useAuthStore()
  authStore.logout()
  if (router.currentRoute.value.name !== 'login') {
    void router.replace({
      name: 'login',
      query: { redirect: router.currentRoute.value.fullPath },
    })
  }
})

app.mount('#app')
