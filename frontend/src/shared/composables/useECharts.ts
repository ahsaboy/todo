import { ref, onUnmounted, nextTick, watch, type Ref } from 'vue'
import type { ECharts } from 'echarts/core'
import { useMediaQuery } from './useMediaQuery'

export function useECharts(
  initFn: () => (ECharts | null)[],
  options?: { isDark?: Ref<boolean> },
) {
  const charts: ECharts[] = []
  const isMobile = useMediaQuery('(max-width: 767px)')
  const reinitFlag = ref(0)

  function handleResize() {
    charts.forEach(chart => chart?.resize())
  }

  function reinitCharts() {
    charts.forEach(c => c.dispose())
    charts.length = 0
    nextTick(() => {
      initCharts()
      reinitFlag.value++
    })
  }

  function initCharts() {
    window.removeEventListener('resize', handleResize)
    const instances = initFn()
    charts.push(...instances.filter(Boolean) as ECharts[])
    window.addEventListener('resize', handleResize)
  }

  watch(isMobile, () => reinitCharts())

  if (options?.isDark) {
    watch(options.isDark, () => reinitCharts())
  }

  onUnmounted(() => {
    charts.forEach(chart => chart.dispose())
    window.removeEventListener('resize', handleResize)
  })

  return { charts, isMobile, reinitCharts, initCharts, reinitFlag }
}
