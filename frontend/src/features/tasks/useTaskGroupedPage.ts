import { onMounted, type Ref, type ComputedRef } from 'vue'
import { useTaskToggle } from './useTaskToggle'
import type { Task, CreateTaskPayload } from '@/entities/task/model'

export interface TaskGroup {
  label: string
  tasks: Task[]
}

interface TaskGroupedPageConfig {
  tasks: Ref<Task[]>
  loading: Ref<boolean>
  error: Ref<string | null>
  fetchTasks: () => Promise<void>
  createTask: (payload: CreateTaskPayload) => Promise<void>
  toggleComplete: (id: number, focusDuration?: number | null) => Promise<void>
  groups: ComputedRef<TaskGroup[]>
}

export function useTaskGroupedPage(config: TaskGroupedPageConfig) {
  const { tasks, loading, error, fetchTasks, createTask, toggleComplete, groups } = config

  const { focusDialogVisible, focusDialogTaskTitle, handleToggle, handleFocusConfirm } =
    useTaskToggle({ tasks, toggleComplete })

  onMounted(() => fetchTasks())

  return {
    tasks,
    loading,
    error,
    fetchTasks,
    createTask,
    groups,
    handleToggle,
    focusDialogVisible,
    focusDialogTaskTitle,
    handleFocusConfirm,
  }
}
