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
      <button class="task-body" type="button" @click="$emit('open', task)">
        <div class="task-top">
          <span class="task-title">{{ task.title }}</span>
          <PriorityTag :priority="task.priority" />
        </div>
        <div v-if="task.dueAt" class="task-due" :class="{ overdue: isOverdue(task) }">
          {{ formatDue(task.dueAt) }}
        </div>
      </button>
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
}>()

function isOverdue(task: Task): boolean {
  if (task.completed || !task.dueAt) return false
  return new Date(task.dueAt) < new Date()
}

function formatDue(dateStr: string): string {
  const date = new Date(dateStr)
  const now = new Date()
  const isToday = date.toDateString() === now.toDateString()
  const tomorrow = new Date(now)
  tomorrow.setDate(tomorrow.getDate() + 1)
  const isTomorrow = date.toDateString() === tomorrow.toDateString()

  const time = date.toLocaleTimeString('zh-CN', { hour: '2-digit', minute: '2-digit' })
  if (isToday) return `今天 ${time}`
  if (isTomorrow) return `明天 ${time}`
  return `${date.getMonth() + 1}月${date.getDate()}日 ${time}`
}
</script>

<style scoped>
.task-card-list {
  display: flex;
  flex-direction: column;
  gap: 8px;
  padding-bottom: 88px;
  min-width: 0;
}

.task-card {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 12px;
  background: var(--color-surface);
  border: 1px solid var(--color-border);
  border-radius: 8px;
  min-width: 0;
}

.task-card.completed {
  opacity: 0.6;
}

.task-checkbox {
  flex: none;
  width: 18px;
  height: 18px;
  accent-color: var(--color-primary);
}

.task-body {
  flex: 1;
  min-width: 0;
  display: flex;
  flex-direction: column;
  gap: 4px;
  padding: 0;
  margin: 0;
  background: none;
  border: none;
  text-align: left;
  color: inherit;
  cursor: pointer;
}

.task-top {
  display: flex;
  align-items: center;
  gap: 8px;
  min-width: 0;
}

.task-title {
  flex: 1;
  min-width: 0;
  font-size: 14px;
  line-height: 1.4;
  color: var(--color-text);
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.task-card.completed .task-title {
  text-decoration: line-through;
}

.task-due {
  font-size: 12px;
  color: var(--color-text-muted);
}

.task-due.overdue {
  color: var(--color-danger);
  font-weight: 500;
}

@media (max-width: 359px) {
  .task-card-list {
    padding-bottom: 72px;
  }
}
</style>
