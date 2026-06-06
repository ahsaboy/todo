<template>
  <TransitionGroup tag="div" name="list-move" class="task-card-list" appear>
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
        <div v-if="task.tags && task.tags.length > 0" class="task-tags">
          <TagChip v-for="t in task.tags" :key="t" :name="t" />
        </div>
        <div v-if="task.dueAt" class="task-due" :class="{ overdue: isOverdue(task.dueAt, task.completed) }">
          {{ formatDue(task.dueAt) }}
        </div>
        <div v-if="task.focusDuration" class="task-focus">
          专注 {{ task.focusDuration }} min
        </div>
      </div>
    </div>
  </TransitionGroup>
</template>

<script setup lang="ts">
import type { Task } from '@/entities/task/model'
import PriorityTag from '@/shared/ui/PriorityTag.vue'
import TagChip from '@/features/tags/TagChip.vue'
import { formatDueRelative as formatDue, isOverdue } from '@/shared/utils/date'

defineProps<{
  tasks: Task[]
}>()

defineEmits<{
  toggle: [id: number]
  open: [task: Task]
}>()
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
  transition:
    background-color var(--motion-duration-fast) var(--motion-ease-standard),
    border-color var(--motion-duration-fast) var(--motion-ease-standard),
    transform var(--motion-duration-fast) var(--motion-ease-standard),
    box-shadow var(--motion-duration-fast) var(--motion-ease-standard);
}

@media (hover: hover) {
  .task-card:hover {
    background-color: var(--color-surface-muted);
    transform: translateY(-1px);
    box-shadow: var(--shadow-panel);
  }
}

@media (hover: none) {
  .task-card:active {
    transform: scale(0.985);
    background-color: var(--color-surface-muted);
  }
}

.task-card.completed {
  opacity: 0.6;
  animation: completion-pulse var(--motion-duration-base) var(--motion-ease-emphasized);
}

@keyframes completion-pulse {
  0% {
    box-shadow: 0 0 0 0 color-mix(in srgb, var(--color-success) 40%, transparent);
    transform: scale(1);
  }
  50% {
    box-shadow: 0 0 0 8px color-mix(in srgb, var(--color-success) 0%, transparent);
    transform: scale(1.02);
  }
  100% {
    box-shadow: 0 0 0 0 color-mix(in srgb, var(--color-success) 0%, transparent);
    transform: scale(1);
  }
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

.task-tags {
  display: flex;
  flex-wrap: wrap;
  gap: 4px;
}

.task-due {
  font-size: 12px;
  color: var(--color-text-muted);
}

.task-due.overdue {
  color: var(--color-danger);
  font-weight: 500;
}

.task-focus {
  font-size: 12px;
  color: var(--color-primary);
}

@media (max-width: 359px) {
  .task-card-list {
    padding-bottom: 72px;
  }
}
</style>
