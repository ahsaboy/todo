<template>
  <div class="page">
    <h2>今日任务</h2>

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

        <TaskGroupedList v-else :groups="todayGroups" @toggle="toggleComplete" />
      </template>
    </Transition>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted } from 'vue'
import { useTasks } from '@/features/tasks/useTasks'
import QuickCreateTask from '@/features/tasks/QuickCreateTask.vue'
import TaskGroupedList from '@/features/tasks/TaskGroupedList.vue'
import TaskGroupedSkeleton from '@/shared/ui/TaskGroupedSkeleton.vue'
import type { Task } from '@/entities/task/model'

const { tasks, loading, error, fetchTasks, createTask, toggleComplete } = useTasks()

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
</script>

<style scoped>
.page h2 {
  margin: 0;
  font-size: 20px;
}
</style>
