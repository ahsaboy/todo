import { API_BASE_URL } from '@/shared/config/api'
import { ApiError } from '@/shared/api/errors'
import type { ApiResponse, PaginatedResponse } from '@/shared/api/types'
import { logger } from '@/shared/logger/logger'

let unauthorizedHandler: (() => void) | null = null
let isHandlingUnauthorized = false

export function setUnauthorizedHandler(handler: () => void): void {
  unauthorizedHandler = handler
}

// From localStorage get API Key
function getApiKey(): string | null {
  return localStorage.getItem('api_key')
}

// Handle 401 Unauthorized - clear auth and redirect to login
function handleUnauthorized(endpoint: string): void {
  if (endpoint.startsWith('/auth/')) {
    return
  }

  localStorage.removeItem('api_key')

  if (!isHandlingUnauthorized) {
    isHandlingUnauthorized = true
    unauthorizedHandler?.()
    setTimeout(() => {
      isHandlingUnauthorized = false
    }, 0)
  }
}

// Base request method
async function request<T>(endpoint: string, options?: RequestInit): Promise<T> {
  const url = `${API_BASE_URL}${endpoint}`
  const method = options?.method ?? 'GET'

  const headers: HeadersInit = {
    'Content-Type': 'application/json',
  }

  const apiKey = getApiKey()
  if (apiKey) {
    headers['Authorization'] = `Bearer ${apiKey}`
  }

  try {
    const response = await fetch(url, {
      ...options,
      headers: {
        ...headers,
        ...options?.headers,
      },
    })

    const data = await response.json().catch(() => ({}))

    // Handle 401 Unauthorized
    if (response.status === 401) {
      const isSessionExpired = !endpoint.startsWith('/auth/')
      handleUnauthorized(endpoint)
      logger.warn('API request failed', {
        method,
        endpoint,
        status: response.status,
        code: 'UNAUTHORIZED',
      })
      throw new ApiError({
        message: isSessionExpired
          ? 'Session expired, please login again'
          : data.error || 'Unauthorized',
        code: 'UNAUTHORIZED',
        status: 401,
      })
    }

    if (!response.ok || data.success === false) {
      logger.warn('API request failed', {
        method,
        endpoint,
        status: response.status,
        code: data.code || 'INTERNAL_ERROR',
      })
      throw new ApiError({
        message: data.error || 'Request failed',
        code: data.code || 'INTERNAL_ERROR',
        status: response.status,
      })
    }

    return data
  } catch (error) {
    if (error instanceof ApiError) {
      throw error
    }

    logger.error('API request failed', {
      method,
      endpoint,
      status: 0,
      code: 'NETWORK_ERROR',
    })
    throw new ApiError({
      message: 'Request failed',
      code: 'NETWORK_ERROR',
      status: 0,
    })
  }
}

// Export request methods
export const api = {
  get: <T>(endpoint: string) => request<T>(endpoint, { method: 'GET' }),

  post: <T>(endpoint: string, body?: unknown) =>
    request<T>(endpoint, {
      method: 'POST',
      body: body ? JSON.stringify(body) : undefined,
    }),

  put: <T>(endpoint: string, body?: unknown) =>
    request<T>(endpoint, {
      method: 'PUT',
      body: body ? JSON.stringify(body) : undefined,
    }),

  patch: <T>(endpoint: string, body?: unknown) =>
    request<T>(endpoint, {
      method: 'PATCH',
      body: body ? JSON.stringify(body) : undefined,
    }),

  delete: <T>(endpoint: string) => request<T>(endpoint, { method: 'DELETE' }),
}

export type { ApiResponse, PaginatedResponse }
