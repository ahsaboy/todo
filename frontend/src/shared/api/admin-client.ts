import { ApiError } from '@/shared/api/errors'
import type { ApiResponse, PaginatedResponse } from '@/shared/api/types'

const ADMIN_API_BASE_URL = 'admin/api'

let isHandlingUnauthorized = false

function getAdminToken(): string | null {
  return sessionStorage.getItem('admin_token')
}

function handleUnauthorized(endpoint: string): void {
  if (endpoint.startsWith('/login')) {
    return
  }

  sessionStorage.removeItem('admin_token')

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

  const token = getAdminToken()
  if (token) {
    headers['X-Admin-Token'] = token
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
      const isLoginAttempt = endpoint.startsWith('/auth/verify')
      handleUnauthorized(endpoint)
      throw new ApiError({
        message: isLoginAttempt
          ? '令牌无效，请检查后重试'
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
