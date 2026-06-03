<template>
  <div class="base-select" :class="{ 'is-block': block }">
    <button
      ref="triggerRef"
      type="button"
      class="base-select__trigger"
      :class="{ 'is-open': open, 'is-disabled': disabled, 'is-placeholder': !selectedLabel }"
      :disabled="disabled"
      role="combobox"
      aria-haspopup="listbox"
      :aria-expanded="open"
      :aria-label="ariaLabel"
      @click="onTriggerClick"
      @keydown="onTriggerKeydown"
    >
      <span class="base-select__value">{{ selectedLabel || placeholder }}</span>
      <ChevronDown class="base-select__chevron" :size="16" aria-hidden="true" />
    </button>

    <Teleport to="body">
      <transition name="base-select-fade">
        <div
          v-if="open"
          ref="panelRef"
          class="base-select__panel"
          :style="panelStyle"
          role="listbox"
        >
          <button
            v-for="(opt, i) in options"
            :key="i"
            type="button"
            role="option"
            class="base-select__option"
            :class="{
              'is-active': i === highlight,
              'is-selected': isSelected(opt),
              'is-disabled': opt.disabled,
            }"
            :aria-selected="isSelected(opt)"
            :disabled="opt.disabled"
            @mouseenter="highlight = i"
            @click="selectIndex(i)"
          >
            <span class="base-select__option-label">{{ opt.label }}</span>
            <Check v-if="isSelected(opt)" class="base-select__check" :size="15" aria-hidden="true" />
          </button>
        </div>
      </transition>
    </Teleport>

    <VanPopup
      v-model:show="mobileVisible"
      class="base-select-popup"
      position="bottom"
      round
      :lazy-render="true"
      teleport="body"
    >
      <VanPicker
        v-model="pickerValue"
        :columns="mobileColumns"
        :title="title || ariaLabel"
        confirm-button-text="确认"
        cancel-button-text="取消"
        @confirm="onMobileConfirm"
        @cancel="mobileVisible = false"
      />
    </VanPopup>
  </div>
</template>

<script lang="ts">
export interface SelectOption<T extends string | number | undefined = string | number | undefined> {
  label: string
  value: T
  disabled?: boolean
}
</script>

<script setup lang="ts" generic="T extends string | number | undefined">
import { computed, nextTick, onBeforeUnmount, ref } from 'vue'
import { Picker as VanPicker, Popup as VanPopup } from 'vant'
import type { PickerConfirmEventParams } from 'vant'
import 'vant/es/picker/style'
import 'vant/es/popup/style'
import { Check, ChevronDown } from 'lucide-vue-next'

const props = withDefaults(
  defineProps<{
    modelValue: T
    options: SelectOption<T>[]
    placeholder?: string
    disabled?: boolean
    block?: boolean
    ariaLabel?: string
    title?: string
  }>(),
  {
    placeholder: '请选择',
    disabled: false,
    block: false,
    ariaLabel: undefined,
    title: undefined,
  },
)

const emit = defineEmits<{
  'update:modelValue': [value: T]
  change: [value: T]
}>()

const triggerRef = ref<HTMLButtonElement>()
const panelRef = ref<HTMLDivElement>()
const open = ref(false)
const highlight = ref(-1)
const panelStyle = ref<Record<string, string>>({})

const selectedLabel = computed(() => {
  const opt = props.options.find((o) => o.value === props.modelValue)
  return opt ? opt.label : ''
})

function isSelected(opt: SelectOption<T>): boolean {
  return opt.value === props.modelValue
}

function commit(value: T) {
  emit('update:modelValue', value)
  emit('change', value)
}

/* ---------- 桌面端自定义下拉 ---------- */

function isMobileViewport(): boolean {
  return typeof window !== 'undefined' && window.matchMedia('(max-width: 767px)').matches
}

function onTriggerClick() {
  if (props.disabled) return
  if (isMobileViewport()) openMobile()
  else toggleDesktop()
}

function toggleDesktop() {
  if (open.value) closeDesktop()
  else openDesktop()
}

function openDesktop() {
  if (props.disabled) return
  open.value = true
  const current = props.options.findIndex((o) => o.value === props.modelValue)
  highlight.value = current >= 0 ? current : firstEnabledIndex()
  nextTick(() => {
    updatePosition()
    scrollHighlightIntoView()
  })
  window.addEventListener('scroll', updatePosition, true)
  window.addEventListener('resize', updatePosition)
  document.addEventListener('mousedown', onDocumentMousedown, true)
}

function closeDesktop() {
  open.value = false
  window.removeEventListener('scroll', updatePosition, true)
  window.removeEventListener('resize', updatePosition)
  document.removeEventListener('mousedown', onDocumentMousedown, true)
}

function onDocumentMousedown(event: MouseEvent) {
  const target = event.target as Node
  if (triggerRef.value?.contains(target) || panelRef.value?.contains(target)) return
  closeDesktop()
}

function updatePosition() {
  const el = triggerRef.value
  if (!el) return
  const rect = el.getBoundingClientRect()
  const gap = 4
  const panelMaxHeight = 280
  const spaceBelow = window.innerHeight - rect.bottom
  const openUp = spaceBelow < Math.min(panelMaxHeight, 220) && rect.top > spaceBelow
  panelStyle.value = {
    position: 'fixed',
    left: `${rect.left}px`,
    width: `${rect.width}px`,
    ...(openUp
      ? { bottom: `${window.innerHeight - rect.top + gap}px` }
      : { top: `${rect.bottom + gap}px` }),
  }
}

function selectIndex(i: number) {
  const opt = props.options[i]
  if (!opt || opt.disabled) return
  commit(opt.value)
  closeDesktop()
  triggerRef.value?.focus({ preventScroll: true })
}

function firstEnabledIndex(): number {
  return props.options.findIndex((o) => !o.disabled)
}

function moveHighlight(step: number) {
  const total = props.options.length
  if (total === 0) return
  let next = highlight.value
  for (let i = 0; i < total; i++) {
    next = (next + step + total) % total
    if (!props.options[next]?.disabled) break
  }
  highlight.value = next
  scrollHighlightIntoView()
}

function scrollHighlightIntoView() {
  const panel = panelRef.value
  if (!panel || highlight.value < 0) return
  const el = panel.children[highlight.value] as HTMLElement | undefined
  if (!el) return
  // 只滚动面板自身，绝不调用 el.scrollIntoView——面板是 teleport 到 body 的
  // fixed 元素，scrollIntoView 会连带滚动整个文档导致页面跳动。
  const panelRect = panel.getBoundingClientRect()
  const elRect = el.getBoundingClientRect()
  if (elRect.top < panelRect.top) {
    panel.scrollTop -= panelRect.top - elRect.top
  } else if (elRect.bottom > panelRect.bottom) {
    panel.scrollTop += elRect.bottom - panelRect.bottom
  }
}

function onTriggerKeydown(event: KeyboardEvent) {
  if (props.disabled) return
  if (!open.value) {
    if (['ArrowDown', 'ArrowUp', 'Enter', ' '].includes(event.key)) {
      event.preventDefault()
      if (isMobileViewport()) openMobile()
      else openDesktop()
    }
    return
  }

  switch (event.key) {
    case 'ArrowDown':
      event.preventDefault()
      moveHighlight(1)
      break
    case 'ArrowUp':
      event.preventDefault()
      moveHighlight(-1)
      break
    case 'Home':
      event.preventDefault()
      highlight.value = firstEnabledIndex()
      scrollHighlightIntoView()
      break
    case 'End':
      event.preventDefault()
      highlight.value = props.options.length - 1
      scrollHighlightIntoView()
      break
    case 'Enter':
    case ' ':
      event.preventDefault()
      if (highlight.value >= 0) selectIndex(highlight.value)
      break
    case 'Escape':
      event.preventDefault()
      closeDesktop()
      break
    case 'Tab':
      closeDesktop()
      break
  }
}

/* ---------- 移动端 Vant Picker ---------- */

const mobileVisible = ref(false)
const pickerValue = ref<number[]>([])

const mobileColumns = computed(() =>
  props.options.map((o, i) => ({ text: o.label, value: i, disabled: o.disabled })),
)

function openMobile() {
  if (props.disabled) return
  const current = props.options.findIndex((o) => o.value === props.modelValue)
  pickerValue.value = [current >= 0 ? current : firstEnabledIndex()]
  mobileVisible.value = true
}

function onMobileConfirm({ selectedOptions }: PickerConfirmEventParams) {
  const index = selectedOptions[0]?.value as number | undefined
  if (typeof index === 'number') {
    const opt = props.options[index]
    if (opt && !opt.disabled) commit(opt.value)
  }
  mobileVisible.value = false
}

onBeforeUnmount(() => {
  if (open.value) closeDesktop()
})
</script>

<style scoped>
.base-select {
  display: inline-flex;
  min-width: 120px;
  vertical-align: middle;
}

.base-select.is-block {
  display: flex;
  width: 100%;
}

.base-select__trigger {
  display: inline-flex;
  align-items: center;
  justify-content: space-between;
  gap: 8px;
  width: 100%;
  min-height: 37px;
  padding: 8px 12px;
  border: 1px solid var(--color-border);
  border-radius: 6px;
  background: var(--color-surface);
  color: var(--color-text);
  font: inherit;
  font-size: 14px;
  line-height: 20px;
  text-align: left;
  cursor: pointer;
  transition:
    border-color var(--motion-duration-fast),
    box-shadow var(--motion-duration-fast);
}

.base-select__trigger:hover:not(.is-disabled),
.base-select__trigger:focus-visible {
  border-color: var(--color-primary);
  outline: none;
}

.base-select__trigger.is-open {
  border-color: var(--color-primary);
  box-shadow: var(--focus-ring);
}

.base-select__trigger.is-disabled {
  opacity: var(--opacity-disabled);
  cursor: not-allowed;
}

.base-select__value {
  flex: 1;
  min-width: 0;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.base-select__trigger.is-placeholder .base-select__value {
  color: var(--color-text-muted);
}

.base-select__chevron {
  flex-shrink: 0;
  color: var(--color-text-muted);
  transition: transform var(--motion-duration-fast) var(--motion-ease-standard);
}

.base-select__trigger.is-open .base-select__chevron {
  transform: rotate(180deg);
}

.base-select__panel {
  z-index: 3000;
  max-height: 280px;
  overflow-y: auto;
  padding: 4px;
  border: 1px solid var(--color-border);
  border-radius: 8px;
  background: var(--color-surface);
  box-shadow: var(--shadow-panel);
}

.base-select__option {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 8px;
  width: 100%;
  padding: 8px 10px;
  border: none;
  border-radius: 6px;
  background: transparent;
  color: var(--color-text);
  font: inherit;
  font-size: 14px;
  text-align: left;
  cursor: pointer;
}

.base-select__option.is-active {
  background: var(--color-surface-muted);
}

.base-select__option.is-selected {
  color: var(--color-primary);
  font-weight: 600;
}

.base-select__option.is-disabled {
  opacity: 0.45;
  cursor: not-allowed;
}

.base-select__option-label {
  min-width: 0;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.base-select__check {
  flex-shrink: 0;
  color: var(--color-primary);
}

.base-select-fade-enter-active,
.base-select-fade-leave-active {
  transition:
    opacity var(--motion-duration-fast) var(--motion-ease-standard),
    transform var(--motion-duration-fast) var(--motion-ease-standard);
}

.base-select-fade-enter-from,
.base-select-fade-leave-to {
  opacity: 0;
  transform: translateY(-4px);
}

:global(.base-select-popup) {
  --van-primary-color: var(--color-primary);
  --van-text-color: var(--color-text);
  --van-text-color-2: var(--color-text-muted);
  --van-border-color: var(--color-border);
  --van-popup-background: var(--color-surface);
  --van-background-2: var(--color-surface);
  --van-picker-background: var(--color-surface);
  --van-picker-option-text-color: var(--color-text);
  --van-picker-mask-color: linear-gradient(180deg, var(--color-surface), transparent),
    linear-gradient(0deg, var(--color-surface), transparent);
  --van-picker-toolbar-height: 48px;
  --van-picker-confirm-action-color: var(--color-primary);
  --van-picker-cancel-action-color: var(--color-text-muted);
}

:global(.base-select-popup .van-picker__toolbar) {
  border-bottom: 1px solid var(--color-border);
}

:global(.base-select-popup .van-picker__confirm) {
  font-weight: 700;
}

:global(.base-select-popup .van-picker__frame) {
  right: 20px;
  left: 20px;
  border-top: 1px solid color-mix(in srgb, var(--color-primary) 55%, transparent);
  border-bottom: 1px solid color-mix(in srgb, var(--color-primary) 55%, transparent);
  border-radius: 8px;
  background: color-mix(in srgb, var(--color-primary) 9%, transparent);
}

:global(.base-select-popup .van-picker-column__item--selected) {
  color: var(--color-primary);
  font-weight: 700;
}
</style>
