export interface UserDto {
  created_at: string
  email: string
  id: number
  username: string
}

export interface User {
  createdAt: string
  email: string
  id: number
  username: string
}

export interface UpdateProfilePayload {
  email?: string
}

export interface ChangePasswordPayload {
  new_password: string
  old_password: string
}
