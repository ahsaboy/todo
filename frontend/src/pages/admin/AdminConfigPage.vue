<script setup lang="ts">
import { ref, computed } from 'vue'
import { adminApi } from '@/shared/api/admin-client'
import { useFetch } from '@/shared/composables/useFetch'
import type { ApiResponse } from '@/shared/api/types'

const { data: configData, isLoading, error } = useFetch({
  fetcher: () => adminApi.get<ApiResponse<unknown>>('/config').then(r => r.data),
  errorPrefix: '加载系统配置',
})

const configJson = computed(() => configData.value ? JSON.stringify(configData.value, null, 2) : '')
const copied = ref(false)

async function copyConfig() {
  try {
    await navigator.clipboard.writeText(configJson.value)
    copied.value = true
    setTimeout(() => { copied.value = false }, 2000)
  } catch {
    // clipboard not available
  }
}
</script>

<template>
  <div class="page-container">
    <div class="config-header">
      <h1 class="admin-page-title">系统配置</h1>
      <button v-if="configJson" class="btn btn-sm" @click="copyConfig">
        {{ copied ? '已复制' : '复制' }}
      </button>
    </div>

    <p class="config-hint">只读视图。敏感字段（token_hash 等）已脱敏显示为 <code>***</code>。</p>

    <div v-if="error" class="error-message">{{ error }}</div>
    <div v-else-if="isLoading" class="loading-hint">加载中...</div>
    <pre v-else class="config-block">{{ configJson }}</pre>
  </div>
</template>

<style scoped>
.config-header {
  display: flex;
  align-items: center;
  gap: 1rem;
  margin-bottom: 0.25rem;
}

.config-header .admin-page-title {
  margin: 0;
}

.config-hint {
  font-size: 0.85rem;
  color: var(--color-text-muted, #888);
  margin-bottom: 1rem;
}

.config-hint code {
  background: var(--color-surface);
  padding: 0.1em 0.35em;
  border-radius: 3px;
  font-size: 0.8rem;
}

.config-block {
  background: var(--color-surface);
  border: 1px solid var(--color-border);
  border-radius: 8px;
  padding: 1.25rem;
  font-size: 0.85rem;
  line-height: 1.6;
  overflow: auto;
  white-space: pre-wrap;
  word-break: break-all;
  color: var(--color-text);
  max-height: 70vh;
}
</style>
