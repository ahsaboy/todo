<script setup lang="ts">
import AppShell from './AppShell.vue'
import AdminSidebar from './AdminSidebar.vue'
import AdminTopbar from './AdminTopbar.vue'
import AdminMobileBottomNav from './AdminMobileBottomNav.vue'
</script>

<template>
  <AppShell
    :sidebar="AdminSidebar"
    :topbar="AdminTopbar"
    :bottom-nav="AdminMobileBottomNav"
    storage-key="admin_sidebar_collapsed"
  />
</template>
