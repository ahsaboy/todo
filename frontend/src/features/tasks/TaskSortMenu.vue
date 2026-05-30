<template>
  <div class="task-sort">
    <BaseSelect
      v-model="localSort.field"
      :options="fieldOptions"
      aria-label="排序字段"
      @change="applySort"
    />

    <button class="btn-order" type="button" @click="toggleOrder">
      {{ localSort.order === 'asc' ? '↑' : '↓' }}
    </button>
  </div>
</template>

<script setup lang="ts">
import { reactive } from 'vue'
import type { TaskSort } from './useTasks'
import BaseSelect, { type SelectOption } from '@/shared/ui/BaseSelect.vue'

const fieldOptions: SelectOption<TaskSort['field']>[] = [
  { label: '创建时间', value: 'created_at' },
  { label: '截止时间', value: 'due_at' },
  { label: '优先级', value: 'priority' },
  { label: '标题', value: 'title' },
]

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
