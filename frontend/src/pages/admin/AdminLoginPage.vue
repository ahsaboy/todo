<script setup lang="ts">
import { ref } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import { useAdminAuthStore } from '@/app/stores/admin-auth.store'
import { adminApi } from '@/shared/api/admin-client'
import type { ApiResponse } from '@/shared/api/types'

const router = useRouter()
const route = useRoute()
const adminAuthStore = useAdminAuthStore()

const token = ref('')
const error = ref('')
const isLoading = ref(false)

async function handleSubmit() {
  error.value = ''
  isLoading.value = true
  try {
    await adminApi.post<ApiResponse<{ ok: boolean }>>('/auth/verify', { token: token.value })
    adminAuthStore.setAuth(token.value)
    const redirect = typeof route.query.redirect === 'string' ? route.query.redirect : ''
    if (redirect) {
      await router.replace(redirect)
    } else {
      router.push({ name: 'admin-dashboard' })
    }
  } catch {
    error.value = '令牌无效，请检查后重试'
  } finally {
    isLoading.value = false
  }
}
</script>

<template>
  <div class="auth-page">
    <section class="auth-card" aria-labelledby="admin-login-title">
      <div class="auth-header">
        <p class="auth-eyebrow">TODO 任务管理系统</p>
        <h1 id="admin-login-title">后台管理</h1>
      </div>
      <form @submit.prevent="handleSubmit">
        <div class="form-group">
          <label for="admin-token">管理令牌</label>
          <input
            id="admin-token"
            v-model="token"
            type="password"
            required
            autocomplete="off"
            placeholder="请输入管理员令牌"
          />
        </div>
        <div v-if="error" class="error-message">{{ error }}</div>
        <button type="submit" :disabled="isLoading">
          {{ isLoading ? '验证中...' : '进入管理后台' }}
        </button>
      </form>
    </section>
  </div>
</template>
