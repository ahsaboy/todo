<script setup lang="ts">
import { ref, computed, watch } from 'vue'
import { adminApi } from '@/shared/api/admin-client'
import { useFetch } from '@/shared/composables/useFetch'
import type { ApiResponse } from '@/shared/api/types'
import * as echarts from 'echarts/core'
import { PieChart, LineChart, BarChart } from 'echarts/charts'
import { TitleComponent, TooltipComponent, LegendComponent, GridComponent } from 'echarts/components'
import { CanvasRenderer } from 'echarts/renderers'
import { Users, ListTodo, CheckCircle, Bell, Zap, BarChart3 } from 'lucide-vue-next'
import { useThemeStore } from '@/app/stores/theme.store'
import { useECharts } from '@/shared/composables/useECharts'
import StatCard from '@/features/admin/StatCard.vue'
import type { Stats, Trends } from './utils/dashboardTypes'
import {
  getThemeColors, buildTooltipTheme, buildAxisDefaults,
  buildLineSeries, buildBarSeries, safeInitChart,
} from './utils/chartConfigs'

echarts.use([PieChart, LineChart, BarChart, TitleComponent, TooltipComponent, LegendComponent, GridComponent, CanvasRenderer])

const themeStore = useThemeStore()

const { data: dashboardData, error, isLoading: loading } = useFetch({
  fetcher: async () => {
    const [statsRes, trendsRes] = await Promise.all([
      adminApi.get<ApiResponse<Stats>>('/stats'),
      adminApi.get<ApiResponse<Trends>>('/stats/trends'),
    ])
    return { stats: statsRes.data, trends: trendsRes.data }
  },
  errorPrefix: '加载统计数据',
})

const stats = computed(() => dashboardData.value?.stats ?? null)
const trends = computed(() => dashboardData.value?.trends ?? null)

const completionRate = computed(() => {
  if (!stats.value || stats.value.total_tasks === 0) return 0
  return Math.round(stats.value.completed_tasks / stats.value.total_tasks * 100)
})

const priorityLabels: Record<number, string> = { 0: '无', 1: '低', 2: '中', 3: '高', 4: '紧急' }

// Chart refs
const completionChartRef = ref<HTMLDivElement | null>(null)
const taskTrendChartRef = ref<HTMLDivElement | null>(null)
const priorityChartRef = ref<HTMLDivElement | null>(null)
const tagChartRef = ref<HTMLDivElement | null>(null)
const reminderStatusChartRef = ref<HTMLDivElement | null>(null)
const dailyTasksChartRef = ref<HTMLDivElement | null>(null)
const userTrendChartRef = ref<HTMLDivElement | null>(null)

const { isMobile, reinitCharts, initCharts } = useECharts(
  () => [
    initCompletionChart(),
    initTaskTrendChart(),
    initPriorityChart(),
    initTagChart(),
    initReminderStatusChart(),
    initDailyTasksChart(),
    initUserTrendChart(),
  ],
  { isDark: computed(() => themeStore.isDark) },
)

watch(dashboardData, async (data) => {
  if (data) initCharts()
})

// -- Chart inits --

function initCompletionChart(): echarts.ECharts | null {
  if (!stats.value) return null
  const tc = getThemeColors()
  return safeInitChart(completionChartRef.value, {
    tooltip: { trigger: 'item', formatter: '{b}: {c} ({d}%)', ...buildTooltipTheme(tc, themeStore.isDark) },
    series: [{
      type: 'pie',
      radius: isMobile.value ? ['45%', '70%'] : ['55%', '80%'],
      avoidLabelOverlap: false,
      itemStyle: { borderRadius: 10, borderColor: tc.bg, borderWidth: 2 },
      label: {
        show: true, position: 'center',
        formatter: [`{a|${completionRate.value}%}`, '{b|完成率}'].join('\n'),
        rich: {
          a: { fontSize: isMobile.value ? 22 : 28, fontWeight: 'bold', color: tc.text, lineHeight: isMobile.value ? 32 : 40 },
          b: { fontSize: isMobile.value ? 11 : 12, color: tc.textMuted, lineHeight: 20 }
        }
      },
      data: [
        { value: stats.value.completed_tasks, name: '已完成', itemStyle: { color: tc.success } },
        { value: stats.value.total_tasks - stats.value.completed_tasks, name: '未完成', itemStyle: { color: tc.surfaceMuted } }
      ]
    }]
  })
}

function initTaskTrendChart(): echarts.ECharts | null {
  if (!stats.value?.completion_trend?.length) return null
  const tc = getThemeColors()
  const ax = buildAxisDefaults(tc.textMuted, isMobile.value)
  return safeInitChart(taskTrendChartRef.value, {
    tooltip: { trigger: 'axis', ...buildTooltipTheme(tc, themeStore.isDark) },
    ...ax.grid,
    xAxis: ax.xAxis(stats.value.completion_trend.map(d => d.date.slice(5)), isMobile.value ? -45 : 0),
    yAxis: ax.yAxis,
    series: [buildLineSeries(stats.value.completion_trend.map(d => d.count), tc.success, isMobile.value)]
  })
}

function initPriorityChart(): echarts.ECharts | null {
  if (!stats.value?.priority_dist?.length) return null
  const tc = getThemeColors()
  const colors: Record<number, string> = { 0: tc.textMuted, 1: tc.success, 2: tc.info, 3: tc.warning, 4: tc.danger }
  return safeInitChart(priorityChartRef.value, {
    tooltip: { trigger: 'item', formatter: '{b}: {c} ({d}%)', ...buildTooltipTheme(tc, themeStore.isDark) },
    legend: { bottom: 0, textStyle: { color: tc.textMuted, fontSize: isMobile.value ? 10 : 12 } },
    series: [{
      type: 'pie',
      radius: isMobile.value ? '48%' : '60%',
      center: ['50%', isMobile.value ? '42%' : '45%'],
      data: stats.value.priority_dist.map(p => ({
        name: priorityLabels[p.priority] || '未知',
        value: p.count,
        itemStyle: { color: colors[p.priority] || tc.textMuted }
      })),
      emphasis: { itemStyle: { shadowBlur: 10, shadowOffsetX: 0, shadowColor: 'rgba(0, 0, 0, 0.2)' } }
    }]
  })
}

function initTagChart(): echarts.ECharts | null {
  if (!stats.value?.top_tags?.length) return null
  const tc = getThemeColors()
  const ax = buildAxisDefaults(tc.textMuted, isMobile.value)
  const palette = [tc.primary, tc.success, tc.info, tc.warning, tc.danger]
  return safeInitChart(tagChartRef.value, {
    tooltip: { trigger: 'axis', ...buildTooltipTheme(tc, themeStore.isDark) },
    grid: isMobile.value
      ? { top: 10, right: 10, bottom: 50, left: 35 }
      : { top: 10, right: 20, bottom: 60, left: 60 },
    xAxis: ax.xAxis(stats.value.top_tags.map(t => t.tag), isMobile.value ? -45 : 0),
    yAxis: ax.yAxis,
    series: [{ ...buildBarSeries(stats.value.top_tags.map(t => t.count), palette, isMobile.value), barWidth: isMobile.value ? '55%' : '40%' }]
  })
}

function initReminderStatusChart(): echarts.ECharts | null {
  if (!trends.value?.reminder_status_dist) return null
  const tc = getThemeColors()
  const statusColors: Record<string, string> = { sent: tc.success, failed: tc.danger, pending: tc.warning, skipped: tc.textMuted }
  const statusLabels: Record<string, string> = { sent: '已发送', failed: '失败', pending: '待发送', skipped: '已跳过' }
  return safeInitChart(reminderStatusChartRef.value, {
    tooltip: { trigger: 'item', formatter: '{b}: {c} ({d}%)', ...buildTooltipTheme(tc, themeStore.isDark) },
    legend: { bottom: 0, textStyle: { color: tc.textMuted, fontSize: isMobile.value ? 10 : 12 } },
    series: [{
      type: 'pie',
      radius: isMobile.value ? ['35%', '60%'] : ['40%', '70%'],
      center: ['50%', isMobile.value ? '42%' : '45%'],
      avoidLabelOverlap: false,
      itemStyle: { borderRadius: 6, borderColor: tc.bg, borderWidth: 2 },
      label: { show: true, formatter: isMobile.value ? '{d}%' : '{b}\n{d}%', color: tc.text, fontSize: isMobile.value ? 10 : 12 },
      data: Object.entries(trends.value.reminder_status_dist).map(([key, value]) => ({
        name: statusLabels[key] || key, value,
        itemStyle: { color: statusColors[key] || tc.info }
      }))
    }]
  })
}

function initDailyTasksChart(): echarts.ECharts | null {
  if (!trends.value?.tasks_per_day?.length) return null
  const tc = getThemeColors()
  const ax = buildAxisDefaults(tc.textMuted, isMobile.value)
  const dates30 = trends.value.tasks_per_day.map(d => d.date.slice(5))
  return safeInitChart(dailyTasksChartRef.value, {
    tooltip: { trigger: 'axis', ...buildTooltipTheme(tc, themeStore.isDark) },
    ...ax.grid,
    xAxis: { ...ax.xAxis(dates30, isMobile.value ? -45 : 0), axisLabel: { ...ax.xAxis(dates30, isMobile.value ? -45 : 0).axisLabel, interval: isMobile.value ? 4 : 0 } },
    yAxis: ax.yAxis,
    series: [buildBarSeries(trends.value.tasks_per_day.map(d => d.count), tc.info, isMobile.value)]
  })
}

function initUserTrendChart(): echarts.ECharts | null {
  if (!trends.value?.users_per_day?.length) return null
  const tc = getThemeColors()
  const ax = buildAxisDefaults(tc.textMuted, isMobile.value)
  const userDates = trends.value.users_per_day.map(d => d.date.slice(5))
  return safeInitChart(userTrendChartRef.value, {
    tooltip: { trigger: 'axis', ...buildTooltipTheme(tc, themeStore.isDark) },
    ...ax.grid,
    xAxis: { ...ax.xAxis(userDates, isMobile.value ? -45 : 0), axisLabel: { ...ax.xAxis(userDates, isMobile.value ? -45 : 0).axisLabel, interval: isMobile.value ? 4 : 0 } },
    yAxis: ax.yAxis,
    series: [buildLineSeries(trends.value.users_per_day.map(d => d.count), tc.primary, isMobile.value)]
  })
}
</script>

<template>
  <div class="page-container">
    <h1 class="page-title">仪表盘</h1>
    <div v-if="error" class="error-message" style="margin-bottom: 1rem;">{{ error }}</div>

    <div v-if="stats" class="admin-stats-grid motion-stagger">
      <StatCard :icon="Users" icon-class="icon-info" label="注册用户" :value="stats.total_users" :badge-text="`今日 +${stats.today_new_users}`" badge-class="badge-info" />
      <StatCard :icon="ListTodo" icon-class="icon-primary" label="任务总数" :value="stats.total_tasks" :badge-text="`今日 +${stats.today_new_tasks}`" badge-class="badge-done" />
      <StatCard :icon="CheckCircle" icon-class="icon-success" label="已完成" :value="stats.completed_tasks" :badge-text="`${completionRate}%`" badge-class="badge-primary" />
      <StatCard :icon="Bell" icon-class="icon-warning" label="提醒配置" :value="stats.total_reminder_configs" />
      <StatCard :icon="Zap" icon-class="icon-info" label="7日活跃用户" :value="stats.active_users_7d" />
      <StatCard :icon="BarChart3" icon-class="icon-danger" label="提醒日志" :value="stats.total_reminder_logs" />
    </div>
    <div v-else-if="loading" class="loading-hint">加载中...</div>

    <template v-if="stats || trends">
      <div class="admin-charts-grid motion-stagger">
        <div class="admin-chart-card">
          <h3 class="admin-chart-title">任务完成率</h3>
          <div ref="completionChartRef" class="admin-chart-container"></div>
        </div>
        <div class="admin-chart-card">
          <h3 class="admin-chart-title">任务优先级分布</h3>
          <div ref="priorityChartRef" class="admin-chart-container"></div>
        </div>
        <div class="admin-chart-card">
          <h3 class="admin-chart-title">提醒状态分布</h3>
          <div ref="reminderStatusChartRef" class="admin-chart-container"></div>
        </div>
        <div class="admin-chart-card">
          <h3 class="admin-chart-title">近7日完成趋势</h3>
          <div ref="taskTrendChartRef" class="admin-chart-container"></div>
        </div>
        <div class="admin-chart-card">
          <h3 class="admin-chart-title">热门标签 Top 8</h3>
          <div ref="tagChartRef" class="admin-chart-container"></div>
        </div>
        <div class="admin-chart-card">
          <h3 class="admin-chart-title">近30日新增任务</h3>
          <div ref="dailyTasksChartRef" class="admin-chart-container"></div>
        </div>
        <div class="admin-chart-card">
          <h3 class="admin-chart-title">近30日新增用户</h3>
          <div ref="userTrendChartRef" class="admin-chart-container"></div>
        </div>
      </div>
    </template>
  </div>
</template>
