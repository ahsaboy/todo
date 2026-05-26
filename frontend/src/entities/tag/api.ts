import type { ApiResponse } from '@/shared/api/types'
import { api } from '@/shared/api/client'
import type { CreateTagPayload, TagDto, UpdateTagPayload } from './model'

export function getTags() {
  return api.get<ApiResponse<TagDto[]>>('/tags')
}

export function getTag(id: number) {
  return api.get<ApiResponse<TagDto>>(`/tags/${id}`)
}

export function createTag(payload: CreateTagPayload) {
  return api.post<ApiResponse<TagDto>>('/tags', payload)
}

export function updateTag(id: number, payload: UpdateTagPayload) {
  return api.put<ApiResponse<TagDto>>(`/tags/${id}`, payload)
}

export function deleteTag(id: number) {
  return api.delete<ApiResponse<{ tasks_affected: number }>>(`/tags/${id}`)
}
