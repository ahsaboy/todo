<script setup lang="ts">
import { ref } from 'vue'
import { adminApi } from '@/shared/api/admin-client'
import { isoToDateTimeLocal, dateTimeLocalToISOString } from '@/shared/utils/date'
import DateTimePicker from '@/shared/ui/DateTimePicker.vue'
import BaseSelect, { type SelectOption } from '@/shared/ui/BaseSelect.vue'
import { useCrudList } from '@/shared/composables/useCrudList'
import { useFormState } from '@/shared/composables/useFormState'
import PagePagination from '@/shared/ui/PagePagination.vue'
import DataTable from '@/shared/ui/DataTable.vue'
import BaseDialog from '@/shared/ui/BaseDialog.vue'
import type { DataTableConfig } from '@/shared/ui/data-table/types'

interface Task {
  id: number
  user_id: number
  username: string
  title: string
  description: string
  priority: number
  completed: boolean
  due_at: string | null
  focus_duration: number | null
  created_at: string
  updated_at: string
}

const priorityText: Record<number, string> = { 1: '高', 2: '中', 3: '低' }

const list = useCrudList<Task>({
  client: adminApi,
  buildEndpoint: ({ page, limit, filters }) => {
    const params = new URLSearchParams({ page: String(page), limit: String(limit) })
    if (filters.user_id) params.set('user_id', filters.user_id)
    if (filters.status) params.set('status', filters.status)
    if (filters.priority) params.set('priority', filters.priority)
    return `/tasks?${params}`
  },
  errorPrefix: '加载任务列表',
})

// 编辑弹窗
const editVisible = ref(false)
const editTask = ref<Task | null>(null)

const editForm = useFormState({
  initialData: { title: '', description: '', priority: 2, due_at: '' },
  validate: (d) => d.title.trim() ? null : '标题不能为空',
  onSubmit: async (data) => {
    if (!editTask.value) return
    const body: Record<string, unknown> = {
      title: data.title,
      description: data.description || null,
      priority: data.priority,
      due_at: data.due_at ? dateTimeLocalToISOString(data.due_at) : null,
    }
    await adminApi.put(`/tasks/${editTask.value.id}`, body)
    editVisible.value = false
    await list.load()
  },
})

const priorityEditOptions: SelectOption<number>[] = [
  { label: '高', value: 1 },
  { label: '中', value: 2 },
  { label: '低', value: 3 },
]

function openEdit(task: Task) {
  editTask.value = task
  editForm.resetTo({
    title: task.title,
    description: task.description || '',
    priority: task.priority,
    due_at: task.due_at ? isoToDateTimeLocal(task.due_at) : '',
  })
  editVisible.value = true
}

function toggleComplete(row: Task) {
  list.mutate(`/tasks/${row.id}/toggle`, 'PATCH')
}

function deleteTask(row: Task) {
  list.deleteItem(`/tasks/${row.id}`, `确定强制删除任务 "${row.title}"？`)
}

const config: DataTableConfig<Task> = {
  columns: [
    { key: 'id', label: 'ID', width: '60px' },
    { key: 'username', label: '用户', formatter: (_, row) => row.username || `用户#${row.user_id}` },
    { key: 'title', label: '标题' },
    { key: 'priority', label: '优先级', formatter: (v) => priorityText[v] || String(v) },
    {
      key: 'completed', label: '状态',
      cellClass: (row) => row.completed ? 'badge badge-done' : 'badge badge-pending',
      formatter: (v) => v ? '已完成' : '待办',
    },
    { key: 'due_at', label: '截止时间', formatter: (v) => v || '—' },
    { key: 'focus_duration', label: '专注', formatter: (v) => v ? `${v} min` : '—' },
  ],
  actions: [
    { id: 'toggle', label: (row) => row.completed ? '标记待办' : '标记完成', onClick: toggleComplete },
    { id: 'edit', label: '编辑', onClick: openEdit },
    { id: 'delete', label: '删除', variant: 'danger', onClick: deleteTask },
  ],
  filters: [
    { id: 'user_id', type: 'number', placeholder: '用户 ID 筛选', value: list.filters.value.user_id ?? '', width: 'narrow' },
    {
      id: 'status', type: 'select', value: list.filters.value.status ?? '',
      options: [
        { label: '全部状态', value: '' },
        { label: '未完成', value: 'pending' },
        { label: '已完成', value: 'completed' },
      ],
    },
    {
      id: 'priority', type: 'select', value: list.filters.value.priority ?? '',
      options: [
        { label: '全部优先级', value: '' },
        { label: '高', value: '1' },
        { label: '中', value: '2' },
        { label: '低', value: '3' },
      ],
    },
  ],
  emptyText: '暂无任务',
  mobileCard: {
    titleKey: 'title',
    subtitleKey: 'username',
    metaKeys: ['priority', 'completed', 'due_at', 'focus_duration'],
  },
}
</script>

<template>
  <div class="page-container">
    <h1 class="page-title">任务管理</h1>
    <div v-if="list.error.value" class="error-message">{{ list.error.value }}</div>
    <DataTable
      :config="config"
      :data="list.items.value"
      :is-loading="list.isLoading.value"
      @filter-change="list.handleFilterChange"
      @apply-filters="list.applyFilters()"
    />
    <PagePagination
      :page="list.page.value"
      :total="list.total.value"
      :total-pages="list.totalPages.value"
      @update:page="list.setPage"
    />

    <!-- 编辑任务弹窗 -->
    <BaseDialog v-model:visible="editVisible" :title="`编辑任务 #${editTask?.id}`" max-width="500px">
      <div class="form-group">
        <label>标题</label>
        <input v-model="editForm.form.title" type="text" class="form-input" maxlength="255" />
      </div>
      <div class="form-group">
        <label>描述</label>
        <textarea v-model="editForm.form.description" class="form-input" rows="3" maxlength="1000"></textarea>
      </div>
      <div class="form-group">
        <label>优先级</label>
        <BaseSelect v-model="editForm.form.priority" :options="priorityEditOptions" block aria-label="优先级" />
      </div>
      <div class="form-group">
        <label>截止时间</label>
        <DateTimePicker
          :model-value="editForm.form.due_at"
          placeholder="选择截止日期"
          default-time="23:59"
          @update:model-value="editForm.form.due_at = $event || ''"
        />
      </div>
      <div v-if="editForm.error.value" class="error-message">{{ editForm.error.value }}</div>
      <template #footer="{ close }">
        <button class="btn" :disabled="editForm.submitting.value" @click="close">取消</button>
        <button class="btn btn-primary" :disabled="editForm.submitting.value" @click="editForm.handleSubmit">
          {{ editForm.submitting.value ? '保存中...' : '保存' }}
        </button>
      </template>
    </BaseDialog>
  </div>
</template>

