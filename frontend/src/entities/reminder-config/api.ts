import type { ApiResponse } from '@/shared/api/types'
import { api } from '@/shared/api/client'
import type {
  ReminderConfigDto,
  ReminderTemplatesDto,
  CreateReminderConfigPayload,
  UpdateReminderConfigPayload,
} from './model'

export function getReminderTemplates() {
  return api.get<ApiResponse<ReminderTemplatesDto>>('/templates')
}

export function getReminderConfigs() {
  return api.get<ApiResponse<ReminderConfigDto[]>>('/user/reminder-configs')
}

export function getReminderConfig(id: number) {
  return api.get<ApiResponse<ReminderConfigDto>>(`/user/reminder-configs/${id}`)
}

export function createReminderConfig(payload: CreateReminderConfigPayload) {
  return api.post<ApiResponse<ReminderConfigDto>>('/user/reminder-configs', payload)
}

export function updateReminderConfig(id: number, payload: UpdateReminderConfigPayload) {
  return api.put<ApiResponse<ReminderConfigDto>>(`/user/reminder-configs/${id}`, payload)
}

export function deleteReminderConfig(id: number) {
  return api.delete<ApiResponse<void>>(`/user/reminder-configs/${id}`)
}
