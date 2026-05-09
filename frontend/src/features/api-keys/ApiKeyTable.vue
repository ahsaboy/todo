<template>
  <div class="key-table">
    <table>
      <thead>
        <tr>
          <th>名称</th>
          <th>Key 前缀</th>
          <th>创建时间</th>
          <th>最后使用</th>
          <th>操作</th>
        </tr>
      </thead>
      <tbody>
        <tr v-for="key in keys" :key="key.id">
          <td>{{ key.name || '未命名' }}</td>
          <td class="key-prefix">{{ getKeyPrefix() }}</td>
          <td>{{ formatDate(key.createdAt) }}</td>
          <td>{{ key.lastUsedAt ? formatDate(key.lastUsedAt) : '从未' }}</td>
          <td>
            <button class="btn-icon btn-danger" type="button" @click="$emit('revoke', key.id)">
              撤销
            </button>
          </td>
        </tr>
      </tbody>
    </table>
  </div>
</template>

<script setup lang="ts">
import type { ApiKeyInfo } from '@/entities/api-key/model'

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
  return new Date(dateStr).toLocaleDateString('zh-CN')
}
</script>

<style scoped>
.key-table {
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

.key-prefix {
  font-family: monospace;
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
