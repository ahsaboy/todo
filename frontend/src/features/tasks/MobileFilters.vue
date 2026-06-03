<template>
  <MobileSheet :visible="isVisible" @update:visible="onVisibleChange">
    <div class="mobile-filters">
      <h3>筛选</h3>

      <div class="filter-group">
        <label>状态</label>
        <BaseSelect
          v-model="localFilters.status"
          :options="statusOptions"
          block
          aria-label="状态"
        />
      </div>

      <div class="filter-group">
        <label>优先级</label>
        <BaseSelect
          v-model="localFilters.priority"
          :options="priorityOptions"
          block
          aria-label="优先级"
        />
      </div>

      <div class="filter-group">
        <label>标签</label>
        <TagPicker v-model="localFilters.tags" placeholder="按标签筛选..." />
      </div>

      <div class="filter-group">
        <label for="mobile-task-filter-search">搜索</label>
        <input
          id="mobile-task-filter-search"
          v-model="localFilters.search"
          name="mobile_task_filter_search"
          type="text"
          placeholder="搜索任务..."
        />
      </div>

      <div class="filter-actions">
        <button class="btn-secondary" type="button" @click="clearFilters">清除</button>
        <button class="btn-primary" type="button" @click="applyFilters">应用</button>
      </div>
    </div>
  </MobileSheet>
</template>

<script setup lang="ts">
import { reactive, watch, computed } from 'vue'
import MobileSheet from '@/shared/ui/MobileSheet.vue'
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
  visible: boolean
  filters: TaskFilters
}>()

const emit = defineEmits<{
  'update:visible': [value: boolean]
  change: [filters: Partial<TaskFilters>]
}>()

const isVisible = computed({
  get: () => props.visible,
  set: (val) => emit('update:visible', val),
})

function onVisibleChange(val: boolean) {
  emit('update:visible', val)
}

const localFilters = reactive({
  status: props.filters.status,
  priority: props.filters.priority,
  search: props.filters.search,
  tags: Array.isArray(props.filters.tags) ? [...props.filters.tags] : [],
})

watch(
  () => props.filters,
  (val) => {
    Object.assign(localFilters, {
      status: val.status,
      priority: val.priority,
      search: val.search,
      tags: Array.isArray(val.tags) ? [...val.tags] : [],
    })
  },
  { deep: true },
)

function clearFilters() {
  localFilters.status = 'all'
  localFilters.priority = undefined
  localFilters.search = ''
  localFilters.tags = []
  applyFilters()
}

function applyFilters() {
  emit('change', { ...localFilters })
  emit('update:visible', false)
}
</script>

<style scoped>
.mobile-filters h3 {
  margin: 0 0 16px;
  font-size: 16px;
}

.filter-group {
  display: flex;
  flex-direction: column;
  gap: 6px;
  margin-bottom: 16px;
}

.filter-group label {
  font-size: 13px;
  color: var(--color-text-muted);
}

.filter-group select,
.filter-group input {
  padding: 10px 12px;
  border: 1px solid var(--color-border);
  border-radius: 6px;
  font-size: 14px;
  background: var(--color-surface);
  color: var(--color-text);
}

.filter-group :deep(.tag-picker) {
  min-width: 0;
}

.filter-actions {
  display: flex;
  gap: 12px;
  margin-top: 24px;
}

.btn-secondary {
  flex: 1;
  padding: 10px;
}

.btn-primary {
  flex: 1;
  padding: 10px;
}
</style>
