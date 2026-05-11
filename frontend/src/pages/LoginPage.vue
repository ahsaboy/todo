<script setup lang="ts">
import { ref } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import { login } from '@/entities/auth/api'
import { useAuthStore } from '@/app/stores/auth.store'
import type { LoginPayload } from '@/entities/auth/model'

const router = useRouter()
const route = useRoute()
const authStore = useAuthStore()

const payload = ref<LoginPayload>({
  account: '',
  password: '',
})
const error = ref('')
const isLoading = ref(false)

async function handleSubmit() {
  error.value = ''
  isLoading.value = true

  try {
    const response = await login(payload.value)
    authStore.setAuth(response.data.api_key, response.data.user)
    const redirect = typeof route.query.redirect === 'string' ? route.query.redirect : ''
    if (redirect) {
      await router.replace(redirect)
    } else {
      router.push({ name: 'tasks' })
    }
  } catch (e) {
    error.value = e instanceof Error ? e.message : '登录失败，请稍后重试'
  } finally {
    isLoading.value = false
  }
}
</script>

<template>
  <div class="auth-page">
    <section class="auth-card" aria-labelledby="login-title">
      <div class="auth-header">
        <p class="auth-eyebrow">任务管理系统</p>
        <h1 id="login-title">欢迎回来</h1>
      </div>

      <form @submit.prevent="handleSubmit">
        <div class="form-group">
          <label for="account">用户名或邮箱</label>
          <input
            id="account"
            v-model="payload.account"
            name="account"
            type="text"
            required
            autocomplete="username"
            placeholder="请输入用户名或邮箱"
          />
        </div>
        <div class="form-group">
          <label for="password">密码</label>
          <input
            id="password"
            v-model="payload.password"
            name="password"
            type="password"
            required
            autocomplete="current-password"
            placeholder="请输入密码"
          />
        </div>
        <div v-if="error" class="error-message">{{ error }}</div>
        <button type="submit" :disabled="isLoading">
          {{ isLoading ? '正在登录...' : '登录' }}
        </button>
      </form>
      <p class="auth-link">
        还没有账号？<router-link to="/register">立即注册</router-link>
      </p>
    </section>
  </div>
</template>

<style scoped>
.auth-page {
  min-height: 100vh;
  display: flex;
  justify-content: center;
  align-items: center;
  padding: 24px;
  background:
    radial-gradient(circle at top left, var(--color-bg-radial-a), transparent 32%),
    radial-gradient(circle at top right, var(--color-bg-radial-b), transparent 28%),
    linear-gradient(
      135deg,
      color-mix(in srgb, var(--color-bg) 96%, var(--color-surface) 4%) 0%,
      color-mix(in srgb, var(--color-bg) 88%, var(--color-surface-muted) 12%) 100%
    );
}

.auth-card {
  width: 100%;
  max-width: 420px;
  padding: 32px;
  border: 1px solid var(--color-border);
  border-radius: 8px;
  background: var(--color-surface);
  box-shadow: var(--shadow-panel);
}

.auth-header {
  margin-bottom: 28px;
}

.auth-eyebrow {
  margin: 0 0 8px;
  color: var(--color-primary);
  font-size: 13px;
  font-weight: 700;
}

.auth-header h1 {
  margin: 0;
  color: var(--color-text);
  font-size: 28px;
  line-height: 36px;
  font-weight: 600;
}

.auth-header p:last-child {
  margin: 10px 0 0;
  color: var(--color-text-muted);
  font-size: 14px;
  line-height: 22px;
}

.form-group {
  margin-bottom: 18px;
}

label {
  display: block;
  margin-bottom: 8px;
  color: var(--color-text);
  font-size: 14px;
  font-weight: 500;
}

input {
  width: 100%;
  min-height: 44px;
  padding: 10px 12px;
  border: 1px solid var(--color-border);
  border-radius: 6px;
  background: var(--color-surface-muted);
  color: var(--color-text);
  font-size: 15px;
  transition:
    background-color 0.2s,
    border-color 0.2s,
    box-shadow 0.2s;
}

input::placeholder {
  color: var(--color-text-muted);
}

input:focus {
  outline: none;
  background: var(--color-surface);
  border-color: var(--color-primary);
  box-shadow: 0 0 0 3px color-mix(in srgb, var(--color-primary) 14%, transparent);
}

button {
  width: 100%;
  min-height: 46px;
  padding: 0 16px;
  margin-top: 4px;
  border: none;
  border-radius: 6px;
  background: var(--color-primary);
  color: #fff;
  font-size: 15px;
  font-weight: 600;
  cursor: pointer;
  transition: background 0.2s;
}

button:hover:not(:disabled) {
  background: var(--color-primary-hover);
}

button:disabled {
  opacity: 0.6;
  cursor: not-allowed;
}

.error-message {
  margin: 0 0 16px;
  padding: 10px 12px;
  border: 1px solid color-mix(in srgb, var(--color-danger) 20%, transparent);
  border-radius: 6px;
  background: color-mix(in srgb, var(--color-glow-danger) 72%, transparent);
  color: var(--color-danger);
  font-size: 14px;
  line-height: 20px;
}

.auth-link {
  margin: 20px 0 0;
  color: var(--color-text-muted);
  font-size: 14px;
  line-height: 22px;
  text-align: center;
}

.auth-link a {
  color: var(--color-primary);
  font-weight: 600;
  text-decoration: none;
}

.auth-link a:hover {
  text-decoration: underline;
}

@media (max-width: 480px) {
  .auth-page {
    align-items: stretch;
    padding: 16px;
  }

  .auth-card {
    display: flex;
    flex-direction: column;
    justify-content: center;
    padding: 28px 20px;
  }
}
</style>
