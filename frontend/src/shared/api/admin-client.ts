import { ApiError } from '@/shared/api/errors'
import type { ApiResponse, PaginatedResponse } from '@/shared/api/types'

const ADMIN_API_BASE_URL = 'admin/api'

let isHandlingUnauthorized = false

function getAdminApiKey(): string | null {
  return sessionStorage.getItem('admin_api_key')
}

function handleUnauthorized(endpoint: string): void {
  if (endpoint.startsWith('/auth/login')) {
    return
  }

  sessionStorage.removeItem('admin_api_key')

  if (!isHandlingUnauthorized) {
    isHandlingUnauthorized = true
    window.location.hash = '#/admin/login'
    setTimeout(() => {
      isHandlingUnauthorized = false
    }, 0)
  }
}

async function request<T>(endpoint: string, options?: RequestInit): Promise<T> {
  const url = `${ADMIN_API_BASE_URL}${endpoint}`

  const headers: HeadersInit = {
    'Content-Type': 'application/json',
  }

  const apiKey = getAdminApiKey()
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

    if (response.status === 401) {
      const isLoginAttempt = endpoint.startsWith('/auth/login')
      handleUnauthorized(endpoint)
      throw new ApiError({
        message: isLoginAttempt
          ? '用户名或密码错误'
          : '会话已过期，请重新登录',
        code: 'UNAUTHORIZED',
        status: 401,
      })
    }

    if (!response.ok || data.success === false) {
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

    throw new ApiError({
      message: 'Request failed',
      code: 'NETWORK_ERROR',
      status: 0,
    })
  }
}

export const adminApi = {
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
