<template>
  <TransitionGroup tag="div" name="list-move" class="task-card-list">
    <div
      v-for="task in tasks"
      :key="task.id"
      class="task-card"
      :class="{ completed: task.completed }"
      @click="$emit('open', task)"
    >
      <input
        :id="`task-card-completed-${task.id}`"
        :name="`task_card_completed_${task.id}`"
        type="checkbox"
        :checked="task.completed"
        class="checkbox-circle"
        @click.stop
        @change="$emit('toggle', task.id)"
      />
      <label class="sr-only" :for="`task-card-completed-${task.id}`">
        {{ task.completed ? '标记为未完成' : '标记为完成' }}：{{ task.title }}
      </label>
      <div class="task-body">
        <div class="task-top">
          <span class="task-title">{{ task.title }}</span>
          <PriorityTag :priority="task.priority" />
        </div>
        <div v-if="task.dueAt" class="task-due" :class="{ overdue: isOverdue(task) }">
          {{ formatDue(task.dueAt) }}
        </div>
      </div>
    </div>
  </TransitionGroup>
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

  const weekdays = ['周日', '周一', '周二', '周三', '周四', '周五', '周六']
  const weekday = weekdays[date.getDay()]
  const time = `${String(date.getHours()).padStart(2, '0')}:${String(date.getMinutes()).padStart(2, '0')}`
  if (isToday) return `今天 ${time} · ${weekday}`
  if (isTomorrow) return `明天 ${time} · ${weekday}`
  return `${date.getMonth() + 1}月${date.getDate()}日 · ${weekday} ${time} `
}
</script>

<style scoped>
.task-card-list {
  display: flex;
  flex-direction: column;
  gap: 12px;
  min-width: 0;
  position: relative;
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
  cursor: pointer;
  transition: background-color var(--motion-duration-fast) var(--motion-ease-standard), border-color var(--motion-duration-fast) var(--motion-ease-standard);
}

.task-card:hover {
  background-color: var(--color-surface-muted);
}

.task-card.completed {
  opacity: 0.6;
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
