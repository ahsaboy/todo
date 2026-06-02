<script setup lang="ts">
import { useRouter, useRoute } from 'vue-router'
import { User, Mail, Lock } from 'lucide-vue-next'
import { register } from '@/entities/auth/api'
import { useAuthStore } from '@/app/stores/auth.store'
import { useFormState } from '@/shared/composables/useFormState'
import AuthBrandPanel from '@/shared/ui/AuthBrandPanel.vue'

const router = useRouter()
const route = useRoute()
const authStore = useAuthStore()

const { form: payload, submitting: isLoading, error, handleSubmit } = useFormState({
  initialData: { username: '', email: '', password: '' },
  onSubmit: async (data) => {
    const response = await register(data)
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
            <label for="email">邮箱（选填）</label>
            <div class="input-icon-wrapper">
              <Mail class="input-icon" :size="18" :stroke-width="1.8" />
              <input
                id="email"
                v-model="payload.email"
                name="email"
                type="email"
                autocomplete="email"
                placeholder="请输入邮箱地址"
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
