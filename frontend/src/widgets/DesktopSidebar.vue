<script setup lang="ts">
import { computed } from 'vue'
import {
  ListTodo,
  CalendarDays,
  AlarmClock,
  LayoutDashboard,
  BellRing,
  ScrollText,
  KeyRound,
  UserCircle,
  Tags,
  PanelLeftClose,
} from 'lucide-vue-next'

type SidebarMode = 'desktop' | 'mobile'

type NavItem = {
  to: string
  label: string
  icon: typeof ListTodo
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

const taskItems: NavItem[] = [
  { to: '/tasks', label: '全部任务', icon: ListTodo },
  { to: '/tasks/today', label: '今日', icon: CalendarDays },
  { to: '/tasks/upcoming', label: '即将到期', icon: AlarmClock },
  { to: '/tasks/board', label: '看板', icon: LayoutDashboard },
]

const configItems: NavItem[] = [
  { to: '/tags', label: '标签管理', icon: Tags },
  { to: '/reminder-configs', label: '提醒配置', icon: BellRing },
  { to: '/reminder-logs', label: '提醒日志', icon: ScrollText },
  { to: '/api-keys', label: 'API Key', icon: KeyRound },
]

const mobileOnlyConfigItems: NavItem[] = [
  { to: '/profile', label: '个人资料', icon: UserCircle },
]

const isDesktopMode = computed(() => props.mode === 'desktop')
const isCollapsed = computed(() => isDesktopMode.value && props.collapsed)
const shouldShowCollapseToggle = computed(() =>
  props.showCollapseToggle ?? isDesktopMode.value,
)
const visibleConfigItems = computed(() =>
  isDesktopMode.value ? configItems : [...configItems, ...mobileOnlyConfigItems],
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
      <span class="sidebar-logo__full">TODO</span>
      <span class="sidebar-logo__mini" aria-hidden="true">T</span>
    </div>

    <nav class="sidebar-nav">
      <div class="nav-section">
        <div class="nav-label">
          任务中心
        </div>
        <router-link
          v-for="item in taskItems"
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
          <span class="nav-text">
            {{ item.label }}
          </span>
        </router-link>
      </div>

      <div class="nav-section">
        <div class="nav-label">
          配置
        </div>
        <router-link
          v-for="item in visibleConfigItems"
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
          <span class="nav-text">
            {{ item.label }}
          </span>
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
