<script setup lang="ts">
import { ref, computed, onMounted, onUnmounted, nextTick, watch } from 'vue'
import { adminApi } from '@/shared/api/admin-client'
import { useFetch } from '@/shared/composables/useFetch'
import type { ApiResponse } from '@/shared/api/types'
import * as echarts from 'echarts/core'
import { PieChart, LineChart, BarChart } from 'echarts/charts'
import {
  TitleComponent,
  TooltipComponent,
  LegendComponent,
  GridComponent
} from 'echarts/components'
import { CanvasRenderer } from 'echarts/renderers'
import { Users, ListTodo, CheckCircle, Bell, Zap, BarChart3 } from 'lucide-vue-next'
import { useThemeStore } from '@/app/stores/theme.store'

echarts.use([
  PieChart,
  LineChart,
  BarChart,
  TitleComponent,
  TooltipComponent,
  LegendComponent,
  GridComponent,
  CanvasRenderer
])

interface DayCount {
  date: string
  count: number
}

interface PriorityCount {
  priority: number
  count: number
}

interface TagCount {
  tag: string
  count: number
}

interface Stats {
  total_users: number
  total_tasks: number
  completed_tasks: number
  total_reminder_configs: number
  total_reminder_logs: number
  today_new_tasks: number
  today_new_users: number
  active_users_7d: number
  priority_dist: PriorityCount[]
  completion_trend: DayCount[]
  top_tags: TagCount[]
}

interface Trends {
  tasks_per_day: DayCount[]
  users_per_day: DayCount[]
  reminder_status_dist: Record<string, number>
}

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

const completionChartRef = ref<HTMLDivElement | null>(null)
const taskTrendChartRef = ref<HTMLDivElement | null>(null)
const priorityChartRef = ref<HTMLDivElement | null>(null)
const tagChartRef = ref<HTMLDivElement | null>(null)
const reminderStatusChartRef = ref<HTMLDivElement | null>(null)
const dailyTasksChartRef = ref<HTMLDivElement | null>(null)
const userTrendChartRef = ref<HTMLDivElement | null>(null)

const charts: echarts.ECharts[] = []

const priorityLabels: Record<number, string> = {
  0: '无', 1: '低', 2: '中', 3: '高', 4: '紧急'
}

const themeStore = useThemeStore()

const completionRate = computed(() => {
  if (!stats.value || stats.value.total_tasks === 0) return 0
  return Math.round(stats.value.completed_tasks / stats.value.total_tasks * 100)
})

// -- Theme colors --

function getThemeColors() {
  const s = getComputedStyle(document.documentElement)
  return {
    text: s.getPropertyValue('--color-text').trim() || '#333',
    textMuted: s.getPropertyValue('--color-text-muted').trim() || '#999',
    bg: s.getPropertyValue('--color-bg').trim() || '#fff',
    surface: s.getPropertyValue('--color-surface').trim() || '#fff',
    surfaceMuted: s.getPropertyValue('--color-surface-muted').trim() || '#f1f3f5',
    border: s.getPropertyValue('--color-border').trim() || '#dfe3e8',
    primary: s.getPropertyValue('--color-primary').trim() || '#256f6c',
    success: s.getPropertyValue('--color-success').trim() || '#1a7f37',
    danger: s.getPropertyValue('--color-danger').trim() || '#d92d20',
    warning: s.getPropertyValue('--color-warning').trim() || '#b76e00',
    info: s.getPropertyValue('--color-info').trim() || '#1769aa',
  }
}

function buildTooltipTheme(c: ReturnType<typeof getThemeColors>) {
  return {
    backgroundColor: c.surface,
    borderColor: c.border,
    borderWidth: 1,
    textStyle: { color: c.text, fontSize: 12 },
    extraCssText: `box-shadow: 0 8px 24px ${themeStore.isDark ? 'rgba(0,0,0,0.4)' : 'rgba(0,0,0,0.1)'};`,
  }
}

// -- ECharts config builders --

function buildAxisDefaults(textMuted: string) {
  return {
    grid: { top: 20, right: 20, bottom: 30, left: 50 },
    xAxis: (dates: string[]) => ({
      type: 'category' as const,
      data: dates,
      axisLabel: { color: textMuted }
    }),
    yAxis: {
      type: 'value' as const,
      minInterval: 1,
      axisLabel: { color: textMuted }
    }
  }
}

function buildLineSeries(counts: number[], color: string) {
  return {
    type: 'line' as const,
    data: counts,
    smooth: true,
    symbol: 'circle',
    symbolSize: 8,
    lineStyle: { width: 3, color },
    itemStyle: { color },
    areaStyle: {
      color: new echarts.graphic.LinearGradient(0, 0, 0, 1, [
        { offset: 0, color: color + '4d' },
        { offset: 1, color: color + '0d' }
      ])
    }
  }
}

function buildBarSeries(counts: number[], palette: string | string[]) {
  return {
    type: 'bar' as const,
    data: counts.map((v, i) => ({
      value: v,
      itemStyle: {
        color: new echarts.graphic.LinearGradient(0, 0, 0, 1, [
          { offset: 0, color: Array.isArray(palette) ? palette[i % palette.length] : palette },
          { offset: 1, color: (Array.isArray(palette) ? palette[i % palette.length] : palette) + '80' }
        ]),
        borderRadius: [4, 4, 0, 0]
      }
    })),
    barWidth: '50%'
  }
}

function safeInitChart(
  el: HTMLDivElement | null,
  opts: echarts.EChartsOption,
  dataCheck?: unknown[] | Record<string, unknown>
): echarts.ECharts | null {
  if (!el) return null
  if (dataCheck && (Array.isArray(dataCheck) ? dataCheck.length === 0 : !dataCheck)) return null
  const chart = echarts.init(el)
  chart.setOption(opts)
  return chart
}

// -- Lifecycle --

watch(() => themeStore.isDark, () => {
  if (stats.value) {
    charts.forEach(c => c.dispose())
    charts.length = 0
    nextTick(() => initCharts())
  }
})

watch(dashboardData, async (data) => {
  if (data) {
    await nextTick()
    initCharts()
  }
})

onUnmounted(() => {
  charts.forEach(chart => chart.dispose())
  window.removeEventListener('resize', handleResize)
})

function initCharts() {
  window.removeEventListener('resize', handleResize)
  const instances = [
    initCompletionChart(),
    initTaskTrendChart(),
    initPriorityChart(),
    initTagChart(),
    initReminderStatusChart(),
    initDailyTasksChart(),
    initUserTrendChart(),
  ]
  charts.push(...instances.filter(Boolean) as echarts.ECharts[])
  window.addEventListener('resize', handleResize)
}

function handleResize() {
  charts.forEach(chart => chart?.resize())
}

// -- Chart inits --

function initCompletionChart(): echarts.ECharts | null {
  if (!stats.value) return null
  const tc = getThemeColors()

  return safeInitChart(completionChartRef.value, {
    tooltip: { trigger: 'item', formatter: '{b}: {c} ({d}%)', ...buildTooltipTheme(tc) },
    series: [{
      type: 'pie',
      radius: ['55%', '80%'],
      avoidLabelOverlap: false,
      itemStyle: { borderRadius: 10, borderColor: tc.bg, borderWidth: 2 },
      label: {
        show: true,
        position: 'center',
        formatter: [`{a|${completionRate.value}%}`, '{b|完成率}'].join('\n'),
        rich: {
          a: { fontSize: 28, fontWeight: 'bold', color: tc.text, lineHeight: 40 },
          b: { fontSize: 12, color: tc.textMuted, lineHeight: 20 }
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
  const ax = buildAxisDefaults(tc.textMuted)

  return safeInitChart(taskTrendChartRef.value, {
    tooltip: { trigger: 'axis', ...buildTooltipTheme(tc) },
    ...ax.grid,
    xAxis: ax.xAxis(stats.value.completion_trend.map(d => d.date.slice(5))),
    yAxis: ax.yAxis,
    series: [buildLineSeries(stats.value.completion_trend.map(d => d.count), tc.success)]
  })
}

function initPriorityChart(): echarts.ECharts | null {
  if (!stats.value?.priority_dist?.length) return null
  const tc = getThemeColors()
  const colors: Record<number, string> = { 0: tc.textMuted, 1: tc.success, 2: tc.info, 3: tc.warning, 4: tc.danger }

  return safeInitChart(priorityChartRef.value, {
    tooltip: { trigger: 'item', formatter: '{b}: {c} ({d}%)', ...buildTooltipTheme(tc) },
    legend: { bottom: 0, textStyle: { color: tc.textMuted } },
    series: [{
      type: 'pie',
      radius: '60%',
      center: ['50%', '45%'],
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
  const ax = buildAxisDefaults(tc.textMuted)
  const palette = [tc.primary, tc.success, tc.info, tc.warning, tc.danger]

  return safeInitChart(tagChartRef.value, {
    tooltip: { trigger: 'axis', ...buildTooltipTheme(tc) },
    grid: { top: 10, right: 20, bottom: 60, left: 60 },
    xAxis: ax.xAxis(stats.value.top_tags.map(t => t.tag)),
    yAxis: ax.yAxis,
    series: [{
      ...buildBarSeries(stats.value.top_tags.map(t => t.count), palette),
      barWidth: '40%'
    }]
  })
}

function initReminderStatusChart(): echarts.ECharts | null {
  if (!trends.value?.reminder_status_dist) return null
  const tc = getThemeColors()

  const statusColors: Record<string, string> = {
    sent: tc.success, failed: tc.danger, pending: tc.warning, skipped: tc.textMuted
  }
  const statusLabels: Record<string, string> = {
    sent: '已发送', failed: '失败', pending: '待发送', skipped: '已跳过'
  }

  return safeInitChart(reminderStatusChartRef.value, {
    tooltip: { trigger: 'item', formatter: '{b}: {c} ({d}%)', ...buildTooltipTheme(tc) },
    legend: { bottom: 0, textStyle: { color: tc.textMuted } },
    series: [{
      type: 'pie',
      radius: ['40%', '70%'],
      center: ['50%', '45%'],
      avoidLabelOverlap: false,
      itemStyle: { borderRadius: 6, borderColor: tc.bg, borderWidth: 2 },
      label: { show: true, formatter: '{b}\n{d}%', color: tc.text },
      data: Object.entries(trends.value.reminder_status_dist).map(([key, value]) => ({
        name: statusLabels[key] || key,
        value,
        itemStyle: { color: statusColors[key] || tc.info }
      }))
    }]
  })
}

function initDailyTasksChart(): echarts.ECharts | null {
  if (!trends.value?.tasks_per_day?.length) return null
  const tc = getThemeColors()
  const ax = buildAxisDefaults(tc.textMuted)

  return safeInitChart(dailyTasksChartRef.value, {
    tooltip: { trigger: 'axis', ...buildTooltipTheme(tc) },
    ...ax.grid,
    xAxis: ax.xAxis(trends.value.tasks_per_day.map(d => d.date.slice(5))),
    yAxis: ax.yAxis,
    series: [buildBarSeries(trends.value.tasks_per_day.map(d => d.count), tc.info)]
  })
}

function initUserTrendChart(): echarts.ECharts | null {
  if (!trends.value?.users_per_day?.length) return null
  const tc = getThemeColors()
  const ax = buildAxisDefaults(tc.textMuted)

  return safeInitChart(userTrendChartRef.value, {
    tooltip: { trigger: 'axis', ...buildTooltipTheme(tc) },
    ...ax.grid,
    xAxis: ax.xAxis(trends.value.users_per_day.map(d => d.date.slice(5))),
    yAxis: ax.yAxis,
    series: [buildLineSeries(trends.value.users_per_day.map(d => d.count), tc.primary)]
  })
}
</script>

<template>
  <div class="page-container">
    <h1 class="page-title">仪表盘</h1>
    <div v-if="error" class="error-message" style="margin-bottom: 1rem;">{{ error }}</div>

    <!-- 概览卡片 -->
    <div v-if="stats" class="admin-stats-grid motion-stagger">
      <div class="admin-stat-card">
        <div class="stat-icon icon-info">
          <Users :size="24" />
        </div>
        <div class="stat-body">
          <div class="stat-label">注册用户</div>
          <div class="stat-value">{{ stats.total_users }}</div>
          <div class="stat-extra">
            <span class="badge badge-info">今日 +{{ stats.today_new_users }}</span>
          </div>
        </div>
      </div>

      <div class="admin-stat-card">
        <div class="stat-icon icon-primary">
          <ListTodo :size="24" />
        </div>
        <div class="stat-body">
          <div class="stat-label">任务总数</div>
          <div class="stat-value">{{ stats.total_tasks }}</div>
          <div class="stat-extra">
            <span class="badge badge-done">今日 +{{ stats.today_new_tasks }}</span>
          </div>
        </div>
      </div>

      <div class="admin-stat-card">
        <div class="stat-icon icon-success">
          <CheckCircle :size="24" />
        </div>
        <div class="stat-body">
          <div class="stat-label">已完成</div>
          <div class="stat-value">{{ stats.completed_tasks }}</div>
          <div class="stat-extra">
            <span class="badge badge-primary">{{ completionRate }}%</span>
          </div>
        </div>
      </div>

      <div class="admin-stat-card">
        <div class="stat-icon icon-warning">
          <Bell :size="24" />
        </div>
        <div class="stat-body">
          <div class="stat-label">提醒配置</div>
          <div class="stat-value">{{ stats.total_reminder_configs }}</div>
        </div>
      </div>

      <div class="admin-stat-card">
        <div class="stat-icon icon-info">
          <Zap :size="24" />
        </div>
        <div class="stat-body">
          <div class="stat-label">7日活跃用户</div>
          <div class="stat-value">{{ stats.active_users_7d }}</div>
        </div>
      </div>

      <div class="admin-stat-card">
        <div class="stat-icon icon-danger">
          <BarChart3 :size="24" />
        </div>
        <div class="stat-body">
          <div class="stat-label">提醒日志</div>
          <div class="stat-value">{{ stats.total_reminder_logs }}</div>
        </div>
      </div>
    </div>
    <div v-else-if="loading" class="loading-hint">加载中...</div>

    <!-- 图表区域 -->
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

