<script setup lang="ts">
import {
  ListTodo,
  CalendarDays,
  AlarmClock,
  LayoutDashboard,
  BellRing,
  ScrollText,
  KeyRound,
  UserCircle,
  PanelLeftClose,
  PanelLeftOpen,
} from 'lucide-vue-next'

defineProps<{
  collapsed: boolean
}>()

defineEmits<{
  toggle: []
}>()
</script>

<template>
  <aside class="app-sidebar" :class="{ collapsed }">
    <div class="sidebar-logo">
      <span v-if="!collapsed">TODO</span>
      <span v-else>T</span>
    </div>

    <nav class="sidebar-nav">
      <div class="nav-section">
        <div class="nav-label" v-if="!collapsed">任务中心</div>
        <router-link to="/tasks" class="nav-item">
          <ListTodo class="nav-icon" :size="18" />
          <span class="nav-text" v-if="!collapsed">全部任务</span>
        </router-link>
        <router-link to="/tasks/today" class="nav-item">
          <CalendarDays class="nav-icon" :size="18" />
          <span class="nav-text" v-if="!collapsed">今日</span>
        </router-link>
        <router-link to="/tasks/upcoming" class="nav-item">
          <AlarmClock class="nav-icon" :size="18" />
          <span class="nav-text" v-if="!collapsed">即将到期</span>
        </router-link>
        <router-link to="/tasks/board" class="nav-item">
          <LayoutDashboard class="nav-icon" :size="18" />
          <span class="nav-text" v-if="!collapsed">看板</span>
        </router-link>
      </div>

      <div class="nav-section">
        <div class="nav-label" v-if="!collapsed">配置</div>
        <router-link to="/reminder-configs" class="nav-item">
          <BellRing class="nav-icon" :size="18" />
          <span class="nav-text" v-if="!collapsed">提醒配置</span>
        </router-link>
        <router-link to="/reminder-logs" class="nav-item">
          <ScrollText class="nav-icon" :size="18" />
          <span class="nav-text" v-if="!collapsed">提醒日志</span>
        </router-link>
        <router-link to="/api-keys" class="nav-item">
          <KeyRound class="nav-icon" :size="18" />
          <span class="nav-text" v-if="!collapsed">API Key</span>
        </router-link>
      </div>

      <div class="nav-section">
        <div class="nav-label" v-if="!collapsed">账户</div>
        <router-link to="/profile" class="nav-item">
          <UserCircle class="nav-icon" :size="18" />
          <span class="nav-text" v-if="!collapsed">个人资料</span>
        </router-link>
      </div>
    </nav>

    <div class="sidebar-footer">
      <button class="collapse-btn" :aria-label="collapsed ? '展开侧边栏' : '折叠侧边栏'" @click="$emit('toggle')">
        <PanelLeftOpen v-if="collapsed" :size="18" />
        <PanelLeftClose v-else :size="18" />
      </button>
    </div>
  </aside>
</template>
