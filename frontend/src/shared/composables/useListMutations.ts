import type { ApiClient } from '@/shared/api/create-client'

export function createListMutations(
  client: ApiClient,
  load: () => Promise<void>,
  error: { value: string },
) {
  async function deleteItem(endpoint: string, confirmMessage?: string) {
    if (confirmMessage && !confirm(confirmMessage)) return
    try {
      await client.delete(endpoint)
      await load()
    } catch {
      error.value = '删除失败'
    }
  }

  async function mutate(endpoint: string, method: 'PATCH' | 'POST' | 'PUT', body?: unknown) {
    try {
      await client[method.toLowerCase() as 'patch' | 'post' | 'put'](endpoint, body)
      await load()
    } catch {
      error.value = '操作失败'
    }
  }

  return { deleteItem, mutate }
}
