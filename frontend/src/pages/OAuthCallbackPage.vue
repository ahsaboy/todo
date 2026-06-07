<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { useRouter } from 'vue-router'
import { AlertCircle } from 'lucide-vue-next'
import { useAuthStore } from '@/app/stores/auth.store'

const router = useRouter()
const authStore = useAuthStore()
const status = ref<'loading' | 'error'>('loading')
const errorMsg = ref('')

onMounted(async () => {
  const hash = window.location.hash
  const keyMatch = hash.match(/[?&]key=([^&]+)/)
  const key = keyMatch ? decodeURIComponent(keyMatch[1]) : null

  if (keyMatch) {
    const cleanHash = hash.replace(/[?&]key=[^&]*/, '').replace(/\?$/, '')
    history.replaceState(null, '', cleanHash)
  }

  if (!key) {
    status.value = 'error'
    errorMsg.value = '登录失败：缺少认证密钥'
    return
  }

  try {
    authStore.setOAuthAuth(key)
    await authStore.fetchProfile()
    const hash2 = window.location.hash
    const redirectMatch = hash2.match(/[?&]redirect=([^&]+)/)
    const redirect = redirectMatch ? decodeURIComponent(redirectMatch[1]) : ''
    if (redirect) {
      await router.replace(redirect)
    } else {
      await router.push({ name: 'tasks' })
    }
  } catch {
    authStore.logout()
    status.value = 'error'
    errorMsg.value = '登录失败：无法获取用户信息'
  }
})
</script>

<template>
  <div class="auth-page">
    <div class="auth-form-panel">
      <section class="auth-card">
        <div class="auth-header">
          <p class="auth-eyebrow">任务管理系统</p>
          <h1>正在登录...</h1>
        </div>
        <div v-if="status === 'loading'" class="oauth-status" role="status" aria-live="polite">
          <div class="spinner"></div>
          <p>正在完成登录，请稍候...</p>
        </div>
        <div v-else class="oauth-status error" role="alert" aria-live="assertive">
          <AlertCircle :size="40" :stroke-width="1.5" style="color: var(--color-danger)" />
          <p class="error-message">{{ errorMsg }}</p>
          <router-link to="/login" class="btn-link">返回登录页</router-link>
        </div>
      </section>
    </div>
  </div>
</template>
