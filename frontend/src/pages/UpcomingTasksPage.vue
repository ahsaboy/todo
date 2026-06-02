<script setup lang="ts">
import { computed, onMounted } from 'vue'
import { RefreshCw } from 'lucide-vue-next'
import { useTasks } from '@/features/tasks/useTasks'
import { useTaskToggle } from '@/features/tasks/useTaskToggle'
import PageShell from '@/shared/ui/PageShell.vue'
import TaskGroupedList from '@/features/tasks/TaskGroupedList.vue'
import FocusDurationDialog from '@/features/tasks/FocusDurationDialog.vue'
import TaskGroupedSkeleton from '@/shared/ui/TaskGroupedSkeleton.vue'
import type { Task } from '@/entities/task/model'

const { tasks, loading, error, fetchTasks, toggleComplete } = useTasks()
const { focusDialogVisible, focusDialogTaskTitle, handleToggle, handleFocusConfirm } =
  useTaskToggle({ tasks, toggleComplete })

const upcomingGroups = computed(() => {
  const now = new Date()
  const today = new Date(now.getFullYear(), now.getMonth(), now.getDate())
  const tomorrow = new Date(today)
  tomorrow.setDate(tomorrow.getDate() + 1)
  const endOfWeek = new Date(today)
  endOfWeek.setDate(endOfWeek.getDate() + (7 - endOfWeek.getDay()))
  const endOfNextWeek = new Date(endOfWeek)
  endOfNextWeek.setDate(endOfNextWeek.getDate() + 7)

  const tomorrowTasks: Task[] = []
  const thisWeekTasks: Task[] = []
  const nextWeekTasks: Task[] = []
  const laterTasks: Task[] = []

  for (const task of tasks.value) {
    if (task.completed || !task.dueAt) continue
    const dueDate = new Date(task.dueAt)
    const dueDay = new Date(dueDate.getFullYear(), dueDate.getMonth(), dueDate.getDate())
    if (dueDay < today) continue

    if (dueDay.getTime() === tomorrow.getTime()) tomorrowTasks.push(task)
    else if (dueDay < endOfWeek) thisWeekTasks.push(task)
    else if (dueDay < endOfNextWeek) nextWeekTasks.push(task)
    else laterTasks.push(task)
  }

  return [
    { label: '明天', tasks: tomorrowTasks },
    { label: '本周', tasks: thisWeekTasks },
    { label: '下周', tasks: nextWeekTasks },
    { label: '更晚', tasks: laterTasks },
  ].filter((g) => g.tasks.length > 0)
})

onMounted(() => fetchTasks())
</script>

<template>
  <div class="page">
    <div class="page-header">
      <h2>即将到期</h2>
      <button class="btn-secondary" type="button" @click="fetchTasks"><RefreshCw :size="14" /></button>
    </div>

    <PageShell
      :loading="loading"
      :error="error"
      :empty="upcomingGroups.length === 0"
      :skeleton="TaskGroupedSkeleton"
      empty-title="暂无即将到期的任务"
      :error-retry="fetchTasks"
    >
      <TaskGroupedList :groups="upcomingGroups" @toggle="handleToggle" />
    </PageShell>

    <FocusDurationDialog
      v-model:visible="focusDialogVisible"
      :task-title="focusDialogTaskTitle"
      @confirm="handleFocusConfirm"
    />
  </div>
</template>
