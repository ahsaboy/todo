<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { useRouter } from 'vue-router'
import { AlertCircle } from 'lucide-vue-next'
import { useAdminAuthStore } from '@/app/stores/admin-auth.store'
import { adminApi } from '@/shared/api/admin-client'
import type { ApiResponse } from '@/shared/api/types'

const router = useRouter()
const adminAuthStore = useAdminAuthStore()
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
    adminAuthStore.setAuth(key)
    await adminApi.get<ApiResponse<{ total_users: number }>>('/stats')
    await router.push({ name: 'admin-dashboard' })
  } catch (err: unknown) {
    adminAuthStore.logout()
    status.value = 'error'
    const isForbidden = err instanceof Error && 'status' in err && (err as { status: number }).status === 403
    errorMsg.value = isForbidden ? '登录失败：账号无管理员权限' : '登录失败：网络错误或服务不可用'
  }
})
</script>

<template>
  <div class="auth-page">
    <div class="auth-form-panel">
      <section class="auth-card">
        <div class="auth-header">
          <p class="auth-eyebrow">TODO 任务管理系统</p>
          <h1>正在登录管理后台...</h1>
        </div>
        <div v-if="status === 'loading'" class="oauth-status" role="status" aria-live="polite">
          <div class="spinner"></div>
          <p>正在验证管理员权限，请稍候...</p>
        </div>
        <div v-else class="oauth-status error" role="alert" aria-live="assertive">
          <AlertCircle :size="40" :stroke-width="1.5" style="color: var(--color-danger)" />
          <p class="error-message">{{ errorMsg }}</p>
          <router-link to="/admin/login" class="btn-link">返回登录页</router-link>
        </div>
      </section>
    </div>
  </div>
</template>
