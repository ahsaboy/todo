<script setup lang="ts">
import { ref } from 'vue'
import { useRouter } from 'vue-router'
import { login } from '@/entities/auth/api'
import { useAuthStore } from '@/app/stores/auth.store'
import type { LoginPayload } from '@/entities/auth/model'

const router = useRouter()
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
    router.push({ name: 'tasks' })
  } catch (e) {
    error.value = e instanceof Error ? e.message : 'Login failed'
  } finally {
    isLoading.value = false
  }
}
</script>

<template>
  <div class="auth-page">
    <div class="auth-card">
      <h1>Login</h1>
      <form @submit.prevent="handleSubmit">
        <div class="form-group">
          <label for="account">Username or Email</label>
          <input
            id="account"
            v-model="payload.account"
            type="text"
            required
            autocomplete="username"
          />
        </div>
        <div class="form-group">
          <label for="password">Password</label>
          <input
            id="password"
            v-model="payload.password"
            type="password"
            required
            autocomplete="current-password"
          />
        </div>
        <div v-if="error" class="error-message">{{ error }}</div>
        <button type="submit" :disabled="isLoading">
          {{ isLoading ? 'Logging in...' : 'Login' }}
        </button>
      </form>
      <p class="auth-link">
        Don't have an account? <router-link to="/register">Register</router-link>
      </p>
    </div>
  </div>
</template>

<style scoped>
.auth-page {
  display: flex;
  justify-content: center;
  align-items: center;
  min-height: 100vh;
  padding: 1rem;
}

.auth-card {
  width: 100%;
  max-width: 400px;
  padding: 2rem;
  border: 1px solid #e5e7eb;
  border-radius: 8px;
  background: #fff;
}

h1 {
  margin: 0 0 1.5rem;
  font-size: 1.5rem;
  font-weight: 600;
}

.form-group {
  margin-bottom: 1rem;
}

label {
  display: block;
  margin-bottom: 0.5rem;
  font-size: 0.875rem;
  font-weight: 500;
}

input {
  width: 100%;
  padding: 0.625rem;
  border: 1px solid #d1d5db;
  border-radius: 4px;
  font-size: 1rem;
  box-sizing: border-box;
}

input:focus {
  outline: none;
  border-color: #3b82f6;
  box-shadow: 0 0 0 2px rgba(59, 130, 246, 0.2);
}

button {
  width: 100%;
  padding: 0.75rem;
  margin-top: 0.5rem;
  border: none;
  border-radius: 4px;
  background: #3b82f6;
  color: #fff;
  font-size: 1rem;
  font-weight: 500;
  cursor: pointer;
  transition: background 0.2s;
}

button:hover:not(:disabled) {
  background: #2563eb;
}

button:disabled {
  opacity: 0.6;
  cursor: not-allowed;
}

.error-message {
  margin: 0.5rem 0;
  padding: 0.5rem;
  border-radius: 4px;
  background: #fef2f2;
  color: #dc2626;
  font-size: 0.875rem;
}

.auth-link {
  margin-top: 1rem;
  font-size: 0.875rem;
  text-align: center;
}

.auth-link a {
  color: #3b82f6;
  text-decoration: none;
}

.auth-link a:hover {
  text-decoration: underline;
}
</style>
