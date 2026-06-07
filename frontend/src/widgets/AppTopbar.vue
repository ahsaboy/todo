<script setup lang="ts">
import { computed, ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useAuthStore } from '@/app/stores/auth.store'
import { useClickOutside } from '@/shared/composables/useClickOutside'
import AppearanceSettingsTrigger from '@/shared/ui/AppearanceSettingsTrigger.vue'
import UserAvatar from '@/shared/ui/UserAvatar.vue'
import { LogOut, PanelLeftClose, UserCircle } from 'lucide-vue-next'

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
const authStore = useAuthStore()
const isUserMenuOpen = ref(false)
const userMenuRef = ref<HTMLElement | null>(null)
let closeTimer: ReturnType<typeof setTimeout> | null = null

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

async function handleLogout() {
  closeUserMenuImmediate()
  authStore.logout()
  await router.replace({ name: 'login' })
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
          aria-label="用户菜单"
          :aria-expanded="isUserMenuOpen"
          aria-haspopup="menu"
          @click="toggleUserMenu"
          @keydown="handleBtnKeydown"
        >
          <UserAvatar :avatar-url="authStore.user?.avatarUrl" :username="authStore.user?.username" size="sm" />
        </button>
        <Transition name="dropdown">
          <div
            v-if="isUserMenuOpen"
            class="user-dropdown"
            role="menu"
          >
            <router-link
              class="user-menu-item"
              to="/profile"
              role="menuitem"
              @click="closeUserMenuImmediate"
            >
              <UserCircle :size="16" />
              <span>个人资料</span>
            </router-link>
            <button
              class="user-menu-item danger"
              type="button"
              role="menuitem"
              @click="handleLogout"
            >
              <LogOut :size="16" />
              <span>退出登录</span>
            </button>
          </div>
        </Transition>
      </div>
    </div>
  </header>
</template>

