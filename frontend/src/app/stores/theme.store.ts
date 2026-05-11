import { computed, ref } from 'vue'
import { defineStore } from 'pinia'

export type ThemeMode = 'light' | 'dark'

const THEME_STORAGE_KEY = 'theme_mode'

function isBrowserEnvironment() {
  return typeof window !== 'undefined' && typeof document !== 'undefined'
}

function resolvePreferredTheme(): ThemeMode {
  if (!isBrowserEnvironment()) {
    return 'light'
  }

  const storedMode = window.localStorage.getItem(THEME_STORAGE_KEY)
  if (storedMode === 'light' || storedMode === 'dark') {
    return storedMode
  }

  return window.matchMedia('(prefers-color-scheme: dark)').matches ? 'dark' : 'light'
}

export const useThemeStore = defineStore('theme', () => {
  const mode = ref<ThemeMode>('light')

  const isDark = computed(() => mode.value === 'dark')

  function setTheme(nextMode: ThemeMode) {
    mode.value = nextMode

    if (!isBrowserEnvironment()) {
      return
    }

    document.documentElement.dataset.theme = nextMode
    window.localStorage.setItem(THEME_STORAGE_KEY, nextMode)
  }

  function initTheme() {
    setTheme(resolvePreferredTheme())
  }

  function toggleTheme() {
    setTheme(mode.value === 'dark' ? 'light' : 'dark')
  }

  return {
    mode,
    isDark,
    initTheme,
    setTheme,
    toggleTheme,
  }
})
