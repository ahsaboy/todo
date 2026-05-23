<script setup lang="ts">
import {
  LayoutDashboard,
  Users,
  ListTodo,
  BellRing,
  ScrollText,
  Settings,
  LogOut,
  Terminal,
} from 'lucide-vue-next'
import { useRouter } from 'vue-router'
import { useAdminAuthStore } from '@/app/stores/admin-auth.store'

const router = useRouter()
const adminAuthStore = useAdminAuthStore()

type NavItem = {
  to: string
  label: string
  icon: typeof LayoutDashboard
}

const navItems: NavItem[] = [
  { to: '/admin/dashboard', label: '仪表盘', icon: LayoutDashboard },
  { to: '/admin/users', label: '用户管理', icon: Users },
  { to: '/admin/tasks', label: '任务管理', icon: ListTodo },
  { to: '/admin/reminder-configs', label: '提醒配置', icon: BellRing },
  { to: '/admin/reminder-logs', label: '提醒日志', icon: ScrollText },
  { to: '/admin/system-logs', label: '系统日志', icon: Terminal },
  { to: '/admin/config', label: '系统配置', icon: Settings },
]

function handleLogout() {
  adminAuthStore.logout()
  router.push({ name: 'admin-login' })
}
</script>

<template>
  <aside class="app-sidebar">
    <div class="sidebar-logo">
      <span>后台管理</span>
    </div>
    <nav class="sidebar-nav">
      <div class="nav-section">
        <router-link
          v-for="item in navItems"
          :key="item.to"
          :to="item.to"
          class="nav-item"
        >
          <component
            :is="item.icon"
            class="nav-icon"
            :size="18"
          />
          <span class="nav-text">{{ item.label }}</span>
        </router-link>
      </div>
    </nav>
    <div class="sidebar-footer">
      <button
        class="nav-item collapse-btn"
        style="width: 100%; border: none; background: none; cursor: pointer; color: inherit;"
        @click="handleLogout"
      >
        <LogOut class="nav-icon" :size="18" />
        <span class="nav-text">退出管理</span>
      </button>
    </div>
  </aside>
</template>
