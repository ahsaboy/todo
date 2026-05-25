<template>
  <div class="page">
    <h2>即将到期</h2>

    <Transition name="sk-fade" mode="out-in">
      <TaskGroupedSkeleton v-if="loading" key="skeleton" />

      <template v-else key="content">
        <div v-if="error" class="page-error">
          <p>{{ error }}</p>
          <button @click="fetchTasks">重试</button>
        </div>

        <div v-else-if="tasks.length === 0" class="page-empty">
          <p>暂无即将到期的任务</p>
        </div>

        <TaskGroupedList v-else :groups="upcomingGroups" @toggle="handleToggle" />
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
import { useTasks } from '@/features/tasks/useTasks'
import TaskGroupedList from '@/features/tasks/TaskGroupedList.vue'
import FocusDurationDialog from '@/features/tasks/FocusDurationDialog.vue'
import TaskGroupedSkeleton from '@/shared/ui/TaskGroupedSkeleton.vue'
import type { Task } from '@/entities/task/model'
import { toggleTaskComplete as apiToggleComplete } from '@/entities/task/api'
import { toTask } from '@/entities/task/mapper'

const { tasks, loading, error, fetchTasks, toggleComplete } = useTasks()

const focusDialogVisible = ref(false)
const focusDialogTaskTitle = ref('')
const pendingToggleTaskId = ref<number | null>(null)

onMounted(() => {
  fetchTasks()
})

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

    if (dueDay < today) continue // 逾期不显示在即将到期

    if (dueDay.getTime() === tomorrow.getTime()) {
      tomorrowTasks.push(task)
    } else if (dueDay < endOfWeek) {
      thisWeekTasks.push(task)
    } else if (dueDay < endOfNextWeek) {
      nextWeekTasks.push(task)
    } else {
      laterTasks.push(task)
    }
  }

  return [
    { label: '明天', tasks: tomorrowTasks },
    { label: '本周', tasks: thisWeekTasks },
    { label: '下周', tasks: nextWeekTasks },
    { label: '更晚', tasks: laterTasks },
  ].filter((group) => group.tasks.length > 0)
})

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

<style scoped>
.page h2 {
  margin: 0;
  font-size: 20px;
}
</style>
