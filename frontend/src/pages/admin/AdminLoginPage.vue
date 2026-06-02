<script setup lang="ts">
import { useRouter, useRoute } from 'vue-router'
import { User, Lock } from 'lucide-vue-next'
import { useAdminAuthStore } from '@/app/stores/admin-auth.store'
import { adminApi } from '@/shared/api/admin-client'
import { useFormState } from '@/shared/composables/useFormState'
import type { ApiResponse } from '@/shared/api/types'
import AuthBrandPanel from '@/shared/ui/AuthBrandPanel.vue'

const router = useRouter()
const route = useRoute()
const adminAuthStore = useAdminAuthStore()

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
      </section>
    </div>
  </div>
</template>
