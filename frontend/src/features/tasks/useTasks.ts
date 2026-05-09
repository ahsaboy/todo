import { ref, reactive } from 'vue'
import type { Task, CreateTaskPayload, UpdateTaskPayload } from '@/entities/task/model'
import { toTask } from '@/entities/task/mapper'
import * as taskApi from '@/entities/task/api'
import type { PageMeta } from '@/shared/api/types'

export interface TaskFilters {
  status: 'all' | 'completed' | 'pending'
  priority?: 1 | 2 | 3
  search: string
}

export interface TaskSort {
  field: string
  order: 'asc' | 'desc'
}

export function useTasks() {
  const tasks = ref<Task[]>([])
  const loading = ref(false)
  const error = ref<string | null>(null)

  const meta = ref<PageMeta>({
    page: 1,
    limit: 20,
    total_items: 0,
    total_pages: 0,
  })

  const filters = reactive<TaskFilters>({
    status: 'all',
    search: '',
  })

  const sort = reactive<TaskSort>({
    field: 'created_at',
    order: 'desc',
  })

  async function fetchTasks() {
    loading.value = true
    error.value = null

    try {
      const response = await taskApi.getTasks({
        page: meta.value.page,
        limit: meta.value.limit,
        status: filters.status,
        priority: filters.priority,
        search: filters.search,
        sort: sort.field,
        order: sort.order,
      })

      tasks.value = response.data.map(toTask)
      meta.value = response.meta
    } catch (e) {
      error.value = e instanceof Error ? e.message : 'Failed to fetch tasks'
    } finally {
      loading.value = false
    }
  }

  async function createTask(payload: CreateTaskPayload) {
    const response = await taskApi.createTask(payload)
    const newTask = toTask(response.data)
    tasks.value.unshift(newTask)
    meta.value.total_items++
    return newTask
  }

  async function updateTask(id: number, payload: UpdateTaskPayload) {
    const response = await taskApi.updateTask(id, payload)
    const updatedTask = toTask(response.data)
    const index = tasks.value.findIndex((t) => t.id === id)
    if (index !== -1) {
      tasks.value[index] = updatedTask
    }
    return updatedTask
  }

  async function deleteTask(id: number) {
    await taskApi.deleteTask(id)
    tasks.value = tasks.value.filter((t) => t.id !== id)
    meta.value.total_items--
  }

  async function toggleComplete(id: number) {
    const task = tasks.value.find((t) => t.id === id)
    if (!task) return

    // Optimistic update
    const originalCompleted = task.completed
    task.completed = !task.completed

    try {
      const response = await taskApi.toggleTaskComplete(id)
      const updatedTask = toTask(response.data)
      const index = tasks.value.findIndex((t) => t.id === id)
      if (index !== -1) {
        tasks.value[index] = updatedTask
      }
    } catch {
      // Rollback on failure
      task.completed = originalCompleted
      throw new Error('Failed to toggle task completion')
    }
  }

  function setPage(page: number) {
    meta.value.page = page
    fetchTasks()
  }

  function setFilters(newFilters: Partial<TaskFilters>) {
    Object.assign(filters, newFilters)
    meta.value.page = 1
    fetchTasks()
  }

  function setSort(field: string, order: 'asc' | 'desc') {
    sort.field = field
    sort.order = order
    meta.value.page = 1
    fetchTasks()
  }

  return {
    tasks,
    loading,
    error,
    meta,
    filters,
    sort,
    fetchTasks,
    createTask,
    updateTask,
    deleteTask,
    toggleComplete,
    setPage,
    setFilters,
    setSort,
  }
}
