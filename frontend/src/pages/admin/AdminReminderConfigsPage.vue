<script setup lang="ts">
import { adminApi } from '@/shared/api/admin-client'
import { useCrudList } from '@/shared/composables/useCrudList'
import PagePagination from '@/shared/ui/PagePagination.vue'
import DataTable from '@/shared/ui/DataTable.vue'
import type { DataTableConfig } from '@/shared/ui/data-table/types'

interface ReminderConfig {
  id: number
  user_id: number
  username: string
  name: string
  channel_type: string
  webhook_url: string
  webhook_method: string
  max_retries: number
  retry_delay_seconds: number
  enabled: boolean
  created_at: string
}

const { items, total, page, totalPages, isLoading, error, setPage, deleteItem, mutate } =
  useCrudList<ReminderConfig>({
    client: adminApi,
    buildEndpoint: ({ page, limit }) => `/reminder-configs?page=${page}&limit=${limit}`,
    errorPrefix: '加载提醒配置',
  })

function toggleEnabled(row: ReminderConfig) {
  const action = row.enabled ? '禁用' : '启用'
  if (!confirm(`确定${action}提醒配置 "${row.name}"？`)) return
  mutate(`/reminder-configs/${row.id}/toggle`, 'PATCH')
}

function deleteConfig(row: ReminderConfig) {
  deleteItem(`/reminder-configs/${row.id}`, `确定删除提醒配置 "${row.name}"？此操作不可恢复！`)
}

const config: DataTableConfig<ReminderConfig> = {
  columns: [
    { key: 'id', label: 'ID', width: '60px' },
    { key: 'username', label: '用户', formatter: (_, row) => row.username || `用户#${row.user_id}` },
    { key: 'name', label: '名称' },
    { key: 'channel_type', label: '渠道类型', cellClass: 'badge badge-primary' },
    { key: 'webhook_url', label: 'Webhook URL', truncate: true, width: '200px' },
    { key: 'max_retries', label: '最大重试' },
    {
      key: 'enabled', label: '状态',
      cellClass: (row) => row.enabled ? 'badge badge-done' : 'badge badge-muted',
      formatter: (v) => v ? '启用' : '禁用',
    },
    { key: 'created_at', label: '创建时间' },
  ],
  actions: [
    { id: 'toggle', label: (row) => row.enabled ? '禁用' : '启用', onClick: toggleEnabled },
    { id: 'delete', label: '删除', variant: 'danger', onClick: deleteConfig },
  ],
  emptyText: '暂无提醒配置',
}
</script>

<template>
  <div class="page-container">
    <h1 class="page-title">提醒配置</h1>
    <div v-if="error" class="error-message">{{ error }}</div>
    <DataTable :config="config" :data="items" :is-loading="isLoading" />
    <PagePagination :page="page" :total="total" :total-pages="totalPages" @update:page="setPage" />
  </div>
</template>

