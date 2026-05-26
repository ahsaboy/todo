<template>
  <div class="icon-picker">
    <div class="category-tabs">
      <button
        type="button"
        class="cat-tab"
        :class="{ active: activeCategory === '' }"
        @click="activeCategory = ''"
      >
        无
      </button>
      <button
        v-for="cat in categories"
        :key="cat.key"
        type="button"
        class="cat-tab"
        :class="{ active: activeCategory === cat.key }"
        @click="activeCategory = cat.key"
      >
        {{ cat.label }}
      </button>
    </div>
    <div v-if="activeCategory === ''" class="icon-empty">
      <button
        type="button"
        class="icon-cell empty"
        :class="{ active: modelValue === '' }"
        @click="onPick('')"
      >
        <span class="dash">—</span>
        <span class="cell-label">无图标</span>
      </button>
    </div>
    <div v-else class="icon-grid">
      <button
        v-for="iconKey in iconsForCategory"
        :key="iconKey"
        type="button"
        class="icon-cell"
        :class="{ active: modelValue === iconKey }"
        :title="iconKey"
        @click="onPick(iconKey)"
      >
        <component :is="iconComponent(iconKey)" :size="20" />
      </button>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, ref } from 'vue'
import { CURATED_ICONS, ICON_CATEGORIES } from '@/shared/icons/curated'
import type { Component } from 'vue'

const props = defineProps<{
  modelValue: string
}>()

const emit = defineEmits<{
  'update:modelValue': [value: string]
}>()

const categories = ICON_CATEGORIES

// 初始化:如果当前选中图标在某个分类内,默认打开那个分类;
// 如果当前无图标(空字符串),默认打开"无" tab。
const initialCategory = (() => {
  if (!props.modelValue) return ''
  const cat = ICON_CATEGORIES.find((c) => c.icons.includes(props.modelValue))
  return cat ? cat.key : ICON_CATEGORIES[0].key
})()

const activeCategory = ref(initialCategory)

const iconsForCategory = computed(() => {
  const cat = ICON_CATEGORIES.find((c) => c.key === activeCategory.value)
  return cat?.icons ?? []
})

function iconComponent(key: string): Component | null {
  return CURATED_ICONS[key] ?? null
}

function onPick(key: string) {
  emit('update:modelValue', key)
}
</script>

<style scoped>
.icon-picker {
  display: flex;
  flex-direction: column;
  gap: 10px;
}

.category-tabs {
  display: flex;
  flex-wrap: wrap;
  gap: 4px;
  border-bottom: 1px solid var(--color-border);
  padding-bottom: 8px;
}

.cat-tab {
  padding: 4px 10px;
  border: 1px solid transparent;
  background: transparent;
  border-radius: 6px;
  font-size: 12px;
  cursor: pointer;
  color: var(--color-text-muted);
}

.cat-tab:hover {
  background: color-mix(in srgb, var(--color-text) 6%, transparent);
}

.cat-tab.active {
  background: color-mix(in srgb, var(--color-primary) 12%, transparent);
  color: var(--color-primary);
  border-color: color-mix(in srgb, var(--color-primary) 30%, transparent);
}

.icon-grid {
  display: grid;
  grid-template-columns: repeat(10, 1fr);
  gap: 4px;
  max-height: 200px;
  overflow-y: auto;
}

.icon-empty {
  display: flex;
}

.icon-cell {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  gap: 2px;
  padding: 8px;
  border: 1px solid var(--color-border);
  background: var(--color-surface);
  border-radius: 6px;
  cursor: pointer;
  color: var(--color-text);
  aspect-ratio: 1;
}

.icon-cell.empty {
  aspect-ratio: auto;
  padding: 8px 16px;
}

.icon-cell:hover {
  border-color: var(--color-primary);
}

.icon-cell.active {
  border-color: var(--color-primary);
  background: color-mix(in srgb, var(--color-primary) 12%, transparent);
  color: var(--color-primary);
}

.cell-label {
  font-size: 11px;
}

.dash {
  font-size: 16px;
  font-weight: 600;
}

@media (max-width: 600px) {
  .icon-grid {
    grid-template-columns: repeat(8, 1fr);
  }
}
</style>
