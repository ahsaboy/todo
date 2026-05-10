<script setup lang="ts">
import { ListTodo, Bell, ScrollText, KeyRound } from 'lucide-vue-next'
</script>

<template>
  <nav class="mobile-bottom-nav">
    <router-link to="/tasks" class="nav-item">
      <ListTodo class="nav-icon" :size="20" />
      <span class="nav-text">任务</span>
    </router-link>
    <router-link to="/reminder-configs" class="nav-item">
      <Bell class="nav-icon" :size="20" />
      <span class="nav-text">提醒</span>
    </router-link>
    <router-link to="/reminder-logs" class="nav-item">
      <ScrollText class="nav-icon" :size="20" />
      <span class="nav-text">日志</span>
    </router-link>
    <router-link to="/api-keys" class="nav-item">
      <KeyRound class="nav-icon" :size="20" />
      <span class="nav-text">Key</span>
    </router-link>
  </nav>
</template>
