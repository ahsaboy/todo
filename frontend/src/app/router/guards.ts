import type { Router } from 'vue-router'
import { useAuthStore } from '@/app/stores/auth.store'

export function setupRouterGuards(router: Router) {
  router.beforeEach((to) => {
    const authStore = useAuthStore()

    if (to.meta.requiresAuth && !authStore.isAuthenticated) {
      return { name: 'login', query: { redirect: to.fullPath } }
    }

    if (to.meta.requiresGuest && authStore.isAuthenticated) {
      return { name: 'tasks' }
    }
  })
}
