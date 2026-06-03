import { computed, ref } from 'vue'
import { defineStore } from 'pinia'

export type FontMode = 'sans' | 'mono' | 'serif'

const FONT_STORAGE_KEY = 'font_mode'

function isBrowserEnvironment() {
  return typeof window !== 'undefined' && typeof document !== 'undefined'
}

function resolvePreferredFont(): FontMode {
  if (!isBrowserEnvironment()) {
    return 'sans'
  }

  const stored = window.localStorage.getItem(FONT_STORAGE_KEY)
  if (stored === 'sans' || stored === 'mono' || stored === 'serif') {
    return stored
  }

  return 'sans'
}

export const useFontStore = defineStore('font', () => {
  const mode = ref<FontMode>('sans')

  const isMono = computed(() => mode.value === 'mono')

  function setFont(nextMode: FontMode) {
    mode.value = nextMode

    if (!isBrowserEnvironment()) {
      return
    }

    document.documentElement.dataset.font = nextMode
    window.localStorage.setItem(FONT_STORAGE_KEY, nextMode)
  }

  function initFont() {
    setFont(resolvePreferredFont())
  }

  function toggleFont() {
    const modes: FontMode[] = ['sans', 'mono', 'serif']
    const next = modes[(modes.indexOf(mode.value) + 1) % modes.length]
    setFont(next)
  }

  return {
    mode,
    isMono,
    initFont,
    setFont,
    toggleFont,
  }
})
