<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import { User, Lock, Github, Globe, Terminal, Chrome, MessageCircle } from 'lucide-vue-next'
import { login, getEmailStatus, getOAuthProviders } from '@/entities/auth/api'
import { useAuthStore } from '@/app/stores/auth.store'
import { useFormState } from '@/shared/composables/useFormState'
import AuthBrandPanel from '@/shared/ui/AuthBrandPanel.vue'
import { API_BASE_URL } from '@/shared/config/api'
import type { OAuthProvider } from '@/entities/auth/model'

const router = useRouter()
const route = useRoute()
const authStore = useAuthStore()

const emailAvailable = ref(false)
const oauthProviders = ref<OAuthProvider[]>([])
const oauthError = ref<string | null>(null)

// lucide 图标映射：后端返回的 icon 名 → 组件
const iconComponents: Record<string, typeof User> = {
  github: Github,
  google: Chrome,
  globe: Globe,
  terminal: Terminal,
  message: MessageCircle,
  user: User,
}

onMounted(async () => {
  try {
    const res = await getEmailStatus()
    emailAvailable.value = res.data?.available ?? false
  } catch {
    emailAvailable.value = false
  }

  try {
    const res = await getOAuthProviders()
    oauthProviders.value = res.data ?? []
  } catch {
    oauthProviders.value = []
  }

  // 检查 URL 中的 OAuth 错误
  const hash = window.location.hash
  const errorMatch = hash.match(/[?&]error=([^&]+)/)
  if (errorMatch) {
    oauthError.value = decodeURIComponent(errorMatch[1])
    history.replaceState(null, '', '#/login')
  }
})

function handleOAuthLogin(provider: string) {
  window.location.href = `${API_BASE_URL}/auth/oauth/${provider}?redirect_uri=${encodeURIComponent(window.location.origin)}`
}

const { form: payload, submitting: isLoading, error, handleSubmit } = useFormState({
  initialData: { account: '', password: '' },
  onSubmit: async (data) => {
    const response = await login(data)
    authStore.setAuth(response.data.api_key, response.data.user)
    const redirect = typeof route.query.redirect === 'string' ? route.query.redirect : ''
    if (redirect) {
      await router.replace(redirect)
    } else {
      router.push({ name: 'tasks' })
    }
  },
})
</script>

<template>
  <div class="auth-page">
    <AuthBrandPanel />

    <div class="auth-form-panel">
      <section class="auth-card" aria-labelledby="login-title">
        <div class="auth-header">
          <p class="auth-eyebrow">任务管理系统</p>
          <h1 id="login-title">欢迎回来</h1>
        </div>

        <form @submit.prevent="handleSubmit">
          <div class="form-group">
            <label for="account">用户名或邮箱</label>
            <div class="input-icon-wrapper">
              <User class="input-icon" :size="18" :stroke-width="1.8" />
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
                autocomplete="current-password"
                placeholder="请输入密码"
              />
            </div>
          </div>
          <Transition name="error-slide">
            <div v-if="error" class="error-message">{{ error }}</div>
          </Transition>
          <button type="submit" :disabled="isLoading">
            {{ isLoading ? '正在登录...' : '登录' }}
          </button>
        </form>
        <p v-if="emailAvailable" class="auth-link">
          <router-link to="/forgot-password">忘记密码？</router-link>
        </p>
        <p class="auth-link">
          还没有账号？<router-link to="/register">立即注册</router-link>
        </p>
        <div v-if="oauthProviders.length > 0" class="oauth-section">
          <div class="oauth-divider">
            <span>或使用以下方式登录</span>
          </div>
          <Transition name="error-slide">
            <div v-if="oauthError" class="error-message">{{ oauthError }}</div>
          </Transition>
          <div class="oauth-buttons">
            <button
              v-for="provider in oauthProviders"
              :key="provider.name"
              class="oauth-btn"
              type="button"
              @click="handleOAuthLogin(provider.name)"
            >
              <component :is="iconComponents[provider.icon] || Globe" :size="18" :stroke-width="1.8" aria-hidden="true" />
              <span>{{ provider.label }}</span>
            </button>
          </div>
        </div>
      </section>
    </div>
  </div>
</template>
