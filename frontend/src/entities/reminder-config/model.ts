export type ChannelType = 'webhook' | 'feishu' | 'dingtalk' | 'wecom' | 'slack'
export type WebhookMethod = 'GET' | 'POST' | 'PUT'

export interface ReminderConfigDto {
  channel_type: string
  created_at: string
  enabled: boolean
  id: number
  max_retries: number
  name: string
  retry_delay_seconds: number
  updated_at: string
  user_id: number
  webhook_body_template: string
  webhook_headers: Record<string, string> | null
  webhook_method: string
  webhook_url: string
}

export interface ReminderTemplateDto {
  channel_type: string
  webhook_body_template: string
  webhook_headers: Record<string, string> | null
  webhook_method: string
  webhook_url: string
}

export type ReminderTemplatesDto = Record<string, ReminderTemplateDto>

export interface ReminderConfig {
  channelType: string
  createdAt: string
  enabled: boolean
  id: number
  maxRetries: number
  name: string
  retryDelaySeconds: number
  updatedAt: string
  userId: number
  webhookBodyTemplate: string
  webhookHeaders: Record<string, string>
  webhookMethod: string
  webhookUrl: string
}

export interface ReminderLogDto {
  id: number
  user_id: number
  task_id: number
  task_title: string
  reminder_config_id: number | null
  channel_name: string
  channel_type: string
  status: 'success' | 'failed'
  attempts: number
  error_message: string
  created_at: string
}

export interface ReminderLog {
  id: number
  userId: number
  taskId: number
  taskTitle: string
  reminderConfigId: number | null
  channelName: string
  channelType: string
  status: 'success' | 'failed'
  attempts: number
  errorMessage: string
  createdAt: string
}

export interface CreateReminderConfigPayload {
  channel_type: ChannelType
  enabled?: boolean
  max_retries?: number
  name: string
  retry_delay_seconds?: number
  webhook_body_template?: string
  webhook_headers?: Record<string, string>
  webhook_method?: WebhookMethod
  webhook_url: string
}

export interface UpdateReminderConfigPayload {
  channel_type?: ChannelType
  enabled?: boolean
  max_retries?: number
  name?: string
  retry_delay_seconds?: number
  webhook_body_template?: string
  webhook_headers?: Record<string, string>
  webhook_method?: WebhookMethod
  webhook_url?: string
}
