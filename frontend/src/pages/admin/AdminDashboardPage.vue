<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { adminApi } from '@/shared/api/admin-client'
import type { ApiResponse } from '@/shared/api/types'

interface Stats {
  total_users: number
  total_tasks: number
  completed_tasks: number
  total_reminder_configs: number
  total_reminder_logs: number
}

interface DayCount {
  date: string
  count: number
}

interface Trends {
  tasks_per_day: DayCount[]
  users_per_day: DayCount[]
  reminder_status_dist: Record<string, number>
}

const stats = ref<Stats | null>(null)
const trends = ref<Trends | null>(null)
const error = ref('')

onMounted(async () => {
  try {
    const [statsRes, trendsRes] = await Promise.all([
      adminApi.get<ApiResponse<Stats>>('/stats'),
      adminApi.get<ApiResponse<Trends>>('/stats/trends'),
    ])
    stats.value = statsRes.data
    trends.value = trendsRes.data
  } catch {
    error.value = '加载统计数据失败'
  }
})

function completionRate(s: Stats): string {
  if (!s.total_tasks) return '0%'
  return Math.round((s.completed_tasks / s.total_tasks) * 100) + '%'
}

const maxTaskCount = computed(() => {
  if (!trends.value?.tasks_per_day?.length) return 1
  return Math.max(...trends.value.tasks_per_day.map(d => d.count), 1)
})

const maxUserCount = computed(() => {
  if (!trends.value?.users_per_day?.length) return 1
  return Math.max(...trends.value.users_per_day.map(d => d.count), 1)
})

function shortDate(d: string): string {
  return d.slice(5)
}
</script>

<template>
  <div class="page-container">
    <h1 class="page-title">仪表盘</h1>
    <div v-if="error" class="error-message" style="margin-bottom: 1rem;">{{ error }}</div>

    <!-- 统计卡片 -->
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

    <!-- 趋势图表 -->
    <template v-if="trends">
      <!-- 提醒状态分布 -->
      <div class="trend-section">
        <h2 class="section-title">提醒状态分布</h2>
        <div class="status-grid">
          <div v-for="(count, status) in trends.reminder_status_dist" :key="status" class="status-card">
            <div class="status-label">{{ status }}</div>
            <div class="status-value">{{ count }}</div>
          </div>
        </div>
      </div>

      <!-- 近 30 天每日新增任务 -->
      <div class="trend-section">
        <h2 class="section-title">近 30 天每日新增任务</h2>
        <div v-if="trends.tasks_per_day.length" class="bar-chart">
          <div v-for="d in trends.tasks_per_day" :key="d.date" class="bar-item">
            <div class="bar-fill" :style="{ height: (d.count / maxTaskCount * 100) + '%' }">
              <span class="bar-tooltip">{{ d.count }}</span>
            </div>
            <div class="bar-label">{{ shortDate(d.date) }}</div>
          </div>
        </div>
        <div v-else class="empty-hint">暂无数据</div>
      </div>

      <!-- 近 30 天每日注册用户 -->
      <div class="trend-section">
        <h2 class="section-title">近 30 天每日注册用户</h2>
        <div v-if="trends.users_per_day.length" class="bar-chart">
          <div v-for="d in trends.users_per_day" :key="d.date" class="bar-item">
            <div class="bar-fill bar-fill-user" :style="{ height: (d.count / maxUserCount * 100) + '%' }">
              <span class="bar-tooltip">{{ d.count }}</span>
            </div>
            <div class="bar-label">{{ shortDate(d.date) }}</div>
          </div>
        </div>
        <div v-else class="empty-hint">暂无数据</div>
      </div>
    </template>
  </div>
</template>

<style scoped>
@import '@/widgets/admin-common.css';

.trend-section {
  margin-top: 2rem;
}

.section-title {
  font-size: 1rem;
  font-weight: 600;
  margin-bottom: 1rem;
  color: var(--color-text);
}

.status-grid {
  display: flex;
  gap: 1rem;
  flex-wrap: wrap;
}

.status-card {
  background: var(--color-surface);
  border: 1px solid var(--color-border);
  border-radius: 8px;
  padding: 1rem 1.5rem;
  min-width: 100px;
  text-align: center;
}

.status-label {
  font-size: 0.8rem;
  color: var(--color-text-muted);
  margin-bottom: 0.25rem;
  text-transform: capitalize;
}

.status-value {
  font-size: 1.4rem;
  font-weight: 700;
}

.bar-chart {
  display: flex;
  align-items: flex-end;
  gap: 2px;
  height: 160px;
  padding-bottom: 24px;
  position: relative;
  overflow-x: auto;
}

.bar-item {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: flex-end;
  flex: 1;
  min-width: 18px;
  height: 100%;
  position: relative;
}

.bar-fill {
  width: 100%;
  max-width: 28px;
  background: var(--color-primary, #4a9eff);
  border-radius: 3px 3px 0 0;
  min-height: 2px;
  position: relative;
  transition: height 0.3s ease;
}

.bar-fill-user {
  background: var(--color-success, #28a745);
}

.bar-tooltip {
  position: absolute;
  top: -20px;
  left: 50%;
  transform: translateX(-50%);
  font-size: 0.65rem;
  color: var(--color-text-muted);
  white-space: nowrap;
  opacity: 0;
  transition: opacity 0.2s;
}

.bar-item:hover .bar-tooltip {
  opacity: 1;
}

.bar-label {
  font-size: 0.6rem;
  color: var(--color-text-muted);
  position: absolute;
  bottom: -20px;
  white-space: nowrap;
  transform: rotate(-45deg);
}

.empty-hint {
  color: var(--color-text-muted);
  font-size: 0.85rem;
  padding: 1rem 0;
}

.loading-hint {
  color: var(--color-text-muted);
  text-align: center;
  padding: 2rem;
}
</style>
