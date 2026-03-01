export default defineNuxtConfig({
  ssr: false,
  modules: ['nuxt-electron'],
  router: {
    options: {
      // Electron runs on file:// in production; hash mode avoids Windows absolute
      // paths being interpreted as route paths (e.g. /C:/.../index.html -> 404).
      hashMode: true,
    },
  },
  electron: {
    build: [
      { entry: 'electron/main.ts' },
      { entry: 'electron/preload.ts' },
    ],
  },
  nitro: {
    // 生成静态 HTML 用于 Electron 加载
    prerender: {
      routes: ['/'],
    },
  },
  devtools: { enabled: false },
  compatibilityDate: '2025-01-01',
})
