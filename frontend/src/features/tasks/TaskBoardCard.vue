<template>
  <div class="board-card" :class="{ completed: task.completed }" @click="$emit('click', task)">
    <div class="card-title">{{ task.title }}</div>
    <div class="card-meta">
      <PriorityTag v-if="task.priority" :priority="task.priority" />
      <span v-if="task.dueAt" class="due-date">
        {{ formatDate(task.dueAt) }}
      </span>
    </div>
  </div>
</template>

<script setup lang="ts">
import type { Task } from '@/entities/task/model'
import PriorityTag from '@/shared/ui/PriorityTag.vue'

defineProps<{
  task: Task
}>()

defineEmits<{
  click: [task: Task]
}>()

function formatDate(dateStr: string): string {
  if (!dateStr) return ''
  const date = new Date(dateStr)
  return date.toLocaleDateString('zh-CN', { month: '2-digit', day: '2-digit' })
}
</script>

<style scoped>
.board-card {
  padding: 12px;
  background: var(--color-surface);
  border: 1px solid var(--color-border);
  border-radius: 6px;
  cursor: pointer;
  transition: box-shadow 200ms;
}

.board-card:hover {
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.1);
}

.board-card.completed {
  opacity: 0.6;
}

.card-title {
  font-size: 14px;
  color: var(--color-text);
  margin-bottom: 8px;
}

.board-card.completed .card-title {
  text-decoration: line-through;
}

.card-meta {
  display: flex;
  gap: 8px;
  font-size: 12px;
  color: var(--color-text-muted);
}
</style>
