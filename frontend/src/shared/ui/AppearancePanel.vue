<script setup lang="ts">
import { computed, nextTick, onBeforeUnmount, ref, watch } from 'vue'
import { X } from 'lucide-vue-next'
import { useMediaQuery } from '@/shared/composables/useMediaQuery'
import MobileSheet from './MobileSheet.vue'
import AppearanceSettingItem from './AppearanceSettingItem.vue'
import { createAppearanceSettings } from './appearance.data'

const props = defineProps<{
  visible: boolean
  triggerRef: HTMLElement | null
}>()

const emit = defineEmits<{
  'update:visible': [value: boolean]
}>()

const isMobile = useMediaQuery('(max-width: 767px)')
const settings = createAppearanceSettings()

/* ---------- 桌面端定位 ---------- */

const panelRef = ref<HTMLDivElement>()
const panelStyle = ref<Record<string, string>>({})

function updatePosition() {
  const el = props.triggerRef
  if (!el) return
  const rect = el.getBoundingClientRect()
  const gap = 8
  const panelEstHeight = 360
  const spaceBelow = window.innerHeight - rect.bottom
  const openUp = spaceBelow < panelEstHeight && rect.top > spaceBelow

  panelStyle.value = {
    position: 'fixed',
    right: `${window.innerWidth - rect.right}px`,
    ...(openUp
      ? { bottom: `${window.innerHeight - rect.top + gap}px` }
      : { top: `${rect.bottom + gap}px` }),
  }
}

function openDesktop() {
  nextTick(() => updatePosition())
  window.addEventListener('scroll', updatePosition, true)
  window.addEventListener('resize', updatePosition)
  document.addEventListener('mousedown', onDocumentMousedown, true)
  document.addEventListener('keydown', onKeydown)
}

function closeDesktop() {
  window.removeEventListener('scroll', updatePosition, true)
  window.removeEventListener('resize', updatePosition)
  document.removeEventListener('mousedown', onDocumentMousedown, true)
  document.removeEventListener('keydown', onKeydown)
}

function onDocumentMousedown(event: MouseEvent) {
  const target = event.target as Node
  if (props.triggerRef?.contains(target) || panelRef.value?.contains(target)) return
  emit('update:visible', false)
}

function onKeydown(event: KeyboardEvent) {
  if (event.key === 'Escape') {
    event.preventDefault()
    emit('update:visible', false)
  }
}

/* ---------- 响应 visible 变化 ---------- */

watch(
  () => props.visible,
  (open) => {
    if (!isMobile.value) {
      if (open) openDesktop()
      else closeDesktop()
    }
  },
)

onBeforeUnmount(() => {
  if (!isMobile.value && props.visible) closeDesktop()
})

/* ---------- 关闭 ---------- */

function close() {
  emit('update:visible', false)
}

const panelTitle = computed(() => '外观设置')
</script>

<template>
  <!-- 桌面端 popover -->
  <Teleport v-if="!isMobile" to="body">
    <Transition name="appearance-fade">
      <div
        v-if="props.visible"
        ref="panelRef"
        class="appearance-panel"
        :style="panelStyle"
        role="dialog"
        aria-label="外观设置"
      >
        <div class="appearance-panel__header">
          <span class="appearance-panel__title">{{ panelTitle }}</span>
          <button
            type="button"
            class="appearance-panel__close"
            aria-label="关闭"
            @click="close"
          >
            <X :size="14" />
          </button>
        </div>
        <div class="appearance-panel__body">
          <AppearanceSettingItem
            v-for="s in settings"
            :key="s.key"
            :setting="s"
          />
        </div>
      </div>
    </Transition>
  </Teleport>

  <!-- 移动端底部抽屉 -->
  <MobileSheet
    v-else
    :visible="props.visible"
    @update:visible="emit('update:visible', $event)"
  >
    <div class="appearance-panel__header">
      <span class="appearance-panel__title">{{ panelTitle }}</span>
      <button
        type="button"
        class="appearance-panel__close"
        aria-label="关闭"
        @click="close"
      >
        <X :size="14" />
      </button>
    </div>
    <div class="appearance-panel__body">
      <AppearanceSettingItem
        v-for="s in settings"
        :key="s.key"
        :setting="s"
      />
    </div>
  </MobileSheet>
</template>

<style scoped>
.appearance-panel {
  z-index: 3000;
  width: 280px;
  padding: 16px;
  border: 1px solid var(--color-border);
  border-radius: 12px;
  background: var(--color-surface);
  box-shadow: var(--shadow-panel);
}

.appearance-panel__header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 16px;
}

.appearance-panel__title {
  font-size: 13px;
  font-weight: 600;
  color: var(--color-text);
}

.appearance-panel__close {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 24px;
  height: 24px;
  padding: 0;
  border: none;
  border-radius: 6px;
  background: transparent;
  color: var(--color-text-muted);
  cursor: pointer;
  transition:
    background var(--motion-duration-fast) var(--motion-ease-standard),
    color var(--motion-duration-fast) var(--motion-ease-standard);
}

.appearance-panel__close:hover {
  background: var(--color-surface-muted);
  color: var(--color-text);
}

.appearance-panel__body {
  display: flex;
  flex-direction: column;
}
</style>

<style>
/* 桌面端入场/退场动画（复用 base-select-fade 同款） */
.appearance-fade-enter-active,
.appearance-fade-leave-active {
  transition:
    opacity var(--motion-duration-fast) var(--motion-ease-standard),
    transform var(--motion-duration-fast) var(--motion-ease-standard);
}

.appearance-fade-enter-from,
.appearance-fade-leave-to {
  opacity: 0;
  transform: translateY(-4px);
}
</style>
