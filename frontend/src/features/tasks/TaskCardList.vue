<template>
  <div class="task-card-list">
    <div
      v-for="task in tasks"
      :key="task.id"
      class="task-card"
      :class="{ completed: task.completed }"
    >
      <div class="task-card-main">
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
        <div class="task-content">
          <button class="task-summary" type="button" @click="$emit('open', task)">
            <div class="task-head">
              <div class="task-title">{{ task.title }}</div>
              <div class="task-tags">
                <PriorityTag :priority="task.priority" />
                <StatusTag :completed="task.completed" :overdue="isOverdue(task)" />
              </div>
            </div>
            <div v-if="task.description" class="task-desc">
              {{ task.description }}
            </div>
            <dl class="task-meta-list">
              <div class="meta-row" :class="{ overdue: isOverdue(task) }">
                <dt>截止</dt>
                <dd>{{ formatDateTime(task.dueAt) }}</dd>
              </div>
              <div class="meta-row">
                <dt>提醒</dt>
                <dd>{{ formatDateTime(task.remindAt) }}</dd>
              </div>
              <div class="meta-row">
                <dt>重复</dt>
                <dd>{{ formatRepeat(task.repeatType, task.repeatInterval) }}</dd>
              </div>
              <div class="meta-row">
                <dt>创建</dt>
                <dd>{{ formatDateTime(task.createdAt) }}</dd>
              </div>
            </dl>
          </button>
        </div>
      </div>
      <div class="task-actions">
        <button class="action-btn" type="button" @click="$emit('edit', task)"><Pencil :size="14" /> 编辑</button>
        <button class="action-btn action-btn-danger" type="button" @click="$emit('delete', task.id)"><Trash2 :size="14" /> 删除</button>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import type { Task } from '@/entities/task/model'
import PriorityTag from '@/shared/ui/PriorityTag.vue'
import StatusTag from '@/shared/ui/StatusTag.vue'
import { Pencil, Trash2 } from 'lucide-vue-next'

defineProps<{
  tasks: Task[]
}>()

defineEmits<{
  toggle: [id: number]
  open: [task: Task]
  edit: [task: Task]
  delete: [id: number]
}>()

function isOverdue(task: Task): boolean {
  if (task.completed || !task.dueAt) return false
  return new Date(task.dueAt) < new Date()
}

function formatDateTime(dateStr: string): string {
  if (!dateStr) return '-'
  const date = new Date(dateStr)
  return date.toLocaleDateString('zh-CN', {
    month: '2-digit',
    day: '2-digit',
    hour: '2-digit',
    minute: '2-digit',
  })
}

function formatRepeat(type: string, interval: number): string {
  if (!type || type === 'none') return '-'
  const labels: Record<string, string> = {
    daily: '每天',
    weekly: '每周',
    monthly: '每月',
    yearly: '每年',
  }
  return `${labels[type] || type}${interval > 1 ? ` ×${interval}` : ''}`
}
</script>

<style scoped>
.task-card-list {
  display: flex;
  flex-direction: column;
  gap: 12px;
  padding-bottom: 88px;
  min-width: 0;
}

.task-card {
  display: flex;
  flex-direction: column;
  gap: 12px;
  padding: 12px;
  background: var(--color-surface);
  border: 1px solid var(--color-border);
  border-radius: 8px;
  min-width: 0;
}

.task-card.completed {
  opacity: 0.72;
}

.task-card-main {
  display: flex;
  align-items: flex-start;
  gap: 12px;
  min-width: 0;
}

.task-checkbox {
  flex: none;
  margin-top: 4px;
}

.task-content {
  flex: 1;
  min-width: 0;
}

.task-summary {
  width: 100%;
  display: flex;
  flex-direction: column;
  gap: 10px;
  padding: 0;
  margin: 0;
  background: none;
  border: none;
  text-align: left;
  color: inherit;
  cursor: pointer;
}

.task-head {
  display: flex;
  flex-direction: column;
  gap: 8px;
  min-width: 0;
}

.task-title {
  font-size: 14px;
  line-height: 1.5;
  color: var(--color-text);
  overflow-wrap: anywhere;
}

.task-card.completed .task-title {
  text-decoration: line-through;
}

.task-tags {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
}

.task-desc {
  font-size: 12px;
  line-height: 1.5;
  color: var(--color-text-muted);
  display: -webkit-box;
  -webkit-line-clamp: 3;
  -webkit-box-orient: vertical;
  overflow: hidden;
  overflow-wrap: anywhere;
}

.task-meta-list {
  display: grid;
  gap: 8px;
  margin: 0;
}

.meta-row {
  display: grid;
  grid-template-columns: 40px minmax(0, 1fr);
  gap: 8px;
  font-size: 12px;
  line-height: 1.5;
  min-width: 0;
}

.meta-row dt {
  color: var(--color-text-muted);
}

.meta-row dd {
  margin: 0;
  color: var(--color-text);
  min-width: 0;
  overflow-wrap: anywhere;
}

.meta-row.overdue dd {
  color: var(--color-danger);
  font-weight: 500;
}

.task-actions {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 8px;
}

.action-btn {
  min-height: 32px;
  padding: 6px 10px;
  border: 1px solid var(--color-border);
  border-radius: 6px;
  background: var(--color-surface);
  color: var(--color-text);
  font-size: 13px;
  cursor: pointer;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.action-btn-danger {
  color: var(--color-danger);
}

@media (max-width: 359px) {
  .task-card-list {
    padding-bottom: 72px;
  }

  .task-actions {
    grid-template-columns: 1fr;
  }
}
</style>
