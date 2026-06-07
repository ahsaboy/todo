<script setup lang="ts">
import { ref, watch, onUnmounted } from 'vue'
import { Mail, Send } from 'lucide-vue-next'
import { useFormState } from '@/shared/composables/useFormState'
import { sendVerificationCode } from '@/entities/auth/api'

const props = defineProps<{
  currentEmail: string
  emailServiceEnabled: boolean
}>()

const emit = defineEmits<{
  submit: [email: string, code?: string]
}>()

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

async function handleSendCode() {
  if (cooldown.value > 0 || sendingCode.value) return
  codeError.value = ''
  sendingCode.value = true
  cooldown.value = 60
  try {
    await sendVerificationCode({ email: state.form.email, purpose: 'change_email' })
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
      <div class="input-icon-wrapper">
        <Mail class="input-icon" :size="18" :stroke-width="1.8" />
        <input
          id="profile-email"
          v-model="state.form.email"
          type="email"
          placeholder="your@email.com"
          autocomplete="email"
        />
      </div>
    </div>

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
          <Send :size="14" />
          {{ cooldown > 0 ? `${cooldown}s` : '发送验证码' }}
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
        :disabled="state.submitting.value || state.form.email === currentEmail"
      >
        {{ state.submitting.value ? '保存中...' : '保存' }}
      </button>
    </div>
  </form>
</template>

<style scoped>
.email-manager {
  display: flex;
  flex-direction: column;
  gap: 1rem;
}

.input-icon-wrapper {
  position: relative;
  width: 100%;
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
  display: flex;
  align-items: center;
  gap: 0.375rem;
  padding: 0 0.875rem;
  font-size: var(--text-sm);
  white-space: nowrap;
  flex-shrink: 0;
}

.code-send-btn:disabled {
  opacity: var(--opacity-disabled);
  cursor: not-allowed;
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
