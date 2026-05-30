<template>
  <div class="page">
    <div class="page-header">
      <h2>全部任务</h2>
      <div class="header-actions">
        <button class="btn-secondary" type="button" @click="fetchTasks"><RefreshCw :size="14" /></button>
        <!-- 桌面端按钮 -->
        <button class="btn-primary desktop-only" @click="openCreate">新建任务</button>
        <!-- 移动端筛选按钮 -->
        <button class="mobile-filter-btn mobile-only" @click="showFilters = true"><SlidersHorizontal :size="16" /> 筛选</button>
      </div>
    </div>

    <div class="page-toolbar">
      <TaskFilters :filters="filters" @change="setFilters" />
    </div>

    <Transition name="sk-fade" mode="out-in">
      <TaskListSkeleton v-if="loading" key="skeleton" />

      <template v-else key="content">
        <div v-if="error" class="page-error">
          <p>{{ error }}</p>
          <button @click="fetchTasks">重试</button>
        </div>

        <div v-else-if="tasks.length === 0 && filters.search" class="page-empty search-empty">
          <p>未找到匹配的任务</p>
          <button @click="setFilters({ search: '' })">清除搜索</button>
        </div>

        <div v-else-if="tasks.length === 0" class="page-empty">
          <p>暂无任务</p>
          <button class="btn-primary" @click="openCreate">创建第一个任务</button>
        </div>

        <template v-else>
          <!-- 移动端卡片列表 -->
          <TaskCardList
            v-if="isMobile"
            :tasks="tasks"
            @toggle="handleToggle"
            @open="openTask"
          />

          <!-- 桌面端表格 -->
          <TaskTable
            v-else
            :tasks="tasks"
            @toggle="handleToggle"
            @open="openTask"
          />

          <div class="pagination">
            <button :disabled="meta.page <= 1" @click="setPage(meta.page - 1)">上一页</button>
            <span>{{ meta.page }} / {{ meta.total_pages }}</span>
            <button :disabled="meta.page >= meta.total_pages" @click="setPage(meta.page + 1)">
              下一页
            </button>
          </div>
        </template>
      </template>
    </Transition>

    <!-- 移动端浮动按钮 -->
    <button v-if="isMobile" class="fab" type="button" aria-label="新建任务" @click="openCreate"><Plus :size="24" /></button>

    <!-- 移动端筛选面板 -->
    <MobileFilters v-model:visible="showFilters" :filters="filters" @change="setFilters" />

    <!-- 抽屉 -->
    <TaskDetailDrawer v-model:visible="drawerVisible" :task="selectedTask" @submit="handleSubmit" @delete="handleDelete" />

    <!-- 专注时长对话框 -->
    <FocusDurationDialog
      v-model:visible="focusDialogVisible"
      :task-title="focusDialogTaskTitle"
      @confirm="handleFocusConfirm"
    />
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { Plus, RefreshCw, SlidersHorizontal } from 'lucide-vue-next'
import { useTasks } from '@/features/tasks/useTasks'
import { useMediaQuery } from '@/shared/composables/useMediaQuery'
import TaskTable from '@/features/tasks/TaskTable.vue'
import TaskCardList from '@/features/tasks/TaskCardList.vue'
import TaskFilters from '@/features/tasks/TaskFilters.vue'
import MobileFilters from '@/features/tasks/MobileFilters.vue'
import TaskDetailDrawer from '@/features/tasks/TaskDetailDrawer.vue'
import FocusDurationDialog from '@/features/tasks/FocusDurationDialog.vue'
import TaskListSkeleton from '@/shared/ui/TaskListSkeleton.vue'
import type { Task, CreateTaskPayload, UpdateTaskPayload } from '@/entities/task/model'
import { toggleTaskComplete as apiToggleComplete } from '@/entities/task/api'
import { toTask } from '@/entities/task/mapper'

const {
  tasks,
  loading,
  error,
  meta,
  filters,
  fetchTasks,
  createTask,
  updateTask,
  toggleComplete,
  deleteTask,
  setPage,
  setFilters,
  setSort,
} = useTasks()

const drawerVisible = ref(false)
const selectedTask = ref<Task | null>(null)
const showFilters = ref(false)
const isMobile = useMediaQuery('(max-width: 767px)')

const focusDialogVisible = ref(false)
const focusDialogTaskTitle = ref('')
const pendingToggleTaskId = ref<number | null>(null)

onMounted(() => {
  setSort('task_center', 'asc')
  fetchTasks()
})

function openCreate() {
  selectedTask.value = null
  drawerVisible.value = true
}

function openTask(task: Task) {
  selectedTask.value = task
  drawerVisible.value = true
}

async function handleSubmit(payload: CreateTaskPayload | UpdateTaskPayload) {
  if (selectedTask.value) {
    await updateTask(selectedTask.value.id, payload)
  } else {
    await createTask(payload)
  }
}

async function handleDelete(id: number) {
  if (window.confirm('确定要删除这个任务吗？')) {
    await deleteTask(id)
    drawerVisible.value = false
  }
}

function handleToggle(id: number) {
  const task = tasks.value.find((t) => t.id === id)
  if (!task) return

  if (!task.completed) {
    // 标记完成 → 弹出专注时长对话框
    pendingToggleTaskId.value = id
    focusDialogTaskTitle.value = task.title
    focusDialogVisible.value = true
  } else {
    // 取消完成 → 直接调用
    toggleComplete(id)
  }
}

async function handleFocusConfirm(duration: number | null) {
  if (pendingToggleTaskId.value == null) return
  const id = pendingToggleTaskId.value
  pendingToggleTaskId.value = null

  const task = tasks.value.find((t) => t.id === id)
  if (!task) return

  const originalCompleted = task.completed
  task.completed = !task.completed

  try {
    const response = await apiToggleComplete(id, duration != null ? duration : undefined)
    const updatedTask = toTask(response.data)
    const index = tasks.value.findIndex((t) => t.id === id)
    if (index !== -1) {
      tasks.value[index] = updatedTask
    }
  } catch {
    task.completed = originalCompleted
  }
}
</script>

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

.pagination {
  display: flex;
  justify-content: center;
  align-items: center;
  gap: 16px;
  padding: 16px 0;
  flex-wrap: wrap;
}

.pagination button {
  padding: 6px 12px;
  border: 1px solid var(--color-border);
  border-radius: 4px;
  background: var(--color-surface);
  cursor: pointer;
}

.pagination button:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

/* 移动端样式 */
.mobile-filter-btn {
  padding: 8px 16px;
  background: var(--color-surface);
  border: 1px solid var(--color-border);
  border-radius: 6px;
  cursor: pointer;
  font-size: 14px;
  color: var(--color-text);
}

.desktop-only {
  display: block;
}

.mobile-only {
  display: none;
}

.fab {
  display: none;
}

@media (max-width: 767px) {
  .desktop-only {
    display: none;
  }

  .mobile-only {
    display: block;
  }

  .fab {
    display: flex;
    position: fixed;
    right: 16px;
    bottom: calc(var(--bottom-nav-height) + 16px);
    width: 56px;
    height: 56px;
    background: var(--color-primary);
    color: white;
    border: none;
    border-radius: 50%;
    font-size: 24px;
    align-items: center;
    justify-content: center;
    box-shadow: var(--shadow-glow-primary);
    z-index: 50;
    cursor: pointer;
  }

  .page-toolbar {
    display: none;
  }

  .pagination {
      padding: 8px 72px 8px 0;
      display: flex;
      text-align: center;
  }

  .pagination span {
    order: -1;
    width: 100%;
    text-align: center;
  }
}

@media (max-width: 359px) {
  .header-actions {
    width: 100%;
    justify-content: flex-end;
  }

  .pagination {
    padding-right: 0;
  }
}
</style>
