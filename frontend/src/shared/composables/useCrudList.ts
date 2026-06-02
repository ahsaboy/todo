import { ref, computed, onMounted, watch, type Ref } from 'vue'
import type { PaginatedResponse } from '@/shared/api/types'
import type { ApiClient } from '@/shared/api/create-client'
import { createListMutations } from './useListMutations'

export interface CrudListConfig<T> {
  client: ApiClient
  buildEndpoint: (params: {
    page: number
    limit: number
    filters: Record<string, string>
  }) => string
  limit?: number
  autoLoad?: boolean
  mapItem?: (raw: unknown) => T
  errorPrefix?: string
}

export interface CrudListReturn<T> {
  items: Ref<T[]>
  total: Ref<number>
  page: Ref<number>
  limit: number
  totalPages: Ref<number>
  isLoading: Ref<boolean>
  error: Ref<string>
  filters: Ref<Record<string, string>>
  load: () => Promise<void>
  resetFilters: () => void
  applyFilters: (newFilters?: Record<string, string>) => void
  handleFilterChange: (id: string, value: string) => void
  setPage: (p: number) => void
  deleteItem: (endpoint: string, confirmMessage?: string) => Promise<void>
  mutate: (endpoint: string, method: 'PATCH' | 'POST' | 'PUT', body?: unknown) => Promise<void>
}

export function useCrudList<T>(config: CrudListConfig<T>): CrudListReturn<T> {
  const items = ref<T[]>([]) as Ref<T[]>
  const total = ref(0)
  const page = ref(1)
  const limit = config.limit ?? 20
  const isLoading = ref(false)
  const error = ref('')
  const filters = ref<Record<string, string>>({})

  const totalPages = computed(() => Math.ceil(total.value / limit))

  async function load() {
    isLoading.value = true
    error.value = ''
    try {
      const endpoint = config.buildEndpoint({
        page: page.value,
        limit,
        filters: filters.value,
      })
      const res = await config.client.get<PaginatedResponse<T>>(endpoint)
      items.value = config.mapItem
        ? (res.data as unknown[]).map(config.mapItem)
        : res.data
      total.value = res.meta.total_items
    } catch {
      error.value = `${config.errorPrefix ?? '加载数据'}失败`
    } finally {
      isLoading.value = false
    }
  }

  function setPage(p: number) {
    page.value = p
  }

  function applyFilters(newFilters?: Record<string, string>) {
    if (newFilters !== undefined) {
      filters.value = newFilters
    }
    const wasAlreadyPage1 = page.value === 1
    page.value = 1
    if (wasAlreadyPage1) load()
  }

  function resetFilters() {
    filters.value = {}
    const wasAlreadyPage1 = page.value === 1
    page.value = 1
    if (wasAlreadyPage1) load()
  }

  function handleFilterChange(id: string, value: string) {
    filters.value[id] = value
  }

  const { deleteItem, mutate } = createListMutations(config.client, load, error)

  if (config.autoLoad !== false) {
    onMounted(load)
  }

  watch(page, load)

  return {
    items,
    total,
    page,
    limit,
    totalPages,
    isLoading,
    error,
    filters,
    load,
    resetFilters,
    applyFilters,
    handleFilterChange,
    setPage,
    deleteItem,
    mutate,
  }
}
