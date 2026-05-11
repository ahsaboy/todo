<template>
  <div class="page">
    <h2>即将到期</h2>

    <div v-if="loading" class="page-loading">加载中...</div>

    <div v-else-if="error" class="page-error">
      <p>{{ error }}</p>
      <button @click="fetchTasks">重试</button>
    </div>

    <div v-else-if="tasks.length === 0" class="page-empty">
      <p>暂无即将到期的任务</p>
    </div>

    <TaskGroupedList v-else :groups="upcomingGroups" @toggle="toggleComplete" />
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted } from 'vue'
import { useTasks } from '@/features/tasks/useTasks'
import TaskGroupedList from '@/features/tasks/TaskGroupedList.vue'
import type { Task } from '@/entities/task/model'

const { tasks, loading, error, fetchTasks, toggleComplete } = useTasks()

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
</script>

<style scoped>
.page h2 {
  margin: 0;
  font-size: 20px;
}
</style>
