<script setup lang="ts">
import { computed, ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useAuthStore } from '@/app/stores/auth.store'
import { useThemeStore } from '@/app/stores/theme.store'
import { LogOut, Menu, Moon, Sun, UserCircle } from 'lucide-vue-next'

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
const themeStore = useThemeStore()
const isUserMenuOpen = ref(false)

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

const themeToggleLabel = computed(() => {
  return themeStore.isDark ? '切换到浅色主题' : '切换到深色主题'
})

const themeToggleIcon = computed(() => {
  return themeStore.isDark ? Sun : Moon
})

const showSidebarToggle = computed(() => props.sidebarToggleMode !== null)

const sidebarToggleLabel = computed(() => {
  if (props.sidebarToggleMode === 'mobile') {
    return props.isSidebarOpen ? '关闭导航菜单' : '打开导航菜单'
  }

  return props.isSidebarOpen ? '折叠侧边栏' : '展开侧边栏'
})

function openUserMenu() {
  isUserMenuOpen.value = true
}

function closeUserMenu() {
  isUserMenuOpen.value = false
}

function toggleUserMenu() {
  isUserMenuOpen.value = !isUserMenuOpen.value
}

async function handleLogout() {
  closeUserMenu()
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
      <Menu
        :size="18"
      />
    </button>
    <h1 class="page-title">{{ pageTitle }}</h1>
    <div class="topbar-actions">
      <button
        class="btn-icon theme-toggle-btn"
        type="button"
        :aria-label="themeToggleLabel"
        @click="themeStore.toggleTheme()"
      >
        <component
          :is="themeToggleIcon"
          :size="16"
        />
      </button>
      <div
        class="user-menu"
        @mouseenter="openUserMenu"
        @mouseleave="closeUserMenu"
      >
        <button
          class="user-btn"
          type="button"
          aria-label="用户菜单"
          :aria-expanded="isUserMenuOpen"
          aria-haspopup="menu"
          @click="toggleUserMenu"
          @focus="openUserMenu"
        >
          {{ userInitial }}
        </button>
        <div
          v-show="isUserMenuOpen"
          class="user-dropdown"
          role="menu"
        >
          <router-link
            class="user-menu-item"
            to="/profile"
            role="menuitem"
            @click="closeUserMenu"
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
      </div>
    </div>
  </header>
</template>

<style scoped>
.theme-toggle-btn {
  width: 32px;
  height: 32px;
  flex-shrink: 0;
  border: 1px solid var(--color-border);
  border-radius: 50%;
  background: var(--color-surface-muted);
  color: var(--color-text-muted);
  transition:
    border-color 150ms,
    background-color 150ms,
    box-shadow 150ms,
    color 150ms;
}

.theme-toggle-btn:hover,
.theme-toggle-btn:focus-visible {
  background: var(--color-surface);
  border-color: var(--color-primary);
  box-shadow: 0 0 0 3px color-mix(in srgb, var(--color-primary) 12%, transparent);
  color: var(--color-text);
  outline: none;
}

.user-menu {
  position: relative;
  display: flex;
  align-items: center;
  flex-shrink: 0;
}

.user-btn {
  transition:
    border-color 150ms,
    background-color 150ms,
    box-shadow 150ms;
}

.user-btn:hover,
.user-btn:focus-visible,
.user-btn[aria-expanded='true'] {
  background: var(--color-surface);
  border-color: var(--color-primary);
  box-shadow: 0 0 0 3px color-mix(in srgb, var(--color-primary) 12%, transparent);
  outline: none;
}

.user-dropdown {
  position: absolute;
  top: calc(100% + 8px);
  right: 0;
  min-width: 148px;
  padding: 6px;
  border: 1px solid var(--color-border);
  border-radius: 8px;
  background: var(--color-surface);
  box-shadow: var(--shadow-panel);
  z-index: 120;
}

.user-dropdown::before {
  position: absolute;
  top: -8px;
  right: 0;
  left: 0;
  height: 8px;
  content: '';
}

.user-menu-item {
  width: 100%;
  min-height: 36px;
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 8px 10px;
  border: none;
  border-radius: 6px;
  background: transparent;
  color: var(--color-text);
  cursor: pointer;
  font-size: 14px;
  line-height: 20px;
  text-align: left;
  text-decoration: none;
}

.user-menu-item:hover,
.user-menu-item:focus-visible {
  background: var(--color-surface-muted);
  outline: none;
}

.user-menu-item.danger {
  color: var(--color-danger);
}

.user-menu-item svg {
  flex-shrink: 0;
  color: currentColor;
}

@media (max-width: 767px) {
  .user-dropdown {
    right: -4px;
  }
}
</style>
