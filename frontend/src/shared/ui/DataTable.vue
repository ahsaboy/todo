<script setup lang="ts" generic="T extends Record<string, unknown>">
import type { DataTableConfig } from './data-table/types'

const props = defineProps<{
  config: DataTableConfig<T>
  data: T[]
  isLoading: boolean
}>()

const emit = defineEmits<{
  'apply-filters': []
  'filter-change': [id: string, value: string]
}>()

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

  <!-- Table -->
  <div class="admin-table-wrap">
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
</template>
