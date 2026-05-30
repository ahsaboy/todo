import { defineConfig } from 'vite'
import vue from '@vitejs/plugin-vue'
import { fileURLToPath, URL } from 'node:url'

// https://vite.dev/config/
export default defineConfig({
  base: './',
  plugins: [vue()],
  resolve: {
    alias: {
      '@': fileURLToPath(new URL('./src', import.meta.url)),
    },
  },
  build: {
    rollupOptions: {
      onwarn(warning, defaultHandler) {
        // @vueuse/core 预构建产物里 /* #__PURE__ */ 注释位置触发 Rolldown 的
        // INVALID_ANNOTATION 告警，属于第三方依赖问题且无法修改其源码，过滤掉避免噪音。
        // 仅屏蔽 node_modules 来源，自己代码若有同类问题仍会提示。
        if (
          warning.code === 'INVALID_ANNOTATION' &&
          (warning.loc?.file ?? warning.id ?? '').includes('node_modules')
        ) {
          return
        }
        defaultHandler(warning)
      },
    },
  },
})

