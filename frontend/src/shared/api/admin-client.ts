import { ADMIN_API_BASE_URL } from '@/shared/config/api'
import { createApiClient } from './create-client'
import type { ApiResponse, PaginatedResponse } from '@/shared/api/types'

let isHandlingUnauthorized = false

export const adminApi = createApiClient({
  baseUrl: ADMIN_API_BASE_URL,
  storage: 'sessionStorage',
  tokenKey: 'admin_api_key',
  authPaths: ['/auth/login'],
  authErrorMessage: '用户名或密码错误',
  sessionExpiredMessage: '会话已过期，请重新登录',
  onUnauthorized(_endpoint) {
    sessionStorage.removeItem('admin_api_key')
    if (!isHandlingUnauthorized) {
      isHandlingUnauthorized = true
      window.location.hash = '#/admin/login'
      setTimeout(() => {
        isHandlingUnauthorized = false
      }, 0)
    }
  },
})

export type { ApiResponse, PaginatedResponse }
