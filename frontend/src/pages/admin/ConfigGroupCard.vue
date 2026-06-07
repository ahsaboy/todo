<script setup lang="ts">
import { computed, ref, watch } from 'vue'
import { Lock, RotateCcw, CheckCircle2, XCircle } from 'lucide-vue-next'
import BaseSelect, { type SelectOption } from '@/shared/ui/BaseSelect.vue'
import BaseDialog from '@/shared/ui/BaseDialog.vue'
import { useFormState } from '@/shared/composables/useFormState'
import { updateConfig, resetConfig, testEmailConnection } from '@/entities/app-config/api'
import type { ConfigField, ConfigGroup, ConfigUpdate } from '@/entities/app-config/model'

const props = defineProps<{ group: ConfigGroup; title: string }>()
const emit = defineEmits<{ saved: [restartRequired: boolean] }>()

const editableFields = computed(() => props.group.fields.filter((f) => f.editable))
const readonlyFields = computed(() => props.group.fields.filter((f) => !f.editable))

function buildInitial(): Record<string, unknown> {
  const o: Record<string, unknown> = {}
  for (const f of props.group.fields) {
    if (f.editable) o[f.key] = f.value
  }
  return o
}

const fs = useFormState<Record<string, unknown>>({
  initialData: buildInitial(),
  onSubmit: async (data) => {
    const updates: ConfigUpdate[] = []
    for (const f of editableFields.value) {
      if (data[f.key] !== f.value) updates.push({ key: f.key, value: data[f.key] })
    }
    if (updates.length === 0) return
    const res = await updateConfig(updates)
    emit('saved', res.restartRequired)
  },
})

watch(
  () => props.group,
  () => fs.resetTo(buildInitial()),
)

const dirty = computed(() => editableFields.value.some((f) => fs.form[f.key] !== f.value))

function setField(key: string, v: unknown) {
  fs.form[key] = v
}

function onNumberInput(key: string, raw: string) {
  if (raw === '') return
  const n = Number(raw)
  if (!Number.isNaN(n)) fs.form[key] = n
}

function enumOptions(f: ConfigField): SelectOption<string>[] {
  return f.enum.map((v) => ({ label: v, value: v }))
}

function formatValue(v: unknown): string {
  if (v === null || v === undefined || v === '') return '—'
  if (typeof v === 'boolean') return v ? '是' : '否'
  return String(v)
}

const resetTarget = ref<ConfigField | null>(null)
const resetting = ref(false)
const resetVisible = computed({
  get: () => resetTarget.value !== null,
  set: (v) => {
    if (!v) resetTarget.value = null
  },
})

function askReset(f: ConfigField) {
  resetTarget.value = f
}

async function confirmReset() {
  if (!resetTarget.value) return
  resetting.value = true
  try {
    await resetConfig(resetTarget.value.key)
    resetTarget.value = null
    emit('saved', true)
  } finally {
    resetting.value = false
  }
}

const testing = ref(false)
const testResult = ref<{ ok: boolean; message: string } | null>(null)

async function handleTestEmail() {
  testing.value = true
  testResult.value = null
  try {
    testResult.value = await testEmailConnection()
  } catch {
    testResult.value = { ok: false, message: '测试请求失败' }
  } finally {
    testing.value = false
  }
}

function isPasswordField(key: string): boolean {
  return key.includes('password')
}

function isWideField(f: ConfigField): boolean {
  return f.key === 'reminder.webhook_body_template'
}
</script>

<template>
  <section class="config-card">
    <div class="config-card__header">
      <h2 class="config-card__title">{{ title }}</h2>
      <div class="config-card__actions">
        <button
          v-if="group.group === 'email'"
          class="btn btn-secondary"
          :disabled="testing"
          @click="handleTestEmail"
        >
          {{ testing ? '测试中...' : '测试连接' }}
        </button>
        <button
          class="btn btn-primary"
          :disabled="!dirty || fs.submitting.value"
          @click="fs.handleSubmit"
        >
          {{ fs.submitting.value ? '保存中...' : '保存' }}
        </button>
      </div>
    </div>

    <!-- 测试结果 -->
    <Transition name="error-slide">
      <div v-if="testResult" class="test-result" :class="testResult.ok ? 'test-success' : 'test-error'">
        <component :is="testResult.ok ? CheckCircle2 : XCircle" :size="16" />
        {{ testResult.message }}
      </div>
    </Transition>

    <div v-if="fs.error.value" class="error-message">{{ fs.error.value }}</div>

    <!-- 可编辑字段列表 -->
    <div class="config-table">
      <div
        v-for="f in editableFields"
        :key="f.key"
        class="config-row"
        :class="{ 'config-row--wide': isWideField(f) }"
      >
        <div class="config-row__info">
          <span class="config-row__label">{{ f.label }}</span>
          <span class="config-row__meta">
            <span v-if="!f.hotReload" class="badge badge-xs badge-info">重启</span>
            <button
              v-if="f.source === 'db'"
              type="button"
              class="reset-btn"
              title="恢复默认"
              @click="askReset(f)"
            >
              <RotateCcw :size="12" />
            </button>
          </span>
        </div>

        <div class="config-row__control">
          <label v-if="f.type === 'bool'" class="toggle">
            <input
              type="checkbox"
              :checked="Boolean(fs.form[f.key])"
              @change="setField(f.key, ($event.target as HTMLInputElement).checked)"
            />
            <span class="toggle__track"><span class="toggle__thumb" /></span>
          </label>

          <BaseSelect
            v-else-if="f.type === 'enum'"
            :model-value="String(fs.form[f.key] ?? '')"
            :options="enumOptions(f)"
            :aria-label="f.label"
            @update:model-value="setField(f.key, $event)"
          />

          <textarea
            v-else-if="f.key === 'reminder.webhook_body_template'"
            class="form-input config-textarea"
            rows="4"
            :value="fs.form[f.key] as string"
            @input="setField(f.key, ($event.target as HTMLTextAreaElement).value)"
          />

          <input
            v-else-if="f.type === 'int' || f.type === 'float'"
            type="number"
            class="form-input config-input"
            :step="f.type === 'float' ? 'any' : '1'"
            :value="fs.form[f.key] as number"
            @input="onNumberInput(f.key, ($event.target as HTMLInputElement).value)"
          />

          <input
            v-else-if="isPasswordField(f.key)"
            type="password"
            class="form-input config-input"
            autocomplete="off"
            :value="fs.form[f.key] as string"
            @input="setField(f.key, ($event.target as HTMLInputElement).value)"
          />

          <input
            v-else
            type="text"
            class="form-input config-input"
            :value="fs.form[f.key] as string"
            @input="setField(f.key, ($event.target as HTMLInputElement).value)"
          />
        </div>
      </div>
    </div>

    <!-- 只读字段 -->
    <div v-if="readonlyFields.length" class="config-readonly-section">
      <div class="config-readonly-header">
        <Lock :size="13" />
        <span>引导类配置（只读）</span>
      </div>
      <div class="config-readonly-grid">
        <div v-for="f in readonlyFields" :key="f.key" class="config-readonly-item">
          <span class="config-readonly-item__label">{{ f.label }}</span>
          <span class="config-readonly-item__value">{{ formatValue(f.value) }}</span>
        </div>
      </div>
    </div>

    <BaseDialog v-model:visible="resetVisible" title="恢复默认">
      <p>
        确定将 <strong>{{ resetTarget?.label }}</strong> 恢复为配置文件 / 默认值？
        删除数据库覆盖后<strong>需重启</strong>生效。
      </p>
      <template #footer="{ close }">
        <button class="btn" :disabled="resetting" @click="close">取消</button>
        <button class="btn btn-danger" :disabled="resetting" @click="confirmReset">
          {{ resetting ? '处理中...' : '恢复默认' }}
        </button>
      </template>
    </BaseDialog>
  </section>
</template>

<style scoped>
.config-card {
  background: var(--color-surface);
  border: 1px solid var(--color-border);
  border-radius: var(--radius-xl);
  padding: 1.25rem 1.5rem;
  margin-bottom: 1rem;
}

.config-card__header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 1rem;
  margin-bottom: 1rem;
}

.config-card__title {
  font-size: 0.95rem;
  font-weight: 600;
  margin: 0;
  color: var(--color-text);
}

.config-card__actions {
  display: flex;
  gap: 0.5rem;
  flex-shrink: 0;
}

/* 表格式字段布局 */
.config-table {
  display: flex;
  flex-direction: column;
}

.config-row {
  display: grid;
  grid-template-columns: 1fr 1fr;
  align-items: center;
  gap: 1rem;
  padding: 0.6rem 0;
  border-bottom: 1px solid var(--color-border);
}

.config-row:last-child {
  border-bottom: none;
}

.config-row--wide {
  grid-template-columns: 1fr;
  align-items: start;
}

.config-row--wide .config-row__control {
  margin-top: 0.25rem;
}

.config-row__info {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  min-width: 0;
}

.config-row__label {
  font-size: var(--text-sm);
  font-weight: 500;
  color: var(--color-text);
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.config-row__meta {
  display: inline-flex;
  align-items: center;
  gap: 0.25rem;
  flex-shrink: 0;
}

.config-row__control {
  display: flex;
  align-items: center;
  justify-content: flex-end;
}

.config-row--wide .config-row__control {
  justify-content: stretch;
}

.config-input {
  width: 100%;
  max-width: 280px;
  height: 34px;
  font-size: var(--text-sm);
}

.config-textarea {
  width: 100%;
  font-size: var(--text-sm);
  font-family: var(--font-mono, monospace);
  resize: vertical;
  min-height: 80px;
}

/* Badge 紧凑版 */
.badge-xs {
  font-size: 0.65rem;
  padding: 0.05rem 0.35rem;
  border-radius: var(--radius-sm);
  font-weight: 600;
  line-height: 1.4;
}

/* 重置按钮 */
.reset-btn {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  padding: 2px;
  border: none;
  background: transparent;
  color: var(--color-text-muted);
  cursor: pointer;
  border-radius: var(--radius-sm);
  transition:
    color var(--motion-duration-fast) var(--motion-ease-standard),
    background var(--motion-duration-fast) var(--motion-ease-standard);
}

.reset-btn:hover {
  color: var(--color-primary);
  background: var(--color-surface-muted);
}

/* 测试结果 */
.test-result {
  display: flex;
  align-items: center;
  gap: 0.4rem;
  margin-bottom: 0.75rem;
  padding: 0.5rem 0.75rem;
  border-radius: var(--radius-md);
  font-size: var(--text-sm);
}

.test-success {
  background: color-mix(in srgb, var(--color-success, #22c55e) 10%, transparent);
  color: var(--color-success, #22c55e);
}

.test-error {
  background: color-mix(in srgb, var(--color-danger, #ef4444) 10%, transparent);
  color: var(--color-danger, #ef4444);
}

/* 只读区域 */
.config-readonly-section {
  margin-top: 1rem;
  padding-top: 1rem;
  border-top: 1px solid var(--color-border);
}

.config-readonly-header {
  display: flex;
  align-items: center;
  gap: 0.35rem;
  font-size: var(--text-xs);
  font-weight: 600;
  color: var(--color-text-muted);
  margin-bottom: 0.75rem;
  text-transform: uppercase;
  letter-spacing: 0.04em;
}

.config-readonly-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(200px, 1fr));
  gap: 0.5rem 1.5rem;
}

.config-readonly-item {
  display: flex;
  align-items: baseline;
  gap: 0.5rem;
  font-size: var(--text-sm);
}

.config-readonly-item__label {
  color: var(--color-text-muted);
  white-space: nowrap;
  flex-shrink: 0;
}

.config-readonly-item__value {
  color: var(--color-text);
  word-break: break-all;
  font-family: var(--font-mono, monospace);
  font-size: var(--text-xs);
}

/* Toggle 开关 */
.toggle {
  display: inline-flex;
  align-items: center;
  cursor: pointer;
  user-select: none;
}

.toggle input {
  position: absolute;
  width: 1px;
  height: 1px;
  opacity: 0;
}

.toggle__track {
  position: relative;
  width: 36px;
  height: 20px;
  background: var(--color-border);
  border-radius: var(--radius-pill);
  transition: background var(--motion-duration-fast) var(--motion-ease-standard);
}

.toggle__thumb {
  position: absolute;
  top: 2px;
  left: 2px;
  width: 16px;
  height: 16px;
  background: var(--color-surface);
  border-radius: 50%;
  box-shadow: 0 1px 3px color-mix(in srgb, var(--color-text) 25%, transparent);
  transition: transform var(--motion-duration-fast) var(--motion-ease-standard);
}

.toggle input:checked + .toggle__track {
  background: var(--color-primary);
}

.toggle input:checked + .toggle__track .toggle__thumb {
  transform: translateX(16px);
}

/* 移动端 */
@media (max-width: 767px) {
  .config-card {
    padding: 1rem;
  }

  .config-card__header {
    flex-direction: column;
    align-items: flex-start;
  }

  .config-card__actions {
    width: 100%;
    justify-content: flex-end;
  }

  .config-row {
    grid-template-columns: 1fr;
    gap: 0.35rem;
  }

  .config-row__info {
    flex-wrap: wrap;
  }

  .config-row__control {
    justify-content: stretch;
  }

  .config-input {
    max-width: none;
  }

  .config-readonly-grid {
    grid-template-columns: 1fr;
  }
}
</style>
