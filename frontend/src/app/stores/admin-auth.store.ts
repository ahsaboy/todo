import { computed, ref } from 'vue'
import { defineStore } from 'pinia'

export const useAdminAuthStore = defineStore('admin-auth', () => {
  const token = ref<string | null>(sessionStorage.getItem('admin_token'))
  const isAuthenticated = computed(() => !!token.value)

  function setAuth(rawToken: string) {
    token.value = rawToken
    sessionStorage.setItem('admin_token', rawToken)
  }

  function logout() {
    token.value = null
    sessionStorage.removeItem('admin_token')
  }

  return { token, isAuthenticated, setAuth, logout }
})
