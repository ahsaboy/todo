import type { Tag, TagDto } from './model'

export function toTag(dto: TagDto): Tag {
  return {
    id: dto.id,
    userId: dto.user_id,
    name: dto.name,
    color: dto.color,
    icon: dto.icon ?? '',
    sortOrder: dto.sort_order,
    createdAt: dto.created_at,
    updatedAt: dto.updated_at,
  }
}
