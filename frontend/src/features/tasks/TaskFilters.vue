<template>
  <div class="task-filters">
    <div class="filter-group">
      <label for="task-filter-status">状态</label>
      <select
        id="task-filter-status"
        v-model="localFilters.status"
        name="task_filter_status"
        @change="applyFilters"
      >
        <option value="all">全部</option>
        <option value="pending">待处理</option>
        <option value="completed">已完成</option>
      </select>
    </div>

    <div class="filter-group">
      <label for="task-filter-priority">优先级</label>
      <select
        id="task-filter-priority"
        v-model="localFilters.priority"
        name="task_filter_priority"
        @change="applyFilters"
      >
        <option :value="undefined">全部</option>
        <option :value="1">高</option>
        <option :value="2">中</option>
        <option :value="3">低</option>
      </select>
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
</style>
