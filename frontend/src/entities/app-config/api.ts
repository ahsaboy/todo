import type { ApiResponse } from '@/shared/api/types'
import { adminApi } from '@/shared/api/admin-client'
import type { ConfigGroup, ConfigGroupDto, ConfigUpdate, UpdateConfigResult } from './model'
import { toConfigGroup } from './mapper'

export async function getConfig(): Promise<ConfigGroup[]> {
  const res = await adminApi.get<ApiResponse<{ groups: ConfigGroupDto[] }>>('/config')
  return (res.data?.groups ?? []).map(toConfigGroup)
}

export async function updateConfig(updates: ConfigUpdate[]): Promise<UpdateConfigResult> {
  const res = await adminApi.put<ApiResponse<{ restart_required: boolean; updated: string[] }>>('/config', { updates })
  return {
    restartRequired: res.data?.restart_required ?? false,
    updated: res.data?.updated ?? [],
  }
}

export async function resetConfig(key: string): Promise<UpdateConfigResult> {
  const res = await adminApi.delete<ApiResponse<{ restart_required: boolean }>>(`/config/${key}`)
  return {
    restartRequired: res.data?.restart_required ?? true,
    updated: [],
  }
}

export async function testEmailConnection(): Promise<{ ok: boolean; message: string }> {
  const res = await adminApi.post<ApiResponse<{ ok: boolean; message: string }>>('/config/test-email')
  return res.data ?? { ok: false, message: '未知错误' }
}
