import type { ReminderConfigDto, ReminderConfig, ReminderLogDto, ReminderLog } from './model'

export function toReminderConfig(dto: ReminderConfigDto): ReminderConfig {
  return {
    channelType: dto.channel_type,
    createdAt: dto.created_at,
    enabled: dto.enabled,
    id: dto.id,
    maxRetries: dto.max_retries,
    name: dto.name,
    retryDelaySeconds: dto.retry_delay_seconds,
    updatedAt: dto.updated_at,
    userId: dto.user_id,
    webhookBodyTemplate: dto.webhook_body_template,
    webhookHeaders: dto.webhook_headers ?? {},
    webhookMethod: dto.webhook_method,
    webhookUrl: dto.webhook_url,
  }
}

export function toReminderLog(dto: ReminderLogDto): ReminderLog {
  return {
    attempts: dto.attempts,
    channelName: dto.channel_name,
    channelType: dto.channel_type,
    createdAt: dto.created_at,
    errorMessage: dto.error_message,
    id: dto.id,
    reminderConfigId: dto.reminder_config_id,
    status: dto.status,
    taskId: dto.task_id,
    taskTitle: dto.task_title,
    userId: dto.user_id,
  }
}
