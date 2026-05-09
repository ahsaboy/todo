import { API_BASE_URL } from '@/shared/config/api'
import { ApiError } from '@/shared/api/errors'
import type { ApiResponse, PaginatedResponse } from '@/shared/api/types'

// From localStorage get API Key
function getApiKey(): string | null {
  return localStorage.getItem('api_key')
}

// Handle 401 Unauthorized - clear auth and redirect to login
function handleUnauthorized(): void {
  localStorage.removeItem('api_key')
  // Redirect to login page using hash mode
  window.location.href = '/#/login'
}

// Base request method
async function request<T>(endpoint: string, options?: RequestInit): Promise<T> {
  const url = `${API_BASE_URL}${endpoint}`

  const headers: HeadersInit = {
    'Content-Type': 'application/json',
  }

  const apiKey = getApiKey()
  if (apiKey) {
    headers['Authorization'] = `Bearer ${apiKey}`
  }

  const response = await fetch(url, {
    ...options,
    headers: {
      ...headers,
      ...options?.headers,
    },
  })

  const data = await response.json()

  // Handle 401 Unauthorized
  if (response.status === 401) {
    handleUnauthorized()
    throw new ApiError({
      message: 'Session expired, please login again',
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
