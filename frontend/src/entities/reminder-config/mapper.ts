import type { ReminderConfigDto, ReminderConfig } from './model'

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
