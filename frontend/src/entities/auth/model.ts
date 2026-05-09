export interface LoginPayload {
  account: string
  password: string
}

export interface RegisterPayload {
  email?: string
  password: string
  username: string
}

export interface AuthResponse {
  api_key: string
  user: {
    created_at: string
    email: string
    id: number
    username: string
  }
}
