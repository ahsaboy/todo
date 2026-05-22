import type { Router } from 'vue-router'
import { useAuthStore } from '@/app/stores/auth.store'
import { useAdminAuthStore } from '@/app/stores/admin-auth.store'

export function setupRouterGuards(router: Router) {
  router.beforeEach((to) => {
    const authStore = useAuthStore()
    const adminAuthStore = useAdminAuthStore()

    if (to.meta.requiresAuth && !authStore.isAuthenticated) {
      return { name: 'login', query: { redirect: to.fullPath } }
    }

    if (to.meta.requiresGuest && authStore.isAuthenticated) {
      return { name: 'tasks' }
    }

    if (to.meta.requiresAdmin && !adminAuthStore.isAuthenticated) {
      return { name: 'admin-login', query: { redirect: to.fullPath } }
    }

    if (to.meta.requiresAdminGuest && adminAuthStore.isAuthenticated) {
      return { name: 'admin-dashboard' }
    }
  })
}
