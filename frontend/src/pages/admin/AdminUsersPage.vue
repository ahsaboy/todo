<script setup lang="ts">
import { ref, onMounted, watch } from 'vue'
import { adminApi } from '@/shared/api/admin-client'
import type { PaginatedResponse } from '@/shared/api/types'

interface User {
  id: number
  username: string
  email: string
  created_at: string
}

const users = ref<User[]>([])
const total = ref(0)
const page = ref(1)
const limit = 20
const search = ref('')
const error = ref('')
const isLoading = ref(false)

const resetDialog = ref<{ show: boolean; userId: number; newPwd: string }>({ show: false, userId: 0, newPwd: '' })
const resetError = ref('')
const isResetting = ref(false)

async function loadUsers() {
  isLoading.value = true
  error.value = ''
  try {
    const res = await adminApi.get<PaginatedResponse<User[]>>(
      `/users?page=${page.value}&limit=${limit}&search=${encodeURIComponent(search.value)}`
    )
    users.value = res.data
    total.value = res.meta.total_items
  } catch {
    error.value = '加载用户列表失败'
  } finally {
    isLoading.value = false
  }
}

onMounted(loadUsers)
watch(page, loadUsers)

function handleSearch() {
  page.value = 1
  loadUsers()
}

async function deleteUser(id: number, username: string) {
  if (!confirm(`确定删除用户 "${username}"？该用户的所有任务、API Key、提醒配置将一并删除，此操作不可恢复！`)) return
  try {
    await adminApi.delete(`/users/${id}`)
    await loadUsers()
  } catch {
    error.value = '删除用户失败'
  }
}

function openReset(id: number) {
  resetDialog.value = { show: true, userId: id, newPwd: '' }
  resetError.value = ''
}

async function confirmReset() {
  resetError.value = ''
  if (resetDialog.value.newPwd.length < 6) {
    resetError.value = '密码至少 6 位'
    return
  }
  if (resetDialog.value.newPwd.length > 72) {
    resetError.value = '密码不能超过 72 位'
    return
  }
  isResetting.value = true
  try {
    await adminApi.post(`/users/${resetDialog.value.userId}/reset-password`, { new_password: resetDialog.value.newPwd })
    resetDialog.value.show = false
  } catch {
    resetError.value = '重置密码失败'
  } finally {
    isResetting.value = false
  }
}

const totalPages = () => Math.ceil(total.value / limit)
</script>

<template>
  <div class="page-container">
    <h1 class="page-title">用户管理</h1>

    <div class="admin-toolbar">
      <input
        v-model="search"
        type="text"
        placeholder="搜索用户名或邮箱..."
        class="admin-search-input"
        @keyup.enter="handleSearch"
      />
      <button class="btn btn-primary" @click="handleSearch">搜索</button>
    </div>

    <div v-if="error" class="error-message">{{ error }}</div>

    <div class="admin-table-wrap">
      <table class="admin-table">
        <thead>
          <tr>
            <th>ID</th>
            <th>用户名</th>
            <th>邮箱</th>
            <th>注册时间</th>
            <th>操作</th>
          </tr>
        </thead>
        <tbody>
          <tr v-if="isLoading">
            <td colspan="5" style="text-align:center; padding: 2rem;">加载中...</td>
          </tr>
          <tr v-else-if="!users.length">
            <td colspan="5" style="text-align:center; padding: 2rem; color: var(--color-text-muted);">暂无用户</td>
          </tr>
          <tr v-for="u in users" :key="u.id">
            <td>{{ u.id }}</td>
            <td>{{ u.username }}</td>
            <td>{{ u.email || '—' }}</td>
            <td>{{ u.created_at }}</td>
            <td class="action-cell">
              <button class="btn btn-sm" @click="openReset(u.id)">重置密码</button>
              <button class="btn btn-sm btn-danger" @click="deleteUser(u.id, u.username)">删除</button>
            </td>
          </tr>
        </tbody>
      </table>
    </div>

    <div class="admin-pagination">
      <span>共 {{ total }} 条</span>
      <button :disabled="page <= 1" class="btn btn-sm" @click="page--">上一页</button>
      <span>{{ page }} / {{ totalPages() }}</span>
      <button :disabled="page >= totalPages()" class="btn btn-sm" @click="page++">下一页</button>
    </div>

    <!-- 重置密码弹窗 -->
    <div v-if="resetDialog.show" class="admin-modal-overlay" @click.self="resetDialog.show = false">
      <div class="admin-modal">
        <h3>重置密码</h3>
        <div class="form-group">
          <label>新密码（至少 6 位）</label>
          <input v-model="resetDialog.newPwd" type="password" class="form-input" />
        </div>
        <div v-if="resetError" class="error-message">{{ resetError }}</div>
        <div class="modal-actions">
          <button class="btn" @click="resetDialog.show = false" :disabled="isResetting">取消</button>
          <button class="btn btn-primary" @click="confirmReset" :disabled="isResetting">
            {{ isResetting ? '重置中...' : '确认重置' }}
          </button>
        </div>
      </div>
    </div>
  </div>
</template>

<style scoped>
@import '@/widgets/admin-common.css';

.admin-modal-overlay {
  position: fixed;
  inset: 0;
  background: rgba(0, 0, 0, 0.45);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 1000;
}

.admin-modal {
  background: var(--color-bg);
  border: 1px solid var(--color-border);
  border-radius: 10px;
  padding: 1.5rem;
  min-width: 320px;
  display: flex;
  flex-direction: column;
  gap: 1rem;
}

.admin-modal h3 {
  margin: 0;
  font-size: 1.1rem;
}

.form-group {
  display: flex;
  flex-direction: column;
  gap: 0.35rem;
}

.form-input {
  padding: 0.4rem 0.7rem;
  border: 1px solid var(--color-border);
  border-radius: 5px;
  background: var(--color-surface);
  color: var(--color-text);
}

.modal-actions {
  display: flex;
  justify-content: flex-end;
  gap: 0.5rem;
}
</style>
