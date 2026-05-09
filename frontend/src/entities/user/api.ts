import { api } from '@/shared/api/client'
import type { ApiResponse } from '@/shared/api/types'
import type { UserDto, UpdateProfilePayload, ChangePasswordPayload } from './model'

export function getProfile() {
  return api.get<ApiResponse<UserDto>>('/user/profile')
}

export function updateProfile(payload: UpdateProfilePayload) {
  return api.put<ApiResponse<UserDto>>('/user/profile', payload)
}

export function changePassword(payload: ChangePasswordPayload) {
  return api.put<ApiResponse<void>>('/user/password', payload)
}
