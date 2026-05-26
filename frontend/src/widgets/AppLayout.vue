<script setup lang="ts">
import { computed, onBeforeUnmount, onMounted, ref, watch } from 'vue'
import { useRoute } from 'vue-router'
import { useTagStore } from '@/entities/tag/store'
import DesktopSidebar from './DesktopSidebar.vue'
import AppTopbar from './AppTopbar.vue'
import AppFooter from './AppFooter.vue'
import MobileBottomNav from './MobileBottomNav.vue'

const MOBILE_MEDIA_QUERY = '(max-width: 767px)'
const DESKTOP_MEDIA_QUERY = '(min-width: 1024px)'

const route = useRoute()
const sidebarCollapsed = ref(false)
const mobileSidebarOpen = ref(false)
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

const toggleDesktopSidebar = () => {
  if (!isDesktop.value) return
  sidebarCollapsed.value = !sidebarCollapsed.value
}

const openMobileSidebar = () => {
  if (!isMobile.value) return
  mobileSidebarOpen.value = true
}

const closeMobileSidebar = () => {
  mobileSidebarOpen.value = false
}

const handleSidebarToggle = () => {
  if (isMobile.value) {
    if (mobileSidebarOpen.value) {
      closeMobileSidebar()
    } else {
      openMobileSidebar()
    }
    return
  }

  toggleDesktopSidebar()
}

const handleWindowKeydown = (event: globalThis.KeyboardEvent) => {
  if (event.key !== 'Escape' || !isMobile.value || !mobileSidebarOpen.value) {
    return
  }

  closeMobileSidebar()
}

onMounted(() => {
  sidebarCollapsed.value = localStorage.getItem('sidebar_collapsed') === 'true'
  mobileMediaQuery = window.matchMedia(MOBILE_MEDIA_QUERY)
  desktopMediaQuery = window.matchMedia(DESKTOP_MEDIA_QUERY)
  syncViewportState()
  mobileMediaQuery.addEventListener('change', syncViewportState)
  desktopMediaQuery.addEventListener('change', syncViewportState)
  window.addEventListener('keydown', handleWindowKeydown)
  // 预加载标签字典,确保 TagChip 在首次渲染时就有颜色/图标数据
  useTagStore().fetchTags()
})

onBeforeUnmount(() => {
  mobileMediaQuery?.removeEventListener('change', syncViewportState)
  desktopMediaQuery?.removeEventListener('change', syncViewportState)
  window.removeEventListener('keydown', handleWindowKeydown)
})

watch(sidebarCollapsed, (value) => {
  if (!isDesktop.value) return
  localStorage.setItem('sidebar_collapsed', String(value))
})

watch(isMobile, (mobile) => {
  if (!mobile) {
    closeMobileSidebar()
  }
})

watch([mobileSidebarOpen, isMobile], ([open, mobile]) => {
  document.body.classList.toggle('mobile-sidebar-open', mobile && open)
})

watch(
  () => route.fullPath,
  () => {
    if (mobileSidebarOpen.value) {
      closeMobileSidebar()
    }
  },
)

onBeforeUnmount(() => {
  document.body.classList.remove('mobile-sidebar-open')
})
</script>

<template>
  <div class="app-layout">
    <DesktopSidebar
      :collapsed="effectiveSidebarCollapsed"
      @toggle="toggleDesktopSidebar"
    />
    <div class="app-main">
      <AppTopbar
        :sidebar-toggle-mode="isMobile ? 'mobile' : isDesktop ? 'desktop' : null"
        :is-sidebar-open="isMobile ? mobileSidebarOpen : !effectiveSidebarCollapsed"
        @toggle-sidebar="handleSidebarToggle"
      />
      <div class="page-content">
        <div class="route-page-stage">
          <router-view v-slot="{ Component }">
            <component :is="Component" />
          </router-view>
        </div>
      </div>
      <AppFooter />
    </div>
    <Transition name="overlay-motion">
      <div
        v-if="isMobile && mobileSidebarOpen"
        class="mobile-sidebar-layer"
      >
        <div
          class="mobile-sidebar-backdrop"
          @click="closeMobileSidebar"
        />
        <DesktopSidebar
          mode="mobile"
          class="motion-panel motion-panel--sidebar"
          :close-on-navigate="true"
          :show-collapse-toggle="false"
          @request-close="closeMobileSidebar"
        />
      </div>
    </Transition>
    <MobileBottomNav />
  </div>
</template>
