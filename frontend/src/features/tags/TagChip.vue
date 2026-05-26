<template>
  <span class="tag-chip" :style="chipStyle" :title="tag?.name ?? name">
    <component v-if="iconComponent" :is="iconComponent" :size="iconSize" />
    <span class="chip-label">{{ tag?.name ?? name }}</span>
    <button
      v-if="removable"
      type="button"
      class="chip-remove"
      :aria-label="`移除标签 ${tag?.name ?? name}`"
      @click.stop="$emit('remove')"
    >
      ×
    </button>
  </span>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { storeToRefs } from 'pinia'
import { useTagStore } from '@/entities/tag/store'
import { CURATED_ICONS, DEFAULT_TAG_COLOR } from '@/shared/icons/curated'

const props = withDefaults(
  defineProps<{
    name: string
    size?: 'sm' | 'md'
    removable?: boolean
  }>(),
  { size: 'sm', removable: false },
)

defineEmits<{
  remove: []
}>()

const store = useTagStore()
const { byName } = storeToRefs(store)

const tag = computed(() => byName.value.get(props.name))

const color = computed(() => tag.value?.color ?? DEFAULT_TAG_COLOR)

const iconComponent = computed(() => {
  const key = tag.value?.icon ?? ''
  if (!key) return null
  return CURATED_ICONS[key] ?? null
})

const iconSize = computed(() => (props.size === 'md' ? 14 : 12))

const chipStyle = computed(() => ({
  color: color.value,
  background: `color-mix(in srgb, ${color.value} 14%, transparent)`,
  borderColor: `color-mix(in srgb, ${color.value} 30%, transparent)`,
}))
</script>

<style scoped>
.tag-chip {
  display: inline-flex;
  align-items: center;
  gap: 4px;
  padding: 1px 8px;
  border: 1px solid;
  border-radius: 999px;
  font-size: 11px;
  font-weight: 500;
  line-height: 1.5;
  flex-shrink: 0;
  white-space: nowrap;
  max-width: 160px;
}

.chip-label {
  overflow: hidden;
  text-overflow: ellipsis;
}

.chip-remove {
  background: transparent;
  border: 0;
  color: inherit;
  cursor: pointer;
  font-size: 14px;
  line-height: 1;
  padding: 0 2px;
  opacity: 0.7;
}
.chip-remove:hover {
  opacity: 1;
}
</style>
