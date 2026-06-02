<script setup lang="ts">
import { adminApi } from '@/shared/api/admin-client'
import { useCrudList } from '@/shared/composables/useCrudList'
import PagePagination from '@/shared/ui/PagePagination.vue'
import DataTable from '@/shared/ui/DataTable.vue'
import type { DataTableConfig } from '@/shared/ui/data-table/types'

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

const { items, total, page, totalPages, isLoading, error, filters, setPage, applyFilters, handleFilterChange } =
  useCrudList<ReminderLog>({
    client: adminApi,
    buildEndpoint: ({ page, limit, filters }) => {
      const params = new URLSearchParams({ page: String(page), limit: String(limit) })
      if (filters.user_id) params.set('user_id', filters.user_id)
      if (filters.status) params.set('status', filters.status)
      return `/reminder-logs?${params}`
    },
    errorPrefix: '加载提醒日志',
  })

const config: DataTableConfig<ReminderLog> = {
  columns: [
    { key: 'id', label: 'ID', width: '60px' },
    { key: 'username', label: '用户', formatter: (_, row) => row.username || `用户#${row.user_id}` },
    {
      key: 'task_title', label: '任务', truncate: true, width: '180px',
      formatter: (_, row) => `#${row.task_id} ${row.task_title}`,
    },
    { key: 'channel_name', label: '渠道', formatter: (_, row) => row.channel_name || row.channel_type },
    {
      key: 'status', label: '状态',
      cellClass: (row) => {
        if (row.status === 'sent') return 'badge badge-done'
        if (row.status === 'failed') return 'badge badge-danger'
        return 'badge badge-pending'
      },
    },
    { key: 'attempts', label: '重试次数' },
    { key: 'error_message', label: '错误信息', truncate: true, width: '200px', cellClass: 'text-truncate-muted', formatter: (v) => v || '—' },
    { key: 'created_at', label: '时间' },
  ],
  filters: [
    { id: 'user_id', type: 'number', placeholder: '用户 ID 筛选', value: filters.value.user_id ?? '', width: 'narrow' },
    {
      id: 'status', type: 'select', value: filters.value.status ?? '',
      options: [
        { label: '全部状态', value: '' },
        { label: '已发送', value: 'sent' },
        { label: '失败', value: 'failed' },
      ],
    },
  ],
  emptyText: '暂无提醒日志',
}
</script>

<template>
  <div class="page-container">
    <h1 class="page-title">提醒日志</h1>
    <div v-if="error" class="error-message">{{ error }}</div>
    <DataTable
      :config="config"
      :data="items"
      :is-loading="isLoading"
      @filter-change="handleFilterChange"
      @apply-filters="applyFilters"
    />
    <PagePagination :page="page" :total="total" :total-pages="totalPages" @update:page="setPage" />
  </div>
</template>

