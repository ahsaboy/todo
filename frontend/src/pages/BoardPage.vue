<template>
  <div class="page">
    <h2>看板</h2>

    <div v-if="loading" class="page-loading">加载中...</div>

    <div v-else-if="error" class="page-error">
      <p>{{ error }}</p>
      <button @click="fetchTasks">重试</button>
    </div>

    <div v-else-if="tasks.length === 0" class="page-empty">
      <p>暂无任务</p>
    </div>

    <TaskBoard v-else :tasks="tasks" @cardClick="openTask" />
  </div>
</template>

<script setup lang="ts">
import { onMounted } from 'vue'
import { useTasks } from '@/features/tasks/useTasks'
import TaskBoard from '@/features/tasks/TaskBoard.vue'

const { tasks, loading, error, fetchTasks } = useTasks()

onMounted(() => {
  fetchTasks()
})

function openTask(_task: { id: number }) {
  // TODO: 打开任务详情抽屉
  void _task
}
</script>

<style scoped>
.page h2 {
  margin: 0;
  font-size: 20px;
}
</style>
