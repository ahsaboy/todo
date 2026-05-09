<script setup lang="ts">
import { ref, watch } from 'vue'
import DesktopSidebar from './DesktopSidebar.vue'
import AppTopbar from './AppTopbar.vue'
import MobileBottomNav from './MobileBottomNav.vue'

const sidebarCollapsed = ref(localStorage.getItem('sidebar_collapsed') === 'true')

watch(sidebarCollapsed, (value) => {
  localStorage.setItem('sidebar_collapsed', String(value))
})
</script>

<template>
  <div class="app-layout">
    <DesktopSidebar :collapsed="sidebarCollapsed" @toggle="sidebarCollapsed = !sidebarCollapsed" />
    <div class="app-main">
      <AppTopbar
        :sidebar-collapsed="sidebarCollapsed"
        @toggle-sidebar="sidebarCollapsed = !sidebarCollapsed"
      />
      <div class="page-content">
        <router-view />
      </div>
    </div>
    <MobileBottomNav />
  </div>
</template>
