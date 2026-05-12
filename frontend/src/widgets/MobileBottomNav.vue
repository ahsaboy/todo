<script setup lang="ts">
import {
  AlarmClock,
  CalendarDays,
  LayoutDashboard,
  ListTodo,
} from 'lucide-vue-next'
import { useRoute } from 'vue-router'

type NavItem = {
  to: string
  label: string
  icon: typeof ListTodo
}

const route = useRoute()

const navItems: NavItem[] = [
  { to: '/tasks', label: '全部任务', icon: ListTodo },
  { to: '/tasks/today', label: '今日', icon: CalendarDays },
  { to: '/tasks/upcoming', label: '即将到期', icon: AlarmClock },
  { to: '/tasks/board', label: '看板', icon: LayoutDashboard },
]

const isNavActive = (to: string) => route.path === to
</script>

<template>
  <nav class="mobile-bottom-nav">
    <router-link
      v-for="item in navItems"
      v-slot="{ href, navigate }"
      :key="item.to"
      :to="item.to"
      custom
    >
      <a
        :href="href"
        class="nav-item"
        :class="{ 'router-link-active': isNavActive(item.to) }"
        @click="navigate"
      >
        <component
          :is="item.icon"
          class="nav-icon"
          :size="20"
        />
        <span class="nav-text">{{ item.label }}</span>
      </a>
    </router-link>
  </nav>
</template>
