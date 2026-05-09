<template>
  <form class="config-form" @submit.prevent="handleSubmit">
    <div class="form-group">
      <label for="reminder-config-name">名称 *</label>
      <input
        id="reminder-config-name"
        v-model="form.name"
        name="reminder_config_name"
        type="text"
        placeholder="配置名称"
        required
      />
    </div>

    <div class="form-group">
      <label for="reminder-config-template">预置模板</label>
      <select
        id="reminder-config-template"
        v-model="selectedTemplate"
        name="reminder_config_template"
        :disabled="loadingTemplates"
        @change="applySelectedTemplate"
      >
        <option value="">自定义</option>
        <option v-for="template in templateOptions" :key="template.name" :value="template.name">
          {{ template.label }}
        </option>
      </select>
    </div>

    <div class="form-row">
      <div class="form-group">
        <label for="reminder-config-channel-type">渠道类型 *</label>
        <select
          id="reminder-config-channel-type"
          v-model="form.channel_type"
          name="reminder_config_channel_type"
          required
          @change="handleChannelTypeChange"
        >
          <option value="webhook">Webhook</option>
          <option value="feishu">飞书</option>
          <option value="dingtalk">钉钉</option>
          <option value="wecom">企业微信</option>
          <option value="slack">Slack</option>
        </select>
      </div>

      <div class="form-group">
        <label for="reminder-config-webhook-method">请求方法</label>
        <select
          id="reminder-config-webhook-method"
          v-model="form.webhook_method"
          name="reminder_config_webhook_method"
        >
          <option value="POST">POST</option>
          <option value="GET">GET</option>
          <option value="PUT">PUT</option>
        </select>
      </div>
    </div>

    <div class="form-group">
      <label for="reminder-config-webhook-url">Webhook URL *</label>
      <input
        id="reminder-config-webhook-url"
        v-model="form.webhook_url"
        name="reminder_config_webhook_url"
        type="url"
        placeholder="https://..."
        required
      />
    </div>

    <div class="form-group">
      <label for="reminder-config-webhook-headers">Webhook Headers (JSON)</label>
      <JsonEditor
        ref="headersEditorRef"
        id="reminder-config-webhook-headers"
        v-model="webhookHeadersText"
        placeholder='{"Authorization": "Bearer xxx"}'
        :rows="3"
        @blur="validateWebhookHeaders"
        @focus="lastFocusedTextarea = 'headers'"
      />
      <p v-if="webhookHeadersError" class="field-error">{{ webhookHeadersError }}</p>
      <p v-else class="field-hint">必须是 JSON 对象，值需为字符串。</p>
    </div>

    <div class="form-group">
      <label for="reminder-config-body-template">Body 模板</label>
      <JsonEditor
        ref="bodyEditorRef"
        id="reminder-config-body-template"
        v-model="webhookBodyTemplateText"
        placeholder='{"text":"{{.Title}}"}'
        :rows="4"
        @blur="validateWebhookBodyTemplate"
        @focus="lastFocusedTextarea = 'body'"
      />
      <p v-if="webhookBodyTemplateError" class="field-error">{{ webhookBodyTemplateError }}</p>
      <p v-else class="field-hint">JSON 对象中的字符串可以使用模板变量，点击即可插入：</p>
      <div class="template-vars">
        <button
          v-for="variable in templateVariables"
          :key="variable"
          type="button"
          class="template-var-clickable"
          @mousedown.prevent
          @click="insertVariable(variable)"
        >
          {{ variable }}
        </button>
      </div>
    </div>

    <div class="form-row">
      <div class="form-group">
        <label for="reminder-config-max-retries">最大重试次数</label>
        <input
          id="reminder-config-max-retries"
          v-model.number="form.max_retries"
          name="reminder_config_max_retries"
          type="number"
          min="0"
          max="10"
        />
      </div>

      <div class="form-group">
        <label for="reminder-config-retry-delay">重试延迟（秒）</label>
        <input
          id="reminder-config-retry-delay"
          v-model.number="form.retry_delay_seconds"
          name="reminder_config_retry_delay_seconds"
          type="number"
          min="1"
          max="300"
        />
      </div>
    </div>

    <label class="checkbox-label" for="reminder-config-enabled">
      <input
        id="reminder-config-enabled"
        v-model="form.enabled"
        name="reminder_config_enabled"
        type="checkbox"
      />
      <span>启用此通知渠道</span>
    </label>

    <div class="form-actions">
      <button type="button" class="btn-secondary" @click="$emit('cancel')">取消</button>
      <button type="submit" class="btn-primary" :disabled="submitting">
        {{ submitting ? '保存中...' : '保存' }}
      </button>
    </div>
  </form>
</template>

<script setup lang="ts">
import { reactive, ref, computed, onMounted } from 'vue'
import { getReminderTemplates } from '@/entities/reminder-config/api'
import JsonEditor from '@/components/JsonEditor.vue'
import type {
  CreateReminderConfigPayload,
  UpdateReminderConfigPayload,
  ChannelType,
  ReminderTemplateDto,
  ReminderTemplatesDto,
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
  enabled: boolean
  webhook_url: string
  webhook_method: WebhookMethod
  webhook_headers: Record<string, string>
  webhook_body_template: string
  max_retries: number
  retry_delay_seconds: number
}>({
  name: props.initialData?.name || '',
  channel_type: props.initialData?.channel_type || 'webhook',
  enabled: props.initialData?.enabled ?? true,
  webhook_url: props.initialData?.webhook_url || '',
  webhook_method: props.initialData?.webhook_method || 'POST',
  webhook_headers: props.initialData?.webhook_headers || {},
  webhook_body_template: props.initialData?.webhook_body_template || '',
  max_retries: props.initialData?.max_retries ?? 3,
  retry_delay_seconds: props.initialData?.retry_delay_seconds ?? 5,
})

const submitting = ref(false)
const loadingTemplates = ref(false)
const selectedTemplate = ref('')
const templates = ref<ReminderTemplatesDto>({})
const webhookHeadersText = ref(JSON.stringify(form.webhook_headers || {}, null, 2))
const webhookBodyTemplateText = ref(form.webhook_body_template || '')
const webhookHeadersError = ref('')
const webhookBodyTemplateError = ref('')
const lastFocusedTextarea = ref<'headers' | 'body'>('body')
const headersEditorRef = ref<InstanceType<typeof JsonEditor>>()
const bodyEditorRef = ref<InstanceType<typeof JsonEditor>>()

const templateVariables = [
  '{{.TaskID}}',
  '{{.Title}}',
  '{{.Description}}',
  '{{.Priority}}',
  '{{.PriorityText}}',
  '{{.DueAt}}',
  '{{.RemindAt}}',
  '{{.RepeatType}}',
  '{{.CreatedAt}}',
]

const templateOptions = computed(() =>
  Object.keys(templates.value)
    .sort()
    .map((name) => ({
      label: formatTemplateName(name),
      name,
    })),
)

onMounted(() => {
  loadTemplates()
})

async function loadTemplates() {
  loadingTemplates.value = true
  try {
    const response = await getReminderTemplates()
    templates.value = response.data ?? {}
  } catch {
    templates.value = {}
  } finally {
    loadingTemplates.value = false
  }
}

function applySelectedTemplate() {
  if (!selectedTemplate.value) return
  applyTemplate(selectedTemplate.value)
}

function handleChannelTypeChange() {
  if (templates.value[form.channel_type]) {
    selectedTemplate.value = form.channel_type
    applyTemplate(form.channel_type)
    return
  }

  selectedTemplate.value = ''
}

function applyTemplate(name: string) {
  const template = templates.value[name]
  if (!template) return

  form.channel_type = resolveTemplateChannelType(name, template)
  form.webhook_method = isWebhookMethod(template.webhook_method) ? template.webhook_method : 'POST'
  form.webhook_headers = template.webhook_headers ?? {}
  form.webhook_body_template = template.webhook_body_template || ''
  webhookHeadersText.value = JSON.stringify(form.webhook_headers, null, 2)
  webhookBodyTemplateText.value = form.webhook_body_template
  webhookHeadersError.value = ''
  webhookBodyTemplateError.value = ''

  if (!form.webhook_url && template.webhook_url) {
    form.webhook_url = template.webhook_url
  }
  if (!form.name) {
    form.name = formatTemplateName(name)
  }
}

function resolveTemplateChannelType(name: string, template: ReminderTemplateDto): ChannelType {
  if (isChannelType(name)) {
    return name
  }
  if (isChannelType(template.channel_type)) {
    return template.channel_type
  }
  return 'webhook'
}

function isChannelType(value: string): value is ChannelType {
  return ['webhook', 'feishu', 'dingtalk', 'wecom', 'slack'].includes(value)
}

function isWebhookMethod(value: string): value is WebhookMethod {
  return ['GET', 'POST', 'PUT'].includes(value)
}

function formatTemplateName(name: string): string {
  const labels: Record<string, string> = {
    dingtalk: '钉钉',
    feishu: '飞书',
    mcp: 'MCP',
    telegram: 'Telegram',
    wecom: '企业微信',
  }
  return labels[name] || name
}

function insertVariable(variable: string) {
  const editor = lastFocusedTextarea.value === 'headers' ? headersEditorRef.value : bodyEditorRef.value
  const textarea = editor?.textareaRef
  if (!textarea) return

  const start = textarea.selectionStart
  const end = textarea.selectionEnd
  const val = textarea.value
  const newVal = val.slice(0, start) + variable + val.slice(end)

  if (lastFocusedTextarea.value === 'headers') {
    webhookHeadersText.value = newVal
  } else {
    webhookBodyTemplateText.value = newVal
  }

  requestAnimationFrame(() => {
    textarea.focus()
    const pos = start + variable.length
    textarea.setSelectionRange(pos, pos)
  })
}

function validateWebhookHeaders(): boolean {
  const raw = webhookHeadersText.value.trim()

  if (!raw) {
    form.webhook_headers = {}
    webhookHeadersText.value = '{}'
    webhookHeadersError.value = ''
    return true
  }

  try {
    const parsed = JSON.parse(raw) as unknown
    if (!isRecord(parsed)) {
      webhookHeadersError.value = 'Headers 必须是 JSON 对象。'
      return false
    }

    const invalidValue = Object.values(parsed).some((value) => typeof value !== 'string')
    if (invalidValue) {
      webhookHeadersError.value = 'Headers 的每个值都必须是字符串。'
      return false
    }

    form.webhook_headers = parsed
    webhookHeadersText.value = JSON.stringify(parsed, null, 2)
    webhookHeadersError.value = ''
    return true
  } catch {
    webhookHeadersError.value = 'Headers 不是合法 JSON。'
    return false
  }
}

function validateWebhookBodyTemplate(): boolean {
  const raw = webhookBodyTemplateText.value.trim()

  if (!raw) {
    form.webhook_body_template = ''
    webhookBodyTemplateError.value = ''
    return true
  }

  try {
    const parsed = JSON.parse(raw) as unknown
    if (!isRecord(parsed)) {
      webhookBodyTemplateError.value = 'Body 模板必须是 JSON 对象。'
      return false
    }

    form.webhook_body_template = JSON.stringify(parsed)
    webhookBodyTemplateText.value = JSON.stringify(parsed, null, 2)
    webhookBodyTemplateError.value = ''
    return true
  } catch {
    webhookBodyTemplateError.value =
      'Body 模板不是合法 JSON。模板变量需要放在 JSON 字符串中，例如 "{{.Title}}"。'
    return false
  }
}

function isRecord(value: unknown): value is Record<string, string> {
  return typeof value === 'object' && value !== null && !Array.isArray(value)
}

async function handleSubmit() {
  if (!validateWebhookHeaders() || !validateWebhookBodyTemplate()) return

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

.field-hint,
.field-error {
  margin: 0;
  font-size: 12px;
  line-height: 18px;
}

.field-hint {
  color: var(--color-text-muted);
}

.field-error {
  color: var(--color-danger);
}

.template-vars {
  display: flex;
  flex-wrap: wrap;
  gap: 6px;
}

.template-var-clickable {
  padding: 4px 8px;
  border: 1px solid var(--color-border);
  border-radius: 999px;
  background: color-mix(in srgb, var(--color-primary) 8%, white);
  color: var(--color-primary);
  font-size: 12px;
  font-family: ui-monospace, SFMono-Regular, Menlo, Consolas, 'Liberation Mono', monospace;
  cursor: pointer;
  transition: background-color 150ms, border-color 150ms, color 150ms, transform 150ms;
}

.template-var-clickable:hover {
  background: var(--color-primary);
  color: white;
  border-color: var(--color-primary);
}

.template-var-clickable:active {
  transform: translateY(1px);
}

.checkbox-label {
  display: flex;
  align-items: center;
  gap: 8px;
  font-size: 13px;
  font-weight: 500;
  color: var(--color-text);
}

.checkbox-label input {
  width: 16px;
  height: 16px;
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
