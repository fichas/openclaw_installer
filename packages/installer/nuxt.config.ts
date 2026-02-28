export default defineNuxtConfig({
  ssr: false,
  modules: ['nuxt-electron'],
  electron: {
    build: [
      { entry: 'electron/main.ts' },
      { entry: 'electron/preload.ts', onstart: { reloadOnChange: true } },
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
