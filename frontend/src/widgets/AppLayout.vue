<script setup lang="ts">
import AppShell from './AppShell.vue'
import DesktopSidebar from './DesktopSidebar.vue'
import AppTopbar from './AppTopbar.vue'
import MobileBottomNav from './MobileBottomNav.vue'
import { useTagStore } from '@/entities/tag/store'

function preloadTags() {
  useTagStore().fetchTags()
}
</script>

<template>
  <AppShell
    :sidebar="DesktopSidebar"
    :topbar="AppTopbar"
    :bottom-nav="MobileBottomNav"
    storage-key="sidebar_collapsed"
    :on-ready="preloadTags"
  />
</template>
