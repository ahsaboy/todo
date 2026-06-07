<script setup lang="ts">
import { ref, computed, onMounted, onUnmounted } from 'vue'
import { useRouter } from 'vue-router'
import { Mail, ShieldCheck, Lock } from 'lucide-vue-next'
import { getEmailStatus, sendVerificationCode, resetPassword } from '@/entities/auth/api'
import { useFormState } from '@/shared/composables/useFormState'
import AuthBrandPanel from '@/shared/ui/AuthBrandPanel.vue'

const router = useRouter()

const step = ref(1)
const email = ref('')
const code = ref('')
const password = ref('')
const confirmPassword = ref('')
const cooldown = ref(0)
const sendingCode = ref(false)
const stepError = ref('')
let cooldownTimer: ReturnType<typeof setInterval> | null = null

onMounted(async () => {
  try {
    const res = await getEmailStatus()
    if (!res.data?.available) {
      router.replace('/login')
    }
  } catch {
    router.replace('/login')
  }
})

onUnmounted(() => {
  if (cooldownTimer) clearInterval(cooldownTimer)
})

const canSendCode = computed(() =>
  email.value.includes('@') && cooldown.value === 0 && !sendingCode.value,
)

// Step 1: 发送验证码
async function handleSendCode() {
  if (!canSendCode.value) return
  stepError.value = ''
  sendingCode.value = true
  try {
    await sendVerificationCode({ email: email.value })
    startCooldown()
    step.value = 2
  } catch (e: unknown) {
    stepError.value = e instanceof Error ? e.message : '发送失败'
  } finally {
    sendingCode.value = false
  }
}

// Step 2: 验证码 + 新密码提交
const { submitting: resetting, error: resetError, handleSubmit } = useFormState({
  initialData: {},
  onSubmit: async () => {
    if (password.value !== confirmPassword.value) {
      throw new Error('两次输入的密码不一致')
    }
    await resetPassword({ email: email.value, code: code.value, password: password.value })
    step.value = 4
  },
})

function goToLogin() {
  router.push('/login')
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
      <section class="auth-card" aria-labelledby="forgot-title">
        <div class="auth-header">
          <p class="auth-eyebrow">密码重置</p>
          <h1 id="forgot-title">
            {{ step === 1 ? '找回密码' : step === 2 ? '验证身份' : step === 3 ? '设置新密码' : '重置成功' }}
          </h1>
        </div>

        <!-- Step 1: 输入邮箱 -->
        <form v-if="step === 1" @submit.prevent="handleSendCode">
          <div class="form-group">
            <label for="email">注册邮箱</label>
            <div class="input-icon-wrapper">
              <Mail class="input-icon" :size="18" :stroke-width="1.8" />
              <input
                id="email"
                v-model="email"
                type="email"
                required
                autocomplete="email"
                placeholder="请输入注册时的邮箱"
              />
            </div>
          </div>
          <Transition name="error-slide">
            <div v-if="stepError" class="error-message">{{ stepError }}</div>
          </Transition>
          <button type="submit" :disabled="!canSendCode">
            {{ sendingCode ? '发送中...' : '发送验证码' }}
          </button>
        </form>

        <!-- Step 2: 输入验证码 -->
        <form v-else-if="step === 2" @submit.prevent="() => { step = 3 }">
          <p class="step-hint">验证码已发送至 <strong>{{ email }}</strong></p>
          <div class="form-group">
            <label for="code">验证码</label>
            <div class="input-icon-wrapper">
              <ShieldCheck class="input-icon" :size="18" :stroke-width="1.8" />
              <input
                id="code"
                v-model="code"
                type="text"
                inputmode="numeric"
                pattern="[0-9]*"
                maxlength="6"
                required
                autocomplete="one-time-code"
                placeholder="6 位验证码"
              />
            </div>
          </div>
          <div class="resend-row">
            <button
              type="button"
              class="btn-link"
              :disabled="!canSendCode"
              @click="handleSendCode"
            >
              {{ cooldown > 0 ? `重新发送(${cooldown}s)` : '重新发送' }}
            </button>
          </div>
          <button type="submit" :disabled="code.length !== 6">
            下一步
          </button>
        </form>

        <!-- Step 3: 设置新密码 -->
        <form v-else-if="step === 3" @submit.prevent="handleSubmit">
          <div class="form-group">
            <label for="new-password">新密码</label>
            <div class="input-icon-wrapper">
              <Lock class="input-icon" :size="18" :stroke-width="1.8" />
              <input
                id="new-password"
                v-model="password"
                type="password"
                required
                minlength="6"
                maxlength="72"
                autocomplete="new-password"
                placeholder="至少 6 个字符"
              />
            </div>
          </div>
          <div class="form-group">
            <label for="confirm-password">确认密码</label>
            <div class="input-icon-wrapper">
              <Lock class="input-icon" :size="18" :stroke-width="1.8" />
              <input
                id="confirm-password"
                v-model="confirmPassword"
                type="password"
                required
                minlength="6"
                maxlength="72"
                autocomplete="new-password"
                placeholder="再次输入密码"
              />
            </div>
          </div>
          <Transition name="error-slide">
            <div v-if="resetError" class="error-message">{{ resetError }}</div>
          </Transition>
          <button type="submit" :disabled="resetting">
            {{ resetting ? '重置中...' : '重置密码' }}
          </button>
        </form>

        <!-- Step 4: 成功 -->
        <div v-else class="success-section">
          <p class="success-text">密码已重置成功！</p>
          <button type="button" @click="goToLogin">返回登录</button>
        </div>

        <p v-if="step < 4" class="auth-link">
          <router-link to="/login">返回登录</router-link>
        </p>
      </section>
    </div>
  </div>
</template>

<style scoped>
.step-hint {
  font-size: var(--text-sm);
  color: var(--color-text-muted);
  margin-bottom: 1rem;
  line-height: 1.5;
}

.step-hint strong {
  color: var(--color-text);
}

.resend-row {
  display: flex;
  justify-content: flex-end;
  margin-bottom: 1rem;
}

.btn-link {
  background: none;
  border: none;
  color: var(--color-primary);
  font-size: var(--text-sm);
  cursor: pointer;
  padding: 0;
  text-decoration: underline;
  text-underline-offset: 2px;
}

.btn-link:disabled {
  opacity: var(--opacity-disabled);
  cursor: not-allowed;
  text-decoration: none;
}

.success-section {
  text-align: center;
  padding: 1.5rem 0;
}

.success-text {
  font-size: var(--text-md);
  color: var(--color-text);
  margin-bottom: 1.5rem;
}
</style>
