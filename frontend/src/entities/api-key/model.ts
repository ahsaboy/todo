export interface ApiKeyInfoDto {
  created_at: string
  id: number
  last_used_at: string
  name: string
}

export interface ApiKeyInfo {
  createdAt: string
  id: number
  lastUsedAt: string
  name: string
}

export interface ApiKeyResponseDto {
  created_at: string
  id: number
  key: string
  name: string
}

export interface ApiKeyResponse {
  createdAt: string
  id: number
  key: string
  name: string
}

export interface CreateKeyPayload {
  name?: string
}
