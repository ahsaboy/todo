<script setup lang="ts">
import { ref, watch, onUnmounted, computed } from 'vue'
import { Mail, Pencil, X } from 'lucide-vue-next'
import { useFormState } from '@/shared/composables/useFormState'
import { sendVerificationCode } from '@/entities/auth/api'

const props = defineProps<{
  currentEmail: string
  emailServiceEnabled: boolean
}>()

const emit = defineEmits<{
  submit: [email: string, code?: string]
}>()

const editing = ref(false)
const cooldown = ref(0)
const sendingCode = ref(false)
const codeError = ref('')
let cooldownTimer: ReturnType<typeof setInterval> | null = null

onUnmounted(() => {
  if (cooldownTimer) clearInterval(cooldownTimer)
})

const state = useFormState({
  initialData: { email: '', code: '' },
  validate: (data) => {
    if (!data.email) return '请输入邮箱'
    if (props.emailServiceEnabled && !data.code) return '请输入验证码'
    return null
  },
  onSubmit: async (data) => {
    emit('submit', data.email, props.emailServiceEnabled ? data.code : undefined)
  },
})

watch(
  () => props.currentEmail,
  (val) => {
    if (val) state.resetTo({ email: val, code: '' })
  },
  { immediate: true },
)

const emailChanged = computed(() => state.form.email !== props.currentEmail)

function startEdit() {
  editing.value = true
}

function cancelEdit() {
  editing.value = false
  state.form.email = props.currentEmail
  state.form.code = ''
  codeError.value = ''
}

async function handleSendCode() {
  if (cooldown.value > 0 || sendingCode.value) return
  codeError.value = ''
  sendingCode.value = true
  cooldown.value = 60
  try {
    await sendVerificationCode({ email: state.form.email })
    cooldownTimer = setInterval(() => {
      cooldown.value--
      if (cooldown.value <= 0) {
        clearInterval(cooldownTimer!)
        cooldownTimer = null
      }
    }, 1000)
  } catch {
    codeError.value = '发送验证码失败'
    cooldown.value = 0
  } finally {
    sendingCode.value = false
  }
}
</script>

<template>
  <form class="email-manager" @submit.prevent="state.handleSubmit">
    <div class="form-group">
      <label for="profile-email">邮箱</label>
      <div class="email-field">
        <div class="input-icon-wrapper" :class="{ 'is-disabled': !editing }">
          <Mail class="input-icon" :size="18" :stroke-width="1.8" />
          <input
            id="profile-email"
            v-model="state.form.email"
            type="email"
            placeholder="your@email.com"
            autocomplete="email"
            :disabled="!editing"
          />
        </div>
        <button
          v-if="!editing"
          type="button"
          class="edit-btn"
          title="修改邮箱"
          @click="startEdit"
        >
          <Pencil :size="14" />
        </button>
        <button
          v-else
          type="button"
          class="edit-btn cancel"
          title="取消修改"
          @click="cancelEdit"
        >
          <X :size="14" />
        </button>
      </div>
    </div>

    <template v-if="editing">
      <div v-if="emailServiceEnabled" class="form-group">
        <label for="profile-code">验证码</label>
        <div class="code-row">
          <input
            id="profile-code"
            v-model="state.form.code"
            type="text"
            maxlength="6"
            placeholder="6位验证码"
            autocomplete="one-time-code"
            class="code-input"
          />
          <button
            type="button"
            class="code-send-btn"
            :disabled="cooldown > 0 || sendingCode || !state.form.email"
            @click="handleSendCode"
          >
            <span v-if="sendingCode" class="send-spinner" />
            {{ cooldown > 0 ? `重新发送 ${cooldown}s` : (sendingCode ? '发送中...' : '获取验证码') }}
          </button>
        </div>
      </div>

      <Transition name="error-slide">
        <div v-if="state.error.value || codeError" class="error-text">{{ state.error.value || codeError }}</div>
      </Transition>

      <div class="form-actions">
        <button
          type="submit"
          class="btn-primary"
          :disabled="state.submitting.value || !emailChanged"
        >
          {{ state.submitting.value ? '保存中...' : '保存' }}
        </button>
      </div>
    </template>
  </form>
</template>

<style scoped>
.email-manager {
  display: flex;
  flex-direction: column;
  gap: 1rem;
}

.email-field {
  display: flex;
  align-items: center;
  gap: 0.5rem;
}

.input-icon-wrapper {
  position: relative;
  flex: 1;
  min-width: 0;
}

.input-icon-wrapper .input-icon {
  position: absolute;
  left: 0.75rem;
  top: 50%;
  transform: translateY(-50%);
  color: var(--color-text-muted);
  pointer-events: none;
}

.input-icon-wrapper input {
  width: 100%;
  padding-left: 2.5rem;
}

.input-icon-wrapper.is-disabled input {
  background: var(--color-surface-muted);
  color: var(--color-text-muted);
  cursor: default;
}

.edit-btn {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 34px;
  height: 34px;
  border: 1px solid var(--color-border);
  border-radius: var(--radius-md);
  background: var(--color-surface);
  color: var(--color-text-muted);
  cursor: pointer;
  flex-shrink: 0;
  transition:
    background-color var(--motion-duration-fast),
    border-color var(--motion-duration-fast),
    color var(--motion-duration-fast);
}

.edit-btn:hover {
  background: var(--color-surface-muted);
  border-color: var(--color-primary);
  color: var(--color-primary);
}

.edit-btn.cancel:hover {
  border-color: var(--color-danger);
  color: var(--color-danger);
}

.code-row {
  display: flex;
  gap: 0.5rem;
}

.code-input {
  flex: 1;
  min-width: 0;
  letter-spacing: 0.2em;
  text-align: center;
}

.code-send-btn {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  gap: 0.375rem;
  padding: 0 1rem;
  min-height: 36px;
  font-size: var(--text-sm);
  font-weight: 500;
  white-space: nowrap;
  flex-shrink: 0;
  color: var(--color-primary);
  border: 1px solid color-mix(in srgb, var(--color-primary) 40%, transparent);
  border-radius: var(--radius-md);
  background: color-mix(in srgb, var(--color-primary) 8%, transparent);
  cursor: pointer;
  transition:
    background-color var(--motion-duration-fast),
    border-color var(--motion-duration-fast),
    color var(--motion-duration-fast);
}

.code-send-btn:hover:not(:disabled) {
  background: color-mix(in srgb, var(--color-primary) 14%, transparent);
  border-color: var(--color-primary);
}

.code-send-btn:disabled {
  opacity: var(--opacity-disabled);
  cursor: not-allowed;
}

.send-spinner {
  width: 14px;
  height: 14px;
  border: 2px solid color-mix(in srgb, var(--color-primary) 30%, transparent);
  border-top-color: var(--color-primary);
  border-radius: 50%;
  animation: spin 0.6s linear infinite;
}

@keyframes spin {
  to { transform: rotate(360deg); }
}

.error-text {
  font-size: var(--text-xs);
  color: var(--color-danger);
  margin: 0;
}

.form-actions {
  display: flex;
  justify-content: flex-end;
  padding-top: 0.25rem;
}

.btn-primary {
  min-height: 36px;
  padding: 0 1.25rem;
}

.btn-primary:disabled {
  opacity: var(--opacity-disabled);
  cursor: not-allowed;
}
</style>
