export interface LoginPayload {
  account: string
  password: string
}

export interface RegisterPayload {
  email?: string
  password: string
  username: string
  code?: string
}

export interface AuthResponse {
  api_key: string
  user: {
    created_at: string
    email: string
    id: number
    username: string
    avatar_url?: string
  }
}

export interface SendCodePayload {
  email: string
  purpose: 'register' | 'reset_password' | 'change_email'
}

export interface VerifyCodePayload {
  email: string
  code: string
  purpose: 'register' | 'reset_password' | 'change_email'
}

export interface ResetPasswordPayload {
  email: string
  code: string
  password: string
}

export interface OAuthProvider {
  name: string
  label: string
  icon: string
}
