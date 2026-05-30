import type { RouteLocationNormalizedLoaded } from 'vue-router'

export function resolveShellTransitionName() {
  return 'route-shell'
}

export function resolveShellKey(route: RouteLocationNormalizedLoaded) {
  if (route.meta.shell === 'app') return 'app-shell'
  if (route.meta.shell === 'admin') return 'admin-shell'
  return route.fullPath
}

export function resolvePageTransitionName(route: RouteLocationNormalizedLoaded) {
  return route.meta.motion === 'board' ? 'route-page-board' : 'route-page'
}
