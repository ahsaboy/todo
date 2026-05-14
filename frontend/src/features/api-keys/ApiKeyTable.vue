<template>
  <div class="key-list motion-stagger">
    <article v-for="key in keys" :key="key.id" class="key-card">
      <div class="key-card-header">
        <div class="key-card-heading">
          <div class="key-card-title">{{ key.name || '未命名' }}</div>
          <span class="key-prefix">{{ getKeyPrefix() }}</span>
        </div>
      </div>

      <dl class="key-meta-list">
        <div class="key-meta-row">
          <dt>创建</dt>
          <dd>{{ formatDate(key.createdAt) }}</dd>
        </div>
        <div class="key-meta-row">
          <dt>使用</dt>
          <dd>{{ key.lastUsedAt ? formatDate(key.lastUsedAt) : '从未' }}</dd>
        </div>
      </dl>

      <button class="action-btn action-btn-danger" type="button" @click="$emit('revoke', key.id)"><Ban :size="14" /> 撤销</button>
    </article>
  </div>

  <div class="key-table-wrap">
    <table class="key-table">
      <thead>
        <tr>
          <th>名称</th>
          <th>Key 前缀</th>
          <th>创建时间</th>
          <th>最后使用</th>
          <th>操作</th>
        </tr>
      </thead>
      <tbody class="motion-stagger">
        <tr v-for="key in keys" :key="key.id">
          <td class="name-cell">{{ key.name || '未命名' }}</td>
          <td>
            <span class="key-prefix">{{ getKeyPrefix() }}</span>
          </td>
          <td>{{ formatDate(key.createdAt) }}</td>
          <td>{{ key.lastUsedAt ? formatDate(key.lastUsedAt) : '从未' }}</td>
          <td>
            <button class="btn-icon btn-icon-text btn-danger" type="button" @click="$emit('revoke', key.id)"><Ban :size="14" /> 撤销</button>
          </td>
        </tr>
      </tbody>
    </table>
  </div>
</template>

<script setup lang="ts">
import type { ApiKeyInfo } from '@/entities/api-key/model'
import { Ban } from 'lucide-vue-next'

defineProps<{
  keys: ApiKeyInfo[]
}>()

defineEmits<{
  revoke: [id: number]
}>()

function getKeyPrefix(): string {
  return 'key_****'
}

function formatDate(dateStr: string): string {
  if (!dateStr) return '-'
  const date = new Date(dateStr)
  if (Number.isNaN(date.getTime())) return dateStr
  return `${date.getMonth() + 1}月${date.getDate()}日 ${String(date.getHours()).padStart(2, '0')}:${String(date.getMinutes()).padStart(2, '0')}`
}
</script>

<style scoped>
.key-table-wrap {
  overflow-x: auto;
  background: var(--color-surface);
  border: 1px solid var(--color-border);
  border-radius: 6px;
}

.key-list {
  display: none;
}

.key-card {
  display: flex;
  flex-direction: column;
  gap: 12px;
  padding: 14px;
  background: var(--color-surface);
  border: 1px solid var(--color-border);
  border-radius: 8px;
  min-width: 0;
  transition: background-color 0.15s ease;
}

.key-card:hover {
  background-color: var(--color-surface-muted);
}

.key-card-header {
  display: flex;
  min-width: 0;
}

.key-card-heading {
  display: grid;
  gap: 8px;
  min-width: 0;
  width: 100%;
}

.key-card-title {
  color: var(--color-text);
  font-size: 15px;
  font-weight: 600;
  line-height: 1.5;
  overflow-wrap: anywhere;
}

.key-meta-list {
  display: grid;
  gap: 8px;
  margin: 0;
}

.key-meta-row {
  display: grid;
  grid-template-columns: 40px minmax(0, 1fr);
  gap: 8px;
  min-width: 0;
  font-size: 13px;
  line-height: 1.5;
}

.key-meta-row dt {
  color: var(--color-text-muted);
}

.key-meta-row dd {
  margin: 0;
  min-width: 0;
  color: var(--color-text);
  overflow-wrap: anywhere;
}

.key-table {
  width: 100%;
  border-collapse: collapse;
  min-width: 720px;
}

.key-table th,
.key-table td {
  padding: 12px;
  text-align: left;
  border-bottom: 1px solid var(--color-border);
  font-size: 14px;
  vertical-align: top;
}

.key-table th {
  background: var(--color-surface-muted);
  color: var(--color-text-muted);
  font-weight: 600;
}

.key-table tr:last-child td {
  border-bottom: none;
}

.name-cell {
  max-width: 220px;
  font-weight: 600;
  overflow-wrap: anywhere;
}

.key-prefix {
  display: inline-flex;
  align-items: center;
  min-height: 24px;
  padding: 2px 8px;
  border-radius: 999px;
  background: var(--color-surface-muted);
  color: var(--color-text-muted);
  font-family: monospace;
  font-size: 12px;
  overflow-wrap: anywhere;
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
  .key-list {
    display: flex;
    flex-direction: column;
    gap: 12px;
  }

  .key-table-wrap {
    display: none;
  }

  .key-card {
    gap: 14px;
    padding: 16px;
    border-radius: 12px;
  }

  .key-meta-list {
    grid-template-columns: repeat(2, minmax(0, 1fr));
    gap: 10px;
  }

  .key-meta-row {
    display: flex;
    flex-direction: column;
    gap: 4px;
    padding: 10px 12px;
    background: var(--color-surface-muted);
    border-radius: 10px;
  }

  .key-meta-row dt,
  .key-meta-row dd {
    font-size: 12px;
    line-height: 1.5;
  }

  .key-meta-row dd {
    font-size: 13px;
    font-weight: 500;
  }

  .key-prefix {
    justify-self: flex-start;
    width: fit-content;
  }

  .action-btn {
    min-height: 40px;
    border-radius: 10px;
    font-weight: 500;
  }
}

@media (max-width: 479px) {
  .key-meta-list {
    grid-template-columns: minmax(0, 1fr);
  }
}

@media (max-width: 359px) {
  .key-prefix {
    max-width: 100%;
  }
}
</style>
