<script setup lang="ts">
import { ref, onMounted, watch } from 'vue'
import { adminApi } from '@/shared/api/admin-client'
import type { PaginatedResponse } from '@/shared/api/types'

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

const configs = ref<ReminderConfig[]>([])
const total = ref(0)
const page = ref(1)
const limit = 20
const error = ref('')
const isLoading = ref(false)

async function loadConfigs() {
  isLoading.value = true
  error.value = ''
  try {
    const res = await adminApi.get<PaginatedResponse<ReminderConfig>>(
      `/reminder-configs?page=${page.value}&limit=${limit}`
    )
    configs.value = res.data
    total.value = res.meta.total_items
  } catch {
    error.value = '加载提醒配置失败'
  } finally {
    isLoading.value = false
  }
}

onMounted(loadConfigs)
watch(page, loadConfigs)

async function toggleEnabled(config: ReminderConfig) {
  const action = config.enabled ? '禁用' : '启用'
  if (!confirm(`确定${action}提醒配置 "${config.name}"？`)) return
  try {
    await adminApi.patch(`/reminder-configs/${config.id}/toggle`)
    await loadConfigs()
  } catch {
    error.value = '切换状态失败'
  }
}

async function deleteConfig(config: ReminderConfig) {
  if (!confirm(`确定删除提醒配置 "${config.name}"？此操作不可恢复！`)) return
  try {
    await adminApi.delete(`/reminder-configs/${config.id}`)
    await loadConfigs()
  } catch {
    error.value = '删除提醒配置失败'
  }
}

const totalPages = () => Math.ceil(total.value / limit)
</script>

<template>
  <div class="page-container">
    <h1 class="page-title">提醒配置</h1>

    <div v-if="error" class="error-message">{{ error }}</div>

    <div class="admin-table-wrap">
      <table class="admin-table">
        <thead>
          <tr>
            <th>ID</th>
            <th>用户</th>
            <th>名称</th>
            <th>渠道类型</th>
            <th>Webhook URL</th>
            <th>最大重试</th>
            <th>状态</th>
            <th>创建时间</th>
            <th>操作</th>
          </tr>
        </thead>
        <tbody>
          <tr v-if="isLoading">
            <td colspan="9" style="text-align:center; padding: 2rem;">加载中...</td>
          </tr>
          <tr v-else-if="!configs.length">
            <td colspan="9" style="text-align:center; padding: 2rem; color: var(--color-text-muted);">暂无提醒配置</td>
          </tr>
          <tr v-for="c in configs" :key="c.id">
            <td>{{ c.id }}</td>
            <td>{{ c.username || '用户#' + c.user_id }}</td>
            <td>{{ c.name }}</td>
            <td><span class="badge badge-channel">{{ c.channel_type }}</span></td>
            <td class="url-cell" :title="c.webhook_url">{{ c.webhook_url }}</td>
            <td>{{ c.max_retries }}</td>
            <td>
              <span :class="c.enabled ? 'badge badge-done' : 'badge badge-disabled'">
                {{ c.enabled ? '启用' : '禁用' }}
              </span>
            </td>
            <td>{{ c.created_at }}</td>
            <td class="action-cell">
              <button class="btn btn-sm" @click="toggleEnabled(c)">
                {{ c.enabled ? '禁用' : '启用' }}
              </button>
              <button class="btn btn-sm btn-danger" @click="deleteConfig(c)">删除</button>
            </td>
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

.badge-disabled { background: var(--color-surface); color: var(--color-text-muted, #888); border: 1px solid var(--color-border); }
.badge-channel { background: var(--color-primary-bg, #e8f0fe); color: var(--color-primary, #4a9eff); }
.url-cell { max-width: 200px; overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }
</style>
