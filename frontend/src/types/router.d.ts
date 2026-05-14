import 'vue-router'

declare module 'vue-router' {
  interface RouteMeta {
    shell?: 'auth' | 'app'
    motion?: 'page' | 'board'
  }
}
