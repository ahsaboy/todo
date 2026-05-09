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
        <label for="task-priority">优先级</label>
        <select id="task-priority" v-model="form.priority" name="task_priority">
          <option :value="undefined">未设置</option>
          <option :value="1">低</option>
          <option :value="2">中</option>
          <option :value="3">高</option>
        </select>
      </div>

      <div class="form-group">
        <label for="task-due-at">截止时间</label>
        <input id="task-due-at" v-model="form.due_at" name="task_due_at" type="datetime-local" />
      </div>
    </div>

    <div class="form-row">
      <div class="form-group">
        <label for="task-remind-at">提醒时间</label>
        <input
          id="task-remind-at"
          v-model="form.remind_at"
          name="task_remind_at"
          type="datetime-local"
        />
      </div>

      <div class="form-group">
        <label for="task-repeat-type">重复类型</label>
        <select id="task-repeat-type" v-model="form.repeat_type" name="task_repeat_type">
          <option value="none">无</option>
          <option value="daily">每天</option>
          <option value="weekly">每周</option>
          <option value="monthly">每月</option>
          <option value="yearly">每年</option>
        </select>
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
        <label for="task-repeat-end-date">重复结束日期</label>
        <input
          id="task-repeat-end-date"
          v-model="form.repeat_end_date"
          name="task_repeat_end_date"
          type="date"
        />
      </div>
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
import { reactive, ref, watch } from 'vue'
import type { CreateTaskPayload, UpdateTaskPayload } from '@/entities/task/model'

const props = defineProps<{
  initialData?: Partial<CreateTaskPayload>
  mode?: 'create' | 'edit'
}>()

const emit = defineEmits<{
  submit: [payload: CreateTaskPayload | UpdateTaskPayload]
  cancel: []
}>()

const form = reactive<CreateTaskPayload>({
  title: props.initialData?.title || '',
  description: props.initialData?.description || undefined,
  priority: props.initialData?.priority || undefined,
  due_at: props.initialData?.due_at || undefined,
  remind_at: props.initialData?.remind_at || undefined,
  repeat_type: props.initialData?.repeat_type || 'none',
  repeat_interval: props.initialData?.repeat_interval || 1,
  repeat_end_date: props.initialData?.repeat_end_date || undefined,
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
    emit('submit', { ...form })
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

.form-group {
  display: flex;
  flex-direction: column;
  gap: 6px;
}

.form-group label {
  font-size: 13px;
  font-weight: 500;
  color: var(--color-text);
}

.form-group input,
.form-group select,
.form-group textarea {
  padding: 8px 12px;
  border: 1px solid var(--color-border);
  border-radius: 6px;
  font-size: 14px;
}

.form-group input:focus,
.form-group select:focus,
.form-group textarea:focus {
  outline: none;
  border-color: var(--color-primary);
}

.form-group input.error,
.form-group textarea.error {
  border-color: var(--color-danger);
}

.error-text {
  font-size: 12px;
  color: var(--color-danger);
}

.form-row {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 16px;
}

.form-actions {
  display: flex;
  justify-content: flex-end;
  gap: 8px;
  padding-top: 16px;
  border-top: 1px solid var(--color-border);
}

.btn-secondary {
  padding: 8px 16px;
  background: var(--color-surface);
  border: 1px solid var(--color-border);
  border-radius: 6px;
  cursor: pointer;
}

.btn-primary {
  padding: 8px 16px;
  background: var(--color-primary);
  color: white;
  border: none;
  border-radius: 6px;
  cursor: pointer;
}

.btn-primary:hover:not(:disabled) {
  background: var(--color-primary-hover);
}

.btn-primary:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}
</style>
