// DTO (后端 snake_case)
export interface TagDto {
  id: number
  user_id: number
  name: string
  color: string
  icon: string
  sort_order: number
  created_at: string
  updated_at: string
}

// Model (前端 camelCase)
export interface Tag {
  id: number
  userId: number
  name: string
  color: string
  icon: string
  sortOrder: number
  createdAt: string
  updatedAt: string
}

export interface CreateTagPayload {
  name: string
  color?: string
  icon?: string
  sort_order?: number
}

export interface UpdateTagPayload {
  name?: string
  color?: string
  icon?: string
  sort_order?: number
}
