import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import type { User, UserDto, ProfileResponse } from '@/entities/user/model'
import { toUser } from '@/entities/user/mapper'
import { api } from '@/shared/api/client'
import type { ApiResponse } from '@/shared/api/types'

export const useAuthStore = defineStore('auth', () => {
  const user = ref<User | null>(null)
  const apiKey = ref<string | null>(localStorage.getItem('api_key'))

  const isAuthenticated = computed(() => !!apiKey.value)

  function setAuth(key: string, userDto: UserDto) {
    apiKey.value = key
    user.value = toUser(userDto)
    localStorage.setItem('api_key', key)
  }

  function setOAuthAuth(key: string) {
    apiKey.value = key
    localStorage.setItem('api_key', key)
  }

  function logout() {
    apiKey.value = null
    user.value = null
    localStorage.removeItem('api_key')
  }

  async function fetchProfile() {
    try {
      const response = await api.get<ApiResponse<ProfileResponse>>('/user/profile')
      user.value = toUser(response.data.user)
    } catch {
      logout()
    }
  }

  return {
    user,
    apiKey,
    isAuthenticated,
    setAuth,
    setOAuthAuth,
    logout,
    fetchProfile,
  }
})
