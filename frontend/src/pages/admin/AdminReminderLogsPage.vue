<script setup lang="ts">
import { ref, onMounted, watch } from 'vue'
import { adminApi } from '@/shared/api/admin-client'
import type { PaginatedResponse } from '@/shared/api/types'

interface ReminderLog {
  id: number
  user_id: number
  username: string
  task_id: number
  task_title: string
  reminder_config_id: number | null
  channel_name: string
  channel_type: string
  status: string
  attempts: number
  error_message: string
  created_at: string
}

const logs = ref<ReminderLog[]>([])
const total = ref(0)
const page = ref(1)
const limit = 20
const filterUserId = ref('')
const filterStatus = ref('')
const error = ref('')
const isLoading = ref(false)

async function loadLogs() {
  isLoading.value = true
  error.value = ''
  try {
    const params = new URLSearchParams({
      page: String(page.value),
      limit: String(limit),
      ...(filterUserId.value ? { user_id: filterUserId.value } : {}),
      ...(filterStatus.value ? { status: filterStatus.value } : {}),
    })
    const res = await adminApi.get<PaginatedResponse<ReminderLog>>(
      `/reminder-logs?${params}`
    )
    logs.value = res.data
    total.value = res.meta.total_items
  } catch {
    error.value = '加载提醒日志失败'
  } finally {
    isLoading.value = false
  }
}

onMounted(loadLogs)
watch(page, loadLogs)

function handleFilter() {
  page.value = 1
  loadLogs()
}

const totalPages = () => Math.ceil(total.value / limit)

function statusClass(status: string): string {
  if (status === 'sent') return 'badge badge-done'
  if (status === 'failed') return 'badge badge-fail'
  return 'badge badge-pending'
}
</script>

<template>
  <div class="page-container">
    <h1 class="page-title">提醒日志</h1>

    <div class="admin-toolbar">
      <input
        v-model="filterUserId"
        type="number"
        placeholder="用户 ID 筛选"
        class="admin-search-input"
        style="max-width: 140px;"
      />
      <select v-model="filterStatus" class="admin-search-input" style="max-width: 120px;">
        <option value="">全部状态</option>
        <option value="sent">已发送</option>
        <option value="failed">失败</option>
      </select>
      <button class="btn btn-primary" @click="handleFilter">筛选</button>
    </div>

    <div v-if="error" class="error-message">{{ error }}</div>

    <div class="admin-table-wrap">
      <table class="admin-table">
        <thead>
          <tr>
            <th>ID</th>
            <th>用户</th>
            <th>任务</th>
            <th>渠道</th>
            <th>状态</th>
            <th>重试次数</th>
            <th>错误信息</th>
            <th>时间</th>
          </tr>
        </thead>
        <tbody>
          <tr v-if="isLoading">
            <td colspan="8" style="text-align:center; padding: 2rem;">加载中...</td>
          </tr>
          <tr v-else-if="!logs.length">
            <td colspan="8" style="text-align:center; padding: 2rem; color: var(--color-text-muted);">暂无提醒日志</td>
          </tr>
          <tr v-for="l in logs" :key="l.id">
            <td>{{ l.id }}</td>
            <td>{{ l.username || '用户#' + l.user_id }}</td>
            <td class="task-cell" :title="l.task_title">#{{ l.task_id }} {{ l.task_title }}</td>
            <td>{{ l.channel_name || l.channel_type }}</td>
            <td><span :class="statusClass(l.status)">{{ l.status }}</span></td>
            <td>{{ l.attempts }}</td>
            <td class="error-cell" :title="l.error_message">{{ l.error_message || '—' }}</td>
            <td>{{ l.created_at }}</td>
          </tr>
        </tbody>
      </table>
    </div>

    <div class="admin-pagination">
      <span>共 {{ total }} 条</span>
      <button :disabled="page <= 1" class="btn btn-sm" @click="page--">上一页</button>
      <span>{{ page }} / {{ totalPages() }}</span>
      <button :disabled="page >= totalPages()" class="btn btn-sm" @click="page++">下一页</button>
    </div>
  </div>
</template>

<style scoped>
@import '@/widgets/admin-common.css';

.badge-fail { background: var(--color-danger-bg, #f8d7da); color: var(--color-danger, #721c24); }
.task-cell { max-width: 180px; overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }
.error-cell { max-width: 200px; overflow: hidden; text-overflow: ellipsis; white-space: nowrap; color: var(--color-text-muted, #888); font-size: 0.8rem; }
</style>
