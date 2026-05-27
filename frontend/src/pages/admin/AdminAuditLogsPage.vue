<script setup lang="ts">
import { ref, onMounted, watch } from 'vue'
import { adminApi } from '@/shared/api/admin-client'
import type { PaginatedResponse } from '@/shared/api/types'

interface AuditLog {
  id: number
  admin_user_id: number
  admin_name: string
  action: string
  target_type: string
  target_id: number | null
  detail: string
  created_at: string
}

const logs = ref<AuditLog[]>([])
const total = ref(0)
const page = ref(1)
const limit = 20
const error = ref('')
const isLoading = ref(false)

async function loadLogs() {
  isLoading.value = true
  error.value = ''
  try {
    const res = await adminApi.get<PaginatedResponse<AuditLog>>(
      `/audit-logs?page=${page.value}&limit=${limit}`
    )
    logs.value = res.data
    total.value = res.meta.total_items
  } catch {
    error.value = '加载操作日志失败'
  } finally {
    isLoading.value = false
  }
}

onMounted(loadLogs)
watch(page, loadLogs)

const totalPages = () => Math.ceil(total.value / limit)

const actionText: Record<string, string> = {
  delete_user: '删除用户',
  reset_password: '重置密码',
  promote_admin: '提升管理员',
  demote_admin: '取消管理员',
  toggle_task_complete: '切换任务状态',
  update_task: '编辑任务',
  delete_task: '删除任务',
  toggle_reminder_config: '切换提醒配置状态',
  delete_reminder_config: '删除提醒配置',
}

const targetText: Record<string, string> = {
  user: '用户',
  task: '任务',
  reminder_config: '提醒配置',
}

function formatAction(action: string): string {
  return actionText[action] || action
}

function formatTarget(type: string): string {
  return targetText[type] || type
}
</script>

<template>
  <div class="page-container">
    <h1 class="page-title">操作日志</h1>

    <div v-if="error" class="error-message">{{ error }}</div>

    <div class="admin-table-wrap">
      <table class="admin-table">
        <thead>
          <tr>
            <th>ID</th>
            <th>管理员</th>
            <th>操作</th>
            <th>目标类型</th>
            <th>目标 ID</th>
            <th>详情</th>
            <th>时间</th>
          </tr>
        </thead>
        <tbody>
          <tr v-if="isLoading">
            <td colspan="7" style="text-align:center; padding: 2rem;">加载中...</td>
          </tr>
          <tr v-else-if="!logs.length">
            <td colspan="7" style="text-align:center; padding: 2rem; color: var(--color-text-muted);">暂无操作日志</td>
          </tr>
          <tr v-for="l in logs" :key="l.id">
            <td>{{ l.id }}</td>
            <td>{{ l.admin_name || '管理员#' + l.admin_user_id }}</td>
            <td><span class="badge badge-action">{{ formatAction(l.action) }}</span></td>
            <td>{{ formatTarget(l.target_type) }}</td>
            <td>{{ l.target_id ?? '—' }}</td>
            <td class="detail-cell" :title="l.detail">{{ l.detail || '—' }}</td>
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

.badge-action {
  background: var(--color-primary-bg, #e8f0fe);
  color: var(--color-primary, #4a9eff);
}
.detail-cell {
  max-width: 200px;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  color: var(--color-text-muted, #888);
  font-size: 0.8rem;
}
</style>
