<template>
  <div class="page">
    <div class="page-header">
      <h2>今日任务</h2>
      <button class="btn-secondary" type="button" @click="fetchTasks"><RefreshCw :size="14" /></button>
    </div>

    <QuickCreateTask @create="handleQuickCreate" />

    <Transition name="sk-fade" mode="out-in">
      <TaskGroupedSkeleton v-if="loading" key="skeleton" />

      <template v-else key="content">
        <div v-if="error" class="page-error">
          <p>{{ error }}</p>
          <button @click="fetchTasks">重试</button>
        </div>

        <div v-else-if="tasks.length === 0" class="page-empty">
          <p>今日暂无任务</p>
        </div>

        <TaskGroupedList v-else :groups="todayGroups" @toggle="handleToggle" />
      </template>
    </Transition>

    <!-- 专注时长对话框 -->
    <FocusDurationDialog
      v-model:visible="focusDialogVisible"
      :task-title="focusDialogTaskTitle"
      @confirm="handleFocusConfirm"
    />
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { RefreshCw } from 'lucide-vue-next'
import { useTasks } from '@/features/tasks/useTasks'
import QuickCreateTask from '@/features/tasks/QuickCreateTask.vue'
import TaskGroupedList from '@/features/tasks/TaskGroupedList.vue'
import FocusDurationDialog from '@/features/tasks/FocusDurationDialog.vue'
import TaskGroupedSkeleton from '@/shared/ui/TaskGroupedSkeleton.vue'
import type { Task } from '@/entities/task/model'
import { toggleTaskComplete as apiToggleComplete } from '@/entities/task/api'
import { toTask } from '@/entities/task/mapper'

const { tasks, loading, error, fetchTasks, createTask, toggleComplete } = useTasks()

const focusDialogVisible = ref(false)
const focusDialogTaskTitle = ref('')
const pendingToggleTaskId = ref<number | null>(null)

onMounted(() => {
  fetchTasks()
})

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

    if (dueDate && dueDate < today) {
      overdue.push(task)
    } else if (dueDate && dueDate < tomorrow) {
      todayDue.push(task)
    } else if (remindDate && remindDate < tomorrow) {
      todayRemind.push(task)
    } else if (!dueDate) {
      pendingNoDue.push(task)
    }
  }

  return [
    { label: '逾期', tasks: overdue },
    { label: '今天截止', tasks: todayDue },
    { label: '今天提醒', tasks: todayRemind },
    { label: '无截止但待处理', tasks: pendingNoDue },
  ].filter((group) => group.tasks.length > 0)
})

async function handleQuickCreate(title: string) {
  await createTask({ title })
}

function handleToggle(id: number) {
  const task = tasks.value.find((t) => t.id === id)
  if (!task) return

  if (!task.completed) {
    pendingToggleTaskId.value = id
    focusDialogTaskTitle.value = task.title
    focusDialogVisible.value = true
  } else {
    toggleComplete(id)
  }
}

async function handleFocusConfirm(duration: number | null) {
  if (pendingToggleTaskId.value == null) return
  const id = pendingToggleTaskId.value
  pendingToggleTaskId.value = null

  const task = tasks.value.find((t) => t.id === id)
  if (!task) return

  const originalCompleted = task.completed
  task.completed = !task.completed

  try {
    const response = await apiToggleComplete(id, duration != null ? duration : undefined)
    const updatedTask = toTask(response.data)
    const index = tasks.value.findIndex((t) => t.id === id)
    if (index !== -1) {
      tasks.value[index] = updatedTask
    }
  } catch {
    task.completed = originalCompleted
  }
}
</script>
