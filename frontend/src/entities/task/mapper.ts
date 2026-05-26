import type { TaskDto, Task } from './model'

export function toTask(dto: TaskDto): Task {
  return {
    completed: dto.completed,
    createdAt: dto.created_at,
    description: dto.description,
    dueAt: dto.due_at,
    focusDuration: dto.focus_duration,
    id: dto.id,
    priority: dto.priority,
    remindAt: dto.remind_at,
    reminderSent: dto.reminder_sent,
    reminderSentAt: dto.reminder_sent_at,
    repeatEndDate: dto.repeat_end_date,
    repeatInterval: dto.repeat_interval,
    repeatType: dto.repeat_type,
    tags: Array.isArray(dto.tags) ? dto.tags : [],
    title: dto.title,
    updatedAt: dto.updated_at,
    userId: dto.user_id,
  }
}
