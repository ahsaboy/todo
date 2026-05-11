<script setup lang="ts">
import { ref } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import { register } from '@/entities/auth/api'
import { useAuthStore } from '@/app/stores/auth.store'
import type { RegisterPayload } from '@/entities/auth/model'

const router = useRouter()
const route = useRoute()
const authStore = useAuthStore()

const payload = ref<RegisterPayload>({
  username: '',
  email: '',
  password: '',
})
const error = ref('')
const isLoading = ref(false)

async function handleSubmit() {
  error.value = ''
  isLoading.value = true

  try {
    const response = await register(payload.value)
    authStore.setAuth(response.data.api_key, response.data.user)
    const redirect = typeof route.query.redirect === 'string' ? route.query.redirect : ''
    if (redirect) {
      await router.replace(redirect)
    } else {
      router.push({ name: 'tasks' })
    }
  } catch (e) {
    error.value = e instanceof Error ? e.message : '注册失败，请稍后重试'
  } finally {
    isLoading.value = false
  }
}
</script>

<template>
  <div class="auth-page">
    <section class="auth-card" aria-labelledby="register-title">
      <div class="auth-header">
        <p class="auth-eyebrow">创建账号</p>
        <h1 id="register-title">开始管理你的任务</h1>
      </div>

      <form @submit.prevent="handleSubmit">
        <div class="form-group">
          <label for="username">用户名</label>
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
        <div class="form-group">
          <label for="email">邮箱（选填）</label>
          <input
            id="email"
            v-model="payload.email"
            name="email"
            type="email"
            autocomplete="email"
            placeholder="请输入邮箱地址"
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
            minlength="6"
            maxlength="72"
            autocomplete="new-password"
            placeholder="至少 6 个字符"
          />
        </div>
        <div v-if="error" class="error-message">{{ error }}</div>
        <button type="submit" :disabled="isLoading">
          {{ isLoading ? '正在注册...' : '注册' }}
        </button>
      </form>
      <p class="auth-link">已有账号？<router-link to="/login">返回登录</router-link></p>
    </section>
  </div>
</template>
