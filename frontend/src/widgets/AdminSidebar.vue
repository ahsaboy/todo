<script setup lang="ts">
import { computed } from 'vue'
import {
  LayoutDashboard,
  Users,
  ListTodo,
  BellRing,
  ScrollText,
  Settings,
  Terminal,
  ShieldCheck,
  PanelLeftClose,
} from 'lucide-vue-next'

type SidebarMode = 'desktop' | 'mobile'

type NavItem = {
  to: string
  label: string
  icon: typeof LayoutDashboard
}

const props = withDefaults(defineProps<{
  collapsed?: boolean
  mode?: SidebarMode
  showCollapseToggle?: boolean
  closeOnNavigate?: boolean
}>(), {
  collapsed: false,
  mode: 'desktop',
})

const emit = defineEmits<{
  toggle: []
  navigate: [to: string]
  requestClose: [to: string]
}>()

const navItems: NavItem[] = [
  { to: '/admin/dashboard', label: '仪表盘', icon: LayoutDashboard },
  { to: '/admin/users', label: '用户管理', icon: Users },
  { to: '/admin/tasks', label: '任务管理', icon: ListTodo },
  { to: '/admin/reminder-configs', label: '提醒配置', icon: BellRing },
  { to: '/admin/reminder-logs', label: '提醒日志', icon: ScrollText },
  { to: '/admin/system-logs', label: '系统日志', icon: Terminal },
  { to: '/admin/audit-logs', label: '操作日志', icon: ShieldCheck },
  { to: '/admin/config', label: '系统配置', icon: Settings },
]

const isDesktopMode = computed(() => props.mode === 'desktop')
const isCollapsed = computed(() => isDesktopMode.value && props.collapsed)
const shouldShowCollapseToggle = computed(() =>
  props.showCollapseToggle ?? isDesktopMode.value,
)

const handleNavigate = (to: string) => {
  emit('navigate', to)

  if (props.closeOnNavigate) {
    emit('requestClose', to)
  }
}
</script>

<template>
  <aside
    class="app-sidebar"
    :class="{
      collapsed: isCollapsed,
      'app-sidebar--mobile': !isDesktopMode,
    }"
  >
    <div class="sidebar-logo">
      <span class="sidebar-logo__full">后台管理</span>
      <span class="sidebar-logo__mini" aria-hidden="true">管</span>
    </div>

    <nav class="sidebar-nav">
      <div class="nav-section">
        <div class="nav-label">
          管理
        </div>
        <router-link
          v-for="item in navItems"
          :key="item.to"
          :to="item.to"
          class="nav-item"
          @click="handleNavigate(item.to)"
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

    <div
      v-if="shouldShowCollapseToggle"
      class="sidebar-footer"
    >
      <button
        class="collapse-btn"
        :aria-label="isCollapsed ? '展开侧边栏' : '折叠侧边栏'"
        @click="emit('toggle')"
      >
        <PanelLeftClose
          :size="18"
          :class="['collapse-btn__icon', { 'is-collapsed': isCollapsed }]"
        />
      </button>
    </div>
  </aside>
</template>
