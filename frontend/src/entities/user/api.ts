import { api } from '@/shared/api/client'
import type { ApiResponse } from '@/shared/api/types'
import type { ProfileResponse, UpdateProfilePayload, ChangePasswordPayload } from './model'

export function getProfile() {
  return api.get<ApiResponse<ProfileResponse>>('/user/profile')
}

export function updateProfile(payload: UpdateProfilePayload) {
  return api.put<ApiResponse<ProfileResponse>>('/user/profile', payload)
}

export function changePassword(payload: ChangePasswordPayload) {
  return api.put<ApiResponse<void>>('/user/password', payload)
}

export function unlinkOAuthAccount(id: number) {
  return api.delete<ApiResponse<void>>(`/user/oauth-accounts/${id}`)
}

export function setPassword(newPassword: string) {
  return api.post<ApiResponse<void>>('/user/set-password', { new_password: newPassword })
}
