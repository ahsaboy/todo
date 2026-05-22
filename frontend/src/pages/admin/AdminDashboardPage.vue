<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { adminApi } from '@/shared/api/admin-client'
import type { ApiResponse } from '@/shared/api/types'

interface Stats {
  total_users: number
  total_tasks: number
  completed_tasks: number
  total_reminder_configs: number
  total_reminder_logs: number
}

const stats = ref<Stats | null>(null)
const error = ref('')

onMounted(async () => {
  try {
    const res = await adminApi.get<ApiResponse<Stats>>('/stats')
    stats.value = res.data
  } catch {
    error.value = '加载统计数据失败'
  }
})

function completionRate(s: Stats): string {
  if (!s.total_tasks) return '0%'
  return Math.round((s.completed_tasks / s.total_tasks) * 100) + '%'
}
</script>

<template>
  <div class="page-container">
    <h1 class="page-title">仪表盘</h1>
    <div v-if="error" class="error-message" style="margin-bottom: 1rem;">{{ error }}</div>
    <div v-if="stats" class="admin-stats-grid">
      <div class="admin-stat-card">
        <div class="stat-label">注册用户</div>
        <div class="stat-value">{{ stats.total_users }}</div>
      </div>
      <div class="admin-stat-card">
        <div class="stat-label">任务总数</div>
        <div class="stat-value">{{ stats.total_tasks }}</div>
      </div>
      <div class="admin-stat-card">
        <div class="stat-label">已完成任务</div>
        <div class="stat-value">{{ stats.completed_tasks }}<span class="stat-sub">（{{ completionRate(stats) }}）</span></div>
      </div>
      <div class="admin-stat-card">
        <div class="stat-label">提醒配置数</div>
        <div class="stat-value">{{ stats.total_reminder_configs }}</div>
      </div>
      <div class="admin-stat-card">
        <div class="stat-label">提醒日志数</div>
        <div class="stat-value">{{ stats.total_reminder_logs }}</div>
      </div>
    </div>
    <div v-else-if="!error" class="loading-hint">加载中...</div>
  </div>
</template>

<style scoped>
@import '@/widgets/admin-common.css';
</style>
