<script setup lang="ts">
import { onMounted } from 'vue'
import { RefreshCw } from 'lucide-vue-next'
import { useTasks } from '@/features/tasks/useTasks'
import PageShell from '@/shared/ui/PageShell.vue'
import TaskBoard from '@/features/tasks/TaskBoard.vue'
import TaskBoardSkeleton from '@/shared/ui/TaskBoardSkeleton.vue'

const { tasks, loading, error, fetchTasks } = useTasks()

onMounted(() => fetchTasks())

function openTask(_task: { id: number }) {
  // TODO: 打开任务详情抽屉
  void _task
}
</script>

<template>
  <div class="page">
    <div class="page-header">
      <h2>看板</h2>
      <button class="btn-secondary" type="button" @click="fetchTasks"><RefreshCw :size="14" /></button>
    </div>

    <PageShell
      :loading="loading"
      :error="error"
      :empty="tasks.length === 0"
      :skeleton="TaskBoardSkeleton"
      empty-title="暂无任务"
      :error-retry="fetchTasks"
    >
      <TaskBoard :tasks="tasks" @card-click="openTask" />
    </PageShell>
  </div>
</template>
