<script setup lang="ts">
import { computed, ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useAdminAuthStore } from '@/app/stores/admin-auth.store'
import { useClickOutside } from '@/shared/composables/useClickOutside'
import AppearanceSettingsTrigger from '@/shared/ui/AppearanceSettingsTrigger.vue'
import { Home, LogOut, PanelLeftClose } from 'lucide-vue-next'

type SidebarToggleMode = 'desktop' | 'mobile' | null

const props = withDefaults(defineProps<{
  sidebarToggleMode?: SidebarToggleMode
  isSidebarOpen?: boolean
}>(), {
  sidebarToggleMode: null,
  isSidebarOpen: false,
})

defineEmits<{
  toggleSidebar: []
}>()

const route = useRoute()
const router = useRouter()
const adminAuthStore = useAdminAuthStore()
const isUserMenuOpen = ref(false)
const userMenuRef = ref<HTMLElement | null>(null)
let closeTimer: ReturnType<typeof setTimeout> | null = null

const pageTitle = computed(() => {
  const titles: Record<string, string> = {
    'admin-dashboard': '仪表盘',
    'admin-users': '用户管理',
    'admin-tasks': '任务管理',
    'admin-reminder-configs': '提醒配置',
    'admin-reminder-logs': '提醒日志',
    'admin-system-logs': '系统日志',
    'admin-audit-logs': '操作日志',
    'admin-config': '系统配置',
  }
  return titles[route.name as string] || '后台管理'
})

const showSidebarToggle = computed(() => props.sidebarToggleMode !== null)

const sidebarToggleLabel = computed(() => {
  if (props.sidebarToggleMode === 'mobile') {
    return props.isSidebarOpen ? '关闭导航菜单' : '打开导航菜单'
  }

  return props.isSidebarOpen ? '折叠侧边栏' : '展开侧边栏'
})

function openUserMenu() {
  if (closeTimer) {
    clearTimeout(closeTimer)
    closeTimer = null
  }
  isUserMenuOpen.value = true
}

function scheduleCloseUserMenu() {
  closeTimer = setTimeout(() => {
    isUserMenuOpen.value = false
    closeTimer = null
  }, 150)
}

function closeUserMenuImmediate() {
  if (closeTimer) {
    clearTimeout(closeTimer)
    closeTimer = null
  }
  isUserMenuOpen.value = false
}

function toggleUserMenu() {
  if (closeTimer) {
    clearTimeout(closeTimer)
    closeTimer = null
  }
  isUserMenuOpen.value = !isUserMenuOpen.value
}

function handleBtnKeydown(e: KeyboardEvent) {
  if (e.key === 'Enter' || e.key === ' ') {
    e.preventDefault()
    toggleUserMenu()
  }
}

useClickOutside(userMenuRef, closeUserMenuImmediate)

function goFrontend() {
  closeUserMenuImmediate()
  router.push('/')
}

function handleLogout() {
  closeUserMenuImmediate()
  adminAuthStore.logout()
  router.push({ name: 'admin-login' })
}
</script>

<template>
  <header class="app-topbar">
    <button
      v-if="showSidebarToggle"
      class="sidebar-toggle"
      type="button"
      :aria-label="sidebarToggleLabel"
      @click="$emit('toggleSidebar')"
    >
      <PanelLeftClose
        :size="18"
        :class="['sidebar-toggle__icon', { 'is-closed': !isSidebarOpen }]"
      />
    </button>
    <h1 class="page-title">{{ pageTitle }}</h1>
    <div class="topbar-actions">
      <AppearanceSettingsTrigger />
      <div
        ref="userMenuRef"
        class="user-menu"
        @mouseenter="openUserMenu"
        @mouseleave="scheduleCloseUserMenu"
      >
        <button
          class="user-btn"
          type="button"
          aria-label="管理员菜单"
          :aria-expanded="isUserMenuOpen"
          aria-haspopup="menu"
          @click="toggleUserMenu"
          @keydown="handleBtnKeydown"
        >
          管
        </button>
        <Transition name="dropdown">
          <div
            v-if="isUserMenuOpen"
            class="user-dropdown"
            role="menu"
          >
            <button
              class="user-menu-item"
              type="button"
              role="menuitem"
              @click="goFrontend"
            >
              <Home :size="16" />
              <span>返回前台</span>
            </button>
            <button
              class="user-menu-item danger"
              type="button"
              role="menuitem"
              @click="handleLogout"
            >
              <LogOut :size="16" />
              <span>退出管理</span>
            </button>
          </div>
        </Transition>
      </div>
    </div>
  </header>
</template>
