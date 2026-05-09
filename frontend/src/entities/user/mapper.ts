import type { UserDto, User } from './model'

export function toUser(dto: UserDto): User {
  return {
    createdAt: dto.created_at,
    email: dto.email,
    id: dto.id,
    username: dto.username,
  }
}
