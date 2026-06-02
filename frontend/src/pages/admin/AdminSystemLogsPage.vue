<script setup lang="ts">
import { computed, ref, watch, onUnmounted } from 'vue'
import { Download, RefreshCw } from 'lucide-vue-next'
import { ADMIN_API_BASE_URL } from '@/shared/config/api'
import { adminApi } from '@/shared/api/admin-client'
import BaseSelect, { type SelectOption } from '@/shared/ui/BaseSelect.vue'
import PagePagination from '@/shared/ui/PagePagination.vue'
import DataTable from '@/shared/ui/DataTable.vue'
import { useFetch } from '@/shared/composables/useFetch'
import { useCrudList } from '@/shared/composables/useCrudList'
import type { DataTableConfig } from '@/shared/ui/data-table/types'
import type { ApiResponse } from '@/shared/api/types'
import ExtraFieldsCell from '@/features/admin/ExtraFieldsCell.vue'

interface LogFile {
  filename: string
  date: string
  size_bytes: number
}

interface LogEntry {
  ts: string
  level: string
  msg: string
  caller?: string
  [key: string]: unknown
}

interface LogEntryRow extends LogEntry {
  _extraFields: [string, string][]
}

const selectedFile = ref('')
const autoRefresh = ref(false)
let refreshTimer: ReturnType<typeof setInterval> | null = null

// 日志文件列表
const { data: logFilesData } = useFetch<LogFile[]>({
  fetcher: () => adminApi.get<ApiResponse<LogFile[]>>('/system-logs').then(r => r.data || []),
  errorPrefix: '加载日志文件列表',
})

const logFiles = computed(() => logFilesData.value ?? [])

// 自动选中第一个文件
watch(logFiles, (files) => {
  if (files.length && !selectedFile.value) {
    selectedFile.value = files[0].filename
  }
}, { immediate: true })

// 日志条目（分页列表）
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

// 文件切换时重置分页并加载
watch(selectedFile, () => {
  entries.setPage(1)
  entries.load()
})

// 自动刷新
watch(autoRefresh, (val) => {
  if (val) {
    refreshTimer = setInterval(() => entries.load(), 5000)
  } else if (refreshTimer) {
    clearInterval(refreshTimer)
    refreshTimer = null
  }
})

onUnmounted(() => { if (refreshTimer) clearInterval(refreshTimer) })

function downloadLog() {
  if (!selectedFile.value) return
  const token = sessionStorage.getItem('admin_api_key')
  const url = `${ADMIN_API_BASE_URL}/system-logs/${encodeURIComponent(selectedFile.value)}/download`
  fetch(url, { headers: { 'Authorization': `Bearer ${token || ''}` } })
    .then((r) => { if (!r.ok) throw new Error(); return r.blob() })
    .then((blob) => {
      const a = document.createElement('a')
      a.href = URL.createObjectURL(blob)
      a.download = selectedFile.value
      a.click()
      URL.revokeObjectURL(a.href)
    })
    .catch(() => { entries.error.value = '下载日志文件失败' })
}

function extraFieldEntries(entry: LogEntry): [string, string][] {
  const reserved = new Set(['ts', 'level', 'msg', 'caller'])
  const pairs: [string, string][] = []
  for (const [k, v] of Object.entries(entry)) {
    if (!reserved.has(k)) pairs.push([k, typeof v === 'string' ? v : JSON.stringify(v)])
  }
  return pairs
}

function formatSize(bytes: number): string {
  if (bytes < 1024) return bytes + ' B'
  if (bytes < 1024 * 1024) return (bytes / 1024).toFixed(1) + ' KB'
  return (bytes / (1024 * 1024)).toFixed(1) + ' MB'
}

function truncateMsg(msg: unknown): string {
  if (typeof msg !== 'string') return String(msg ?? '')
  return msg.length > 120 ? msg.slice(0, 120) + '…' : msg
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

function levelBadgeClass(row: LogEntryRow): string {
  switch (row.level?.toLowerCase()) {
    case 'debug': return 'badge badge-level-debug'
    case 'info': return 'badge badge-level-info'
    case 'warn': return 'badge badge-level-warn'
    case 'error': return 'badge badge-level-error'
    default: return 'badge'
  }
}

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
    <h1 class="page-title">系统日志</h1>

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
