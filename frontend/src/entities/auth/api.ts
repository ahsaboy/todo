import { api } from '@/shared/api/client'
import type { ApiResponse } from '@/shared/api/types'
import type { LoginPayload, RegisterPayload, AuthResponse } from './model'

export function login(payload: LoginPayload) {
  return api.post<ApiResponse<AuthResponse>>('/auth/login', payload)
}

export function register(payload: RegisterPayload) {
  return api.post<ApiResponse<AuthResponse>>('/auth/register', payload)
}
