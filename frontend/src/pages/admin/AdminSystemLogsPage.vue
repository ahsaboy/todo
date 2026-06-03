<script setup lang="ts">
import { computed, ref, watch, onUnmounted } from 'vue'
import { Download, RefreshCw } from 'lucide-vue-next'
import { adminApi } from '@/shared/api/admin-client'
import BaseSelect, { type SelectOption } from '@/shared/ui/BaseSelect.vue'
import PagePagination from '@/shared/ui/PagePagination.vue'
import DataTable from '@/shared/ui/DataTable.vue'
import { useFetch } from '@/shared/composables/useFetch'
import { useCrudList } from '@/shared/composables/useCrudList'
import type { DataTableConfig } from '@/shared/ui/data-table/types'
import type { ApiResponse } from '@/shared/api/types'
import ExtraFieldsCell from '@/features/admin/ExtraFieldsCell.vue'
import {
  type LogFile, type LogEntry, type LogEntryRow,
  extraFieldEntries, formatSize, truncateMsg, levelBadgeClass,
} from './utils/logFormatters'

const selectedFile = ref('')
const autoRefresh = ref(false)
let refreshTimer: ReturnType<typeof setInterval> | null = null

const { data: logFilesData } = useFetch<LogFile[]>({
  fetcher: () => adminApi.get<ApiResponse<LogFile[]>>('/system-logs').then(r => r.data || []),
  errorPrefix: '加载日志文件列表',
})

const logFiles = computed(() => logFilesData.value ?? [])

watch(logFiles, (files) => {
  if (files.length && !selectedFile.value) {
    selectedFile.value = files[0].filename
  }
}, { immediate: true })

const entries = useCrudList<LogEntryRow>({
  client: adminApi,
  autoLoad: false,
  limit: 50,
  buildEndpoint: ({ page, limit, filters }) => {
    if (!selectedFile.value) return ''
    const params = new URLSearchParams({ page: String(page), limit: String(limit) })
    if (filters.level) params.set('level', filters.level)
    return `/system-logs/${encodeURIComponent(selectedFile.value)}/entries?${params}`
  },
  mapItem: (raw) => {
    const entry = raw as LogEntry
    return { ...entry, _extraFields: extraFieldEntries(entry) }
  },
  errorPrefix: '加载日志条目',
})

watch(selectedFile, () => {
  entries.setPage(1)
  entries.load()
})

watch(autoRefresh, (val) => {
  if (val) {
    refreshTimer = setInterval(() => entries.load(), 5000)
  } else if (refreshTimer) {
    clearInterval(refreshTimer)
    refreshTimer = null
  }
})

onUnmounted(() => { if (refreshTimer) clearInterval(refreshTimer) })

async function downloadLog() {
  if (!selectedFile.value) return
  try {
    const blob = await adminApi.download(`/system-logs/${encodeURIComponent(selectedFile.value)}/download`)
    const a = document.createElement('a')
    a.href = URL.createObjectURL(blob)
    a.download = selectedFile.value
    a.click()
    URL.revokeObjectURL(a.href)
  } catch {
    entries.error.value = '下载日志文件失败'
  }
}

const fileOptions = computed<SelectOption<string>[]>(() =>
  logFiles.value.map((f) => ({ label: `${f.date} (${formatSize(f.size_bytes)})`, value: f.filename })),
)

const levelOptions: SelectOption<string>[] = [
  { label: '全部级别', value: '' },
  { label: 'debug', value: 'debug' },
  { label: 'info', value: 'info' },
  { label: 'warn', value: 'warn' },
  { label: 'error', value: 'error' },
]

const config: DataTableConfig<LogEntryRow> = {
  columns: [
    { key: 'ts', label: '时间', width: '170px', cellClass: 'log-ts' },
    { key: 'level', label: '级别', width: '70px', cellClass: levelBadgeClass },
    { key: 'msg', label: '消息', truncate: true, width: '360px', cellClass: 'log-msg', formatter: (v) => truncateMsg(v) },
    { key: 'caller', label: '调用位置', width: '180px', cellClass: 'log-caller', formatter: (v) => v || '—' },
    { key: '_extraFields', label: '附加字段', width: '200px', component: ExtraFieldsCell, componentProps: (row) => ({ fields: row._extraFields }) },
  ],
  emptyText: selectedFile.value ? '暂无日志条目' : '请选择日志文件',
  mobileCard: {
    titleKey: 'msg',
    subtitleKey: 'ts',
    metaKeys: ['level', 'caller', '_extraFields'],
  },
}
</script>

<template>
  <div class="page-container">
    <h1 class="admin-page-title">系统日志</h1>

    <div class="admin-toolbar">
      <BaseSelect
        v-model="selectedFile"
        :options="fileOptions"
        placeholder="选择日志文件"
        aria-label="日志文件"
        style="width: 200px;"
      />
      <BaseSelect v-model="entries.filters.value.level" :options="levelOptions" aria-label="日志级别" class="toolbar-select" @update:model-value="entries.applyFilters()" />
      <label class="toolbar-check">
        <input v-model="autoRefresh" type="checkbox" />
        <RefreshCw :size="14" :class="{ 'loading-spin': autoRefresh }" />
        <span>自动刷新</span>
      </label>
      <button class="btn btn-sm" :disabled="!selectedFile" @click="downloadLog">
        <Download :size="14" /><span>下载</span>
      </button>
    </div>

    <div v-if="entries.error.value" class="error-message">{{ entries.error.value }}</div>

    <DataTable :config="config" :data="entries.items.value" :is-loading="entries.isLoading.value" />

    <PagePagination :page="entries.page.value" :total="entries.total.value" :total-pages="entries.totalPages.value" @update:page="entries.setPage" />
  </div>
</template>

<style scoped>
.toolbar-check {
  display: flex;
  align-items: center;
  gap: 0.35rem;
  font-size: 0.85rem;
  color: var(--color-text);
  cursor: pointer;
  user-select: none;
}

.toolbar-check input[type='checkbox'] {
  width: 14px;
  height: 14px;
  accent-color: var(--color-primary);
}

.log-ts {
  white-space: nowrap;
  font-size: 0.8rem;
  font-family: monospace;
}

.log-msg {
  max-width: 360px;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  font-size: 0.85rem;
}

.log-caller {
  font-size: 0.8rem;
  font-family: monospace;
  color: var(--color-text-muted);
  max-width: 180px;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.badge-level-debug {
  background: var(--color-surface-muted);
  color: var(--color-text-muted);
}

.badge-level-info {
  background: color-mix(in srgb, var(--color-primary) 15%, transparent);
  color: var(--color-primary);
}

.badge-level-warn {
  background: color-mix(in srgb, var(--color-warning) 15%, transparent);
  color: var(--color-warning);
}

.badge-level-error {
  background: color-mix(in srgb, var(--color-danger) 15%, transparent);
  color: var(--color-danger);
}
</style>
