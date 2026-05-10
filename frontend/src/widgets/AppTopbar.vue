<script setup lang="ts">
import { computed } from 'vue'
import { useRoute } from 'vue-router'
import { useAuthStore } from '@/app/stores/auth.store'

defineProps<{
  showSidebarToggle: boolean
}>()

defineEmits<{
  toggleSidebar: []
}>()

const route = useRoute()
const authStore = useAuthStore()

const pageTitle = computed(() => {
  const titles: Record<string, string> = {
    tasks: '任务中心',
    'tasks-today': '今日任务',
    'tasks-upcoming': '即将到期',
    'tasks-board': '看板',
    'reminder-configs': '提醒配置',
    'reminder-logs': '提醒日志',
    'api-keys': 'API Key',
    profile: '个人资料',
  }
  return titles[route.name as string] || 'TODO'
})

const userInitial = computed(() => {
  return authStore.user?.username?.charAt(0).toUpperCase() || '?'
})
</script>

<template>
  <header class="app-topbar">
    <button v-if="showSidebarToggle" class="sidebar-toggle" type="button" @click="$emit('toggleSidebar')">
      <span>☰</span>
    </button>
    <h1 class="page-title">{{ pageTitle }}</h1>
    <div class="topbar-actions">
      <div class="user-menu">
        <button class="user-btn" type="button" aria-label="用户菜单">{{ userInitial }}</button>
      </div>
    </div>
  </header>
</template>

<style scoped>
.user-menu {
  display: flex;
  align-items: center;
  flex-shrink: 0;
}
</style>
