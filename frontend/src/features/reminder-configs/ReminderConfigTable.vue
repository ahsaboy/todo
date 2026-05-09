<template>
  <div class="config-table">
    <table>
      <thead>
        <tr>
          <th>名称</th>
          <th>渠道类型</th>
          <th>启用状态</th>
          <th>请求方法</th>
          <th>Webhook URL</th>
          <th>操作</th>
        </tr>
      </thead>
      <tbody>
        <tr v-for="config in configs" :key="config.id">
          <td>{{ config.name }}</td>
          <td>{{ formatChannelType(config.channelType) }}</td>
          <td>
            <span class="status-badge" :class="{ enabled: config.enabled }">
              {{ config.enabled ? '启用' : '禁用' }}
            </span>
          </td>
          <td>{{ config.webhookMethod || 'POST' }}</td>
          <td class="url-cell">{{ maskUrl(config.webhookUrl) }}</td>
          <td class="actions">
            <button class="btn-icon" type="button" @click="$emit('edit', config)">编辑</button>
            <button class="btn-icon btn-danger" type="button" @click="$emit('delete', config.id)">
              删除
            </button>
          </td>
        </tr>
      </tbody>
    </table>
  </div>
</template>

<script setup lang="ts">
import type { ReminderConfig } from '@/entities/reminder-config/model'

defineProps<{
  configs: ReminderConfig[]
}>()

defineEmits<{
  edit: [config: ReminderConfig]
  delete: [id: number]
}>()

function formatChannelType(type: string): string {
  const types: Record<string, string> = {
    webhook: 'Webhook',
    feishu: '飞书',
    dingtalk: '钉钉',
    wecom: '企业微信',
    slack: 'Slack',
  }
  return types[type] || type
}

function maskUrl(url: string): string {
  if (!url) return '-'
  try {
    const parsed = new URL(url)
    const path =
      parsed.pathname.length > 20 ? parsed.pathname.slice(0, 20) + '...' : parsed.pathname
    return parsed.hostname + path
  } catch {
    return url.slice(0, 30) + '...'
  }
}
</script>

<style scoped>
.config-table {
  overflow-x: auto;
}

table {
  width: 100%;
  border-collapse: collapse;
}

th,
td {
  padding: 12px;
  text-align: left;
  border-bottom: 1px solid var(--color-border);
}

th {
  font-weight: 500;
  color: var(--color-text-muted);
  font-size: 13px;
}

.status-badge {
  padding: 2px 8px;
  border-radius: 4px;
  font-size: 12px;
  background: var(--color-surface-muted);
  color: var(--color-text-muted);
}

.status-badge.enabled {
  background: #dcfce7;
  color: #16a34a;
}

.url-cell {
  max-width: 200px;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.actions {
  display: flex;
  gap: 8px;
}

.btn-icon {
  background: none;
  border: none;
  padding: 4px 8px;
  cursor: pointer;
  color: var(--color-text-muted);
  font-size: 13px;
}

.btn-danger:hover {
  color: var(--color-danger);
}
</style>
