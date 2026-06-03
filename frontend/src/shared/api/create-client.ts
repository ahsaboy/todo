import { ApiError } from '@/shared/api/errors'

export interface ApiClientConfig {
  baseUrl: string
  storage: 'localStorage' | 'sessionStorage'
  tokenKey: string
  /** Paths that should NOT trigger 401 handling */
  authPaths: string[]
  /** Called on 401 for non-auth endpoints */
  onUnauthorized: (endpoint: string) => void
  /** Error message for auth endpoint 401. Default: uses server error or 'Unauthorized' */
  authErrorMessage?: string
  /** Error message for session expired 401. Default: 'Session expired, please login again' */
  sessionExpiredMessage?: string
}

export interface ApiClient {
  get: <T>(endpoint: string) => Promise<T>
  post: <T>(endpoint: string, body?: unknown) => Promise<T>
  put: <T>(endpoint: string, body?: unknown) => Promise<T>
  patch: <T>(endpoint: string, body?: unknown) => Promise<T>
  delete: <T>(endpoint: string) => Promise<T>
  download: (endpoint: string) => Promise<Blob>
}

export function createApiClient(config: ApiClientConfig): ApiClient {
  let isHandlingUnauthorized = false

  function getToken(): string | null {
    return window[config.storage].getItem(config.tokenKey)
  }

  function isAuthEndpoint(endpoint: string): boolean {
    return config.authPaths.some((p) => endpoint.startsWith(p))
  }

  async function request<T>(endpoint: string, options?: RequestInit): Promise<T> {
    const url = `${config.baseUrl}${endpoint}`

    const headers: HeadersInit = {
      'Content-Type': 'application/json',
    }

    const token = getToken()
    if (token) {
      headers['Authorization'] = `Bearer ${token}`
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
        if (!isAuthEndpoint(endpoint)) {
          config.onUnauthorized(endpoint)
        }
        throw new ApiError({
          message: isAuthEndpoint(endpoint)
            ? config.authErrorMessage ?? data.error ?? 'Unauthorized'
            : config.sessionExpiredMessage ?? 'Session expired, please login again',
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

  async function download(endpoint: string): Promise<Blob> {
    const url = `${config.baseUrl}${endpoint}`
    const headers: HeadersInit = {}
    const token = getToken()
    if (token) headers['Authorization'] = `Bearer ${token}`

    const response = await fetch(url, { headers })
    if (response.status === 401) {
      if (!isAuthEndpoint(endpoint)) config.onUnauthorized(endpoint)
      throw new ApiError({ message: config.sessionExpiredMessage ?? 'Unauthorized', code: 'UNAUTHORIZED', status: 401 })
    }
    if (!response.ok) throw new ApiError({ message: 'Download failed', code: 'INTERNAL_ERROR', status: response.status })
    return response.blob()
  }

  return {
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
    download,
  }
}
