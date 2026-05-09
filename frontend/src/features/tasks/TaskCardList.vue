<template>
  <div class="task-card-list">
    <div
      v-for="task in tasks"
      :key="task.id"
      class="task-card"
      :class="{ completed: task.completed }"
    >
      <input
        :id="`task-card-completed-${task.id}`"
        :name="`task_card_completed_${task.id}`"
        type="checkbox"
        :checked="task.completed"
        class="task-checkbox"
        @change="$emit('toggle', task.id)"
      />
      <label class="sr-only" :for="`task-card-completed-${task.id}`">
        {{ task.completed ? '标记为未完成' : '标记为完成' }}：{{ task.title }}
      </label>
      <div class="task-content" @click="$emit('open', task)">
        <div class="task-title">{{ task.title }}</div>
        <div v-if="task.description" class="task-desc">
          {{ truncate(task.description, 50) }}
        </div>
        <div class="task-meta">
          <PriorityTag v-if="task.priority" :priority="task.priority" />
          <span v-if="task.dueAt" class="meta-item">
            {{ formatDate(task.dueAt) }}
          </span>
        </div>
      </div>
      <button class="btn-more" type="button" @click="$emit('more', task)">···</button>
    </div>
  </div>
</template>

<script setup lang="ts">
import type { Task } from '@/entities/task/model'
import PriorityTag from '@/shared/ui/PriorityTag.vue'

defineProps<{
  tasks: Task[]
}>()

defineEmits<{
  toggle: [id: number]
  open: [task: Task]
  more: [task: Task]
}>()

function truncate(str: string, len: number): string {
  return str.length > len ? str.slice(0, len) + '...' : str
}

function formatDate(dateStr: string): string {
  if (!dateStr) return ''
  const date = new Date(dateStr)
  return date.toLocaleDateString('zh-CN', { month: '2-digit', day: '2-digit' })
}
</script>

<style scoped>
.task-card-list {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.task-card {
  display: flex;
  align-items: flex-start;
  gap: 12px;
  padding: 12px;
  background: var(--color-surface);
  border: 1px solid var(--color-border);
  border-radius: 8px;
}

.task-card.completed {
  opacity: 0.6;
}

.task-checkbox {
  margin-top: 2px;
}

.task-content {
  flex: 1;
  min-width: 0;
  cursor: pointer;
}

.task-title {
  font-size: 14px;
  color: var(--color-text);
}

.task-card.completed .task-title {
  text-decoration: line-through;
}

.task-desc {
  font-size: 12px;
  color: var(--color-text-muted);
  margin-top: 4px;
}

.task-meta {
  display: flex;
  gap: 8px;
  margin-top: 6px;
  font-size: 12px;
  color: var(--color-text-muted);
}

.btn-more {
  background: none;
  border: none;
  padding: 4px 8px;
  cursor: pointer;
  color: var(--color-text-muted);
}
</style>
