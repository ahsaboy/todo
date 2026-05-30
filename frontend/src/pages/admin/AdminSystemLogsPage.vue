<script setup lang="ts">
import { computed, ref, watch, onMounted, onUnmounted } from 'vue'
import { Download, RefreshCw } from 'lucide-vue-next'
import { adminApi } from '@/shared/api/admin-client'
import BaseSelect, { type SelectOption } from '@/shared/ui/BaseSelect.vue'
import type { ApiResponse } from '@/shared/api/types'

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

const logFiles = ref<LogFile[]>([])
const selectedFile = ref('')
const entries = ref<LogEntry[]>([])
const total = ref(0)
const page = ref(1)
const limit = 50
const levelFilter = ref('')
const error = ref('')
const isLoading = ref(false)
const autoRefresh = ref(false)
let refreshTimer: ReturnType<typeof setInterval> | null = null

async function loadLogFiles() {
  try {
    const res = await adminApi.get<ApiResponse<LogFile[]>>('/system-logs')
    logFiles.value = res.data || []
    if (logFiles.value.length && !selectedFile.value) {
      selectedFile.value = logFiles.value[0].filename
    }
  } catch {
    error.value = '加载日志文件列表失败'
  }
}

async function loadEntries() {
  if (!selectedFile.value) {
    entries.value = []
    total.value = 0
    return
  }
  isLoading.value = true
  error.value = ''
  try {
    const params = new URLSearchParams({
      page: String(page.value),
      limit: String(limit),
    })
    if (levelFilter.value) {
      params.set('level', levelFilter.value)
    }
    const res = await adminApi.get<{
      success: boolean
      data: LogEntry[]
      meta: { page: number; limit: number; total_items: number; total_pages: number }
    }>(`/system-logs/${encodeURIComponent(selectedFile.value)}/entries?${params}`)
    entries.value = res.data || []
    total.value = res.meta?.total_items || 0
  } catch {
    error.value = '加载日志条目失败'
  } finally {
    isLoading.value = false
  }
}

function resetPageAndLoad() {
  page.value = 1
  loadEntries()
}

watch(selectedFile, resetPageAndLoad)
watch(levelFilter, resetPageAndLoad)
watch(page, loadEntries)

watch(autoRefresh, (val) => {
  if (val) {
    refreshTimer = setInterval(loadEntries, 5000)
  } else if (refreshTimer) {
    clearInterval(refreshTimer)
    refreshTimer = null
  }
})

onMounted(() => {
  loadLogFiles()
})

onUnmounted(() => {
  if (refreshTimer) {
    clearInterval(refreshTimer)
  }
})

function downloadLog() {
  if (!selectedFile.value) return
  const token = sessionStorage.getItem('admin_api_key')
  const url = `admin/api/system-logs/${encodeURIComponent(selectedFile.value)}/download`
  fetch(url, { headers: { 'Authorization': `Bearer ${token || ''}` } })
    .then((r) => {
      if (!r.ok) throw new Error('download failed')
      return r.blob()
    })
    .then((blob) => {
      const a = document.createElement('a')
      a.href = URL.createObjectURL(blob)
      a.download = selectedFile.value
      a.click()
      URL.revokeObjectURL(a.href)
    })
    .catch(() => {
      error.value = '下载日志文件失败'
    })
}

function levelBadgeClass(level: string): string {
  switch (level?.toLowerCase()) {
    case 'debug': return 'badge badge-debug'
    case 'info': return 'badge badge-info'
    case 'warn': return 'badge badge-warn'
    case 'error': return 'badge badge-error'
    default: return 'badge'
  }
}

function extraFieldEntries(entry: LogEntry): [string, string][] {
  const reserved = new Set(['ts', 'level', 'msg', 'caller'])
  const pairs: [string, string][] = []
  for (const [k, v] of Object.entries(entry)) {
    if (!reserved.has(k)) {
      pairs.push([k, typeof v === 'string' ? v : JSON.stringify(v)])
    }
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
  logFiles.value.map((f) => ({
    label: `${f.date} (${formatSize(f.size_bytes)})`,
    value: f.filename,
  })),
)

const levelOptions: SelectOption<string>[] = [
  { label: '全部级别', value: '' },
  { label: 'debug', value: 'debug' },
  { label: 'info', value: 'info' },
  { label: 'warn', value: 'warn' },
  { label: 'error', value: 'error' },
]

const totalPages = () => Math.ceil(total.value / limit)
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

      <BaseSelect
        v-model="levelFilter"
        :options="levelOptions"
        aria-label="日志级别"
        style="width: 120px;"
      />

      <label class="toolbar-check">
        <input
          v-model="autoRefresh"
          type="checkbox"
        />
        <RefreshCw
          :size="14"
          :class="{ spinning: autoRefresh }"
        />
        <span>自动刷新</span>
      </label>

      <button
        class="btn btn-sm"
        :disabled="!selectedFile"
        @click="downloadLog"
      >
        <Download :size="14" />
        <span>下载</span>
      </button>
    </div>

    <div v-if="error" class="error-message">{{ error }}</div>

    <div class="admin-table-wrap">
      <table class="admin-table log-table">
        <thead>
          <tr>
            <th class="col-ts">时间</th>
            <th class="col-level">级别</th>
            <th class="col-msg">消息</th>
            <th class="col-caller">调用位置</th>
            <th class="col-extra">附加字段</th>
          </tr>
        </thead>
        <tbody>
          <tr v-if="isLoading">
            <td colspan="5" style="text-align:center; padding: 2rem;">加载中...</td>
          </tr>
          <tr v-else-if="!entries.length">
            <td colspan="5" style="text-align:center; padding: 2rem; color: var(--color-text-muted);">
              {{ selectedFile ? '暂无日志条目' : '请选择日志文件' }}
            </td>
          </tr>
          <tr v-for="(entry, idx) in entries" :key="idx">
            <td class="log-ts">{{ entry.ts || '—' }}</td>
            <td>
              <span :class="levelBadgeClass(entry.level)">{{ entry.level || '—' }}</span>
            </td>
            <td class="log-msg" :title="String(entry.msg ?? '')">
              {{ truncateMsg(entry.msg) }}
            </td>
            <td class="log-caller">{{ entry.caller || '—' }}</td>
            <td class="log-extra">
              <template v-if="extraFieldEntries(entry).length">
                <div
                  v-for="([k, v], fi) in extraFieldEntries(entry)"
                  :key="fi"
                  class="extra-field"
                >
                  <span class="extra-key">{{ k }}</span>
                  <span class="extra-val" :title="v">{{ v }}</span>
                </div>
              </template>
              <span v-else>—</span>
            </td>
          </tr>
        </tbody>
      </table>
    </div>

    <div v-if="total > 0" class="admin-pagination">
      <span>共 {{ total }} 条</span>
      <button
        :disabled="page <= 1"
        class="btn btn-sm"
        @click="page--"
      >上一页</button>
      <span>{{ page }} / {{ totalPages() }}</span>
      <button
        :disabled="page >= totalPages()"
        class="btn btn-sm"
        @click="page++"
      >下一页</button>
    </div>
  </div>
</template>

<style scoped>
@import '@/widgets/admin-common.css';

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

@keyframes spin {
  from { transform: rotate(0deg); }
  to { transform: rotate(360deg); }
}

.spinning {
  animation: spin 1s linear infinite;
}

.log-table .col-ts { width: 170px; }
.log-table .col-level { width: 70px; }
.log-table .col-caller { width: 180px; }

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

.log-extra {
  font-size: 0.75rem;
  font-family: monospace;
  vertical-align: top;
}

.extra-field {
  display: flex;
  gap: 4px;
  line-height: 1.5;
}

.extra-key {
  color: var(--color-text-muted);
  flex-shrink: 0;
}

.extra-val {
  color: var(--color-text);
  word-break: break-all;
}

/* 级别徽章 */
.badge-debug {
  background: var(--color-surface-muted);
  color: var(--color-text-muted);
}

.badge-info {
  background: color-mix(in srgb, var(--color-primary) 15%, transparent);
  color: var(--color-primary);
}

.badge-warn {
  background: color-mix(in srgb, var(--color-warning) 15%, transparent);
  color: var(--color-warning);
}

.badge-error {
  background: color-mix(in srgb, var(--color-danger) 15%, transparent);
  color: var(--color-danger);
}
</style>
