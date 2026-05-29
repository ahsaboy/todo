<template>
  <div class="board-card" :class="{ completed: task.completed }" @click="$emit('click', task)">
    <div class="card-title">{{ task.title }}</div>
    <div v-if="task.tags && task.tags.length > 0" class="card-tags">
      <TagChip v-for="t in task.tags" :key="t" :name="t" />
    </div>
    <div class="card-meta">
      <PriorityTag v-if="task.priority" :priority="task.priority" />
      <span v-if="task.focusDuration" class="focus-tag">
        专注 {{ task.focusDuration }} min
      </span>
      <span v-if="task.dueAt" class="due-date">
        {{ formatDate(task.dueAt) }}
      </span>
    </div>
  </div>
</template>

<script setup lang="ts">
import type { Task } from '@/entities/task/model'
import PriorityTag from '@/shared/ui/PriorityTag.vue'
import TagChip from '@/features/tags/TagChip.vue'
import { formatDateShort as formatDate } from '@/shared/utils/date'

defineProps<{
  task: Task
}>()

defineEmits<{
  click: [task: Task]
}>()
</script>

<style scoped>
.board-card {
  padding: 12px;
  background: var(--color-surface);
  border: 1px solid var(--color-border);
  border-radius: 6px;
  cursor: pointer;
  transition:
    box-shadow var(--motion-duration-base),
    border-color var(--motion-duration-base),
    transform var(--motion-duration-base);
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

.card-tags {
  display: flex;
  flex-wrap: wrap;
  gap: 4px;
  margin-bottom: 8px;
}

.card-meta {
  display: flex;
  gap: 8px;
  font-size: 12px;
  color: var(--color-text-muted);
}

.focus-tag {
  color: var(--color-primary);
}
</style>
