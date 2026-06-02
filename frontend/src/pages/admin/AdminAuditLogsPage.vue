<script setup lang="ts">
import { adminApi } from '@/shared/api/admin-client'
import { useCrudList } from '@/shared/composables/useCrudList'
import PagePagination from '@/shared/ui/PagePagination.vue'
import DataTable from '@/shared/ui/DataTable.vue'
import type { DataTableConfig } from '@/shared/ui/data-table/types'

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

const { items, total, page, totalPages, isLoading, error, setPage } = useCrudList<AuditLog>({
  client: adminApi,
  buildEndpoint: ({ page, limit }) => `/audit-logs?page=${page}&limit=${limit}`,
  errorPrefix: '加载操作日志',
})

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

const config: DataTableConfig<AuditLog> = {
  columns: [
    { key: 'id', label: 'ID', width: '60px' },
    { key: 'admin_name', label: '管理员', formatter: (_, row) => row.admin_name || `管理员#${row.admin_user_id}` },
    { key: 'action', label: '操作', cellClass: 'badge badge-primary', formatter: (v) => actionText[v] || v },
    { key: 'target_type', label: '目标类型', formatter: (v) => targetText[v] || v },
    { key: 'target_id', label: '目标 ID', formatter: (v) => v != null ? String(v) : '—' },
    { key: 'detail', label: '详情', truncate: true, width: '200px', cellClass: 'text-truncate-muted', formatter: (v) => v || '—' },
    { key: 'created_at', label: '时间' },
  ],
  emptyText: '暂无操作日志',
}
</script>

<template>
  <div class="page-container">
    <h1 class="page-title">操作日志</h1>
    <div v-if="error" class="error-message">{{ error }}</div>
    <DataTable :config="config" :data="items" :is-loading="isLoading" />
    <PagePagination :page="page" :total="total" :total-pages="totalPages" @update:page="setPage" />
  </div>
</template>

