import { defineStore } from 'pinia'
import { computed, ref } from 'vue'
import type { CreateTagPayload, Tag, UpdateTagPayload } from '@/entities/tag/model'
import { toTag } from '@/entities/tag/mapper'
import * as tagApi from '@/entities/tag/api'

export const useTagStore = defineStore('tags', () => {
  const tags = ref<Tag[]>([])
  const loaded = ref(false)
  const loading = ref(false)
  const error = ref<string | null>(null)

  const byName = computed(() => {
    const m = new Map<string, Tag>()
    for (const t of tags.value) m.set(t.name, t)
    return m
  })

  async function fetchTags(force = false) {
    if (loaded.value && !force) return
    loading.value = true
    error.value = null
    try {
      const res = await tagApi.getTags()
      tags.value = (res.data ?? []).map(toTag)
      loaded.value = true
    } catch (e) {
      error.value = e instanceof Error ? e.message : 'failed to load tags'
    } finally {
      loading.value = false
    }
  }

  async function createTag(payload: CreateTagPayload): Promise<Tag> {
    const res = await tagApi.createTag(payload)
    const t = toTag(res.data)
    tags.value.push(t)
    sortInPlace()
    return t
  }

  async function updateTag(id: number, payload: UpdateTagPayload): Promise<Tag | null> {
    const res = await tagApi.updateTag(id, payload)
    const updated = toTag(res.data)
    const prev = tags.value.find((t) => t.id === id)
    const prevName = prev?.name

    const idx = tags.value.findIndex((t) => t.id === id)
    if (idx !== -1) tags.value[idx] = updated
    sortInPlace()

    // 若改名,通知监听者按 prevName -> updated.name 同步本地任务
    if (prevName && prevName !== updated.name) {
      window.dispatchEvent(
        new CustomEvent('tag-renamed', { detail: { oldName: prevName, newName: updated.name } }),
      )
    }
    return updated
  }

  async function deleteTag(id: number): Promise<number> {
    const prev = tags.value.find((t) => t.id === id)
    const res = await tagApi.deleteTag(id)
    tags.value = tags.value.filter((t) => t.id !== id)
    if (prev) {
      window.dispatchEvent(new CustomEvent('tag-deleted', { detail: { name: prev.name } }))
    }
    return res.data?.tasks_affected ?? 0
  }

  function sortInPlace() {
    tags.value.sort((a, b) => {
      if (a.sortOrder !== b.sortOrder) return a.sortOrder - b.sortOrder
      return a.id - b.id
    })
  }

  function getByName(name: string): Tag | undefined {
    return byName.value.get(name)
  }

  function reset() {
    tags.value = []
    loaded.value = false
    error.value = null
  }

  return {
    tags,
    loaded,
    loading,
    error,
    byName,
    fetchTags,
    createTag,
    updateTag,
    deleteTag,
    getByName,
    reset,
  }
})
