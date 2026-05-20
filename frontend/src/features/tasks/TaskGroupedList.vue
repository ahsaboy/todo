<template>
  <MotionStagger class="task-grouped-list">
    <div v-for="group in groups" :key="group.label" class="task-group">
      <div class="group-header" @click="toggleGroup(group.label)">
        <span class="group-arrow" :class="{ collapsed: collapsedGroups.has(group.label) }">▼</span>
        <span class="group-label">{{ group.label }}</span>
        <span class="group-count">{{ group.tasks.length }}</span>
      </div>

      <div v-if="!collapsedGroups.has(group.label)" class="group-tasks">
        <div
          v-for="task in group.tasks"
          :key="task.id"
          class="task-card"
          :class="{ completed: task.completed }"
        >
          <input
            :id="`task-group-completed-${task.id}`"
            :name="`task_group_completed_${task.id}`"
            type="checkbox"
            :checked="task.completed"
            class="checkbox-circle"
            @change="$emit('toggle', task.id)"
          />
          <label class="sr-only" :for="`task-group-completed-${task.id}`">
            {{ task.completed ? '标记为未完成' : '标记为完成' }}：{{ task.title }}
          </label>
          <div class="task-content">
            <div class="task-title">{{ task.title }}</div>
            <div class="task-meta">
              <PriorityTag v-if="task.priority" :priority="task.priority" />
              <span v-if="task.dueAt" class="meta-item">
                {{ formatDate(task.dueAt) }}
              </span>
            </div>
          </div>
        </div>
      </div>
    </div>
  </MotionStagger>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import type { Task } from '@/entities/task/model'
import PriorityTag from '@/shared/ui/PriorityTag.vue'
import MotionStagger from '@/shared/ui/MotionStagger.vue'

interface TaskGroup {
  label: string
  tasks: Task[]
}

defineProps<{
  groups: TaskGroup[]
}>()

defineEmits<{
  toggle: [id: number]
}>()

const collapsedGroups = ref(new Set<string>())

function toggleGroup(label: string) {
  if (collapsedGroups.value.has(label)) {
    collapsedGroups.value.delete(label)
  } else {
    collapsedGroups.value.add(label)
  }
}

function formatDate(dateStr: string): string {
  if (!dateStr) return ''
  const date = new Date(dateStr)
  const now = new Date()
  const today = new Date(now.getFullYear(), now.getMonth(), now.getDate())
  const target = new Date(date.getFullYear(), date.getMonth(), date.getDate())

  if (target.getTime() === today.getTime()) {
    return '今天'
  }

  const tomorrow = new Date(today)
  tomorrow.setDate(tomorrow.getDate() + 1)
  if (target.getTime() === tomorrow.getTime()) {
    return '明天'
  }

  return `${date.getMonth() + 1}月${date.getDate()}日`
}
</script>

<style scoped>
.task-grouped-list {
  display: flex;
  flex-direction: column;
  gap: 16px;
}

.group-header {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 8px 12px;
  background: var(--color-surface-muted);
  border-radius: 6px;
  cursor: pointer;
  user-select: none;
}

.group-arrow {
  font-size: 12px;
  transition: transform var(--motion-duration-base);
  transform-origin: center center;
}

.group-arrow.collapsed {
  transform: rotate(-90deg);
}

.group-label {
  font-weight: 500;
  color: var(--color-text);
}

.group-count {
  font-size: 12px;
  color: var(--color-text-muted);
}

.group-tasks {
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.task-card {
  display: flex;
  align-items: flex-start;
  gap: 12px;
  padding: 12px;
  background: var(--color-surface);
  border: 1px solid var(--color-border);
  border-radius: 6px;
}

.task-card.completed {
  opacity: 0.6;
}

.task-content {
  flex: 1;
  min-width: 0;
}

.task-title {
  font-size: 14px;
  color: var(--color-text);
}

.task-card.completed .task-title {
  text-decoration: line-through;
}

.task-meta {
  display: flex;
  gap: 8px;
  margin-top: 4px;
  font-size: 12px;
  color: var(--color-text-muted);
}
</style>
