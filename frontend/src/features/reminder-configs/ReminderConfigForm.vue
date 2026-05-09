<template>
  <form class="config-form" @submit.prevent="handleSubmit">
    <div class="form-group">
      <label>名称 *</label>
      <input v-model="form.name" type="text" placeholder="配置名称" required />
    </div>

    <div class="form-row">
      <div class="form-group">
        <label>渠道类型 *</label>
        <select v-model="form.channel_type" required>
          <option value="webhook">Webhook</option>
          <option value="feishu">飞书</option>
          <option value="dingtalk">钉钉</option>
          <option value="wecom">企业微信</option>
          <option value="slack">Slack</option>
        </select>
      </div>

      <div class="form-group">
        <label>请求方法</label>
        <select v-model="form.webhook_method">
          <option value="POST">POST</option>
          <option value="GET">GET</option>
          <option value="PUT">PUT</option>
        </select>
      </div>
    </div>

    <div class="form-group">
      <label>Webhook URL *</label>
      <input v-model="form.webhook_url" type="url" placeholder="https://..." required />
    </div>

    <div class="form-group">
      <label>Webhook Headers (JSON)</label>
      <textarea
        v-model="webhookHeadersStr"
        placeholder='{"Authorization": "Bearer xxx"}'
        rows="3"
        class="code-input"
      ></textarea>
    </div>

    <div class="form-group">
      <label>Body 模板</label>
      <textarea
        v-model="form.webhook_body_template"
        placeholder='{"text": "{{task.title}}"}'
        rows="4"
        class="code-input"
      ></textarea>
    </div>

    <div class="form-row">
      <div class="form-group">
        <label>最大重试次数</label>
        <input v-model.number="form.max_retries" type="number" min="0" max="10" />
      </div>

      <div class="form-group">
        <label>重试延迟（秒）</label>
        <input v-model.number="form.retry_delay_seconds" type="number" min="1" max="300" />
      </div>
    </div>

    <div class="form-actions">
      <button type="button" class="btn-secondary" @click="$emit('cancel')">取消</button>
      <button type="submit" class="btn-primary" :disabled="submitting">
        {{ submitting ? '保存中...' : '保存' }}
      </button>
    </div>
  </form>
</template>

<script setup lang="ts">
import { reactive, ref, computed } from 'vue'
import type {
  CreateReminderConfigPayload,
  UpdateReminderConfigPayload,
  ChannelType,
  WebhookMethod,
} from '@/entities/reminder-config/model'

const props = defineProps<{
  initialData?: Partial<CreateReminderConfigPayload>
}>()

const emit = defineEmits<{
  submit: [payload: CreateReminderConfigPayload | UpdateReminderConfigPayload]
  cancel: []
}>()

const form = reactive<{
  name: string
  channel_type: ChannelType
  webhook_url: string
  webhook_method: WebhookMethod
  webhook_headers: Record<string, string>
  webhook_body_template: string
  max_retries: number
  retry_delay_seconds: number
}>({
  name: props.initialData?.name || '',
  channel_type: props.initialData?.channel_type || 'webhook',
  webhook_url: props.initialData?.webhook_url || '',
  webhook_method: props.initialData?.webhook_method || 'POST',
  webhook_headers: props.initialData?.webhook_headers || {},
  webhook_body_template: props.initialData?.webhook_body_template || '',
  max_retries: props.initialData?.max_retries ?? 3,
  retry_delay_seconds: props.initialData?.retry_delay_seconds ?? 5,
})

const webhookHeadersStr = computed({
  get: () => JSON.stringify(form.webhook_headers || {}, null, 2),
  set: (val: string) => {
    try {
      form.webhook_headers = JSON.parse(val)
    } catch {
      // 无效 JSON，保持原值
    }
  },
})

const submitting = ref(false)

async function handleSubmit() {
  submitting.value = true
  try {
    emit('submit', { ...form })
  } finally {
    submitting.value = false
  }
}
</script>

<style scoped>
.config-form {
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
}

.form-group input,
.form-group select,
.form-group textarea {
  padding: 8px 12px;
  border: 1px solid var(--color-border);
  border-radius: 6px;
  font-size: 14px;
}

.code-input {
  font-family: monospace;
  font-size: 13px;
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

.btn-primary:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}
</style>
