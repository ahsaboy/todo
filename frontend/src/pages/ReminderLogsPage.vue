<template>
  <div class="page">
    <div class="page-header">
      <h2>提醒日志</h2>
      <button class="btn-secondary" type="button" @click="fetchLogs"><RefreshCw :size="14" /> 刷新</button>
    </div>

    <Transition name="sk-fade" mode="out-in">
      <TableSkeleton v-if="loading" key="skeleton" :columns="6" :col-widths="['140px', '120px', '100px', '60px', '60px', '200px']" />

      <template v-else key="content">
        <div v-if="error" class="page-error">
          <p>{{ error }}</p>
          <button type="button" @click="fetchLogs">重试</button>
        </div>

        <div v-else-if="logs.length === 0" class="page-empty">
          <p>暂无提醒日志</p>
        </div>

        <div v-else>
          <div class="log-card-list">
            <article v-for="item in logs" :key="item.id" class="log-card">
              <div class="log-card-header">
                <div class="log-card-heading">
                  <span class="status-tag" :class="item.status">
                    {{ item.status === 'success' ? '成功' : '失败' }}
                  </span>
                  <div class="log-card-title">{{ item.taskTitle || `任务 #${item.taskId}` }}</div>
                </div>
                <time class="log-card-time">{{ formatDate(item.createdAt) }}</time>
              </div>

              <dl class="log-meta-list">
                <div class="log-meta-row">
                  <dt>渠道</dt>
                  <dd>{{ item.channelName }}</dd>
                </div>
                <div class="log-meta-row">
                  <dt>类型</dt>
                  <dd>{{ item.channelType }}</dd>
                </div>
                <div class="log-meta-row">
                  <dt>尝试</dt>
                  <dd>{{ item.attempts }}</dd>
                </div>
              </dl>

              <div v-if="item.errorMessage" class="log-card-error">
                <div class="log-card-error-label">错误信息</div>
                <p>{{ item.errorMessage }}</p>
              </div>
            </article>
          </div>

          <div class="log-table-wrap">
            <table class="log-table">
              <thead>
                <tr>
                  <th>时间</th>
                  <th>任务</th>
                  <th>渠道</th>
                  <th>状态</th>
                  <th>尝试</th>
                  <th>错误</th>
                </tr>
              </thead>
              <tbody>
                <tr v-for="item in logs" :key="item.id">
                  <td>{{ formatDate(item.createdAt) }}</td>
                  <td class="task-cell">{{ item.taskTitle || `任务 #${item.taskId}` }}</td>
                  <td>{{ item.channelName }}</td>
                  <td>
                    <span class="status-tag" :class="item.status">
                      {{ item.status === 'success' ? '成功' : '失败' }}
                    </span>
                  </td>
                  <td>{{ item.attempts }}</td>
                  <td class="error-cell">{{ item.errorMessage || '-' }}</td>
                </tr>
              </tbody>
            </table>
          </div>
        </div>
      </template>
    </Transition>

    <div v-if="meta.total_pages > 1" class="pager">
      <button type="button" :disabled="page <= 1" @click="changePage(page - 1)">上一页</button>
      <span>{{ page }} / {{ meta.total_pages }}</span>
      <button type="button" :disabled="page >= meta.total_pages" @click="changePage(page + 1)">
        下一页
      </button>
    </div>
  </div>
</template>

<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { RefreshCw } from 'lucide-vue-next'
import TableSkeleton from '@/shared/ui/TableSkeleton.vue'
import { getReminderLogs } from '@/entities/reminder-config/api'
import { toReminderLog } from '@/entities/reminder-config/mapper'
import type { ReminderLog } from '@/entities/reminder-config/model'
import type { PageMeta } from '@/shared/api/types'

const logs = ref<ReminderLog[]>([])
const loading = ref(false)
const error = ref<string | null>(null)
const page = ref(1)
const limit = 20
const meta = ref<PageMeta>({
  page: 1,
  limit,
  total_items: 0,
  total_pages: 0,
})

onMounted(() => {
  fetchLogs()
})

async function fetchLogs() {
  loading.value = true
  error.value = null
  try {
    const response = await getReminderLogs(page.value, limit)
    logs.value = Array.isArray(response.data) ? response.data.map(toReminderLog) : []
    meta.value = response.meta
  } catch (e) {
    error.value = e instanceof Error ? e.message : '加载失败'
  } finally {
    loading.value = false
  }
}

async function changePage(nextPage: number) {
  page.value = nextPage
  await fetchLogs()
}

function formatDate(value: string) {
  if (!value) return '-'
  const date = new Date(value)
  if (Number.isNaN(date.getTime())) {
    return value
  }
  return `${date.getFullYear()}年${date.getMonth() + 1}月${date.getDate()}日 ${String(date.getHours()).padStart(2, '0')}:${String(date.getMinutes()).padStart(2, '0')}`
}
</script>

<style scoped>
.btn-secondary,
.pager button {
  padding: 8px 14px;
  background: var(--color-surface);
  border: 1px solid var(--color-border);
  border-radius: 6px;
  color: var(--color-text);
  cursor: pointer;
}

.pager button:disabled {
  cursor: not-allowed;
  opacity: 0.5;
}

.log-table-wrap {
  overflow-x: auto;
  background: var(--color-surface);
  border: 1px solid var(--color-border);
  border-radius: 6px;
}

.log-card-list {
  display: none;
}

.log-card {
  display: flex;
  flex-direction: column;
  gap: 12px;
  padding: 14px;
  background: var(--color-surface);
  border: 1px solid var(--color-border);
  border-radius: 8px;
  transition: background-color var(--motion-duration-fast) var(--motion-ease-standard);
}

.log-card:hover {
  background-color: var(--color-surface-muted);
}

.log-card-header {
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
  gap: 12px;
}

.log-card-heading {
  display: grid;
  gap: 10px;
  min-width: 0;
  flex: 1;
}

.log-card-time {
  color: var(--color-text-muted);
  font-size: 12px;
  line-height: 1.5;
  text-align: right;
  white-space: nowrap;
}

.log-card-title {
  color: var(--color-text);
  font-size: 15px;
  font-weight: 600;
  line-height: 1.5;
  overflow-wrap: anywhere;
}

.log-meta-list {
  display: grid;
  gap: 8px;
  margin: 0;
}

.log-meta-row {
  display: grid;
  grid-template-columns: 40px minmax(0, 1fr);
  gap: 8px;
  min-width: 0;
  font-size: 13px;
  line-height: 1.5;
}

.log-meta-row dt {
  color: var(--color-text-muted);
}

.log-meta-row dd {
  margin: 0;
  min-width: 0;
  color: var(--color-text);
  overflow-wrap: anywhere;
}

.log-card-error {
  padding: 10px 12px;
  background: var(--color-surface-muted);
  border-radius: 6px;
}

.log-card-error-label {
  margin-bottom: 4px;
  color: var(--color-text-muted);
  font-size: 12px;
  font-weight: 600;
}

.log-card-error p {
  margin: 0;
  color: var(--color-danger);
  font-size: 13px;
  line-height: 1.5;
  white-space: pre-wrap;
  overflow-wrap: anywhere;
  word-break: break-word;
}

.log-table {
  width: 100%;
  border-collapse: collapse;
  min-width: 860px;
}

.log-table th,
.log-table td {
  padding: 12px;
  text-align: left;
  border-bottom: 1px solid var(--color-border);
  font-size: 14px;
  vertical-align: top;
}

.log-table th {
  background: var(--color-surface-muted);
  color: var(--color-text-muted);
  font-weight: 600;
}

.log-table tr:last-child td {
  border-bottom: none;
}

.task-cell {
  max-width: 220px;
  font-weight: 600;
}

.error-cell {
  max-width: 360px;
  color: var(--color-text-muted);
  word-break: break-word;
}

.status-tag {
  display: inline-flex;
  align-items: center;
  min-width: 44px;
  justify-content: center;
  padding: 2px 8px;
  border-radius: 999px;
  font-size: 12px;
  font-weight: 600;
}

.status-tag.success {
  background: color-mix(in srgb, var(--color-glow-success) 72%, transparent);
  border: 1px solid color-mix(in srgb, var(--color-success) 18%, transparent);
  color: var(--color-success);
}

.status-tag.failed {
  background: color-mix(in srgb, var(--color-glow-danger) 84%, transparent);
  border: 1px solid color-mix(in srgb, var(--color-danger) 26%, transparent);
  color: var(--color-danger);
}

.pager {
  display: flex;
  justify-content: flex-end;
  align-items: center;
  gap: 12px;
  color: var(--color-text-muted);
  flex-wrap: wrap;
}

@media (max-width: 767px) {
  .log-card-list {
    display: flex;
    flex-direction: column;
    gap: 12px;
  }

  .log-table-wrap {
    display: none;
  }

  .log-card {
    gap: 14px;
    padding: 16px;
    border-radius: 12px;
  }

  .log-meta-list {
    grid-template-columns: repeat(3, minmax(0, 1fr));
    gap: 10px;
  }

  .log-meta-row {
    display: flex;
    flex-direction: column;
    gap: 4px;
    padding: 10px 12px;
    background: var(--color-surface-muted);
    border-radius: 10px;
  }

  .log-meta-row dt,
  .log-meta-row dd {
    font-size: 12px;
    line-height: 1.5;
  }

  .log-meta-row dd {
    font-size: 13px;
    font-weight: 500;
  }

  .log-card-error {
    padding: 12px;
    border-left: 3px solid color-mix(in srgb, var(--color-danger) 55%, transparent);
    border-radius: 10px;
  }

  .pager {
    display: grid;
    grid-template-columns: minmax(0, 1fr) auto minmax(0, 1fr);
    gap: 8px;
    align-items: center;
  }

  .pager button {
    width: 100%;
    min-height: 36px;
  }

  .pager button:last-child {
    justify-self: end;
  }

  .pager span {
    text-align: center;
    white-space: nowrap;
  }
}

@media (max-width: 479px) {
  .log-card-header {
    flex-direction: column;
  }

  .log-card-time {
    text-align: left;
  }

  .log-meta-list {
    grid-template-columns: minmax(0, 1fr);
  }
}

@media (max-width: 359px) {
  .page-header .btn-secondary {
    width: 100%;
  }
}
</style>
