import { ref, onMounted, type Ref } from 'vue'
import type { ApiResponse } from '@/shared/api/types'
import type { ApiClient } from '@/shared/api/create-client'
import { createListMutations } from './useListMutations'

export interface SimpleListConfig<T> {
  client: ApiClient
  endpoint: string | (() => string)
  autoLoad?: boolean
  mapItem?: (raw: unknown) => T
  errorPrefix?: string
}

export interface SimpleListReturn<T> {
  items: Ref<T[]>
  isLoading: Ref<boolean>
  error: Ref<string>
  load: () => Promise<void>
  deleteItem: (endpoint: string, confirmMessage?: string) => Promise<void>
  mutate: (endpoint: string, method: 'PATCH' | 'POST' | 'PUT', body?: unknown) => Promise<void>
}

export function useSimpleList<T>(config: SimpleListConfig<T>): SimpleListReturn<T> {
  const items = ref<T[]>([]) as Ref<T[]>
  const isLoading = ref(false)
  const error = ref('')

  function resolveEndpoint(): string {
    return typeof config.endpoint === 'function' ? config.endpoint() : config.endpoint
  }

  async function load() {
    isLoading.value = true
    error.value = ''
    try {
      const res = await config.client.get<ApiResponse<T[]>>(resolveEndpoint())
      items.value = config.mapItem
        ? (res.data as unknown[]).map(config.mapItem)
        : res.data
    } catch {
      error.value = `${config.errorPrefix ?? '加载数据'}失败`
    } finally {
      isLoading.value = false
    }
  }

  const { deleteItem, mutate } = createListMutations(config.client, load, error)

  if (config.autoLoad !== false) {
    onMounted(load)
  }

  return { items, isLoading, error, load, deleteItem, mutate }
}
