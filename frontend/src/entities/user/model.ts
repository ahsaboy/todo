export interface UserDto {
  created_at: string
  email: string
  id: number
  username: string
  avatar_url?: string
}

export interface User {
  createdAt: string
  email: string
  id: number
  username: string
  avatarUrl?: string
}

export interface UpdateProfilePayload {
  email?: string
  code?: string
}

export interface ChangePasswordPayload {
  new_password: string
  old_password: string
}

export interface OAuthAccountDto {
  id: number
  provider: string
  display_name: string
  avatar_url: string
  linked_at: string
}

export interface OAuthAccount {
  id: number
  provider: string
  displayName: string
  avatarUrl: string
  linkedAt: string
}

export interface ProfileResponse {
  user: UserDto
  oauth_accounts: OAuthAccountDto[]
  has_password: boolean
}

export interface Profile {
  user: User
  oauthAccounts: OAuthAccount[]
  hasPassword: boolean
}
