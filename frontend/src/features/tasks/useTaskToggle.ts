import { ref } from 'vue'
import type { Task } from '@/entities/task/model'

export interface TaskToggleDeps {
  tasks: { value: Task[] }
  toggleComplete: (id: number, focusDuration?: number | null) => Promise<void>
}

export function useTaskToggle(deps: TaskToggleDeps) {
  const focusDialogVisible = ref(false)
  const focusDialogTaskTitle = ref('')
  const pendingToggleTaskId = ref<number | null>(null)

  function handleToggle(id: number) {
    const task = deps.tasks.value.find((t) => t.id === id)
    if (!task) return

    if (!task.completed) {
      pendingToggleTaskId.value = id
      focusDialogTaskTitle.value = task.title
      focusDialogVisible.value = true
    } else {
      deps.toggleComplete(id)
    }
  }

  async function handleFocusConfirm(duration: number | null) {
    if (pendingToggleTaskId.value == null) return
    const id = pendingToggleTaskId.value
    pendingToggleTaskId.value = null

    try {
      await deps.toggleComplete(id, duration ?? undefined)
    } catch {
      // toggleComplete already handles rollback
    }
  }

  return {
    focusDialogVisible,
    focusDialogTaskTitle,
    handleToggle,
    handleFocusConfirm,
  }
}
