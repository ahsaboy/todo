<script setup lang="ts">
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
          <span class="nav-icon">📋</span>
          <span class="nav-text" v-if="!collapsed">全部任务</span>
        </router-link>
        <router-link to="/tasks/today" class="nav-item">
          <span class="nav-icon">📅</span>
          <span class="nav-text" v-if="!collapsed">今日</span>
        </router-link>
        <router-link to="/tasks/upcoming" class="nav-item">
          <span class="nav-icon">⏰</span>
          <span class="nav-text" v-if="!collapsed">即将到期</span>
        </router-link>
        <router-link to="/tasks/board" class="nav-item">
          <span class="nav-icon">📊</span>
          <span class="nav-text" v-if="!collapsed">看板</span>
        </router-link>
      </div>

      <div class="nav-section">
        <div class="nav-label" v-if="!collapsed">配置</div>
        <router-link to="/reminder-configs" class="nav-item">
          <span class="nav-icon">🔔</span>
          <span class="nav-text" v-if="!collapsed">提醒配置</span>
        </router-link>
        <router-link to="/reminder-logs" class="nav-item">
          <span class="nav-icon">🧾</span>
          <span class="nav-text" v-if="!collapsed">提醒日志</span>
        </router-link>
        <router-link to="/api-keys" class="nav-item">
          <span class="nav-icon">🔑</span>
          <span class="nav-text" v-if="!collapsed">API Key</span>
        </router-link>
      </div>

      <div class="nav-section">
        <div class="nav-label" v-if="!collapsed">账户</div>
        <router-link to="/profile" class="nav-item">
          <span class="nav-icon">👤</span>
          <span class="nav-text" v-if="!collapsed">个人资料</span>
        </router-link>
      </div>
    </nav>

    <div class="sidebar-footer">
      <button class="collapse-btn" @click="$emit('toggle')">
        <span v-if="collapsed">→</span>
        <span v-else>←</span>
      </button>
    </div>
  </aside>
</template>
