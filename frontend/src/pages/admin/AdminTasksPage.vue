<script setup lang="ts">
import { ref, onMounted, watch } from 'vue'
import { adminApi } from '@/shared/api/admin-client'
import type { PaginatedResponse } from '@/shared/api/types'

interface Task {
  id: number
  user_id: number
  username: string
  title: string
  priority: number
  completed: boolean
  due_at: string | null
  focus_duration: number | null
  created_at: string
}

const tasks = ref<Task[]>([])
const total = ref(0)
const page = ref(1)
const limit = 20
const filterUserId = ref('')
const filterStatus = ref('')
const error = ref('')
const isLoading = ref(false)

const priorityText: Record<number, string> = { 1: '高', 2: '中', 3: '低' }

async function loadTasks() {
  isLoading.value = true
  error.value = ''
  try {
    const params = new URLSearchParams({
      page: String(page.value),
      limit: String(limit),
      ...(filterUserId.value ? { user_id: filterUserId.value } : {}),
      ...(filterStatus.value ? { status: filterStatus.value } : {}),
    })
    const res = await adminApi.get<PaginatedResponse<Task[]>>(`/tasks?${params}`)
    tasks.value = res.data
    total.value = res.meta.total_items
  } catch {
    error.value = '加载任务列表失败'
  } finally {
    isLoading.value = false
  }
}

onMounted(loadTasks)
watch(page, loadTasks)

function handleFilter() {
  page.value = 1
  loadTasks()
}

async function deleteTask(id: number, title: string) {
  if (!confirm(`确定强制删除任务 "${title}"？`)) return
  try {
    await adminApi.delete(`/tasks/${id}`)
    await loadTasks()
  } catch {
    error.value = '删除任务失败'
  }
}

const totalPages = () => Math.ceil(total.value / limit)
</script>

<template>
  <div class="page-container">
    <h1 class="page-title">任务管理</h1>

    <div class="admin-toolbar">
      <input
        v-model="filterUserId"
        type="number"
        placeholder="用户 ID 筛选"
        class="admin-search-input"
        style="max-width: 140px;"
      />
      <select v-model="filterStatus" class="admin-search-input" style="max-width: 120px;">
        <option value="">全部状态</option>
        <option value="pending">未完成</option>
        <option value="completed">已完成</option>
      </select>
      <button class="btn btn-primary" @click="handleFilter">筛选</button>
    </div>

    <div v-if="error" class="error-message">{{ error }}</div>

    <div class="admin-table-wrap">
      <table class="admin-table">
        <thead>
          <tr>
            <th>ID</th>
            <th>用户</th>
            <th>标题</th>
            <th>优先级</th>
            <th>状态</th>
            <th>截止时间</th>
            <th>专注</th>
            <th>操作</th>
          </tr>
        </thead>
        <tbody>
          <tr v-if="isLoading">
            <td colspan="8" style="text-align:center; padding: 2rem;">加载中...</td>
          </tr>
          <tr v-else-if="!tasks.length">
            <td colspan="8" style="text-align:center; padding: 2rem; color: var(--color-text-muted);">暂无任务</td>
          </tr>
          <tr v-for="t in tasks" :key="t.id">
            <td>{{ t.id }}</td>
            <td>{{ t.username || '用户#' + t.user_id }}</td>
            <td>{{ t.title }}</td>
            <td>{{ priorityText[t.priority] || t.priority }}</td>
            <td>
              <span :class="t.completed ? 'badge badge-done' : 'badge badge-pending'">
                {{ t.completed ? '已完成' : '待办' }}
              </span>
            </td>
            <td>{{ t.due_at || '—' }}</td>
            <td>{{ t.focus_duration ? t.focus_duration + ' min' : '—' }}</td>
            <td>
              <button class="btn btn-sm btn-danger" @click="deleteTask(t.id, t.title)">删除</button>
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
  </div>
</template>

<style scoped>
@import '@/widgets/admin-common.css';
</style>
