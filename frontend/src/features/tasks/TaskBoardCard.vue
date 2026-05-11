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
  return `${date.getMonth() + 1}月${date.getDate()}日`
}
</script>

<style scoped>
.board-card {
  padding: 12px;
  background: var(--color-surface);
  border: 1px solid var(--color-border);
  border-radius: 6px;
  cursor: pointer;
  transition:
    box-shadow 200ms,
    border-color 200ms,
    transform 200ms;
}

.board-card:hover {
  border-color: color-mix(in srgb, var(--color-primary) 20%, var(--color-border));
  box-shadow: 0 10px 24px color-mix(in srgb, var(--color-primary) 10%, transparent);
  transform: translateY(-1px);
}

.board-card.completed {
  opacity: 0.72;
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
