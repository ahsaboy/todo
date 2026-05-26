<template>
  <div ref="rootEl" class="tag-picker" :class="{ open }">
    <div class="picker-trigger" @click="toggleOpen" role="button" tabindex="0" @keydown.enter.prevent="toggleOpen">
      <template v-if="modelValue.length === 0">
        <span class="placeholder">{{ placeholder }}</span>
      </template>
      <template v-else>
        <TagChip
          v-for="name in modelValue"
          :key="name"
          :name="name"
          removable
          @remove.stop="removeTag(name)"
        />
      </template>
      <span class="chevron">▾</span>
    </div>

    <div v-if="open" class="picker-panel" @click.stop>
      <div class="panel-search">
        <input
          ref="searchInput"
          v-model="query"
          type="text"
          placeholder="搜索或新建标签..."
          @keydown.enter.prevent="onEnter"
          @keydown.escape.prevent="open = false"
        />
      </div>
      <ul class="panel-list">
        <li
          v-for="tag in filtered"
          :key="tag.id"
          class="panel-item"
          :class="{ selected: modelValue.includes(tag.name) }"
          @click="toggleTag(tag.name)"
        >
          <TagChip :name="tag.name" size="md" />
          <span v-if="modelValue.includes(tag.name)" class="check">✓</span>
        </li>
        <li
          v-if="canCreate"
          class="panel-item create"
          @click="onCreateNew"
        >
          <span>+ 新建标签 "{{ query.trim() }}"</span>
        </li>
        <li
          v-if="filtered.length === 0 && !canCreate"
          class="panel-empty"
        >
          {{ store.tags.length === 0 ? '还没有标签,先去 /tags 创建' : '没有匹配的标签' }}
        </li>
      </ul>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, nextTick, onBeforeUnmount, onMounted, ref, watch } from 'vue'
import { storeToRefs } from 'pinia'
import { useTagStore } from '@/entities/tag/store'
import TagChip from './TagChip.vue'

const props = withDefaults(
  defineProps<{
    modelValue: string[]
    placeholder?: string
  }>(),
  { placeholder: '选择标签...' },
)

const emit = defineEmits<{
  'update:modelValue': [value: string[]]
}>()

const store = useTagStore()
const { tags } = storeToRefs(store)

const open = ref(false)
const query = ref('')
const searchInput = ref<HTMLInputElement | null>(null)
const rootEl = ref<HTMLElement | null>(null)

onMounted(() => {
  store.fetchTags()
  document.addEventListener('click', onDocumentClick)
  window.addEventListener('tag-renamed', onTagRenamed)
  window.addEventListener('tag-deleted', onTagDeleted)
})

onBeforeUnmount(() => {
  document.removeEventListener('click', onDocumentClick)
  window.removeEventListener('tag-renamed', onTagRenamed)
  window.removeEventListener('tag-deleted', onTagDeleted)
})

function onDocumentClick(e: MouseEvent) {
  const target = e.target as Node
  if (rootEl.value && !rootEl.value.contains(target)) {
    open.value = false
  }
}

function toggleOpen() {
  open.value = !open.value
  if (open.value) {
    nextTick(() => searchInput.value?.focus())
  }
}

const filtered = computed(() => {
  const q = query.value.trim().toLowerCase()
  if (!q) return tags.value
  return tags.value.filter((t) => t.name.toLowerCase().includes(q))
})

const canCreate = computed(() => {
  const q = query.value.trim()
  if (!q) return false
  return !tags.value.some((t) => t.name === q)
})

function toggleTag(name: string) {
  const next = [...props.modelValue]
  const idx = next.indexOf(name)
  if (idx >= 0) next.splice(idx, 1)
  else next.push(name)
  emit('update:modelValue', next)
}

function removeTag(name: string) {
  emit(
    'update:modelValue',
    props.modelValue.filter((n) => n !== name),
  )
}

async function onCreateNew() {
  const name = query.value.trim()
  if (!name) return
  try {
    const created = await store.createTag({ name })
    emit('update:modelValue', [...props.modelValue, created.name])
    query.value = ''
  } catch (e) {
    alert(e instanceof Error ? e.message : '新建标签失败')
  }
}

function onEnter() {
  if (canCreate.value) {
    onCreateNew()
    return
  }
  // 若仅一个匹配,直接选中/取消
  if (filtered.value.length === 1) {
    toggleTag(filtered.value[0].name)
    query.value = ''
  }
}

// 监听 store 的改名/删除事件,同步本地 modelValue
function onTagRenamed(e: Event) {
  const detail = (e as CustomEvent<{ oldName: string; newName: string }>).detail
  if (props.modelValue.includes(detail.oldName)) {
    emit(
      'update:modelValue',
      props.modelValue.map((n) => (n === detail.oldName ? detail.newName : n)),
    )
  }
}
function onTagDeleted(e: Event) {
  const detail = (e as CustomEvent<{ name: string }>).detail
  if (props.modelValue.includes(detail.name)) {
    emit(
      'update:modelValue',
      props.modelValue.filter((n) => n !== detail.name),
    )
  }
}

watch(open, (v) => {
  if (!v) query.value = ''
})
</script>

<style scoped>
.tag-picker {
  position: relative;
  width: 100%;
}

.picker-trigger {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  gap: 6px;
  min-height: 36px;
  padding: 6px 28px 6px 10px;
  border: 1px solid var(--color-border);
  border-radius: 6px;
  background: var(--color-surface);
  cursor: pointer;
  position: relative;
}

.picker-trigger:hover {
  border-color: color-mix(in srgb, var(--color-primary) 50%, var(--color-border));
}

.placeholder {
  color: var(--color-text-muted);
  font-size: 13px;
}

.chevron {
  position: absolute;
  right: 10px;
  top: 50%;
  transform: translateY(-50%);
  color: var(--color-text-muted);
  font-size: 12px;
  pointer-events: none;
}

.picker-panel {
  position: absolute;
  top: calc(100% + 4px);
  left: 0;
  right: 0;
  z-index: 50;
  background: var(--color-bg);
  border: 1px solid var(--color-border);
  border-radius: 6px;
  box-shadow: 0 8px 20px rgba(0, 0, 0, 0.12);
  display: flex;
  flex-direction: column;
  max-height: 280px;
  overflow: hidden;
  min-width: 220px;
}

@media (max-width: 600px) {
  .picker-panel {
    left: -8px;
    right: -8px;
    min-width: auto;
  }
}

.panel-search {
  padding: 8px;
  border-bottom: 1px solid var(--color-border);
}

.panel-search input {
  width: 100%;
  padding: 6px 10px;
  border: 1px solid var(--color-border);
  border-radius: 4px;
  font-size: 13px;
  background: var(--color-surface);
}

.panel-list {
  list-style: none;
  margin: 0;
  padding: 4px;
  overflow-y: auto;
}

.panel-item {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 8px;
  padding: 6px 8px;
  border-radius: 4px;
  cursor: pointer;
  font-size: 13px;
}

.panel-item:hover {
  background: color-mix(in srgb, var(--color-text) 6%, transparent);
}

.panel-item.selected {
  background: color-mix(in srgb, var(--color-primary) 8%, transparent);
}

.panel-item.create {
  color: var(--color-primary);
  font-weight: 500;
}

.check {
  color: var(--color-primary);
}

.panel-empty {
  padding: 12px 8px;
  text-align: center;
  font-size: 12px;
  color: var(--color-text-muted);
}
</style>
