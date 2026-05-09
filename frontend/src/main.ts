import { createApp } from 'vue'
import { createPinia } from 'pinia'
import router from '@/app/router'
import { setupRouterGuards } from '@/app/router/guards'
import '@/styles/index.css'
import App from './App.vue'

const app = createApp(App)
const pinia = createPinia()

app.use(pinia)
app.use(router)

setupRouterGuards(router)

app.mount('#app')
