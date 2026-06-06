import { api } from '@/shared/api/client'
import type { ApiResponse } from '@/shared/api/types'
import type { LoginPayload, RegisterPayload, AuthResponse, SendCodePayload, VerifyCodePayload, ResetPasswordPayload } from './model'

export function login(payload: LoginPayload) {
  return api.post<ApiResponse<AuthResponse>>('/auth/login', payload)
}

export function register(payload: RegisterPayload) {
  return api.post<ApiResponse<AuthResponse>>('/auth/register', payload)
}

export function getEmailStatus() {
  return api.get<ApiResponse<{ available: boolean }>>('/auth/email-status')
}

export function sendVerificationCode(payload: SendCodePayload) {
  return api.post<ApiResponse<{ sent: boolean }>>('/auth/send-code', payload)
}

export function verifyCode(payload: VerifyCodePayload) {
  return api.post<ApiResponse<{ verified: boolean }>>('/auth/verify-code', payload)
}

export function resetPassword(payload: ResetPasswordPayload) {
  return api.post<ApiResponse<{ reset: boolean }>>('/auth/reset-password', payload)
}
