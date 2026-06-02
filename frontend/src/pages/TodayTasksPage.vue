<script setup lang="ts">
import { computed, onMounted } from 'vue'
import { RefreshCw } from 'lucide-vue-next'
import { useTasks } from '@/features/tasks/useTasks'
import { useTaskToggle } from '@/features/tasks/useTaskToggle'
import PageShell from '@/shared/ui/PageShell.vue'
import QuickCreateTask from '@/features/tasks/QuickCreateTask.vue'
import TaskGroupedList from '@/features/tasks/TaskGroupedList.vue'
import FocusDurationDialog from '@/features/tasks/FocusDurationDialog.vue'
import TaskGroupedSkeleton from '@/shared/ui/TaskGroupedSkeleton.vue'
import type { Task } from '@/entities/task/model'

const { tasks, loading, error, fetchTasks, createTask, toggleComplete } = useTasks()
const { focusDialogVisible, focusDialogTaskTitle, handleToggle, handleFocusConfirm } =
  useTaskToggle({ tasks, toggleComplete })

const todayGroups = computed(() => {
  const now = new Date()
  const today = new Date(now.getFullYear(), now.getMonth(), now.getDate())
  const tomorrow = new Date(today)
  tomorrow.setDate(tomorrow.getDate() + 1)

  const overdue: Task[] = []
  const todayDue: Task[] = []
  const todayRemind: Task[] = []
  const pendingNoDue: Task[] = []

  for (const task of tasks.value) {
    if (task.completed) continue
    const dueDate = task.dueAt ? new Date(task.dueAt) : null
    const remindDate = task.remindAt ? new Date(task.remindAt) : null

    if (dueDate && dueDate < today) overdue.push(task)
    else if (dueDate && dueDate < tomorrow) todayDue.push(task)
    else if (remindDate && remindDate < tomorrow) todayRemind.push(task)
    else if (!dueDate) pendingNoDue.push(task)
  }

  return [
    { label: '逾期', tasks: overdue },
    { label: '今天截止', tasks: todayDue },
    { label: '今天提醒', tasks: todayRemind },
    { label: '无截止但待处理', tasks: pendingNoDue },
  ].filter((g) => g.tasks.length > 0)
})

onMounted(() => fetchTasks())
</script>

<template>
  <div class="page">
    <div class="page-header">
      <h2>今日任务</h2>
      <button class="btn-secondary" type="button" @click="fetchTasks"><RefreshCw :size="14" /></button>
    </div>

    <QuickCreateTask @create="(title) => createTask({ title })" />

    <PageShell
      :loading="loading"
      :error="error"
      :empty="todayGroups.length === 0"
      :skeleton="TaskGroupedSkeleton"
      empty-title="今日暂无任务"
      :error-retry="fetchTasks"
    >
      <TaskGroupedList :groups="todayGroups" @toggle="handleToggle" />
    </PageShell>

    <FocusDurationDialog
      v-model:visible="focusDialogVisible"
      :task-title="focusDialogTaskTitle"
      @confirm="handleFocusConfirm"
    />
  </div>
</template>
