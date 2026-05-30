<template>
  <div class="date-time-picker">
    <div class="date-time-picker__desktop">
      <VueDatePicker
        :model-value="modelValue || ''"
        :dark="themeStore.isDark"
        :locale="zhCN"
        :placeholder="placeholder"
        :config="{ allowPreventDefault: true }"
        model-type="format"
        format="yyyy-MM-dd HH:mm"
        clearable
        teleport
        :flow="desktopFlow"
        :start-time="desktopStartTime"
        :auto-apply="false"
        :enable-time-picker="true"
        @update:model-value="handleDesktopChange"
      />
    </div>

    <div class="date-time-picker__mobile">
      <button
        type="button"
        class="date-time-picker__trigger"
        @click="openMobilePicker"
      >
        <CalendarClock
          :size="16"
          aria-hidden="true"
        />
        <span :class="{ 'date-time-picker__placeholder': !displayValue }">
          {{ displayValue || placeholder }}
        </span>
      </button>
      <button
        v-if="modelValue"
        type="button"
        class="date-time-picker__clear"
        title="清除时间"
        aria-label="清除时间"
        @click="clearValue"
      >
        <X
          :size="16"
          aria-hidden="true"
        />
      </button>
    </div>

    <VanPopup
      v-model:show="mobilePickerVisible"
      class="date-time-picker-popup"
      position="bottom"
      round
      :lazy-render="true"
      teleport="body"
    >
      <VanPickerGroup
        title="选择时间"
        :tabs="['日期', '时间']"
        next-step-text="下一步"
        confirm-button-text="确认"
        cancel-button-text="取消"
        @confirm="confirmMobilePicker"
        @cancel="mobilePickerVisible = false"
      >
        <VanDatePicker
          v-model="draftDate"
          :min-date="minDate"
          :max-date="maxDate"
          :show-toolbar="false"
          :formatter="formatDateColumn"
        />
        <VanTimePicker
          v-model="draftTime"
          :show-toolbar="false"
          :columns-type="['hour', 'minute']"
        />
      </VanPickerGroup>
    </VanPopup>
  </div>
</template>

<script setup lang="ts">
import { computed, ref } from 'vue'
import { VueDatePicker } from '@vuepic/vue-datepicker'
import type { FlowConfig, TimeModel } from '@vuepic/vue-datepicker'
import { DatePicker as VanDatePicker, PickerGroup as VanPickerGroup, Popup as VanPopup, TimePicker as VanTimePicker } from 'vant'
import type { PickerOption } from 'vant'
import 'vant/es/date-picker/style'
import 'vant/es/popup/style'
import 'vant/es/time-picker/style'
import { CalendarClock, X } from 'lucide-vue-next'
import { useThemeStore } from '@/app/stores/theme.store'
import { combineDateTimeLocal, splitDateTimeLocal } from '@/shared/utils/date'
import { zhCN } from '@/shared/utils/date-locale'

const props = withDefaults(defineProps<{
  modelValue?: string | null
  placeholder?: string
  defaultTime?: string
}>(), {
  modelValue: undefined,
  placeholder: '选择日期',
  defaultTime: '00:00',
})

const emit = defineEmits<{
  'update:modelValue': [value?: string]
}>()

const themeStore = useThemeStore()
const pad = (n: number) => String(n).padStart(2, '0')
const minDate = new Date(2000, 0, 1)
const maxDate = new Date(2100, 11, 31)
const today = new Date()
const initialDate = normalizeDateParts([
  String(today.getFullYear()),
  String(today.getMonth() + 1),
  String(today.getDate()),
])
const initialTime = normalizeTimeParts(props.defaultTime.split(':'))

const mobilePickerVisible = ref(false)
const draftDate = ref<string[]>(initialDate)
const draftTime = ref<string[]>(initialTime)

const displayValue = computed(() => props.modelValue || '')
const desktopFlow = {
  steps: ['calendar', 'time'],
  partial: false,
} satisfies Partial<FlowConfig>

const desktopStartTime = computed<TimeModel>(() => {
  const [hours, minutes] = normalizeTimeParts(props.defaultTime.split(':'))
  return {
    hours: Number(hours),
    minutes: Number(minutes),
    seconds: 0,
  }
})

function formatDateColumn(type: string, option: PickerOption): PickerOption {
  const suffixMap: Record<string, string> = {
    year: '年',
    month: '月',
    day: '日',
  }
  const text = String(option.text).replace(/[年月日]$/, '')
  return {
    ...option,
    text: `${text}${suffixMap[type] || ''}`,
  }
}

function formatDateTime(date: Date): string {
  return [
    `${date.getFullYear()}-${pad(date.getMonth() + 1)}-${pad(date.getDate())}`,
    `${pad(date.getHours())}:${pad(date.getMinutes())}`,
  ].join(' ')
}

function clampNumber(value: string | undefined, min: number, max: number, fallback: number): number {
  const parsed = Number(value)
  if (!Number.isFinite(parsed)) return fallback
  return Math.min(Math.max(parsed, min), max)
}

function normalizeDateParts(parts: string[]): string[] {
  return [
    String(clampNumber(parts[0], minDate.getFullYear(), maxDate.getFullYear(), today.getFullYear())),
    String(clampNumber(parts[1], 1, 12, today.getMonth() + 1)),
    String(clampNumber(parts[2], 1, 31, today.getDate())),
  ]
}

function normalizeTimeParts(parts: string[]): string[] {
  const [defaultHour = '0', defaultMinute = '0'] = props.defaultTime.split(':')
  return [
    String(clampNumber(parts[0], 0, 23, Number(defaultHour) || 0)),
    String(clampNumber(parts[1], 0, 59, Number(defaultMinute) || 0)),
  ]
}

function normalizeDateTimeValue(value: unknown): string {
  if (!value) return ''
  if (typeof value === 'string') {
    const { date, time } = splitDateTimeLocal(value)
    return date ? combineDateTimeLocal(date, props.modelValue ? time || props.defaultTime : props.defaultTime) : ''
  }
  if (value instanceof Date && !Number.isNaN(value.getTime())) {
    const { date, time } = splitDateTimeLocal(formatDateTime(value))
    return combineDateTimeLocal(date, props.modelValue ? time : props.defaultTime)
  }
  return ''
}

function handleDesktopChange(value: unknown) {
  const nextValue = normalizeDateTimeValue(value)
  emit('update:modelValue', nextValue || undefined)
}

function resolvePickerSeed(): { date: string; time: string } {
  const current = splitDateTimeLocal(props.modelValue)
  if (current.date) {
    return {
      date: current.date,
      time: current.time || props.defaultTime,
    }
  }

  const now = new Date()
  return {
    date: `${now.getFullYear()}-${pad(now.getMonth() + 1)}-${pad(now.getDate())}`,
    time: props.defaultTime,
  }
}

function openMobilePicker() {
  const { date, time } = resolvePickerSeed()
  const [year, month, day] = date.split('-')
  const [hour = '00', minute = '00'] = time.split(':')
  draftDate.value = normalizeDateParts([year, month, day])
  draftTime.value = normalizeTimeParts([hour, minute])
  mobilePickerVisible.value = true
}

function confirmMobilePicker() {
  const [year, month, day] = draftDate.value
  const [hour, minute] = draftTime.value

  if (!year || !month || !day || !hour || !minute) {
    emit('update:modelValue', undefined)
    mobilePickerVisible.value = false
    return
  }

  emit('update:modelValue', combineDateTimeLocal(
    `${year}-${pad(Number(month))}-${pad(Number(day))}`,
    `${pad(Number(hour))}:${pad(Number(minute))}`,
  ))
  mobilePickerVisible.value = false
}

function clearValue() {
  emit('update:modelValue', undefined)
}
</script>

<style scoped>
.date-time-picker :deep(.dp__main) {
  min-width: 0;
}

.date-time-picker__mobile {
  display: none;
}

.date-time-picker__trigger {
  display: inline-flex;
  align-items: center;
  gap: 8px;
  width: 100%;
  min-height: 37px;
  padding: 8px 12px;
  border: 1px solid var(--color-border);
  border-radius: 6px;
  background: var(--color-surface-muted);
  color: var(--color-text);
  font: inherit;
  font-size: 14px;
  line-height: 20px;
  text-align: left;
  cursor: pointer;
}

.date-time-picker__trigger span {
  min-width: 0;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.date-time-picker__placeholder {
  color: var(--color-text-muted);
}

.date-time-picker__clear {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 37px;
  min-width: 37px;
  min-height: 37px;
  border: 1px solid var(--color-border);
  border-radius: 6px;
  background: var(--color-surface-muted);
  color: var(--color-text-muted);
  cursor: pointer;
}

.date-time-picker__trigger:hover,
.date-time-picker__trigger:focus-visible,
.date-time-picker__clear:hover,
.date-time-picker__clear:focus-visible {
  border-color: var(--color-primary);
  outline: none;
}

:global(.date-time-picker-popup) {
  --van-primary-color: var(--color-primary);
  --van-text-color: var(--color-text);
  --van-text-color-2: var(--color-text-muted);
  --van-border-color: var(--color-border);
  --van-popup-background: var(--color-surface);
  --van-background-2: var(--color-surface);
  --van-picker-group-background: var(--color-surface);
  --van-picker-background: var(--color-surface);
  --van-picker-option-text-color: var(--color-text);
  --van-picker-mask-color: linear-gradient(
    180deg,
    var(--color-surface),
    transparent
  ),
    linear-gradient(
      0deg,
      var(--color-surface),
      transparent
    );
  --van-picker-toolbar-height: 48px;
  --van-picker-title-font-size: 15px;
  --van-picker-loading-icon-color: var(--color-primary);
  --van-picker-confirm-action-color: var(--color-primary);
  --van-picker-cancel-action-color: var(--color-text-muted);
  --van-tabs-default-color: var(--color-primary);
  --van-tabs-nav-background: var(--color-surface);
  --van-tabs-bottom-bar-color: var(--color-primary);
  --van-tab-text-color: var(--color-text-muted);
  --van-tab-active-text-color: var(--color-primary);
}

:global(.date-time-picker-popup .van-picker__toolbar) {
  border-bottom: 1px solid var(--color-border);
}

:global(.date-time-picker-popup .van-picker-group__tabs),
:global(.date-time-picker-popup .van-tabs__wrap),
:global(.date-time-picker-popup .van-tabs__nav) {
  background: var(--color-surface);
}

:global(.date-time-picker-popup .van-tab) {
  color: var(--color-text-muted);
}

:global(.date-time-picker-popup .van-tab--active) {
  color: var(--color-primary);
}

:global(.date-time-picker-popup .van-picker__frame) {
  right: 20px;
  left: 20px;
  border-top: 1px solid color-mix(in srgb, var(--color-primary) 55%, transparent);
  border-bottom: 1px solid color-mix(in srgb, var(--color-primary) 55%, transparent);
  border-radius: 8px;
  background: color-mix(in srgb, var(--color-primary) 9%, transparent);
}

:global(.date-time-picker-popup .van-picker-column__item--selected) {
  color: var(--color-primary);
  font-weight: 700;
}

:global(.date-time-picker-popup .van-picker__confirm) {
  font-weight: 700;
}

:global(.date-time-picker-popup .van-tabs__line) {
  box-shadow: 0 0 10px color-mix(in srgb, var(--color-primary) 38%, transparent);
}

@media (max-width: 767px) {
  .date-time-picker__desktop {
    display: none;
  }

  .date-time-picker__mobile {
    display: flex;
    gap: 8px;
  }
}
</style>
