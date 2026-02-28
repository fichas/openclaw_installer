export default defineNuxtConfig({
  ssr: false,
  modules: ['nuxt-electron'],
  electron: {
    build: [{ entry: 'electron/main.ts' }],
  },
  devtools: { enabled: false },
  compatibilityDate: '2025-01-01',
})
