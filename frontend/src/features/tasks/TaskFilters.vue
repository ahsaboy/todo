<template>
  <div class="task-filters">
    <div class="filter-group">
      <label>状态</label>
      <BaseSelect
        v-model="localFilters.status"
        :options="statusOptions"
        aria-label="状态"
        @change="applyFilters"
      />
    </div>

    <div class="filter-group">
      <label>优先级</label>
      <BaseSelect
        v-model="localFilters.priority"
        :options="priorityOptions"
        aria-label="优先级"
        @change="applyFilters"
      />
    </div>

    <div class="filter-group tags">
      <label>标签</label>
      <TagPicker v-model="localFilters.tags" placeholder="按标签筛选..." @update:modelValue="applyFilters" />
    </div>

    <div class="filter-group search">
      <label class="sr-only" for="task-filter-search">搜索任务</label>
      <input
        id="task-filter-search"
        v-model="localFilters.search"
        name="task_filter_search"
        type="text"
        placeholder="搜索任务..."
        @input="applyFilters"
      />
    </div>
  </div>
</template>

<script setup lang="ts">
import { reactive } from 'vue'
import type { TaskFilters } from './useTasks'
import TagPicker from '@/features/tags/TagPicker.vue'
import BaseSelect, { type SelectOption } from '@/shared/ui/BaseSelect.vue'

const statusOptions: SelectOption<TaskFilters['status']>[] = [
  { label: '全部', value: 'all' },
  { label: '待处理', value: 'pending' },
  { label: '已完成', value: 'completed' },
]

const priorityOptions: SelectOption<TaskFilters['priority']>[] = [
  { label: '全部', value: undefined },
  { label: '高', value: 1 },
  { label: '中', value: 2 },
  { label: '低', value: 3 },
]

const props = defineProps<{
  filters: TaskFilters
}>()

const emit = defineEmits<{
  change: [filters: Partial<TaskFilters>]
}>()

const localFilters = reactive({
  status: props.filters.status,
  priority: props.filters.priority,
  search: props.filters.search,
  tags: Array.isArray(props.filters.tags) ? [...props.filters.tags] : [],
})

function applyFilters() {
  emit('change', { ...localFilters })
}
</script>

<style scoped>
.task-filters {
  display: flex;
  gap: 16px;
  align-items: center;
  flex-wrap: wrap;
}

.filter-group {
  display: flex;
  align-items: center;
  gap: 8px;
}

.filter-group label {
  font-size: 13px;
  color: var(--color-text-muted);
}

.filter-group select,
.filter-group input {
  padding: 6px 12px;
  border: 1px solid var(--color-border);
  border-radius: 4px;
  font-size: 13px;
  background: var(--color-surface);
}

.filter-group.tags {
  min-width: 220px;
}

.filter-group.tags :deep(.tag-picker) {
  width: auto;
  min-width: 200px;
}

.filter-group.search {
  flex: 1;
  min-width: 200px;
}

.filter-group.search input {
  width: 100%;
}

@media (max-width: 767px) {
  .filter-group.tags {
    min-width: 0;
  }

  .filter-group.tags :deep(.tag-picker) {
    min-width: 0;
  }

  .filter-group.search {
    min-width: 0;
  }
}
</style>
