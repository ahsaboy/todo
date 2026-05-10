<script setup lang="ts">
import { computed, onBeforeUnmount, onMounted, ref, watch } from 'vue'
import DesktopSidebar from './DesktopSidebar.vue'
import AppTopbar from './AppTopbar.vue'
import MobileBottomNav from './MobileBottomNav.vue'

const MOBILE_MEDIA_QUERY = '(max-width: 767px)'
const DESKTOP_MEDIA_QUERY = '(min-width: 1024px)'

const sidebarCollapsed = ref(false)
const isMobile = ref(false)
const isDesktop = ref(false)

let mobileMediaQuery: MediaQueryList | null = null
let desktopMediaQuery: MediaQueryList | null = null

const syncViewportState = () => {
  isMobile.value = mobileMediaQuery?.matches ?? false
  isDesktop.value = desktopMediaQuery?.matches ?? false
}

const effectiveSidebarCollapsed = computed(() => {
  if (isMobile.value) return false
  if (!isDesktop.value) return true
  return sidebarCollapsed.value
})

const toggleSidebar = () => {
  if (!isDesktop.value) return
  sidebarCollapsed.value = !sidebarCollapsed.value
}

onMounted(() => {
  sidebarCollapsed.value = localStorage.getItem('sidebar_collapsed') === 'true'
  mobileMediaQuery = window.matchMedia(MOBILE_MEDIA_QUERY)
  desktopMediaQuery = window.matchMedia(DESKTOP_MEDIA_QUERY)
  syncViewportState()
  mobileMediaQuery.addEventListener('change', syncViewportState)
  desktopMediaQuery.addEventListener('change', syncViewportState)
})

onBeforeUnmount(() => {
  mobileMediaQuery?.removeEventListener('change', syncViewportState)
  desktopMediaQuery?.removeEventListener('change', syncViewportState)
})

watch(sidebarCollapsed, (value) => {
  if (isMobile.value) return
  localStorage.setItem('sidebar_collapsed', String(value))
})
</script>

<template>
  <div class="app-layout">
    <DesktopSidebar :collapsed="effectiveSidebarCollapsed" @toggle="toggleSidebar" />
    <div class="app-main">
      <AppTopbar
        :show-sidebar-toggle="isDesktop"
        @toggle-sidebar="toggleSidebar"
      />
      <div class="page-content">
        <router-view />
      </div>
    </div>
    <MobileBottomNav />
  </div>
</template>
