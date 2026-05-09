import type { ApiKeyInfoDto, ApiKeyInfo, ApiKeyResponseDto, ApiKeyResponse } from './model'

export function toApiKeyInfo(dto: ApiKeyInfoDto): ApiKeyInfo {
  return {
    createdAt: dto.created_at,
    id: dto.id,
    lastUsedAt: dto.last_used_at,
    name: dto.name,
  }
}

export function toApiKeyResponse(dto: ApiKeyResponseDto): ApiKeyResponse {
  return {
    createdAt: dto.created_at,
    id: dto.id,
    key: dto.key,
    name: dto.name,
  }
}
