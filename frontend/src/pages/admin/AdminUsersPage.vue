<script setup lang="ts">
import { ref } from 'vue'
import { adminApi } from '@/shared/api/admin-client'
import { useCrudList } from '@/shared/composables/useCrudList'
import { useFormState } from '@/shared/composables/useFormState'
import PagePagination from '@/shared/ui/PagePagination.vue'
import DataTable from '@/shared/ui/DataTable.vue'
import BaseDialog from '@/shared/ui/BaseDialog.vue'
import type { DataTableConfig } from '@/shared/ui/data-table/types'
import type { ApiResponse } from '@/shared/api/types'

interface User {
  id: number
  username: string
  email: string
  created_at: string
  is_admin: boolean
}

interface UserDetail {
  id: number
  username: string
  email: string
  created_at: string
  is_admin: boolean
  task_count: number
  api_key_count: number
}

const list = useCrudList<User>({
  client: adminApi,
  buildEndpoint: ({ page, limit, filters }) =>
    `/users?page=${page}&limit=${limit}&search=${encodeURIComponent(filters.search || '')}`,
  errorPrefix: '加载用户列表',
})

// 重置密码弹窗
const resetVisible = ref(false)
const resetUserId = ref(0)

const resetForm = useFormState({
  initialData: { password: '' },
  validate: (d) => {
    if (d.password.length < 6) return '密码至少 6 位'
    if (d.password.length > 72) return '密码不能超过 72 位'
    return null
  },
  onSubmit: async (data) => {
    await adminApi.post(`/users/${resetUserId.value}/reset-password`, { new_password: data.password })
    resetVisible.value = false
  },
})

// 用户详情弹窗
const detailVisible = ref(false)
const detailLoading = ref(false)
const detailUser = ref<UserDetail | null>(null)

function openReset(id: number) {
  resetUserId.value = id
  resetForm.resetTo({ password: '' })
  resetVisible.value = true
}

function toggleAdmin(row: User) {
  const action = row.is_admin ? '取消管理员权限' : '设为管理员'
  if (!confirm(`确定将用户 "${row.username}" ${action}？`)) return
  list.mutate(`/users/${row.id}/admin`, 'PATCH', { is_admin: !row.is_admin })
}

async function openDetail(userId: number) {
  detailVisible.value = true
  detailLoading.value = true
  detailUser.value = null
  try {
    const res = await adminApi.get<ApiResponse<UserDetail>>(`/users/${userId}`)
    detailUser.value = res.data
  } catch { detailVisible.value = false; list.error.value = '加载用户详情失败' }
  finally { detailLoading.value = false }
}

const config: DataTableConfig<User> = {
  columns: [
    { key: 'id', label: 'ID', width: '60px' },
    { key: 'username', label: '用户名', cellClass: 'username-link' },
    { key: 'email', label: '邮箱', formatter: (v) => v || '—' },
    {
      key: 'is_admin', label: '管理员',
      cellClass: (row) => row.is_admin ? 'badge badge-done' : 'badge badge-muted',
      formatter: (v) => v ? '管理员' : '普通用户',
    },
    { key: 'created_at', label: '注册时间' },
  ],
  actions: [
    { id: 'reset', label: '重置密码', onClick: (row) => openReset(row.id) },
    {
      id: 'toggle', label: (row) => row.is_admin ? '取消管理员' : '设为管理员',
      onClick: toggleAdmin,
    },
    {
      id: 'delete', label: '删除', variant: 'danger',
      onClick: (row) => list.deleteItem(`/users/${row.id}`, `确定删除用户 "${row.username}"？该用户的所有任务、API Key、提醒配置将一并删除，此操作不可恢复！`),
    },
  ],
  filters: [
    { id: 'search', type: 'text', placeholder: '搜索用户名或邮箱...', value: list.filters.value.search ?? '' },
  ],
  filterButtonText: '搜索',
  emptyText: '暂无用户',
  mobileCard: {
    titleKey: 'username',
    subtitleKey: 'email',
    metaKeys: ['is_admin', 'created_at'],
  },
}

// 用户名点击打开详情（通过 DataTable 的 cellClass + 自定义事件处理）
function handleTableClick(e: MouseEvent) {
  const target = e.target as HTMLElement
  if (target.classList.contains('username-link')) {
    const row = (target.closest('tr') as HTMLElement)?.dataset?.idx
    if (row != null) {
      const user = list.items.value[Number(row)]
      if (user) openDetail(user.id)
    }
  }
}
</script>

<template>
  <div class="page-container">
    <h1 class="page-title">用户管理</h1>
    <div v-if="list.error.value" class="error-message">{{ list.error.value }}</div>
    <div @click="handleTableClick">
      <DataTable
        :config="config"
        :data="list.items.value"
        :is-loading="list.isLoading.value"
        @filter-change="list.handleFilterChange"
        @apply-filters="list.applyFilters()"
      />
    </div>
    <PagePagination
      :page="list.page.value"
      :total="list.total.value"
      :total-pages="list.totalPages.value"
      @update:page="list.setPage"
    />

    <!-- 重置密码弹窗 -->
    <BaseDialog v-model:visible="resetVisible" title="重置密码">
      <div class="form-group">
        <label>新密码（至少 6 位）</label>
        <input v-model="resetForm.form.password" type="password" class="form-input" />
      </div>
      <div v-if="resetForm.error.value" class="error-message">{{ resetForm.error.value }}</div>
      <template #footer="{ close }">
        <button class="btn" :disabled="resetForm.submitting.value" @click="close">取消</button>
        <button class="btn btn-primary" :disabled="resetForm.submitting.value" @click="resetForm.handleSubmit">
          {{ resetForm.submitting.value ? '重置中...' : '确认重置' }}
        </button>
      </template>
    </BaseDialog>

    <!-- 用户详情弹窗 -->
    <BaseDialog v-model:visible="detailVisible" title="用户详情">
      <div v-if="detailLoading" class="table-cell-status">加载中...</div>
      <template v-else-if="detailUser">
        <div class="detail-grid">
          <div class="detail-label">ID</div><div>{{ detailUser.id }}</div>
          <div class="detail-label">用户名</div><div>{{ detailUser.username }}</div>
          <div class="detail-label">邮箱</div><div>{{ detailUser.email || '—' }}</div>
          <div class="detail-label">管理员</div>
          <div><span :class="detailUser.is_admin ? 'badge badge-done' : 'badge badge-muted'">{{ detailUser.is_admin ? '是' : '否' }}</span></div>
          <div class="detail-label">任务数</div><div>{{ detailUser.task_count }}</div>
          <div class="detail-label">API Key 数</div><div>{{ detailUser.api_key_count }}</div>
          <div class="detail-label">注册时间</div><div>{{ detailUser.created_at }}</div>
        </div>
      </template>
      <template #footer="{ close }">
        <button class="btn" @click="close">关闭</button>
      </template>
    </BaseDialog>
  </div>
</template>

<style scoped>
.username-link {
  color: var(--color-primary);
  cursor: pointer;
  text-decoration: underline;
  text-decoration-style: dotted;
  text-underline-offset: 2px;
}

.username-link:hover {
  text-decoration-style: solid;
}

.detail-grid {
  display: grid;
  grid-template-columns: auto 1fr;
  gap: 0.5rem 1rem;
  align-items: center;
}

.detail-label {
  font-weight: 600;
  color: var(--color-text-muted);
  font-size: 0.85rem;
}
</style>
