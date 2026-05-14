import type { RouteLocationNormalizedLoaded } from 'vue-router'

export function resolveShellTransitionName() {
  return 'route-shell'
}

export function resolveShellKey(route: RouteLocationNormalizedLoaded) {
  return route.meta.shell === 'app' ? 'app-shell' : route.fullPath
}

export function resolvePageTransitionName(route: RouteLocationNormalizedLoaded) {
  return route.meta.motion === 'board' ? 'route-page-board' : 'route-page'
}
