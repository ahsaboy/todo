<script setup lang="ts">
import {
  LayoutDashboard,
  ListTodo,
  Settings,
  Users,
} from 'lucide-vue-next'
import { useRoute } from 'vue-router'

type NavItem = {
  to: string
  label: string
  icon: typeof LayoutDashboard
}

const route = useRoute()

const navItems: NavItem[] = [
  { to: '/admin/dashboard', label: '仪表盘', icon: LayoutDashboard },
  { to: '/admin/users', label: '用户', icon: Users },
  { to: '/admin/tasks', label: '任务', icon: ListTodo },
  { to: '/admin/config', label: '配置', icon: Settings },
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
