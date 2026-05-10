<template>
  <div class="keys-page">
    <div class="page-header">
      <h2>API Key</h2>
      <button class="btn-primary" type="button" @click="showCreate = true">生成 API Key</button>
    </div>

    <div class="info-banner">
      API Key 用于第三方应用访问你的任务数据。请妥善保管，创建后仅显示一次。
      登录时自动生成的 "login" 类型 Key 在最后使用超过 24 小时后会自动清理。
    </div>

    <div v-if="loading" class="page-loading">加载中...</div>

    <div v-else-if="error" class="page-error">
      <p>{{ error }}</p>
      <button type="button" @click="fetchKeys">重试</button>
    </div>

    <div v-else-if="keys.length === 0" class="page-empty">
      <p>暂无 API Key</p>
    </div>

    <ApiKeyTable v-else :keys="keys" @revoke="handleRevoke" />

    <ApiKeyCreateDialog v-model:visible="showCreate" @created="fetchKeys" />
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { getApiKeys, deleteApiKey } from '@/entities/api-key/api'
import { toApiKeyInfo } from '@/entities/api-key/mapper'
import type { ApiKeyInfo } from '@/entities/api-key/model'
import ApiKeyTable from '@/features/api-keys/ApiKeyTable.vue'
import ApiKeyCreateDialog from '@/features/api-keys/ApiKeyCreateDialog.vue'

const keys = ref<ApiKeyInfo[]>([])
const loading = ref(false)
const error = ref<string | null>(null)
const showCreate = ref(false)

onMounted(() => {
  fetchKeys()
})

async function fetchKeys() {
  loading.value = true
  error.value = null
  try {
    const response = await getApiKeys()
    const data = Array.isArray(response.data) ? response.data : []
    keys.value = data.map(toApiKeyInfo)
  } catch (e) {
    error.value = e instanceof Error ? e.message : '加载失败'
  } finally {
    loading.value = false
  }
}

async function handleRevoke(id: number) {
  if (!confirm('确定要撤销这个 API Key 吗？撤销后将无法恢复。')) return
  await deleteApiKey(id)
  await fetchKeys()
}
</script>

<style scoped>
.keys-page {
  display: flex;
  flex-direction: column;
  gap: 16px;
}

.page-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.page-header h2 {
  margin: 0;
  font-size: 20px;
}

.info-banner {
  background: #f0f9ff;
  border: 1px solid #bae6fd;
  border-radius: 6px;
  padding: 12px;
  color: #0369a1;
  font-size: 14px;
}

.btn-primary {
  padding: 8px 16px;
  background: var(--color-primary);
  color: white;
  border: none;
  border-radius: 6px;
  cursor: pointer;
}

.page-loading,
.page-error,
.page-empty {
  text-align: center;
  padding: 48px 24px;
  color: var(--color-text-muted);
}

.page-error button {
  margin-top: 12px;
  padding: 8px 16px;
  background: var(--color-surface);
  border: 1px solid var(--color-border);
  border-radius: 6px;
  cursor: pointer;
}
</style>
