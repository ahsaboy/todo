import { api } from '@/shared/api/client'
import type { ApiResponse, PaginatedResponse } from '@/shared/api/types'
import type { TaskDto, CreateTaskPayload, UpdateTaskPayload } from './model'

export interface TaskListParams {
  page?: number
  limit?: number
  sort?: string
  order?: 'asc' | 'desc'
  status?: 'all' | 'completed' | 'pending'
  priority?: 1 | 2 | 3
  due_before?: string
  due_after?: string
  search?: string
}

export function getTasks(params?: TaskListParams) {
  const searchParams = new URLSearchParams()
  if (params) {
    Object.entries(params).forEach(([key, value]) => {
      if (value !== undefined && value !== null && value !== '') {
        searchParams.append(key, String(value))
      }
    })
  }
  const query = searchParams.toString()
  return api.get<PaginatedResponse<TaskDto>>(`/tasks${query ? `?${query}` : ''}`)
}

export function getTask(id: number) {
  return api.get<ApiResponse<TaskDto>>(`/tasks/${id}`)
}

export function createTask(payload: CreateTaskPayload) {
  return api.post<ApiResponse<TaskDto>>('/tasks', payload)
}

export function updateTask(id: number, payload: UpdateTaskPayload) {
  return api.put<ApiResponse<TaskDto>>(`/tasks/${id}`, payload)
}

export function deleteTask(id: number) {
  return api.delete<ApiResponse<void>>(`/tasks/${id}`)
}

export function toggleTaskComplete(id: number) {
  return api.patch<ApiResponse<TaskDto>>(`/tasks/${id}/complete`)
}
