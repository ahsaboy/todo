const DEFAULT_API_BASE_URL = 'api/v1'
const DEFAULT_ADMIN_API_BASE_URL = 'admin/api'

function normalizeApiBaseUrl(baseUrl: string): string {
  return baseUrl.replace(/\/+$/, '')
}

export const API_BASE_URL = normalizeApiBaseUrl(
  import.meta.env.VITE_API_BASE_URL?.trim() || DEFAULT_API_BASE_URL,
)

export const ADMIN_API_BASE_URL = normalizeApiBaseUrl(
  import.meta.env.VITE_ADMIN_API_BASE_URL?.trim() || DEFAULT_ADMIN_API_BASE_URL,
)
