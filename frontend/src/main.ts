import { createApp } from 'vue'
import { createPinia } from 'pinia'
import router from '@/app/router'
import { setupRouterGuards } from '@/app/router/guards'
import { setUnauthorizedHandler } from '@/shared/api/client'
import { useAuthStore } from '@/app/stores/auth.store'
import '@/styles/index.css'
import App from './App.vue'

const app = createApp(App)
const pinia = createPinia()

app.use(pinia)
app.use(router)

setupRouterGuards(router)

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
