<template>
  <div class="config-list">
    <article v-for="config in configs" :key="config.id" class="config-card">
      <div class="config-card-header">
        <div class="config-card-title">{{ config.name }}</div>
        <span class="status-badge" :class="{ enabled: config.enabled }">
          {{ config.enabled ? '启用' : '禁用' }}
        </span>
      </div>

      <dl class="config-meta-list">
        <div class="config-meta-row">
          <dt>渠道</dt>
          <dd>{{ formatChannelType(config.channelType) }}</dd>
        </div>
        <div class="config-meta-row">
          <dt>方法</dt>
          <dd>{{ config.webhookMethod || 'POST' }}</dd>
        </div>
        <div class="config-meta-row">
          <dt>重试</dt>
          <dd>{{ formatRetry(config.maxRetries, config.retryDelaySeconds) }}</dd>
        </div>
        <div class="config-meta-row">
          <dt>地址</dt>
          <dd>{{ maskUrl(config.webhookUrl) }}</dd>
        </div>
      </dl>

      <div class="card-actions">
        <button class="action-btn" type="button" @click="$emit('edit', config)">编辑</button>
        <button class="action-btn action-btn-danger" type="button" @click="$emit('delete', config.id)">
          删除
        </button>
      </div>
    </article>
  </div>

  <div class="config-table-wrap">
    <table class="config-table">
      <thead>
        <tr>
          <th>名称</th>
          <th>渠道类型</th>
          <th>状态</th>
          <th>请求方法</th>
          <th>重试策略</th>
          <th>Webhook URL</th>
          <th>操作</th>
        </tr>
      </thead>
      <tbody>
        <tr v-for="config in configs" :key="config.id">
          <td class="name-cell">{{ config.name }}</td>
          <td>{{ formatChannelType(config.channelType) }}</td>
          <td>
            <span class="status-badge" :class="{ enabled: config.enabled }">
              {{ config.enabled ? '启用' : '禁用' }}
            </span>
          </td>
          <td>{{ config.webhookMethod || 'POST' }}</td>
          <td>{{ formatRetry(config.maxRetries, config.retryDelaySeconds) }}</td>
          <td class="url-cell">{{ maskUrl(config.webhookUrl) }}</td>
          <td>
            <div class="table-actions">
              <button class="btn-icon" type="button" @click="$emit('edit', config)">编辑</button>
              <button class="btn-icon btn-danger" type="button" @click="$emit('delete', config.id)">
                删除
              </button>
            </div>
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

function formatRetry(maxRetries: number, retryDelaySeconds: number): string {
  if (!maxRetries) return '不重试'
  return `${maxRetries} 次 / ${retryDelaySeconds || 0} 秒`
}
</script>

<style scoped>
.config-table-wrap {
  overflow-x: auto;
  background: var(--color-surface);
  border: 1px solid var(--color-border);
  border-radius: 6px;
}

.config-list {
  display: none;
}

.config-card {
  display: flex;
  flex-direction: column;
  gap: 12px;
  padding: 14px;
  background: var(--color-surface);
  border: 1px solid var(--color-border);
  border-radius: 8px;
  min-width: 0;
}

.config-card-header {
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
  gap: 12px;
  min-width: 0;
}

.config-card-title {
  color: var(--color-text);
  font-size: 15px;
  font-weight: 600;
  line-height: 1.5;
  overflow-wrap: anywhere;
}

.config-meta-list {
  display: grid;
  gap: 8px;
  margin: 0;
}

.config-meta-row {
  display: grid;
  grid-template-columns: 40px minmax(0, 1fr);
  gap: 8px;
  min-width: 0;
  font-size: 13px;
  line-height: 1.5;
}

.config-meta-row dt {
  color: var(--color-text-muted);
}

.config-meta-row dd {
  margin: 0;
  min-width: 0;
  color: var(--color-text);
  overflow-wrap: anywhere;
}

.config-table {
  width: 100%;
  border-collapse: collapse;
  min-width: 920px;
}

.config-table th,
.config-table td {
  padding: 12px;
  text-align: left;
  border-bottom: 1px solid var(--color-border);
  font-size: 14px;
  vertical-align: top;
}

.config-table th {
  background: var(--color-surface-muted);
  color: var(--color-text-muted);
  font-weight: 600;
}

.config-table tr:last-child td {
  border-bottom: none;
}

.status-badge {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  min-width: 44px;
  padding: 2px 8px;
  border-radius: 999px;
  font-size: 12px;
  font-weight: 600;
  background: var(--color-surface-muted);
  color: var(--color-text-muted);
}

.status-badge.enabled {
  background: #dcfce7;
  color: #16a34a;
}

.name-cell {
  max-width: 180px;
  font-weight: 600;
  overflow-wrap: anywhere;
}

.url-cell {
  max-width: 260px;
  color: var(--color-text-muted);
  overflow-wrap: anywhere;
}

.table-actions {
  display: flex;
  gap: 8px;
  flex-wrap: wrap;
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

.card-actions {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 8px;
}

.action-btn {
  min-height: 36px;
  padding: 6px 10px;
  border: 1px solid var(--color-border);
  border-radius: 6px;
  background: var(--color-surface);
  color: var(--color-text);
  font-size: 13px;
  cursor: pointer;
}

.action-btn-danger {
  color: var(--color-danger);
}

@media (max-width: 767px) {
  .config-list {
    display: flex;
    flex-direction: column;
    gap: 12px;
  }

  .config-table-wrap {
    display: none;
  }
}

@media (max-width: 359px) {
  .config-card-header {
    flex-direction: column;
  }
}
</style>
