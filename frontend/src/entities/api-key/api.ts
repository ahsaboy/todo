import { api, type ApiResponse } from '@/shared/api/client'
import type { ApiKeyInfoDto, ApiKeyResponseDto, CreateKeyPayload } from './model'

export function getApiKeys() {
  return api.get<ApiResponse<ApiKeyInfoDto[]>>('/user/keys')
}

export function createApiKey(payload: CreateKeyPayload) {
  return api.post<ApiResponse<ApiKeyResponseDto>>('/user/keys', payload)
}

export function deleteApiKey(id: number) {
  return api.delete<ApiResponse<void>>(`/user/keys/${id}`)
}
