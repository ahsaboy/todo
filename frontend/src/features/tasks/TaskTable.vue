<template>
  <div class="task-table">
    <table>
      <thead>
        <tr>
          <th class="col-completed">完成</th>
          <th class="col-title">标题</th>
          <th class="col-priority">优先级</th>
          <th class="col-due">截止时间</th>
          <th class="col-remind">提醒时间</th>
          <th class="col-repeat">重复</th>
          <th class="col-created">创建时间</th>
          <th class="col-actions">操作</th>
        </tr>
      </thead>
      <tbody>
        <tr v-for="task in tasks" :key="task.id" :class="{ completed: task.completed }">
          <td class="col-completed">
            <input
              :id="`task-table-completed-${task.id}`"
              :name="`task_table_completed_${task.id}`"
              type="checkbox"
              :checked="task.completed"
              @change="$emit('toggle', task.id)"
            />
            <label class="sr-only" :for="`task-table-completed-${task.id}`">
              {{ task.completed ? '标记为未完成' : '标记为完成' }}：{{ task.title }}
            </label>
          </td>
          <td class="col-title">
            <button class="task-title-btn" type="button" @click="$emit('open', task)">
              {{ task.title }}
            </button>
          </td>
          <td class="col-priority">
            <PriorityTag :priority="task.priority" />
          </td>
          <td class="col-due" :class="{ overdue: isOverdue(task) }">
            {{ formatDate(task.dueAt) }}
          </td>
          <td class="col-remind">
            {{ formatDate(task.remindAt) }}
          </td>
          <td class="col-repeat">
            {{ formatRepeat(task.repeatType, task.repeatInterval) }}
          </td>
          <td class="col-created">
            {{ formatDate(task.createdAt) }}
          </td>
          <td class="col-actions">
            <button class="btn-icon btn-icon-text" type="button" @click="$emit('edit', task)"><Pencil :size="14" /> 编辑</button>
            <button class="btn-icon btn-icon-text btn-danger" type="button" @click="$emit('delete', task.id)"><Trash2 :size="14" /> 删除</button>
          </td>
        </tr>
      </tbody>
    </table>
  </div>
</template>

<script setup lang="ts">
import type { Task } from '@/entities/task/model'
import PriorityTag from '@/shared/ui/PriorityTag.vue'
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

function formatDate(dateStr: string): string {
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
.task-table {
  overflow-x: auto;
}

table {
  width: 100%;
  border-collapse: collapse;
}

th,
td {
  padding: 0 12px;
  height: 44px;
  text-align: left;
  border-bottom: 1px solid var(--color-border);
}

th {
  font-weight: 500;
  color: var(--color-text-muted);
  font-size: 13px;
}

.col-completed {
  width: 48px;
}

.col-title {
  min-width: 200px;
}

.col-actions {
  width: 120px;
}

.task-title-btn {
  background: none;
  border: none;
  padding: 0;
  color: var(--color-text);
  cursor: pointer;
  text-align: left;
}

.task-title-btn:hover {
  color: var(--color-primary);
}

tr.completed {
  opacity: 0.72;
}

tr.completed .task-title-btn {
  text-decoration: line-through;
}

.overdue {
  color: var(--color-danger);
}

.btn-icon {
  background: none;
  border: none;
  padding: 4px 8px;
  cursor: pointer;
  color: var(--color-text-muted);
  font-size: 13px;
  border-radius: 6px;
  transition:
    color 150ms,
    background-color 150ms,
    box-shadow 150ms;
}

.btn-icon:hover {
  color: var(--color-text);
  background: color-mix(in srgb, var(--color-primary) 8%, transparent);
  box-shadow: 0 0 0 1px color-mix(in srgb, var(--color-primary) 10%, transparent);
}

.btn-danger:hover {
  color: var(--color-danger);
  background: color-mix(in srgb, var(--color-danger) 10%, transparent);
  box-shadow: 0 0 0 1px color-mix(in srgb, var(--color-danger) 14%, transparent);
}
</style>
