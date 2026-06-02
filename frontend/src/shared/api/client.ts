import { API_BASE_URL } from '@/shared/config/api'
import { createApiClient } from './create-client'
import type { ApiResponse, PaginatedResponse } from '@/shared/api/types'

let unauthorizedHandler: (() => void) | null = null
let isHandlingUnauthorized = false

export function setUnauthorizedHandler(handler: () => void): void {
  unauthorizedHandler = handler
}

export const api = createApiClient({
  baseUrl: API_BASE_URL,
  storage: 'localStorage',
  tokenKey: 'api_key',
  authPaths: ['/auth/'],
  onUnauthorized(_endpoint) {
    localStorage.removeItem('api_key')
    if (!isHandlingUnauthorized) {
      isHandlingUnauthorized = true
      unauthorizedHandler?.()
      setTimeout(() => {
        isHandlingUnauthorized = false
      }, 0)
    }
  },
})

export type { ApiResponse, PaginatedResponse }
