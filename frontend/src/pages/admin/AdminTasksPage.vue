<script setup lang="ts">
import { ref, onMounted, watch } from 'vue'
import { VueDatePicker } from '@vuepic/vue-datepicker'
import { adminApi } from '@/shared/api/admin-client'
import { useThemeStore } from '@/app/stores/theme.store'
import { isoToDateTimeLocal, dateTimeLocalToISOString } from '@/shared/utils/date'
import type { PaginatedResponse } from '@/shared/api/types'

const themeStore = useThemeStore()

interface Task {
  id: number
  user_id: number
  username: string
  title: string
  description: string
  priority: number
  completed: boolean
  due_at: string | null
  remind_at: string | null
  repeat_type: string
  repeat_interval: number
  repeat_end_date: string | null
  reminder_sent: boolean
  reminder_sent_at: string | null
  focus_duration: number | null
  created_at: string
  updated_at: string
}

const tasks = ref<Task[]>([])
const total = ref(0)
const page = ref(1)
const limit = 20
const filterUserId = ref('')
const filterStatus = ref('')
const filterPriority = ref('')
const error = ref('')
const isLoading = ref(false)

const editDialog = ref<{ show: boolean; task: Task | null; saving: boolean; err: string }>({
  show: false, task: null, saving: false, err: ''
})
const editForm = ref({ title: '', description: '', priority: 2, due_at: '' })

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
      ...(filterPriority.value ? { priority: filterPriority.value } : {}),
    })
    const res = await adminApi.get<PaginatedResponse<Task>>(`/tasks?${params}`)
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

async function toggleComplete(task: Task) {
  try {
    await adminApi.patch(`/tasks/${task.id}/toggle`)
    await loadTasks()
  } catch {
    error.value = '切换任务状态失败'
  }
}

function openEdit(task: Task) {
  editForm.value = {
    title: task.title,
    description: task.description || '',
    priority: task.priority,
    due_at: task.due_at ? isoToDateTimeLocal(task.due_at) : '',
  }
  editDialog.value = { show: true, task, saving: false, err: '' }
}

async function saveEdit() {
  if (!editDialog.value.task) return
  if (!editForm.value.title.trim()) {
    editDialog.value.err = '标题不能为空'
    return
  }
  editDialog.value.saving = true
  editDialog.value.err = ''
  try {
    const body: Record<string, unknown> = {
      title: editForm.value.title,
      description: editForm.value.description || null,
      priority: editForm.value.priority,
    }
    if (editForm.value.due_at) {
      body.due_at = dateTimeLocalToISOString(editForm.value.due_at)
    } else {
      body.due_at = null
    }
    await adminApi.put(`/tasks/${editDialog.value.task.id}`, body)
    editDialog.value.show = false
    await loadTasks()
  } catch {
    editDialog.value.err = '保存失败'
  } finally {
    editDialog.value.saving = false
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
      <select v-model="filterPriority" class="admin-search-input" style="max-width: 120px;">
        <option value="">全部优先级</option>
        <option value="1">高</option>
        <option value="2">中</option>
        <option value="3">低</option>
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
            <td class="action-cell">
              <button class="btn btn-sm" @click="toggleComplete(t)">
                {{ t.completed ? '标记待办' : '标记完成' }}
              </button>
              <button class="btn btn-sm" @click="openEdit(t)">编辑</button>
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

    <!-- 编辑任务弹窗 -->
    <div v-if="editDialog.show" class="admin-modal-overlay" @click.self="editDialog.show = false">
      <div class="admin-modal">
        <h3>编辑任务 #{{ editDialog.task?.id }}</h3>
        <div class="form-group">
          <label>标题</label>
          <input v-model="editForm.title" type="text" class="form-input" maxlength="255" />
        </div>
        <div class="form-group">
          <label>描述</label>
          <textarea v-model="editForm.description" class="form-input" rows="3" maxlength="1000"></textarea>
        </div>
        <div class="form-group">
          <label>优先级</label>
          <select v-model.number="editForm.priority" class="form-input">
            <option :value="1">高</option>
            <option :value="2">中</option>
            <option :value="3">低</option>
          </select>
        </div>
        <div class="form-group">
          <label>截止时间</label>
          <VueDatePicker
            v-model="editForm.due_at"
            :dark="themeStore.isDark"
            model-type="format"
            format="yyyy-MM-dd HH:mm"
            locale="zh-CN"
            auto-apply
            clearable
            time-picker-inline
            placeholder="选择截止时间"
            teleport
            :config="{ allowPreventDefault: true }"
          />
        </div>
        <div v-if="editDialog.err" class="error-message">{{ editDialog.err }}</div>
        <div class="modal-actions">
          <button class="btn" @click="editDialog.show = false" :disabled="editDialog.saving">取消</button>
          <button class="btn btn-primary" @click="saveEdit" :disabled="editDialog.saving">
            {{ editDialog.saving ? '保存中...' : '保存' }}
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
  min-width: 400px;
  max-width: 500px;
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

.form-group label {
  font-size: 0.85rem;
  font-weight: 600;
  color: var(--color-text-muted);
}

.form-input {
  padding: 0.4rem 0.7rem;
  border: 1px solid var(--color-border);
  border-radius: 5px;
  background: var(--color-surface);
  color: var(--color-text);
  font-family: inherit;
}

textarea.form-input {
  resize: vertical;
}

.modal-actions {
  display: flex;
  justify-content: flex-end;
  gap: 0.5rem;
}
</style>
