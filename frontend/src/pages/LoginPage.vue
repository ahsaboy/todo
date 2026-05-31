<script setup lang="ts">
import { ref } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import { User, Lock } from 'lucide-vue-next'
import { login } from '@/entities/auth/api'
import { useAuthStore } from '@/app/stores/auth.store'
import type { LoginPayload } from '@/entities/auth/model'
import AuthBrandPanel from '@/shared/ui/AuthBrandPanel.vue'

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
        <p class="auth-link">
          还没有账号？<router-link to="/register">立即注册</router-link>
        </p>
      </section>
    </div>
  </div>
</template>
