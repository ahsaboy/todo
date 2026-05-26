import { computed, ref } from 'vue'
import { defineStore } from 'pinia'

export const useAdminAuthStore = defineStore('admin-auth', () => {
  const apiKey = ref<string | null>(sessionStorage.getItem('admin_api_key'))
  const isAuthenticated = computed(() => !!apiKey.value)

  function setAuth(key: string) {
    apiKey.value = key
    sessionStorage.setItem('admin_api_key', key)
  }

  function logout() {
    apiKey.value = null
    sessionStorage.removeItem('admin_api_key')
  }

  return { apiKey, isAuthenticated, setAuth, logout }
})
