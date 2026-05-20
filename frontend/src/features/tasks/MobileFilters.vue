<template>
  <MobileSheet :visible="isVisible" @update:visible="onVisibleChange">
    <div class="mobile-filters">
      <h3>筛选</h3>

      <div class="filter-group">
        <label for="mobile-task-filter-status">状态</label>
        <select
          id="mobile-task-filter-status"
          v-model="localFilters.status"
          name="mobile_task_filter_status"
        >
          <option value="all">全部</option>
          <option value="pending">待处理</option>
          <option value="completed">已完成</option>
        </select>
      </div>

      <div class="filter-group">
        <label for="mobile-task-filter-priority">优先级</label>
        <select
          id="mobile-task-filter-priority"
          v-model="localFilters.priority"
          name="mobile_task_filter_priority"
        >
          <option :value="undefined">全部</option>
          <option :value="1">高</option>
          <option :value="2">中</option>
          <option :value="3">低</option>
        </select>
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
})

watch(
  () => props.filters,
  (val) => {
    Object.assign(localFilters, {
      status: val.status,
      priority: val.priority,
      search: val.search,
    })
  },
  { deep: true },
)

function clearFilters() {
  localFilters.status = 'all'
  localFilters.priority = undefined
  localFilters.search = ''
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

.filter-actions {
  display: flex;
  gap: 12px;
  margin-top: 24px;
}

.btn-secondary {
  flex: 1;
  padding: 10px;
  background: var(--color-surface);
  border: 1px solid var(--color-border);
  border-radius: 6px;
  cursor: pointer;
  color: var(--color-text);
}

.btn-primary {
  flex: 1;
  padding: 10px;
  background: var(--color-primary);
  color: white;
  border: none;
  border-radius: 6px;
  cursor: pointer;
}
</style>
