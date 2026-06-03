<template>
  <form class="task-form" @submit.prevent="handleSubmit">
    <div class="form-group">
      <label for="task-title">标题 *</label>
      <input
        id="task-title"
        v-model="form.title"
        name="task_title"
        type="text"
        placeholder="任务标题"
        :class="{ error: errors.title }"
      />
      <span v-if="errors.title" class="error-text">{{ errors.title }}</span>
    </div>

    <div class="form-group">
      <label for="task-description">描述</label>
      <textarea
        id="task-description"
        v-model="form.description"
        name="task_description"
        placeholder="任务描述"
        rows="3"
      ></textarea>
    </div>

    <div class="form-row">
      <div class="form-group">
        <label>优先级</label>
        <BaseSelect
          v-model="form.priority"
          :options="priorityOptions"
          placeholder="未设置"
          block
          aria-label="优先级"
        />
      </div>

      <div class="form-group">
        <label>截止时间</label>
        <DateTimePicker
          v-model="form.due_at"
          placeholder="选择截止日期"
        />
      </div>
    </div>

    <div class="form-row">
      <div class="form-group">
        <label>提醒时间</label>
        <DateTimePicker
          v-model="form.remind_at"
          placeholder="选择提醒日期"
        />
      </div>

      <div class="form-group">
        <label>重复类型</label>
        <BaseSelect
          v-model="form.repeat_type"
          :options="repeatTypeOptions"
          block
          aria-label="重复类型"
        />
      </div>
    </div>

    <div v-if="form.repeat_type !== 'none'" class="form-row">
      <div class="form-group">
        <label for="task-repeat-interval">重复间隔</label>
        <input
          id="task-repeat-interval"
          v-model.number="form.repeat_interval"
          name="task_repeat_interval"
          type="number"
          min="1"
          max="365"
        />
      </div>

      <div class="form-group">
        <label>重复结束日期</label>
        <VueDatePicker
          v-bind="datePickerProps"
          v-model="form.repeat_end_date"
          placeholder="选择结束日期"
        />
      </div>
    </div>

    <div class="form-group">
      <label>标签</label>
      <TagPicker v-model="form.tags" placeholder="选择或新建标签..." />
    </div>

    <div class="form-actions">
      <slot name="actions">
        <button type="button" class="btn-secondary" @click="$emit('cancel')">取消</button>
        <button type="submit" class="btn-primary" :disabled="submitting">
          {{ submitting ? '保存中...' : '保存' }}
        </button>
      </slot>
    </div>
  </form>
</template>

<script setup lang="ts">
import { computed, reactive, ref, watch } from 'vue'
import { VueDatePicker } from '@vuepic/vue-datepicker'
import type { CreateTaskPayload, Task, UpdateTaskPayload } from '@/entities/task/model'
import { useThemeStore } from '@/app/stores/theme.store'
import { isoToDateTimeLocal, isoToDateLocal, dateTimeLocalToISOString, dateToEndOfDayISOString } from '@/shared/utils/date'
import { zhCN } from '@/shared/utils/date-locale'
import TagPicker from '@/features/tags/TagPicker.vue'
import DateTimePicker from '@/shared/ui/DateTimePicker.vue'
import BaseSelect, { type SelectOption } from '@/shared/ui/BaseSelect.vue'

const themeStore = useThemeStore()

const priorityOptions: SelectOption<CreateTaskPayload['priority']>[] = [
  { label: '未设置', value: undefined },
  { label: '高', value: 1 },
  { label: '中', value: 2 },
  { label: '低', value: 3 },
]

const repeatTypeOptions: SelectOption<CreateTaskPayload['repeat_type']>[] = [
  { label: '无', value: 'none' },
  { label: '每天', value: 'daily' },
  { label: '每周', value: 'weekly' },
  { label: '每月', value: 'monthly' },
  { label: '每年', value: 'yearly' },
]

// 仅日期模式的 VueDatePicker props
const datePickerProps = computed(() => ({
  dark: themeStore.isDark,
  'model-type': 'format' as const,
  format: 'yyyy-MM-dd',
  locale: zhCN,
  'auto-apply': true,
  clearable: true,
  'enable-time-picker': false,
  teleport: true,
  config: { allowPreventDefault: true },
}))

type TaskFormInitialData = Partial<CreateTaskPayload> &
  Partial<Pick<Task, 'dueAt' | 'remindAt' | 'repeatEndDate' | 'repeatInterval' | 'repeatType' | 'tags'>> & {
    title?: string
    description?: string
    priority?: 1 | 2 | 3
  }

const props = defineProps<{
  initialData?: TaskFormInitialData
  mode?: 'create' | 'edit'
}>()

const emit = defineEmits<{
  submit: [payload: CreateTaskPayload | UpdateTaskPayload]
  cancel: []
}>()

// 初始化表单：将后端 UTC RFC3339 转为本地输入格式
// initialData 可能是 snake_case（CreateTaskPayload）或 camelCase（Task model）
const d = props.initialData
const form = reactive<CreateTaskPayload>({
  title: d?.title || '',
  description: d?.description || undefined,
  priority: d?.priority || undefined,
  due_at: (d?.due_at || d?.dueAt) ? isoToDateTimeLocal(d.due_at || d.dueAt) : undefined,
  remind_at: (d?.remind_at || d?.remindAt) ? isoToDateTimeLocal(d.remind_at || d.remindAt) : undefined,
  repeat_type: d?.repeat_type || d?.repeatType || 'none',
  repeat_interval: d?.repeat_interval || d?.repeatInterval || 1,
  repeat_end_date: (d?.repeat_end_date || d?.repeatEndDate) ? isoToDateLocal(d.repeat_end_date || d.repeatEndDate) : undefined,
  tags: Array.isArray(d?.tags) ? [...d.tags] : [],
})

const errors = reactive({
  title: '',
})

const submitting = ref(false)

watch(
  () => form.title,
  () => {
    errors.title = ''
  },
)

function validate(): boolean {
  let valid = true

  if (!form.title.trim()) {
    errors.title = '标题不能为空'
    valid = false
  }

  return valid
}

async function handleSubmit() {
  if (!validate()) return

  submitting.value = true
  try {
    const payload = { ...form }
    // 将 datetime-local 本地时间转为 RFC3339 UTC
    if (payload.due_at) {
      payload.due_at = dateTimeLocalToISOString(payload.due_at)
    }
    if (payload.remind_at) {
      payload.remind_at = dateTimeLocalToISOString(payload.remind_at)
    }
    if (payload.repeat_end_date) {
      payload.repeat_end_date = dateToEndOfDayISOString(payload.repeat_end_date)
    }
    emit('submit', payload)
  } finally {
    submitting.value = false
  }
}
</script>

<style scoped>
.task-form {
  display: flex;
  flex-direction: column;
  gap: 16px;
}
</style>
