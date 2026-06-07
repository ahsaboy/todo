import type { UserDto, User, OAuthAccountDto, OAuthAccount, ProfileResponse, Profile } from './model'

export function toUser(dto: UserDto): User {
  return {
    createdAt: dto.created_at,
    email: dto.email,
    id: dto.id,
    username: dto.username,
    avatarUrl: dto.avatar_url,
  }
}

export function toOAuthAccount(dto: OAuthAccountDto): OAuthAccount {
  return {
    id: dto.id,
    provider: dto.provider,
    displayName: dto.display_name,
    avatarUrl: dto.avatar_url,
    linkedAt: dto.linked_at,
  }
}

export function toProfile(dto: ProfileResponse): Profile {
  return {
    user: toUser(dto.user),
    oauthAccounts: (dto.oauth_accounts ?? []).map(toOAuthAccount),
    hasPassword: dto.has_password,
  }
}
