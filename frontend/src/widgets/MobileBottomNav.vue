<script setup lang="ts"></script>

<template>
  <nav class="mobile-bottom-nav">
    <router-link to="/tasks" class="nav-item">
      <span class="nav-icon">📋</span>
      <span class="nav-text">任务</span>
    </router-link>
    <router-link to="/reminder-configs" class="nav-item">
      <span class="nav-icon">🔔</span>
      <span class="nav-text">提醒</span>
    </router-link>
    <router-link to="/reminder-logs" class="nav-item">
      <span class="nav-icon">🧾</span>
      <span class="nav-text">日志</span>
    </router-link>
    <router-link to="/api-keys" class="nav-item">
      <span class="nav-icon">🔑</span>
      <span class="nav-text">Key</span>
    </router-link>
    <router-link to="/profile" class="nav-item">
      <span class="nav-icon">👤</span>
      <span class="nav-text">我的</span>
    </router-link>
  </nav>
</template>
