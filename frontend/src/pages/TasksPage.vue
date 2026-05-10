<template>
  <div class="tasks-page">
    <div class="page-header">
      <h2>全部任务</h2>
      <div class="header-actions">
        <!-- 桌面端按钮 -->
        <button class="btn-primary desktop-only" @click="openCreate">新建任务</button>
        <!-- 移动端筛选按钮 -->
        <button class="mobile-filter-btn mobile-only" @click="showFilters = true"><SlidersHorizontal :size="16" /> 筛选</button>
      </div>
    </div>

    <div class="page-toolbar">
      <TaskFilters :filters="filters" @change="setFilters" />
      <TaskSortMenu :sort="sort" @change="setSort" />
    </div>

    <div v-if="loading" class="page-loading">加载中...</div>

    <div v-else-if="error" class="page-error">
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
        @toggle="toggleComplete"
        @open="openTask"
        @edit="editTask"
        @delete="handleDelete"
      />

      <!-- 桌面端表格 -->
      <TaskTable
        v-else
        :tasks="tasks"
        @toggle="toggleComplete"
        @open="openTask"
        @edit="editTask"
        @delete="handleDelete"
      />

      <div class="pagination">
        <button :disabled="meta.page <= 1" @click="setPage(meta.page - 1)">上一页</button>
        <span>{{ meta.page }} / {{ meta.total_pages }}</span>
        <button :disabled="meta.page >= meta.total_pages" @click="setPage(meta.page + 1)">
          下一页
        </button>
      </div>
    </template>

    <!-- 移动端浮动按钮 -->
    <button v-if="isMobile" class="fab" type="button" aria-label="新建任务" @click="openCreate"><Plus :size="24" /></button>

    <!-- 移动端筛选面板 -->
    <MobileFilters v-model:visible="showFilters" :filters="filters" @change="setFilters" />

    <!-- 抽屉 -->
    <TaskDetailDrawer v-model:visible="drawerVisible" :task="selectedTask" @submit="handleSubmit" />
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, onUnmounted } from 'vue'
import { Plus, SlidersHorizontal } from 'lucide-vue-next'
import { useTasks } from '@/features/tasks/useTasks'
import TaskTable from '@/features/tasks/TaskTable.vue'
import TaskCardList from '@/features/tasks/TaskCardList.vue'
import TaskFilters from '@/features/tasks/TaskFilters.vue'
import TaskSortMenu from '@/features/tasks/TaskSortMenu.vue'
import MobileFilters from '@/features/tasks/MobileFilters.vue'
import TaskDetailDrawer from '@/features/tasks/TaskDetailDrawer.vue'
import type { Task, CreateTaskPayload, UpdateTaskPayload } from '@/entities/task/model'

const {
  tasks,
  loading,
  error,
  meta,
  filters,
  sort,
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
const windowWidth = ref(window.innerWidth)

const isMobile = computed(() => windowWidth.value < 768)

function handleResize() {
  windowWidth.value = window.innerWidth
}

onMounted(() => {
  window.addEventListener('resize', handleResize)
  fetchTasks()
})

onUnmounted(() => {
  window.removeEventListener('resize', handleResize)
})

function openCreate() {
  selectedTask.value = null
  drawerVisible.value = true
}

function openTask(task: Task) {
  selectedTask.value = task
  drawerVisible.value = true
}

function editTask(task: Task) {
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
  }
}
</script>

<style scoped>
.tasks-page {
  display: flex;
  flex-direction: column;
  gap: 16px;
  min-width: 0;
}

.page-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  gap: 12px;
  min-width: 0;
}

.page-header h2 {
  margin: 0;
  font-size: 20px;
  min-width: 0;
  overflow-wrap: anywhere;
}

.header-actions {
  display: flex;
  gap: 8px;
  flex-shrink: 0;
}

.btn-primary {
  padding: 8px 16px;
  background: var(--color-primary);
  color: white;
  border: none;
  border-radius: 6px;
  cursor: pointer;
  font-size: 14px;
}

.btn-primary:hover {
  background: var(--color-primary-hover);
}

.page-toolbar {
  display: flex;
  justify-content: space-between;
  align-items: center;
  flex-wrap: wrap;
  gap: 16px;
}

.page-loading,
.page-error,
.page-empty {
  text-align: center;
  padding: 48px 24px;
  color: var(--color-text-muted);
}

.page-error {
  color: var(--color-danger);
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
    box-shadow: 0 4px 12px rgba(0, 0, 0, 0.2);
    z-index: 50;
    cursor: pointer;
  }

  .page-toolbar {
    display: none;
  }

  .page-header {
    align-items: flex-start;
  }

  .page-header h2 {
    font-size: 18px;
    line-height: 1.4;
  }

  .pagination {
    justify-content: space-between;
    gap: 8px;
    padding: 8px 72px 8px 0;
  }

  .pagination span {
    order: -1;
    width: 100%;
    text-align: center;
  }
}

@media (max-width: 359px) {
  .page-header {
    flex-wrap: wrap;
  }

  .header-actions {
    width: 100%;
    justify-content: flex-end;
  }

  .pagination {
    padding-right: 0;
  }
}
</style>
