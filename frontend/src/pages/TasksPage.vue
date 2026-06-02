<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { Plus, RefreshCw, SlidersHorizontal } from 'lucide-vue-next'
import { useTasks } from '@/features/tasks/useTasks'
import { useTaskToggle } from '@/features/tasks/useTaskToggle'
import { useMediaQuery } from '@/shared/composables/useMediaQuery'
import PageShell from '@/shared/ui/PageShell.vue'
import PagePagination from '@/shared/ui/PagePagination.vue'
import MobileFab from '@/shared/ui/MobileFab.vue'
import TaskTable from '@/features/tasks/TaskTable.vue'
import TaskCardList from '@/features/tasks/TaskCardList.vue'
import TaskFilters from '@/features/tasks/TaskFilters.vue'
import MobileFilters from '@/features/tasks/MobileFilters.vue'
import TaskDetailDrawer from '@/features/tasks/TaskDetailDrawer.vue'
import FocusDurationDialog from '@/features/tasks/FocusDurationDialog.vue'
import TaskListSkeleton from '@/shared/ui/TaskListSkeleton.vue'
import type { Task, CreateTaskPayload, UpdateTaskPayload } from '@/entities/task/model'

const {
  tasks, loading, error, meta, filters,
  fetchTasks, createTask, updateTask, toggleComplete, deleteTask,
  setPage, setFilters, setSort,
} = useTasks()

const { focusDialogVisible, focusDialogTaskTitle, handleToggle, handleFocusConfirm } =
  useTaskToggle({ tasks, toggleComplete })

const drawerVisible = ref(false)
const selectedTask = ref<Task | null>(null)
const showFilters = ref(false)
const isMobile = useMediaQuery('(max-width: 767px)')

onMounted(() => { setSort('task_center', 'asc'); fetchTasks() })

function openCreate() { selectedTask.value = null; drawerVisible.value = true }
function openTask(task: Task) { selectedTask.value = task; drawerVisible.value = true }

async function handleSubmit(payload: CreateTaskPayload | UpdateTaskPayload) {
  if (selectedTask.value) { await updateTask(selectedTask.value.id, payload) }
  else { await createTask(payload) }
}

async function handleDelete(id: number) {
  if (window.confirm('确定要删除这个任务吗？')) { await deleteTask(id); drawerVisible.value = false }
}
</script>

<template>
  <div class="page">
    <div class="page-header">
      <h2>全部任务</h2>
      <div class="header-actions">
        <button class="btn-secondary" type="button" @click="fetchTasks"><RefreshCw :size="14" /></button>
        <button class="btn-primary desktop-only" @click="openCreate">新建任务</button>
        <button class="mobile-filter-btn mobile-only" @click="showFilters = true"><SlidersHorizontal :size="16" /> 筛选</button>
      </div>
    </div>

    <div class="page-toolbar">
      <TaskFilters :filters="filters" @change="setFilters" />
    </div>

    <PageShell
      :loading="loading"
      :error="error"
      :empty="tasks.length === 0 && !filters.search"
      :skeleton="TaskListSkeleton"
      empty-title="暂无任务"
      :empty-action="{ label: '创建第一个任务', onClick: openCreate }"
      :error-retry="fetchTasks"
    >
      <div v-if="tasks.length === 0 && filters.search" class="page-empty search-empty">
        <p>未找到匹配的任务</p>
        <button @click="setFilters({ search: '' })">清除搜索</button>
      </div>
      <template v-else>
        <TaskCardList v-if="isMobile" :tasks="tasks" @toggle="handleToggle" @open="openTask" />
        <TaskTable v-else :tasks="tasks" @toggle="handleToggle" @open="openTask" />
        <PagePagination
          v-if="meta.total_pages > 1"
          :page="meta.page"
          :total="meta.total_items"
          :total-pages="meta.total_pages"
          @update:page="setPage"
        />
      </template>
    </PageShell>

    <MobileFab label="新建任务" @click="openCreate"><Plus :size="24" /></MobileFab>

    <MobileFilters v-model:visible="showFilters" :filters="filters" @change="setFilters" />
    <TaskDetailDrawer v-model:visible="drawerVisible" :task="selectedTask" @submit="handleSubmit" @delete="handleDelete" />
    <FocusDurationDialog v-model:visible="focusDialogVisible" :task-title="focusDialogTaskTitle" @confirm="handleFocusConfirm" />
  </div>
</template>

<style scoped>
.header-actions {
  display: flex;
  gap: 8px;
  flex-shrink: 0;
}

.page-toolbar {
  display: flex;
  justify-content: space-between;
  align-items: center;
  flex-wrap: wrap;
  gap: 16px;
}

.mobile-filter-btn {
  padding: 8px 16px;
  background: var(--color-surface);
  border: 1px solid var(--color-border);
  border-radius: 6px;
  cursor: pointer;
  font-size: 14px;
  color: var(--color-text);
}

.desktop-only { display: block; }
.mobile-only { display: none; }

@media (max-width: 767px) {
  .desktop-only { display: none; }
  .mobile-only { display: block; }
  .page-toolbar { display: none; }
}

@media (max-width: 359px) {
  .header-actions { width: 100%; justify-content: flex-end; }
}
</style>
