export interface ApiResponse<T> {
  success: boolean
  data: T
}

export interface ApiErrorResponse {
  success: boolean
  error: string
  code: string
}

export interface PageMeta {
  page: number
  limit: number
  total_items: number
  total_pages: number
}

export interface PaginatedResponse<T> {
  success: boolean
  data: T[]
  meta: PageMeta
}
