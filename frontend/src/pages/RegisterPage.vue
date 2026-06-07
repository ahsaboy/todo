<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import { User, Mail, Lock, ShieldCheck } from 'lucide-vue-next'
import { register, getEmailStatus, sendVerificationCode } from '@/entities/auth/api'
import { useAuthStore } from '@/app/stores/auth.store'
import { useFormState } from '@/shared/composables/useFormState'
import AuthBrandPanel from '@/shared/ui/AuthBrandPanel.vue'

const router = useRouter()
const route = useRoute()
const authStore = useAuthStore()

const emailRequired = ref(false)
const code = ref('')
const cooldown = ref(0)
const sendingCode = ref(false)
const sendCodeError = ref('')
let cooldownTimer: ReturnType<typeof setInterval> | null = null

onMounted(async () => {
  try {
    const res = await getEmailStatus()
    emailRequired.value = res.data?.available ?? false
  } catch {
    emailRequired.value = false
  }
})

const emailLabel = computed(() => emailRequired.value ? '邮箱' : '邮箱（选填）')

const { form: payload, submitting: isLoading, error, handleSubmit } = useFormState({
  initialData: { username: '', email: '', password: '' },
  onSubmit: async (data) => {
    const submitData = emailRequired.value
      ? { ...data, code: code.value }
      : data
    const response = await register(submitData)
    authStore.setAuth(response.data.api_key, response.data.user)
    const redirect = typeof route.query.redirect === 'string' ? route.query.redirect : ''
    if (redirect) {
      await router.replace(redirect)
    } else {
      router.push({ name: 'tasks' })
    }
  },
})

const canSendCode = computed(() => {
  const email = payload.email
  return email && email.includes('@') && cooldown.value === 0 && !sendingCode.value
})

async function handleSendCode() {
  if (!canSendCode.value) return
  sendCodeError.value = ''
  sendingCode.value = true
  try {
    await sendVerificationCode({ email: payload.email })
    startCooldown()
  } catch (e: unknown) {
    sendCodeError.value = e instanceof Error ? e.message : '验证码发送失败，请稍后重试'
  } finally {
    sendingCode.value = false
  }
}

function startCooldown() {
  cooldown.value = 60
  cooldownTimer = setInterval(() => {
    cooldown.value--
    if (cooldown.value <= 0) {
      if (cooldownTimer) clearInterval(cooldownTimer)
      cooldownTimer = null
    }
  }, 1000)
}
</script>

<template>
  <div class="auth-page">
    <AuthBrandPanel />

    <div class="auth-form-panel">
      <section class="auth-card" aria-labelledby="register-title">
        <div class="auth-header">
          <p class="auth-eyebrow">创建账号</p>
          <h1 id="register-title">开始管理你的任务</h1>
        </div>

        <form @submit.prevent="handleSubmit">
          <div class="form-group">
            <label for="username">用户名</label>
            <div class="input-icon-wrapper">
              <User class="input-icon" :size="18" :stroke-width="1.8" />
              <input
                id="username"
                v-model="payload.username"
                name="username"
                type="text"
                required
                minlength="3"
                maxlength="32"
                autocomplete="username"
                placeholder="3-32 个字符"
              />
            </div>
          </div>
          <div class="form-group">
            <label for="email">{{ emailLabel }}</label>
            <div class="input-icon-wrapper">
              <Mail class="input-icon" :size="18" :stroke-width="1.8" />
              <input
                id="email"
                v-model="payload.email"
                name="email"
                type="email"
                :required="emailRequired"
                autocomplete="email"
                placeholder="请输入邮箱地址"
              />
            </div>
          </div>
          <div v-if="emailRequired" class="form-group">
            <label for="code">验证码</label>
            <div class="code-row">
              <div class="input-icon-wrapper code-input-wrapper">
                <ShieldCheck class="input-icon" :size="18" :stroke-width="1.8" />
                <input
                  id="code"
                  v-model="code"
                  name="code"
                  type="text"
                  inputmode="numeric"
                  pattern="[0-9]*"
                  maxlength="6"
                  required
                  autocomplete="one-time-code"
                  placeholder="6 位验证码"
                />
              </div>
              <button
                type="button"
                class="btn-send-code"
                :disabled="!canSendCode"
                @click="handleSendCode"
              >
                {{ cooldown > 0 ? `重新发送(${cooldown}s)` : sendingCode ? '发送中...' : '获取验证码' }}
              </button>
            </div>
            <p v-if="sendCodeError" class="code-error">{{ sendCodeError }}</p>
          </div>
          <div class="form-group">
            <label for="password">密码</label>
            <div class="input-icon-wrapper">
              <Lock class="input-icon" :size="18" :stroke-width="1.8" />
              <input
                id="password"
                v-model="payload.password"
                name="password"
                type="password"
                required
                minlength="6"
                maxlength="72"
                autocomplete="new-password"
                placeholder="至少 6 个字符"
              />
            </div>
          </div>
          <Transition name="error-slide">
            <div v-if="error" class="error-message">{{ error }}</div>
          </Transition>
          <button type="submit" :disabled="isLoading">
            {{ isLoading ? '正在注册...' : '注册' }}
          </button>
        </form>
        <p class="auth-link">已有账号？<router-link to="/login">返回登录</router-link></p>
      </section>
    </div>
  </div>
</template>

<style scoped>
.code-row {
  display: flex;
  gap: 0.5rem;
}

.code-input-wrapper {
  flex: 1;
}

.btn-send-code {
  flex-shrink: 0;
  padding: 0 1rem;
  height: 42px;
  border: 1px solid var(--color-primary);
  border-radius: var(--radius-md);
  background: transparent;
  color: var(--color-primary);
  font-size: var(--text-sm);
  font-weight: 500;
  cursor: pointer;
  white-space: nowrap;
  transition: background var(--motion-duration-fast) var(--motion-ease-standard),
    opacity var(--motion-duration-fast) var(--motion-ease-standard);
}

.btn-send-code:hover:not(:disabled) {
  background: color-mix(in srgb, var(--color-primary) 10%, transparent);
}

.btn-send-code:disabled {
  opacity: var(--opacity-disabled);
  cursor: not-allowed;
}

.code-error {
  font-size: var(--text-xs);
  color: var(--color-danger, #ef4444);
  margin-top: 0.35rem;
}
</style>
