<script setup lang="ts" generic="T extends Record<string, unknown>">
import { computed } from 'vue'
import type { DataTableConfig } from './data-table/types'
import { useMediaQuery } from '@/shared/composables/useMediaQuery'

const props = defineProps<{
  config: DataTableConfig<T>
  data: T[]
  isLoading: boolean
}>()

const emit = defineEmits<{
  'apply-filters': []
  'filter-change': [id: string, value: string]
}>()

const isMobile = useMediaQuery('(max-width: 767px)')

function getCellValue(row: T, key: string): unknown {
  return row[key as keyof T]
}

function formatCell(row: T, col: DataTableConfig<T>['columns'][number]): string {
  const raw = getCellValue(row, col.key)
  if (col.formatter) {
    return col.formatter(raw as never, row)
  }
  return raw != null ? String(raw) : '—'
}

function getCellClass(row: T, col: DataTableConfig<T>['columns'][number]): string {
  if (typeof col.cellClass === 'function') return col.cellClass(row)
  return col.cellClass ?? ''
}

function colSpan(): number {
  let count = props.config.columns.length
  if (props.config.actions?.length) count++
  return count
}

function getActionLabel(label: string | ((row: T) => string), row: T): string {
  return typeof label === 'string' ? label : label(row)
}

// Card mode helpers
const titleCol = computed(() => {
  const key = props.config.mobileCard?.titleKey
  return key ? props.config.columns.find(c => c.key === key) : props.config.columns[0]
})

const subtitleCol = computed(() => {
  const key = props.config.mobileCard?.subtitleKey
  if (key) return props.config.columns.find(c => c.key === key)
  return props.config.columns.length > 1 ? props.config.columns[1] : undefined
})

const metaColumns = computed(() => {
  const exclude = new Set<string>()
  if (titleCol.value) exclude.add(titleCol.value.key)
  if (subtitleCol.value) exclude.add(subtitleCol.value.key)

  const keys = props.config.mobileCard?.metaKeys
  if (keys) {
    return props.config.columns.filter(c => keys.includes(c.key))
  }
  return props.config.columns.filter(c => !exclude.has(c.key))
})
</script>

<template>
  <!-- Toolbar -->
  <div v-if="config.filters?.length" class="admin-toolbar">
    <template v-for="filter in config.filters" :key="filter.id">
      <input
        v-if="filter.type === 'text' || filter.type === 'number'"
        :type="filter.type"
        :value="filter.value"
        :placeholder="filter.placeholder"
        class="admin-search-input"
        :class="{
          'toolbar-input-narrow': filter.width === 'narrow',
        }"
        @input="emit('filter-change', filter.id, ($event.target as HTMLInputElement).value)"
        @keyup.enter="emit('apply-filters')"
      />
      <select
        v-else-if="filter.type === 'select'"
        :value="filter.value"
        class="admin-search-input toolbar-select"
        @change="emit('filter-change', filter.id, ($event.target as HTMLSelectElement).value)"
      >
        <option v-for="opt in filter.options" :key="opt.value" :value="opt.value">
          {{ opt.label }}
        </option>
      </select>
    </template>
    <button class="btn btn-primary" @click="emit('apply-filters')">
      {{ config.filterButtonText ?? '筛选' }}
    </button>
  </div>

  <!-- Desktop Table -->
  <div v-if="!isMobile" class="admin-table-wrap">
    <table class="admin-table">
      <colgroup>
        <col
          v-for="col in config.columns"
          :key="col.key"
          :style="col.width ? { width: col.width } : undefined"
        />
        <col
          v-if="config.actions?.length"
          :style="{ width: config.actionsWidth ?? '120px' }"
        />
      </colgroup>
      <thead>
        <tr>
          <th
            v-for="col in config.columns"
            :key="col.key"
          >
            {{ col.label }}
          </th>
          <th v-if="config.actions?.length">操作</th>
        </tr>
      </thead>
      <tbody>
        <tr v-if="isLoading">
          <td :colspan="colSpan()" class="table-cell-status">
            {{ config.loadingText ?? '加载中...' }}
          </td>
        </tr>
        <tr v-else-if="!data.length">
          <td :colspan="colSpan()" class="table-cell-status">
            {{ config.emptyText ?? '暂无数据' }}
          </td>
        </tr>
        <tr v-for="(row, idx) in data" :key="idx">
          <td
            v-for="col in config.columns"
            :key="col.key"
            :class="[
              getCellClass(row, col),
              { 'text-truncate': col.truncate },
            ]"
            :title="col.truncate ? String(getCellValue(row, col.key) ?? '') : undefined"
            :style="col.truncate && col.width ? { maxWidth: col.width } : undefined"
          >
            <component
              :is="col.component"
              v-if="col.component"
              v-bind="col.componentProps?.(row) ?? {}"
            />
            <template v-else>
              {{ formatCell(row, col) }}
            </template>
          </td>
          <td v-if="config.actions?.length" class="action-cell">
            <template v-for="action in config.actions" :key="action.id">
              <button
                v-if="!action.visible || action.visible(row)"
                class="btn btn-sm"
                :class="{
                  'btn-primary': action.variant === 'primary',
                  'btn-danger': action.variant === 'danger',
                }"
                @click="action.onClick(row)"
              >
                <component :is="action.icon" v-if="action.icon" :size="14" />
                <span v-if="action.icon && getActionLabel(action.label, row)">&nbsp;</span>
                {{ getActionLabel(action.label, row) }}
              </button>
            </template>
          </td>
        </tr>
      </tbody>
    </table>
  </div>

  <!-- Mobile Cards -->
  <div v-else class="data-card-list">
    <div v-if="isLoading" class="data-card-status">
      {{ config.loadingText ?? '加载中...' }}
    </div>
    <div v-else-if="!data.length" class="data-card-status">
      {{ config.emptyText ?? '暂无数据' }}
    </div>
    <article v-for="(row, idx) in data" :key="idx" class="data-card">
      <div v-if="titleCol" class="data-card-header">
        <div class="data-card-title">
          <component
            :is="titleCol.component"
            v-if="titleCol.component"
            v-bind="titleCol.componentProps?.(row) ?? {}"
          />
          <template v-else>{{ formatCell(row, titleCol) }}</template>
        </div>
        <div v-if="subtitleCol" class="data-card-subtitle">
          <component
            :is="subtitleCol.component"
            v-if="subtitleCol.component"
            v-bind="subtitleCol.componentProps?.(row) ?? {}"
          />
          <template v-else>{{ formatCell(row, subtitleCol) }}</template>
        </div>
      </div>
      <dl v-if="metaColumns.length" class="data-card-meta">
        <div v-for="col in metaColumns" :key="col.key" class="data-card-meta-row">
          <dt>{{ col.label }}</dt>
          <dd :class="getCellClass(row, col)">
            <component
              :is="col.component"
              v-if="col.component"
              v-bind="col.componentProps?.(row) ?? {}"
            />
            <template v-else>{{ formatCell(row, col) }}</template>
          </dd>
        </div>
      </dl>
      <div v-if="config.actions?.length" class="data-card-actions">
        <template v-for="action in config.actions" :key="action.id">
          <button
            v-if="!action.visible || action.visible(row)"
            class="btn btn-sm"
            :class="{
              'btn-primary': action.variant === 'primary',
              'btn-danger': action.variant === 'danger',
            }"
            @click="action.onClick(row)"
          >
            <component :is="action.icon" v-if="action.icon" :size="14" />
            <span v-if="action.icon && getActionLabel(action.label, row)">&nbsp;</span>
            {{ getActionLabel(action.label, row) }}
          </button>
        </template>
      </div>
    </article>
  </div>
</template>

<style scoped>
.data-card-list {
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.data-card-status {
  text-align: center;
  padding: 2rem 1rem;
  color: var(--color-text-muted);
  font-size: 0.9rem;
}

.data-card {
  display: flex;
  flex-direction: column;
  gap: 12px;
  padding: 14px 16px;
  background: var(--color-surface);
  border: 1px solid var(--color-border);
  border-radius: 12px;
  transition: background var(--motion-duration-fast) var(--motion-ease-standard);
}

.data-card:active {
  background: var(--color-surface-muted);
}

.data-card-header {
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.data-card-title {
  font-size: 0.95rem;
  font-weight: 600;
  color: var(--color-text);
  line-height: 1.4;
}

.data-card-subtitle {
  font-size: 0.8rem;
  color: var(--color-text-muted);
}

.data-card-meta {
  display: flex;
  flex-direction: column;
  gap: 8px;
  margin: 0;
}

.data-card-meta-row {
  display: grid;
  grid-template-columns: 72px minmax(0, 1fr);
  gap: 8px;
  align-items: baseline;
  padding: 6px 10px;
  background: var(--color-surface-muted);
  border-radius: 8px;
}

.data-card-meta-row dt {
  font-size: 0.75rem;
  color: var(--color-text-muted);
  font-weight: 500;
  white-space: nowrap;
}

.data-card-meta-row dd {
  margin: 0;
  font-size: 0.85rem;
  color: var(--color-text);
  min-width: 0;
  word-break: break-all;
}

.data-card-actions {
  display: flex;
  gap: 8px;
  flex-wrap: wrap;
  padding-top: 4px;
  border-top: 1px solid var(--color-border);
}
</style>
