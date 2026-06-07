<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import { User, Lock, Github, Globe, Terminal, Chrome, MessageCircle } from 'lucide-vue-next'
import { useAdminAuthStore } from '@/app/stores/admin-auth.store'
import { adminApi } from '@/shared/api/admin-client'
import { useFormState } from '@/shared/composables/useFormState'
import type { ApiResponse } from '@/shared/api/types'
import type { OAuthProvider } from '@/entities/auth/model'
import AuthBrandPanel from '@/shared/ui/AuthBrandPanel.vue'
import { ADMIN_API_BASE_URL } from '@/shared/config/api'
import { getBaseRedirectUri } from '@/shared/utils/url'

const router = useRouter()
const route = useRoute()
const adminAuthStore = useAdminAuthStore()

const oauthProviders = ref<OAuthProvider[]>([])
const oauthError = ref<string | null>(null)

const iconComponents: Record<string, typeof User> = {
  github: Github,
  google: Chrome,
  globe: Globe,
  terminal: Terminal,
  message: MessageCircle,
  user: User,
}

onMounted(async () => {
  // 管理后台复用用户端的 providers 接口（通过 localhost 访问）
  try {
    const res = await adminApi.get<ApiResponse<OAuthProvider[]>>('/auth/oauth/providers')
    oauthProviders.value = res.data ?? []
  } catch {
    oauthProviders.value = []
  }

  const hash = window.location.hash
  const errorMatch = hash.match(/[?&]error=([^&]+)/)
  if (errorMatch) {
    oauthError.value = decodeURIComponent(errorMatch[1])
    history.replaceState(null, '', '#/admin/login')
  }
})

function handleOAuthLogin(provider: string) {
  window.location.href = `${ADMIN_API_BASE_URL}/auth/oauth/${provider}?redirect_uri=${encodeURIComponent(getBaseRedirectUri())}`
}

const { form, submitting: isLoading, error, handleSubmit } = useFormState({
  initialData: { account: '', password: '' },
  onSubmit: async (data) => {
    const res = await adminApi.post<ApiResponse<{ api_key: string }>>('/auth/login', {
      account: data.account,
      password: data.password,
    })
    adminAuthStore.setAuth(res.data.api_key)
    const redirect = typeof route.query.redirect === 'string' ? route.query.redirect : ''
    if (redirect) {
      await router.replace(redirect)
    } else {
      router.push({ name: 'admin-dashboard' })
    }
  },
})
</script>

<template>
  <div class="auth-page">
    <AuthBrandPanel tagline="管理后台" />

    <div class="auth-form-panel">
      <section class="auth-card" aria-labelledby="admin-login-title">
        <div class="auth-header">
          <p class="auth-eyebrow">TODO 任务管理系统</p>
          <h1 id="admin-login-title">后台管理</h1>
        </div>
        <form @submit.prevent="handleSubmit">
          <div class="form-group">
            <label for="admin-account">用户名</label>
            <div class="input-icon-wrapper">
              <User class="input-icon" :size="18" :stroke-width="1.8" />
              <input
                id="admin-account"
                v-model="form.account"
                type="text"
                required
                autocomplete="username"
                placeholder="请输入管理员用户名"
              />
            </div>
          </div>
          <div class="form-group">
            <label for="admin-password">密码</label>
            <div class="input-icon-wrapper">
              <Lock class="input-icon" :size="18" :stroke-width="1.8" />
              <input
                id="admin-password"
                v-model="form.password"
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
            {{ isLoading ? '登录中...' : '登录管理后台' }}
          </button>
        </form>
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
