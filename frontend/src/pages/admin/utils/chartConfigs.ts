import * as echarts from 'echarts/core'

export interface ThemeColors {
  text: string
  textMuted: string
  bg: string
  surface: string
  surfaceMuted: string
  border: string
  primary: string
  success: string
  danger: string
  warning: string
  info: string
}

export function getThemeColors(): ThemeColors {
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

export function buildTooltipTheme(c: ThemeColors, isDark: boolean) {
  return {
    backgroundColor: c.surface,
    borderColor: c.border,
    borderWidth: 1,
    textStyle: { color: c.text, fontSize: 12 },
    extraCssText: `box-shadow: 0 8px 24px ${isDark ? 'rgba(0,0,0,0.4)' : 'rgba(0,0,0,0.1)'};`,
  }
}

export function buildAxisDefaults(textMuted: string, isMobile: boolean) {
  return {
    grid: isMobile
      ? { top: 15, right: 10, bottom: 30, left: 35 }
      : { top: 20, right: 20, bottom: 30, left: 50 },
    xAxis: (dates: string[], rotate?: number) => ({
      type: 'category' as const,
      data: dates,
      axisLabel: {
        color: textMuted,
        fontSize: isMobile ? 10 : 12,
        rotate: rotate ?? 0,
      }
    }),
    yAxis: {
      type: 'value' as const,
      minInterval: 1,
      axisLabel: { color: textMuted, fontSize: isMobile ? 10 : 12 }
    }
  }
}

export function buildLineSeries(counts: number[], color: string, isMobile: boolean) {
  return {
    type: 'line' as const,
    data: counts,
    smooth: true,
    symbol: 'circle',
    symbolSize: isMobile ? 5 : 8,
    lineStyle: { width: isMobile ? 2 : 3, color },
    itemStyle: { color },
    areaStyle: {
      color: new echarts.graphic.LinearGradient(0, 0, 0, 1, [
        { offset: 0, color: color + '4d' },
        { offset: 1, color: color + '0d' }
      ])
    }
  }
}

export function buildBarSeries(counts: number[], palette: string | string[], isMobile: boolean) {
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
    barWidth: isMobile ? '60%' : '50%'
  }
}

export function safeInitChart(
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
