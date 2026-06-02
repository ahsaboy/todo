import { ref, onMounted, type Ref } from 'vue'

export interface UseFetchConfig<T> {
  fetcher: () => Promise<T>
  autoLoad?: boolean
  errorPrefix?: string
}

export interface UseFetchReturn<T> {
  data: Ref<T | null>
  isLoading: Ref<boolean>
  error: Ref<string>
  load: () => Promise<void>
}

export function useFetch<T>(config: UseFetchConfig<T>): UseFetchReturn<T> {
  const data = ref<T | null>(null) as Ref<T | null>
  const isLoading = ref(false)
  const error = ref('')

  async function load() {
    isLoading.value = true
    error.value = ''
    try {
      data.value = await config.fetcher()
    } catch {
      error.value = `${config.errorPrefix ?? '加载数据'}失败`
    } finally {
      isLoading.value = false
    }
  }

  if (config.autoLoad !== false) {
    onMounted(load)
  }

  return { data, isLoading, error, load }
}
