<template>
  <div class="task-sort">
    <label class="sr-only" for="task-sort-field">排序字段</label>
    <select
      id="task-sort-field"
      v-model="localSort.field"
      name="task_sort_field"
      @change="applySort"
    >
      <option value="created_at">创建时间</option>
      <option value="due_at">截止时间</option>
      <option value="priority">优先级</option>
      <option value="title">标题</option>
    </select>

    <button class="btn-order" type="button" @click="toggleOrder">
      {{ localSort.order === 'asc' ? '↑' : '↓' }}
    </button>
  </div>
</template>

<script setup lang="ts">
import { reactive } from 'vue'
import type { TaskSort } from './useTasks'

const props = defineProps<{
  sort: TaskSort
}>()

const emit = defineEmits<{
  change: [field: string, order: 'asc' | 'desc']
}>()

const localSort = reactive({
  field: props.sort.field,
  order: props.sort.order,
})

function applySort() {
  emit('change', localSort.field, localSort.order)
}

function toggleOrder() {
  localSort.order = localSort.order === 'asc' ? 'desc' : 'asc'
  applySort()
}
</script>

<style scoped>
.task-sort {
  display: flex;
  align-items: center;
  gap: 8px;
}

.task-sort select {
  padding: 6px 12px;
  border: 1px solid var(--color-border);
  border-radius: 4px;
  font-size: 13px;
  background: var(--color-surface);
}

.btn-order {
  padding: 6px 10px;
  border: 1px solid var(--color-border);
  border-radius: 4px;
  background: var(--color-surface);
  cursor: pointer;
}
</style>
