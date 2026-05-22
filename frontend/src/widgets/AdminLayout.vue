<script setup lang="ts">
import AdminSidebar from './AdminSidebar.vue'
</script>

<template>
  <div class="app-layout">
    <AdminSidebar />
    <div class="app-main">
      <div class="page-content">
        <div class="route-page-stage">
          <router-view />
        </div>
      </div>
    </div>
  </div>
</template>
