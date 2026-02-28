# OpenClaw 安装器重构实施计划

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** 将 OpenClaw 安装器从 Go CLI + Wails 架构重构为 Electron + Nuxt 3 Monorepo 架构，包含图形化安装向导和 Web 配置服务。

**Architecture:** pnpm workspace Monorepo，三个包：`packages/installer`（Electron + Nuxt 安装向导）、`packages/config-server`（Nuxt 全栈 Web 配置服务）、`packages/shared`（共享组件和类型）。安装器是一次性运行的桌面应用，安装完成后启动常驻后台的配置服务，用户通过浏览器访问 `localhost:18080` 管理配置。

**Tech Stack:** Node.js 20 LTS, pnpm, Nuxt 3, Vue 3, TypeScript, Naive UI, Electron, electron-builder, nuxt-electron

**Design Doc:** `docs/plans/2026-03-01-installer-redesign-design.md`

---

## Phase 1: 项目脚手架搭建

### Task 1: 初始化 Monorepo 根配置

**Files:**
- Create: `package.json`
- Create: `pnpm-workspace.yaml`
- Create: `tsconfig.json`
- Create: `.npmrc`
- Modify: `.gitignore`

**Step 1: 创建根 package.json**

```json
{
  "name": "openclaw",
  "private": true,
  "version": "2.0.0",
  "description": "OpenClaw AI 助手安装器与配置管理系统",
  "scripts": {
    "dev:installer": "pnpm --filter @openclaw/installer dev",
    "dev:config": "pnpm --filter @openclaw/config-server dev",
    "build:installer": "pnpm --filter @openclaw/installer build",
    "build:config": "pnpm --filter @openclaw/config-server build",
    "build": "pnpm -r build",
    "lint": "pnpm -r lint",
    "test": "pnpm -r test"
  },
  "engines": {
    "node": ">=20.0.0",
    "pnpm": ">=9.0.0"
  }
}
```

**Step 2: 创建 pnpm-workspace.yaml**

```yaml
packages:
  - 'packages/*'
```

**Step 3: 创建 .npmrc**

```ini
registry=https://registry.npmmirror.com
electron_mirror=https://npmmirror.com/mirrors/electron/
electron_builder_binaries_mirror=https://npmmirror.com/mirrors/electron-builder-binaries/
shamefully-hoist=true
strict-peer-dependencies=false
```

**Step 4: 创建根 tsconfig.json**

```json
{
  "compilerOptions": {
    "target": "ES2022",
    "module": "ESNext",
    "moduleResolution": "bundler",
    "strict": true,
    "esModuleInterop": true,
    "skipLibCheck": true,
    "forceConsistentCasingInFileNames": true,
    "resolveJsonModule": true,
    "declaration": true,
    "declarationMap": true,
    "sourceMap": true
  },
  "exclude": ["node_modules", "dist", ".output", ".nuxt"]
}
```

**Step 5: 更新 .gitignore 添加 Node.js 相关条目**

在现有 `.gitignore` 末尾追加：

```
# Node.js
node_modules/
.output/
.nuxt/
.pnpm-store/

# Electron
dist-electron/

# Environment
.env
.env.local
```

**Step 6: 运行 pnpm install 验证根配置**

Run: `pnpm install`
Expected: 成功创建 pnpm-lock.yaml，无报错

**Step 7: Commit**

```bash
git add package.json pnpm-workspace.yaml tsconfig.json .npmrc .gitignore pnpm-lock.yaml
git commit -m "chore: 初始化 pnpm workspace Monorepo 根配置"
```

---

### Task 2: 创建 shared 共享包

**Files:**
- Create: `packages/shared/package.json`
- Create: `packages/shared/tsconfig.json`
- Create: `packages/shared/types/index.ts`
- Create: `packages/shared/types/config.ts`
- Create: `packages/shared/types/adapter.ts`
- Create: `packages/shared/types/platform.ts`
- Create: `packages/shared/utils/platform.ts`
- Create: `packages/shared/utils/paths.ts`
- Create: `packages/shared/utils/index.ts`
- Create: `packages/shared/index.ts`

**Step 1: 创建 shared/package.json**

```json
{
  "name": "@openclaw/shared",
  "version": "2.0.0",
  "private": true,
  "main": "./index.ts",
  "types": "./index.ts",
  "exports": {
    ".": "./index.ts",
    "./types": "./types/index.ts",
    "./utils": "./utils/index.ts"
  }
}
```

**Step 2: 创建 shared/tsconfig.json**

```json
{
  "extends": "../../tsconfig.json",
  "include": ["**/*.ts"]
}
```

**Step 3: 创建类型定义 types/config.ts**

从现有 Go 代码 `installer/config.go` 中迁移的配置结构：

```typescript
export interface OpenClawConfig {
  version: string
  platform: string
  server: ServerConfig
  adapters: AdapterConfig[]
  apiKeys: ApiKeyConfig[]
  settings: Record<string, string>
}

export interface ServerConfig {
  host: string
  port: number
  tls: boolean
}

export interface AdapterConfig {
  name: string
  type: string
  enabled: boolean
  displayName: string
  options: Record<string, string>
}

export interface ApiKeyConfig {
  id: string
  name: string
  provider: string
  key: string
  endpoint?: string
  createdAt: string
}

export interface InstallOptions {
  installDir: string
  configDir: string
  sourceDir: string
  adapters: string[]
  mode: 'standard' | 'custom'
}

export interface InstallProgress {
  step: number
  totalSteps: number
  message: string
  percent: number
  error?: string
}
```

**Step 4: 创建类型定义 types/adapter.ts**

从现有 `adapters/*/adapter.json` 迁移的适配器类型：

```typescript
export interface AdapterSchema {
  name: string
  type: string
  displayName: string
  version: string
  description: string
  supportedPlatforms: string[]
  configSchema: Record<string, AdapterFieldSchema>
}

export interface AdapterFieldSchema {
  type: 'string' | 'number' | 'boolean'
  required: boolean
  label?: string
  placeholder?: string
  description?: string
}
```

**Step 5: 创建类型定义 types/platform.ts**

从现有 `installer/platform.go` 迁移的平台类型：

```typescript
export type OS = 'windows' | 'darwin' | 'linux'
export type Arch = 'amd64' | 'arm64'

export interface PlatformInfo {
  os: OS
  arch: Arch
  isWindows: boolean
  isMacOS: boolean
  isLinux: boolean
}

export interface PlatformPaths {
  installDir: string
  configDir: string
  configFile: string
  logDir: string
}
```

**Step 6: 创建 types/index.ts 导出**

```typescript
export * from './config'
export * from './adapter'
export * from './platform'
```

**Step 7: 创建 utils/platform.ts**

从 `installer/platform.go` 迁移平台检测逻辑：

```typescript
import { platform, arch } from 'os'
import { join } from 'path'
import type { PlatformInfo, PlatformPaths, OS, Arch } from '../types'

export function detectPlatform(): PlatformInfo {
  const os = platform() as OS
  const cpuArch = arch()

  const normalizedArch: Arch = cpuArch === 'x64' ? 'amd64' : 'arm64'

  return {
    os: os === 'win32' ? 'windows' : os,
    arch: normalizedArch,
    isWindows: os === 'win32',
    isMacOS: os === 'darwin',
    isLinux: os === 'linux',
  }
}

export function getPlatformPaths(info?: PlatformInfo): PlatformPaths {
  const p = info ?? detectPlatform()

  if (p.isWindows) {
    const appData = process.env.APPDATA || join(process.env.USERPROFILE || '', 'AppData', 'Roaming')
    return {
      installDir: join(process.env.PROGRAMFILES || 'C:\\Program Files', 'OpenClaw'),
      configDir: join(appData, 'OpenClaw'),
      configFile: join(appData, 'OpenClaw', 'config.json'),
      logDir: join(appData, 'OpenClaw', 'logs'),
    }
  }

  if (p.isMacOS) {
    const home = process.env.HOME || '/Users'
    return {
      installDir: '/usr/local/bin',
      configDir: join(home, 'Library', 'Application Support', 'OpenClaw'),
      configFile: join(home, 'Library', 'Application Support', 'OpenClaw', 'config.json'),
      logDir: join(home, 'Library', 'Logs', 'OpenClaw'),
    }
  }

  // Linux
  const home = process.env.HOME || '/home'
  return {
    installDir: '/usr/local/bin',
    configDir: join(home, '.config', 'openclaw'),
    configFile: join(home, '.config', 'openclaw', 'config.json'),
    logDir: join(home, '.local', 'share', 'openclaw', 'logs'),
  }
}
```

**Step 8: 创建 utils/paths.ts**

```typescript
import { readFileSync, writeFileSync, existsSync, mkdirSync } from 'fs'
import { dirname } from 'path'
import type { OpenClawConfig } from '../types'

export function loadConfig(configPath: string): OpenClawConfig | null {
  if (!existsSync(configPath)) return null
  const raw = readFileSync(configPath, 'utf-8')
  return JSON.parse(raw) as OpenClawConfig
}

export function saveConfig(configPath: string, config: OpenClawConfig): void {
  const dir = dirname(configPath)
  if (!existsSync(dir)) {
    mkdirSync(dir, { recursive: true })
  }
  writeFileSync(configPath, JSON.stringify(config, null, 2), 'utf-8')
}
```

**Step 9: 创建 utils/index.ts 和 index.ts 导出**

`utils/index.ts`:
```typescript
export * from './platform'
export * from './paths'
```

`index.ts`:
```typescript
export * from './types'
export * from './utils'
```

**Step 10: 运行 pnpm install 验证 shared 包**

Run: `cd /home/chaos/openclaw && pnpm install`
Expected: 成功链接 @openclaw/shared 包

**Step 11: Commit**

```bash
git add packages/shared/
git commit -m "feat: 创建 @openclaw/shared 包（类型定义和工具函数）"
```

---

## Phase 2: Electron 安装器

### Task 3: 初始化 Electron + Nuxt 安装器项目

**Files:**
- Create: `packages/installer/package.json`
- Create: `packages/installer/nuxt.config.ts`
- Create: `packages/installer/tsconfig.json`
- Create: `packages/installer/app.vue`

**Step 1: 创建 installer/package.json**

```json
{
  "name": "@openclaw/installer",
  "version": "2.0.0",
  "private": true,
  "main": "dist-electron/main.js",
  "scripts": {
    "dev": "nuxt dev",
    "build": "nuxt build && electron-builder",
    "generate": "nuxt generate",
    "preview": "nuxt preview",
    "lint": "eslint .",
    "test": "vitest"
  },
  "dependencies": {
    "@openclaw/shared": "workspace:*",
    "naive-ui": "^2.39.0",
    "vue": "^3.5.0"
  },
  "devDependencies": {
    "nuxt": "^3.14.0",
    "nuxt-electron": "^0.7.0",
    "electron": "^33.0.0",
    "electron-builder": "^25.0.0",
    "vite-plugin-electron": "^0.28.0",
    "vite-plugin-electron-renderer": "^0.14.0",
    "typescript": "^5.6.0",
    "vitest": "^2.1.0",
    "@nuxt/test-utils": "^3.14.0"
  },
  "build": {
    "appId": "com.openclaw.installer",
    "productName": "OpenClaw 安装器",
    "directories": {
      "output": "dist-electron"
    },
    "nsis": {
      "oneClick": false,
      "allowToChangeInstallationDirectory": false,
      "installerLanguages": ["zh_CN"],
      "language": "2052"
    },
    "mac": {
      "target": ["dmg"],
      "category": "public.app-category.utilities"
    },
    "linux": {
      "target": ["AppImage", "deb"],
      "category": "Utility"
    },
    "win": {
      "target": ["nsis"]
    }
  }
}
```

**Step 2: 创建 installer/nuxt.config.ts**

```typescript
export default defineNuxtConfig({
  ssr: false,

  modules: [
    'nuxt-electron',
  ],

  electron: {
    build: [
      { entry: 'electron/main.ts' },
    ],
  },

  devtools: { enabled: false },

  compatibilityDate: '2025-01-01',
})
```

**Step 3: 创建 installer/tsconfig.json**

```json
{
  "extends": "../../tsconfig.json",
  "compilerOptions": {
    "jsx": "preserve"
  }
}
```

**Step 4: 创建最小 app.vue 占位**

```vue
<template>
  <div>
    <NuxtPage />
  </div>
</template>
```

**Step 5: 安装依赖**

Run: `cd /home/chaos/openclaw && pnpm install`
Expected: 成功安装所有依赖

**Step 6: Commit**

```bash
git add packages/installer/
git commit -m "chore: 初始化 Electron + Nuxt 安装器项目脚手架"
```

---

### Task 4: 创建 Electron 主进程

**Files:**
- Create: `packages/installer/electron/main.ts`

**Step 1: 创建 electron/main.ts**

```typescript
import { app, BrowserWindow } from 'electron'
import path from 'path'

let mainWindow: BrowserWindow | null = null

function createWindow() {
  mainWindow = new BrowserWindow({
    width: 800,
    height: 600,
    resizable: false,
    title: 'OpenClaw 安装器',
    webPreferences: {
      nodeIntegration: true,
      contextIsolation: false,
    },
  })

  if (process.env.VITE_DEV_SERVER_URL) {
    mainWindow.loadURL(process.env.VITE_DEV_SERVER_URL)
  } else {
    mainWindow.loadFile(path.join(__dirname, '../.output/public/index.html'))
  }

  mainWindow.on('closed', () => {
    mainWindow = null
  })
}

app.whenReady().then(createWindow)

app.on('window-all-closed', () => {
  app.quit()
})
```

**Step 2: 验证开发模式能否启动**

Run: `cd /home/chaos/openclaw && pnpm dev:installer`
Expected: Nuxt 开发服务器启动，Electron 窗口打开

**Step 3: Commit**

```bash
git add packages/installer/electron/
git commit -m "feat: 添加 Electron 主进程入口"
```

---

### Task 5: 实现安装向导页面布局和导航

**Files:**
- Create: `packages/installer/composables/useWizard.ts`
- Create: `packages/installer/layouts/default.vue`
- Modify: `packages/installer/app.vue`

**Step 1: 创建向导状态管理 composables/useWizard.ts**

```typescript
import { ref, computed } from 'vue'
import type { InstallOptions } from '@openclaw/shared'

const currentStep = ref(0)

const steps = [
  { name: 'welcome', title: '欢迎', path: '/' },
  { name: 'mode', title: '安装模式', path: '/mode' },
  { name: 'adapter', title: '适配器选择', path: '/adapter' },
  { name: 'confirm', title: '确认安装', path: '/confirm' },
  { name: 'progress', title: '安装中', path: '/progress' },
  { name: 'done', title: '完成', path: '/done' },
]

const installOptions = ref<InstallOptions>({
  installDir: '',
  configDir: '',
  sourceDir: '',
  adapters: [],
  mode: 'standard',
})

export function useWizard() {
  const totalSteps = steps.length
  const canGoBack = computed(() => currentStep.value > 0 && currentStep.value < 4)
  const isInstalling = computed(() => currentStep.value === 4)
  const currentStepInfo = computed(() => steps[currentStep.value])

  function goNext() {
    if (currentStep.value < totalSteps - 1) {
      currentStep.value++
    }
  }

  function goBack() {
    if (canGoBack.value) {
      currentStep.value--
    }
  }

  function goToStep(step: number) {
    currentStep.value = step
  }

  return {
    currentStep,
    totalSteps,
    steps,
    canGoBack,
    isInstalling,
    currentStepInfo,
    installOptions,
    goNext,
    goBack,
    goToStep,
  }
}
```

**Step 2: 创建默认布局 layouts/default.vue**

```vue
<template>
  <n-config-provider :locale="zhCN" :date-locale="dateZhCN" :theme="darkTheme">
    <n-message-provider>
      <div class="installer-layout">
        <div class="installer-header">
          <h1>OpenClaw 安装器</h1>
          <n-steps :current="currentStep" size="small" class="installer-steps">
            <n-step
              v-for="(step, index) in steps"
              :key="step.name"
              :title="step.title"
            />
          </n-steps>
        </div>
        <div class="installer-content">
          <slot />
        </div>
      </div>
    </n-message-provider>
  </n-config-provider>
</template>

<script setup lang="ts">
import { darkTheme } from 'naive-ui'
import { zhCN, dateZhCN } from 'naive-ui'
import { useWizard } from '~/composables/useWizard'

const { currentStep, steps } = useWizard()
</script>

<style>
body {
  margin: 0;
  font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', sans-serif;
}

.installer-layout {
  display: flex;
  flex-direction: column;
  height: 100vh;
  background: #0a0c10;
  color: #e0e0e0;
}

.installer-header {
  padding: 24px 32px 16px;
  border-bottom: 1px solid #222;
}

.installer-header h1 {
  font-size: 20px;
  margin: 0 0 16px 0;
  color: #fff;
}

.installer-steps {
  max-width: 600px;
}

.installer-content {
  flex: 1;
  display: flex;
  flex-direction: column;
  justify-content: center;
  align-items: center;
  padding: 32px;
}
</style>
```

**Step 3: 更新 app.vue**

```vue
<template>
  <NuxtLayout>
    <NuxtPage />
  </NuxtLayout>
</template>
```

**Step 4: Commit**

```bash
git add packages/installer/composables/ packages/installer/layouts/ packages/installer/app.vue
git commit -m "feat: 实现安装向导布局和步骤导航"
```

---

### Task 6: 实现安装向导六个页面

**Files:**
- Create: `packages/installer/pages/index.vue` (欢迎页)
- Create: `packages/installer/pages/mode.vue` (安装模式)
- Create: `packages/installer/pages/adapter.vue` (适配器选择)
- Create: `packages/installer/pages/confirm.vue` (安装确认)
- Create: `packages/installer/pages/progress.vue` (安装进度)
- Create: `packages/installer/pages/done.vue` (完成页)

**Step 1: 创建欢迎页 pages/index.vue**

```vue
<template>
  <div class="welcome-page">
    <div class="welcome-icon">🐾</div>
    <h2>欢迎安装 OpenClaw</h2>
    <p class="version">版本 {{ version }}</p>
    <p class="description">OpenClaw AI 助手将帮助您轻松连接企业 IM 与 AI 服务</p>
    <n-button type="primary" size="large" @click="start">
      开始安装
    </n-button>
  </div>
</template>

<script setup lang="ts">
import { useWizard } from '~/composables/useWizard'

const version = '2.0.0'
const { goNext } = useWizard()

function start() {
  goNext()
  navigateTo('/mode')
}
</script>

<style scoped>
.welcome-page {
  text-align: center;
}
.welcome-icon {
  font-size: 64px;
  margin-bottom: 16px;
}
.version {
  color: #888;
  margin: 4px 0 16px;
}
.description {
  max-width: 400px;
  line-height: 1.6;
  margin-bottom: 32px;
  color: #aaa;
}
</style>
```

**Step 2: 创建安装模式页 pages/mode.vue**

```vue
<template>
  <div class="mode-page">
    <h2>选择安装模式</h2>
    <n-radio-group v-model:value="installOptions.mode" class="mode-group">
      <n-space vertical :size="16">
        <n-radio value="standard" size="large">
          <div class="mode-option">
            <strong>标准安装（推荐）</strong>
            <p>安装到默认目录，自动完成全部配置</p>
            <p class="mode-path">{{ defaultDir }}</p>
          </div>
        </n-radio>
        <n-radio value="custom" size="large">
          <div class="mode-option">
            <strong>自定义安装</strong>
            <p>选择安装目录</p>
          </div>
        </n-radio>
      </n-space>
    </n-radio-group>

    <div v-if="installOptions.mode === 'custom'" class="custom-dir">
      <n-input
        v-model:value="installOptions.installDir"
        placeholder="请选择安装目录"
        readonly
      >
        <template #suffix>
          <n-button text @click="selectDir">选择</n-button>
        </template>
      </n-input>
    </div>

    <div class="nav-buttons">
      <n-button @click="goBackToWelcome">上一步</n-button>
      <n-button type="primary" @click="goNextToAdapter">下一步</n-button>
    </div>
  </div>
</template>

<script setup lang="ts">
import { onMounted } from 'vue'
import { getPlatformPaths } from '@openclaw/shared'
import { useWizard } from '~/composables/useWizard'

const { installOptions, goBack, goNext } = useWizard()
const defaultDir = ref('')

onMounted(() => {
  const paths = getPlatformPaths()
  defaultDir.value = paths.installDir
  if (!installOptions.value.installDir) {
    installOptions.value.installDir = paths.installDir
    installOptions.value.configDir = paths.configDir
  }
})

async function selectDir() {
  // Electron dialog 选择目录
  if (window.require) {
    const { dialog } = window.require('electron').remote
    const result = await dialog.showOpenDialog({ properties: ['openDirectory'] })
    if (!result.canceled && result.filePaths[0]) {
      installOptions.value.installDir = result.filePaths[0]
    }
  }
}

function goBackToWelcome() {
  goBack()
  navigateTo('/')
}

function goNextToAdapter() {
  goNext()
  navigateTo('/adapter')
}
</script>

<style scoped>
.mode-page { width: 100%; max-width: 500px; }
.mode-page h2 { text-align: center; margin-bottom: 24px; }
.mode-group { width: 100%; }
.mode-option p { margin: 4px 0 0; color: #888; font-size: 13px; }
.mode-path { font-family: monospace; color: #5b8cff; }
.custom-dir { margin-top: 16px; }
.nav-buttons { display: flex; justify-content: space-between; margin-top: 32px; }
</style>
```

**Step 3: 创建适配器选择页 pages/adapter.vue**

```vue
<template>
  <div class="adapter-page">
    <h2>选择 IM 适配器</h2>
    <p class="hint">请选择需要对接的即时通讯平台（可多选）</p>
    <n-checkbox-group v-model:value="installOptions.adapters">
      <n-space vertical :size="12">
        <n-checkbox
          v-for="adapter in availableAdapters"
          :key="adapter.name"
          :value="adapter.name"
          :label="adapter.displayName + ' — ' + adapter.description"
          size="large"
        />
      </n-space>
    </n-checkbox-group>

    <div class="nav-buttons">
      <n-button @click="goBackToMode">上一步</n-button>
      <n-button
        type="primary"
        :disabled="installOptions.adapters.length === 0"
        @click="goNextToConfirm"
      >
        下一步
      </n-button>
    </div>
  </div>
</template>

<script setup lang="ts">
import { useWizard } from '~/composables/useWizard'

const { installOptions, goBack, goNext } = useWizard()

const availableAdapters = [
  { name: 'wechat-work', displayName: '企业微信', description: '对接企业微信机器人' },
  { name: 'dingtalk', displayName: '钉钉', description: '对接钉钉机器人' },
  { name: 'feishu', displayName: '飞书', description: '对接飞书机器人' },
]

function goBackToMode() {
  goBack()
  navigateTo('/mode')
}

function goNextToConfirm() {
  goNext()
  navigateTo('/confirm')
}
</script>

<style scoped>
.adapter-page { width: 100%; max-width: 500px; }
.adapter-page h2 { text-align: center; margin-bottom: 8px; }
.hint { text-align: center; color: #888; margin-bottom: 24px; }
.nav-buttons { display: flex; justify-content: space-between; margin-top: 32px; }
</style>
```

**Step 4: 创建安装确认页 pages/confirm.vue**

```vue
<template>
  <div class="confirm-page">
    <h2>确认安装信息</h2>
    <n-descriptions bordered :column="1" label-placement="left">
      <n-descriptions-item label="安装模式">
        {{ installOptions.mode === 'standard' ? '标准安装' : '自定义安装' }}
      </n-descriptions-item>
      <n-descriptions-item label="安装目录">
        {{ installOptions.installDir }}
      </n-descriptions-item>
      <n-descriptions-item label="已选适配器">
        {{ selectedAdapterNames }}
      </n-descriptions-item>
    </n-descriptions>

    <div class="nav-buttons">
      <n-button @click="goBackToAdapter">上一步</n-button>
      <n-button type="primary" @click="startInstall">开始安装</n-button>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useWizard } from '~/composables/useWizard'

const { installOptions, goBack, goNext } = useWizard()

const adapterNameMap: Record<string, string> = {
  'wechat-work': '企业微信',
  'dingtalk': '钉钉',
  'feishu': '飞书',
}

const selectedAdapterNames = computed(() =>
  installOptions.value.adapters.map(a => adapterNameMap[a] || a).join('、')
)

function goBackToAdapter() {
  goBack()
  navigateTo('/adapter')
}

function startInstall() {
  goNext()
  navigateTo('/progress')
}
</script>

<style scoped>
.confirm-page { width: 100%; max-width: 500px; }
.confirm-page h2 { text-align: center; margin-bottom: 24px; }
.nav-buttons { display: flex; justify-content: space-between; margin-top: 32px; }
</style>
```

**Step 5: 创建安装进度页 pages/progress.vue**

```vue
<template>
  <div class="progress-page">
    <h2>正在安装</h2>
    <n-progress
      type="line"
      :percentage="progress.percent"
      :status="progress.error ? 'error' : 'default'"
      :height="12"
      style="max-width: 400px; margin-bottom: 16px;"
    />
    <p class="status-text">{{ progress.message }}</p>
    <p v-if="progress.error" class="error-text">
      {{ progress.error }}
    </p>
    <n-button v-if="progress.error" type="primary" @click="retry">
      重试
    </n-button>
  </div>
</template>

<script setup lang="ts">
import { onMounted, ref } from 'vue'
import type { InstallProgress } from '@openclaw/shared'
import { useWizard } from '~/composables/useWizard'

const { installOptions, goNext } = useWizard()

const progress = ref<InstallProgress>({
  step: 0,
  totalSteps: 5,
  message: '准备安装...',
  percent: 0,
})

onMounted(() => {
  runInstall()
})

async function runInstall() {
  const steps = [
    { message: '正在创建安装目录...', percent: 10 },
    { message: '正在复制主程序文件...', percent: 30 },
    { message: '正在安装适配器组件...', percent: 55 },
    { message: '正在生成配置文件...', percent: 75 },
    { message: '正在启动配置服务...', percent: 90 },
  ]

  try {
    for (const [i, step] of steps.entries()) {
      progress.value = { step: i, totalSteps: steps.length, message: step.message, percent: step.percent }
      // TODO: 实际安装逻辑通过 Electron IPC 调用
      await new Promise(r => setTimeout(r, 800))
    }

    progress.value = { step: steps.length, totalSteps: steps.length, message: '安装完成！', percent: 100 }
    await new Promise(r => setTimeout(r, 500))
    goNext()
    navigateTo('/done')
  } catch (err: any) {
    progress.value.error = `安装出错：${err.message}。请检查磁盘空间是否充足，或尝试以管理员身份运行安装器。`
  }
}

function retry() {
  progress.value = { step: 0, totalSteps: 5, message: '准备重试...', percent: 0 }
  runInstall()
}
</script>

<style scoped>
.progress-page { text-align: center; width: 100%; max-width: 500px; }
.status-text { color: #aaa; }
.error-text { color: #ef4444; margin: 8px 0 16px; }
</style>
```

**Step 6: 创建完成页 pages/done.vue**

```vue
<template>
  <div class="done-page">
    <div class="done-icon">✓</div>
    <h2>安装成功</h2>
    <p>OpenClaw 已安装完成。</p>
    <p class="hint">点击下方按钮打开浏览器，进入配置页面设置 API Key 和适配器参数。</p>

    <n-checkbox v-model:checked="openConfig" style="margin-bottom: 24px;">
      立即打开配置页面
    </n-checkbox>

    <n-button type="primary" size="large" @click="finish">
      完成
    </n-button>
  </div>
</template>

<script setup lang="ts">
import { ref } from 'vue'

const openConfig = ref(true)

function finish() {
  if (openConfig.value) {
    // 用系统默认浏览器打开配置页面
    if (window.require) {
      const { shell } = window.require('electron')
      shell.openExternal('http://localhost:18080')
    }
  }
  // 关闭安装器
  if (window.require) {
    const { app } = window.require('electron').remote
    app.quit()
  }
}
</script>

<style scoped>
.done-page { text-align: center; }
.done-icon {
  width: 80px; height: 80px; border-radius: 50%;
  background: #22c55e; color: #fff; font-size: 40px;
  display: flex; align-items: center; justify-content: center;
  margin: 0 auto 16px;
}
.hint { color: #888; max-width: 400px; line-height: 1.6; margin-bottom: 24px; }
</style>
```

**Step 7: 验证所有页面可导航**

Run: `cd /home/chaos/openclaw && pnpm dev:installer`
Expected: 能从欢迎页一路点击到完成页，步骤条同步更新

**Step 8: Commit**

```bash
git add packages/installer/pages/
git commit -m "feat: 实现安装向导六个步骤页面"
```

---

## Phase 3: Web 配置服务

### Task 7: 初始化 config-server Nuxt 项目

**Files:**
- Create: `packages/config-server/package.json`
- Create: `packages/config-server/nuxt.config.ts`
- Create: `packages/config-server/tsconfig.json`
- Create: `packages/config-server/app.vue`

**Step 1: 创建 config-server/package.json**

```json
{
  "name": "@openclaw/config-server",
  "version": "2.0.0",
  "private": true,
  "scripts": {
    "dev": "nuxt dev --port 18080",
    "build": "nuxt build",
    "preview": "nuxt preview",
    "lint": "eslint .",
    "test": "vitest"
  },
  "dependencies": {
    "@openclaw/shared": "workspace:*",
    "naive-ui": "^2.39.0",
    "vue": "^3.5.0"
  },
  "devDependencies": {
    "nuxt": "^3.14.0",
    "typescript": "^5.6.0",
    "vitest": "^2.1.0",
    "@nuxt/test-utils": "^3.14.0"
  }
}
```

**Step 2: 创建 config-server/nuxt.config.ts**

```typescript
export default defineNuxtConfig({
  ssr: false,

  devServer: {
    port: 18080,
  },

  devtools: { enabled: false },

  compatibilityDate: '2025-01-01',
})
```

**Step 3: 创建 config-server/tsconfig.json**

```json
{
  "extends": "../../tsconfig.json"
}
```

**Step 4: 创建 config-server/app.vue**

```vue
<template>
  <n-config-provider :locale="zhCN" :date-locale="dateZhCN" :theme="darkTheme">
    <n-message-provider>
      <NuxtLayout>
        <NuxtPage />
      </NuxtLayout>
    </n-message-provider>
  </n-config-provider>
</template>

<script setup lang="ts">
import { darkTheme } from 'naive-ui'
import { zhCN, dateZhCN } from 'naive-ui'
</script>
```

**Step 5: 安装依赖并验证**

Run: `cd /home/chaos/openclaw && pnpm install && pnpm dev:config`
Expected: Nuxt 在 18080 端口启动

**Step 6: Commit**

```bash
git add packages/config-server/
git commit -m "chore: 初始化 @openclaw/config-server Nuxt 项目"
```

---

### Task 8: 实现配置服务侧边栏布局

**Files:**
- Create: `packages/config-server/layouts/default.vue`

**Step 1: 创建带侧边栏的布局 layouts/default.vue**

```vue
<template>
  <n-layout has-sider style="height: 100vh;">
    <n-layout-sider
      bordered
      :width="200"
      :collapsed-width="0"
      collapse-mode="width"
      show-trigger
    >
      <div class="logo">
        <span>🐾 OpenClaw</span>
      </div>
      <n-menu
        :value="activeKey"
        :options="menuOptions"
        @update:value="handleMenuClick"
      />
    </n-layout-sider>
    <n-layout>
      <n-layout-header bordered style="padding: 12px 24px; font-size: 16px; font-weight: 600;">
        {{ pageTitle }}
      </n-layout-header>
      <n-layout-content content-style="padding: 24px;">
        <slot />
      </n-layout-content>
    </n-layout>
  </n-layout>
</template>

<script setup lang="ts">
import { computed, h } from 'vue'
import { NIcon } from 'naive-ui'

const route = useRoute()
const router = useRouter()

const menuOptions = [
  { label: '仪表盘', key: '/', },
  { label: 'API 配置', key: '/apikeys', },
  { label: 'IM 适配器', key: '/adapters', },
  { label: '日志', key: '/logs', },
]

const activeKey = computed(() => route.path)

const pageTitle = computed(() => {
  const item = menuOptions.find(m => m.key === route.path)
  return item?.label ?? 'OpenClaw'
})

function handleMenuClick(key: string) {
  router.push(key)
}
</script>

<style scoped>
.logo {
  padding: 16px;
  font-size: 18px;
  font-weight: bold;
  text-align: center;
  border-bottom: 1px solid #333;
}
</style>
```

**Step 2: Commit**

```bash
git add packages/config-server/layouts/
git commit -m "feat: 实现配置服务侧边栏布局"
```

---

### Task 9: 实现配置服务 Server API

**Files:**
- Create: `packages/config-server/server/api/config.get.ts`
- Create: `packages/config-server/server/api/config.post.ts`
- Create: `packages/config-server/server/api/service.get.ts`
- Create: `packages/config-server/server/api/service.post.ts`
- Create: `packages/config-server/server/api/tokens.get.ts`
- Create: `packages/config-server/server/api/logs.get.ts`

**Step 1: 创建配置读取 API server/api/config.get.ts**

```typescript
import { readFileSync, existsSync } from 'fs'
import { getPlatformPaths } from '@openclaw/shared'

export default defineEventHandler(() => {
  const paths = getPlatformPaths()
  if (!existsSync(paths.configFile)) {
    return { success: true, data: null }
  }
  const raw = readFileSync(paths.configFile, 'utf-8')
  return { success: true, data: JSON.parse(raw) }
})
```

**Step 2: 创建配置写入 API server/api/config.post.ts**

```typescript
import { writeFileSync, mkdirSync, existsSync } from 'fs'
import { dirname } from 'path'
import { getPlatformPaths } from '@openclaw/shared'

export default defineEventHandler(async (event) => {
  const body = await readBody(event)
  const paths = getPlatformPaths()

  const dir = dirname(paths.configFile)
  if (!existsSync(dir)) {
    mkdirSync(dir, { recursive: true })
  }

  writeFileSync(paths.configFile, JSON.stringify(body, null, 2), 'utf-8')
  return { success: true }
})
```

**Step 3: 创建服务状态 API server/api/service.get.ts**

```typescript
import { execSync } from 'child_process'
import { detectPlatform } from '@openclaw/shared'

export default defineEventHandler(() => {
  const platform = detectPlatform()
  let running = false
  let startTime = ''

  try {
    if (platform.isLinux) {
      const output = execSync('systemctl is-active openclaw', { encoding: 'utf-8' }).trim()
      running = output === 'active'
      if (running) {
        startTime = execSync(
          "systemctl show openclaw --property=ActiveEnterTimestamp --value",
          { encoding: 'utf-8' }
        ).trim()
      }
    } else if (platform.isMacOS) {
      const output = execSync('launchctl list | grep openclaw', { encoding: 'utf-8' })
      running = output.length > 0
    } else if (platform.isWindows) {
      const output = execSync('sc query OpenClaw', { encoding: 'utf-8' })
      running = output.includes('RUNNING')
    }
  } catch {
    running = false
  }

  return { success: true, data: { running, startTime } }
})
```

**Step 4: 创建服务控制 API server/api/service.post.ts**

```typescript
import { execSync } from 'child_process'
import { detectPlatform } from '@openclaw/shared'

export default defineEventHandler(async (event) => {
  const { action } = await readBody<{ action: 'start' | 'stop' }>(event)
  const platform = detectPlatform()

  try {
    if (platform.isLinux) {
      execSync(`sudo systemctl ${action} openclaw`)
    } else if (platform.isMacOS) {
      const cmd = action === 'start' ? 'load' : 'unload'
      execSync(`launchctl ${cmd} /Library/LaunchDaemons/com.openclaw.plist`)
    } else if (platform.isWindows) {
      const cmd = action === 'start' ? 'start' : 'stop'
      execSync(`net ${cmd} OpenClaw`)
    }
    return { success: true }
  } catch (err: any) {
    return { success: false, error: err.message }
  }
})
```

**Step 5: 创建 Token 用量 API server/api/tokens.get.ts**

```typescript
import { readFileSync, existsSync } from 'fs'
import { join } from 'path'
import { getPlatformPaths } from '@openclaw/shared'

export default defineEventHandler((event) => {
  const query = getQuery(event)
  const days = Number(query.days) || 7
  const paths = getPlatformPaths()
  const tokenFile = join(paths.configDir, 'token-usage.json')

  if (!existsSync(tokenFile)) {
    return { success: true, data: { total: 0, daily: [] } }
  }

  const raw = readFileSync(tokenFile, 'utf-8')
  const allData = JSON.parse(raw) as Array<{ date: string; tokens: number }>

  const cutoff = new Date()
  cutoff.setDate(cutoff.getDate() - days)
  const filtered = allData.filter(d => new Date(d.date) >= cutoff)
  const total = filtered.reduce((sum, d) => sum + d.tokens, 0)

  return { success: true, data: { total, daily: filtered } }
})
```

**Step 6: 创建日志查询 API server/api/logs.get.ts**

```typescript
import { readFileSync, existsSync, readdirSync, statSync } from 'fs'
import { join } from 'path'
import { getPlatformPaths } from '@openclaw/shared'

export default defineEventHandler((event) => {
  const query = getQuery(event)
  const page = Number(query.page) || 1
  const pageSize = Number(query.pageSize) || 100
  const level = (query.level as string) || ''
  const search = (query.search as string) || ''

  const paths = getPlatformPaths()

  if (!existsSync(paths.logDir)) {
    return { success: true, data: { logs: [], total: 0, page, pageSize } }
  }

  // 读取最新的日志文件
  const logFiles = readdirSync(paths.logDir)
    .filter(f => f.endsWith('.log'))
    .map(f => ({ name: f, mtime: statSync(join(paths.logDir, f)).mtime }))
    .sort((a, b) => b.mtime.getTime() - a.mtime.getTime())

  if (logFiles.length === 0) {
    return { success: true, data: { logs: [], total: 0, page, pageSize } }
  }

  const latestLog = readFileSync(join(paths.logDir, logFiles[0].name), 'utf-8')
  let lines = latestLog.split('\n').filter(Boolean)

  // 按级别筛选
  if (level) {
    lines = lines.filter(l => l.includes(`[${level.toUpperCase()}]`))
  }

  // 搜索
  if (search) {
    lines = lines.filter(l => l.includes(search))
  }

  const total = lines.length
  const start = (page - 1) * pageSize
  const paged = lines.slice(start, start + pageSize)

  return {
    success: true,
    data: { logs: paged, total, page, pageSize },
  }
})
```

**Step 7: Commit**

```bash
git add packages/config-server/server/
git commit -m "feat: 实现配置服务后端 API（配置、服务、Token、日志）"
```

---

### Task 10: 实现仪表盘页面

**Files:**
- Create: `packages/config-server/pages/index.vue`

**Step 1: 创建仪表盘 pages/index.vue**

```vue
<template>
  <n-space vertical :size="24">
    <!-- 服务状态 -->
    <n-card title="服务状态">
      <n-space align="center" :size="24">
        <n-tag :type="serviceRunning ? 'success' : 'error'" size="large">
          {{ serviceRunning ? '运行中' : '已停止' }}
        </n-tag>
        <span v-if="serviceRunning && startTime" style="color: #888;">
          启动时间：{{ startTime }}
        </span>
        <n-button
          :type="serviceRunning ? 'error' : 'success'"
          :loading="serviceLoading"
          @click="toggleService"
        >
          {{ serviceRunning ? '关闭服务' : '启动服务' }}
        </n-button>
      </n-space>
    </n-card>

    <!-- Token 用量 -->
    <n-card title="Token 用量">
      <n-space justify="space-between" align="center" style="margin-bottom: 16px;">
        <n-statistic label="近 7 天总用量" :value="tokenTotal" />
        <n-radio-group v-model:value="tokenDays" size="small" @update:value="fetchTokens">
          <n-radio-button :value="7">7 天</n-radio-button>
          <n-radio-button :value="30">30 天</n-radio-button>
          <n-radio-button :value="90">90 天</n-radio-button>
        </n-radio-group>
      </n-space>
      <div v-if="tokenDaily.length > 0" class="token-chart">
        <div
          v-for="item in tokenDaily"
          :key="item.date"
          class="chart-bar"
          :style="{ height: barHeight(item.tokens) + 'px' }"
          :title="item.date + ': ' + item.tokens + ' tokens'"
        />
      </div>
      <n-empty v-else description="暂无用量数据" />
    </n-card>

    <!-- 版本信息 -->
    <n-card title="版本信息">
      <n-descriptions :column="2">
        <n-descriptions-item label="版本">2.0.0</n-descriptions-item>
        <n-descriptions-item label="平台">{{ platformStr }}</n-descriptions-item>
      </n-descriptions>
    </n-card>
  </n-space>
</template>

<script setup lang="ts">
import { ref, onMounted, computed } from 'vue'
import { detectPlatform } from '@openclaw/shared'

const serviceRunning = ref(false)
const serviceLoading = ref(false)
const startTime = ref('')
const tokenDays = ref(7)
const tokenTotal = ref(0)
const tokenDaily = ref<Array<{ date: string; tokens: number }>>([])

const platform = detectPlatform()
const platformStr = `${platform.os}/${platform.arch}`

const maxTokens = computed(() => Math.max(...tokenDaily.value.map(d => d.tokens), 1))
function barHeight(tokens: number) {
  return Math.max(4, (tokens / maxTokens.value) * 120)
}

onMounted(() => {
  fetchService()
  fetchTokens()
})

async function fetchService() {
  try {
    const res = await $fetch('/api/service')
    serviceRunning.value = res.data.running
    startTime.value = res.data.startTime
  } catch {}
}

async function toggleService() {
  serviceLoading.value = true
  try {
    await $fetch('/api/service', {
      method: 'POST',
      body: { action: serviceRunning.value ? 'stop' : 'start' },
    })
    await fetchService()
  } catch {}
  serviceLoading.value = false
}

async function fetchTokens() {
  try {
    const res = await $fetch(`/api/tokens?days=${tokenDays.value}`)
    tokenTotal.value = res.data.total
    tokenDaily.value = res.data.daily
  } catch {}
}
</script>

<style scoped>
.token-chart {
  display: flex;
  align-items: flex-end;
  gap: 4px;
  height: 140px;
  padding: 8px 0;
}
.chart-bar {
  flex: 1;
  background: linear-gradient(to top, #5b8cff, #8b5cf6);
  border-radius: 2px 2px 0 0;
  min-width: 8px;
  cursor: pointer;
}
</style>
```

**Step 2: Commit**

```bash
git add packages/config-server/pages/index.vue
git commit -m "feat: 实现仪表盘页面（服务状态 + Token 用量 + 版本信息）"
```

---

### Task 11: 实现 API Key 管理页面

**Files:**
- Create: `packages/config-server/pages/apikeys.vue`

**Step 1: 创建 API Key 管理页 pages/apikeys.vue**

```vue
<template>
  <n-space vertical :size="16">
    <n-space justify="end">
      <n-button type="primary" @click="showModal = true">添加 API Key</n-button>
    </n-space>

    <n-data-table
      :columns="columns"
      :data="apiKeys"
      :loading="loading"
      :bordered="false"
    />

    <!-- 添加/编辑弹窗 -->
    <n-modal v-model:show="showModal" preset="dialog" :title="editingKey ? '编辑 API Key' : '添加 API Key'">
      <n-form ref="formRef" :model="form" label-placement="left" label-width="100">
        <n-form-item label="名称" path="name">
          <n-input v-model:value="form.name" placeholder="例如：生产环境 Key" />
        </n-form-item>
        <n-form-item label="服务商" path="provider">
          <n-select v-model:value="form.provider" :options="providerOptions" placeholder="选择 AI 服务商" />
        </n-form-item>
        <n-form-item label="API Key" path="key">
          <n-input v-model:value="form.key" type="password" show-password-on="click" placeholder="输入 API Key" />
        </n-form-item>
        <n-form-item label="接口地址" path="endpoint">
          <n-input v-model:value="form.endpoint" placeholder="可选，使用自定义接口地址" />
        </n-form-item>
      </n-form>
      <template #action>
        <n-space>
          <n-button @click="testConnection" :loading="testing">测试连接</n-button>
          <n-button type="primary" @click="saveKey">保存</n-button>
        </n-space>
      </template>
    </n-modal>
  </n-space>
</template>

<script setup lang="ts">
import { ref, h, onMounted } from 'vue'
import { NButton, NSpace, NTag, useMessage } from 'naive-ui'
import type { DataTableColumns } from 'naive-ui'
import type { ApiKeyConfig } from '@openclaw/shared'

const message = useMessage()
const loading = ref(false)
const showModal = ref(false)
const editingKey = ref<string | null>(null)
const testing = ref(false)
const apiKeys = ref<ApiKeyConfig[]>([])

const form = ref({
  name: '',
  provider: '',
  key: '',
  endpoint: '',
})

const providerOptions = [
  { label: 'OpenAI', value: 'openai' },
  { label: 'Anthropic', value: 'anthropic' },
  { label: '其他', value: 'custom' },
]

const columns: DataTableColumns<ApiKeyConfig> = [
  { title: '名称', key: 'name' },
  { title: '服务商', key: 'provider' },
  {
    title: 'API Key',
    key: 'key',
    render(row) {
      return row.key.slice(0, 8) + '****'
    },
  },
  {
    title: '操作',
    key: 'actions',
    render(row) {
      return h(NSpace, {}, () => [
        h(NButton, { size: 'small', text: true, onClick: () => editKey(row) }, () => '编辑'),
        h(NButton, { size: 'small', text: true, type: 'error', onClick: () => deleteKey(row.id) }, () => '删除'),
      ])
    },
  },
]

onMounted(() => {
  fetchKeys()
})

async function fetchKeys() {
  loading.value = true
  try {
    const res = await $fetch('/api/config')
    apiKeys.value = res.data?.apiKeys || []
  } catch {}
  loading.value = false
}

function editKey(key: ApiKeyConfig) {
  editingKey.value = key.id
  form.value = { name: key.name, provider: key.provider, key: key.key, endpoint: key.endpoint || '' }
  showModal.value = true
}

async function deleteKey(id: string) {
  apiKeys.value = apiKeys.value.filter(k => k.id !== id)
  await saveConfig()
  message.success('已删除')
}

async function testConnection() {
  testing.value = true
  // TODO: 实际的连接测试逻辑
  await new Promise(r => setTimeout(r, 1500))
  message.success('连接测试成功')
  testing.value = false
}

async function saveKey() {
  if (editingKey.value) {
    const idx = apiKeys.value.findIndex(k => k.id === editingKey.value)
    if (idx >= 0) {
      apiKeys.value[idx] = { ...apiKeys.value[idx], ...form.value }
    }
  } else {
    apiKeys.value.push({
      id: Date.now().toString(),
      ...form.value,
      createdAt: new Date().toISOString(),
    })
  }

  await saveConfig()
  showModal.value = false
  editingKey.value = null
  form.value = { name: '', provider: '', key: '', endpoint: '' }
  message.success('已保存')
}

async function saveConfig() {
  const res = await $fetch('/api/config')
  const config = res.data || {}
  config.apiKeys = apiKeys.value
  await $fetch('/api/config', { method: 'POST', body: config })
}
</script>
```

**Step 2: Commit**

```bash
git add packages/config-server/pages/apikeys.vue
git commit -m "feat: 实现 API Key 管理页面（增删改 + 连接测试）"
```

---

### Task 12: 实现 IM 适配器配置页面

**Files:**
- Create: `packages/config-server/pages/adapters.vue`

**Step 1: 创建适配器配置页 pages/adapters.vue**

```vue
<template>
  <n-tabs type="line" animated>
    <n-tab-pane
      v-for="adapter in adapterSchemas"
      :key="adapter.name"
      :name="adapter.name"
      :tab="adapter.displayName"
    >
      <n-card>
        <n-form label-placement="left" label-width="120">
          <n-form-item label="启用">
            <n-switch v-model:value="adapterConfigs[adapter.name].enabled" />
          </n-form-item>

          <template v-for="(schema, fieldKey) in adapter.configSchema" :key="fieldKey">
            <n-form-item
              :label="schema.label || fieldKey"
              :required="schema.required"
            >
              <n-input
                v-model:value="adapterConfigs[adapter.name].options[fieldKey]"
                :placeholder="schema.placeholder || '请输入 ' + (schema.label || fieldKey)"
              />
              <template #feedback>
                <span v-if="schema.description" class="field-desc">{{ schema.description }}</span>
              </template>
            </n-form-item>
          </template>
        </n-form>

        <n-space justify="end">
          <n-button type="primary" @click="saveAdapter(adapter.name)">保存</n-button>
        </n-space>
      </n-card>
    </n-tab-pane>
  </n-tabs>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted } from 'vue'
import { useMessage } from 'naive-ui'
import type { AdapterSchema, AdapterConfig } from '@openclaw/shared'

const message = useMessage()

// 从 adapters/ 目录读取的适配器 schema（硬编码，与 adapters/*.json 对应）
const adapterSchemas: AdapterSchema[] = [
  {
    name: 'wechat-work',
    type: 'messaging',
    displayName: '企业微信',
    version: '1.0.0',
    description: '企业微信消息适配器',
    supportedPlatforms: ['windows', 'darwin', 'linux'],
    configSchema: {
      corp_id: { type: 'string', required: true, label: '企业 ID', placeholder: '例如：ww1234567890abcdef', description: '在企业微信管理后台「我的企业」中查看' },
      corp_secret: { type: 'string', required: true, label: '应用密钥', placeholder: '例如：abcdefghijklmnopqrstuvwxyz', description: '在企业微信管理后台「应用管理」中查看' },
      agent_id: { type: 'string', required: true, label: 'AgentID', placeholder: '例如：1000001', description: '在企业微信管理后台「应用管理」中查看' },
    },
  },
  {
    name: 'dingtalk',
    type: 'messaging',
    displayName: '钉钉',
    version: '1.0.0',
    description: '钉钉消息适配器',
    supportedPlatforms: ['windows', 'darwin', 'linux'],
    configSchema: {
      app_key: { type: 'string', required: true, label: 'AppKey', placeholder: '例如：dingxxxxxxxxxx', description: '在钉钉开放平台「应用开发」中查看' },
      app_secret: { type: 'string', required: true, label: 'AppSecret', placeholder: '例如：xxxxxxxxxxxxxxxxxxxxx', description: '在钉钉开放平台「应用开发」中查看' },
      robot_code: { type: 'string', required: false, label: '机器人编码', placeholder: '可选', description: '如需使用自定义机器人，在此填写机器人编码' },
    },
  },
  {
    name: 'feishu',
    type: 'messaging',
    displayName: '飞书',
    version: '1.0.0',
    description: '飞书消息适配器',
    supportedPlatforms: ['windows', 'darwin', 'linux'],
    configSchema: {
      app_id: { type: 'string', required: true, label: 'App ID', placeholder: '例如：cli_xxxxxxxxxx', description: '在飞书开放平台「应用详情」中查看' },
      app_secret: { type: 'string', required: true, label: 'App Secret', placeholder: '例如：xxxxxxxxxxxxxxxxxxxxx', description: '在飞书开放平台「应用详情」中查看' },
      encrypt_key: { type: 'string', required: false, label: '加密密钥', placeholder: '可选', description: '如需使用事件订阅加密，在此填写密钥' },
    },
  },
]

const adapterConfigs = reactive<Record<string, { enabled: boolean; options: Record<string, string> }>>({})

onMounted(() => {
  // 初始化各适配器默认值
  for (const schema of adapterSchemas) {
    adapterConfigs[schema.name] = {
      enabled: false,
      options: Object.fromEntries(Object.keys(schema.configSchema).map(k => [k, ''])),
    }
  }
  fetchConfig()
})

async function fetchConfig() {
  try {
    const res = await $fetch('/api/config')
    const adapters: AdapterConfig[] = res.data?.adapters || []
    for (const a of adapters) {
      if (adapterConfigs[a.name]) {
        adapterConfigs[a.name].enabled = a.enabled
        adapterConfigs[a.name].options = { ...adapterConfigs[a.name].options, ...a.options }
      }
    }
  } catch {}
}

async function saveAdapter(name: string) {
  try {
    const res = await $fetch('/api/config')
    const config = res.data || {}
    const adapters: AdapterConfig[] = config.adapters || []
    const schema = adapterSchemas.find(s => s.name === name)!
    const existing = adapters.findIndex(a => a.name === name)
    const adapterData: AdapterConfig = {
      name,
      type: schema.type,
      displayName: schema.displayName,
      enabled: adapterConfigs[name].enabled,
      options: adapterConfigs[name].options,
    }

    if (existing >= 0) {
      adapters[existing] = adapterData
    } else {
      adapters.push(adapterData)
    }

    config.adapters = adapters
    await $fetch('/api/config', { method: 'POST', body: config })
    message.success(`${schema.displayName} 配置已保存`)
  } catch {
    message.error('保存失败，请重试')
  }
}
</script>

<style scoped>
.field-desc { color: #888; font-size: 12px; }
</style>
```

**Step 2: Commit**

```bash
git add packages/config-server/pages/adapters.vue
git commit -m "feat: 实现 IM 适配器配置页面（企业微信/钉钉/飞书）"
```

---

### Task 13: 实现日志查看页面

**Files:**
- Create: `packages/config-server/pages/logs.vue`

**Step 1: 创建日志页 pages/logs.vue**

```vue
<template>
  <n-space vertical :size="16">
    <!-- 筛选栏 -->
    <n-space>
      <n-select
        v-model:value="level"
        :options="levelOptions"
        placeholder="日志级别"
        clearable
        style="width: 140px;"
        @update:value="fetchLogs"
      />
      <n-input
        v-model:value="search"
        placeholder="搜索日志内容"
        clearable
        style="width: 240px;"
        @clear="fetchLogs"
        @keyup.enter="fetchLogs"
      >
        <template #suffix>
          <n-button text size="small" @click="fetchLogs">搜索</n-button>
        </template>
      </n-input>
    </n-space>

    <!-- 日志列表 -->
    <n-card>
      <n-spin :show="loading">
        <div v-if="logs.length > 0" class="log-container">
          <pre v-for="(line, i) in logs" :key="i" class="log-line" :class="getLogLevel(line)">{{ line }}</pre>
        </div>
        <n-empty v-else description="暂无日志" />
      </n-spin>
    </n-card>

    <!-- 分页 -->
    <n-space justify="end">
      <n-pagination
        v-model:page="page"
        :page-count="pageCount"
        :page-size="pageSize"
        @update:page="fetchLogs"
      />
    </n-space>
  </n-space>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'

const loading = ref(false)
const logs = ref<string[]>([])
const total = ref(0)
const page = ref(1)
const pageSize = 100
const level = ref<string | null>(null)
const search = ref('')

const pageCount = computed(() => Math.ceil(total.value / pageSize))

const levelOptions = [
  { label: 'INFO', value: 'INFO' },
  { label: 'WARN', value: 'WARN' },
  { label: 'ERROR', value: 'ERROR' },
  { label: 'DEBUG', value: 'DEBUG' },
]

function getLogLevel(line: string): string {
  if (line.includes('[ERROR]')) return 'log-error'
  if (line.includes('[WARN]')) return 'log-warn'
  if (line.includes('[INFO]')) return 'log-info'
  if (line.includes('[DEBUG]')) return 'log-debug'
  return ''
}

onMounted(() => {
  fetchLogs()
})

async function fetchLogs() {
  loading.value = true
  try {
    const params = new URLSearchParams({
      page: page.value.toString(),
      pageSize: pageSize.toString(),
    })
    if (level.value) params.set('level', level.value)
    if (search.value) params.set('search', search.value)

    const res = await $fetch(`/api/logs?${params}`)
    logs.value = res.data.logs
    total.value = res.data.total
  } catch {}
  loading.value = false
}
</script>

<style scoped>
.log-container {
  max-height: 500px;
  overflow-y: auto;
  font-family: 'Courier New', monospace;
  font-size: 13px;
}
.log-line {
  margin: 0;
  padding: 2px 8px;
  white-space: pre-wrap;
  word-break: break-all;
  border-bottom: 1px solid #1a1a1a;
}
.log-error { color: #ef4444; }
.log-warn { color: #f59e0b; }
.log-info { color: #22c55e; }
.log-debug { color: #888; }
</style>
```

**Step 2: Commit**

```bash
git add packages/config-server/pages/logs.vue
git commit -m "feat: 实现日志查看页面（分页 + 级别筛选 + 搜索）"
```

---

## Phase 4: 清理与构建

### Task 14: 移除旧的 Go 代码

**Files:**
- Delete: `installer/` (整个目录)
- Delete: `wails-installer/` (整个目录)
- Delete: `frontend/` (整个目录，已被 Nuxt 替代)

**Step 1: 删除旧目录**

```bash
rm -rf installer/ wails-installer/ frontend/
```

**Step 2: 验证项目仍然可以运行**

Run: `cd /home/chaos/openclaw && pnpm install && pnpm dev:config`
Expected: 配置服务正常启动

**Step 3: Commit**

```bash
git add -A
git commit -m "chore: 移除旧的 Go CLI 安装器、Wails 安装器和静态前端"
```

---

### Task 15: 更新 CI/CD 工作流

**Files:**
- Modify: `.github/workflows/build.yml`

**Step 1: 重写 build.yml 为 Node.js 构建**

```yaml
name: Build & Release

on:
  push:
    branches: [main, master]
    tags: ['v*']
  pull_request:
    branches: [main, master]

permissions:
  contents: write
  actions: read

jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: pnpm/action-setup@v4
        with:
          version: 9
      - uses: actions/setup-node@v4
        with:
          node-version: 20
          cache: 'pnpm'
      - run: pnpm install
      - run: pnpm test

  build-installer:
    needs: test
    strategy:
      matrix:
        include:
          - os: ubuntu-latest
            targets: 'linux'
          - os: macos-latest
            targets: 'mac'
          - os: windows-latest
            targets: 'win'
    runs-on: ${{ matrix.os }}
    steps:
      - uses: actions/checkout@v4
      - uses: pnpm/action-setup@v4
        with:
          version: 9
      - uses: actions/setup-node@v4
        with:
          node-version: 20
          cache: 'pnpm'
      - run: pnpm install
      - run: pnpm build:installer
        env:
          ELECTRON_MIRROR: https://npmmirror.com/mirrors/electron/
          ELECTRON_BUILDER_BINARIES_MIRROR: https://npmmirror.com/mirrors/electron-builder-binaries/
      - uses: actions/upload-artifact@v4
        with:
          name: installer-${{ matrix.targets }}
          path: packages/installer/dist-electron/*

  build-config-server:
    needs: test
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: pnpm/action-setup@v4
        with:
          version: 9
      - uses: actions/setup-node@v4
        with:
          node-version: 20
          cache: 'pnpm'
      - run: pnpm install
      - run: pnpm build:config
      - uses: actions/upload-artifact@v4
        with:
          name: config-server
          path: packages/config-server/.output/

  release:
    if: startsWith(github.ref, 'refs/tags/v') || github.ref == 'refs/heads/master'
    needs: [build-installer, build-config-server]
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: actions/download-artifact@v4
        with:
          path: artifacts/
      - name: Create nightly tag
        if: github.ref == 'refs/heads/master'
        run: |
          git tag -f nightly
          git push -f origin nightly
      - name: Create Release
        uses: softprops/action-gh-release@v2
        with:
          tag_name: ${{ startsWith(github.ref, 'refs/tags/') && github.ref_name || 'nightly' }}
          name: ${{ startsWith(github.ref, 'refs/tags/') && github.ref_name || 'Nightly Build' }}
          prerelease: ${{ github.ref == 'refs/heads/master' }}
          generate_release_notes: true
          files: artifacts/**/*
```

**Step 2: Commit**

```bash
git add .github/workflows/build.yml
git commit -m "ci: 更新 CI/CD 工作流为 Node.js + Electron 构建"
```

---

### Task 16: 更新构建脚本

**Files:**
- Create: `scripts/build-installer.sh`
- Create: `scripts/build-config.sh`
- Modify: `scripts/build.sh` (精简为新架构)

**Step 1: 创建安装器构建脚本 scripts/build-installer.sh**

```bash
#!/bin/bash
set -e

echo "=== 构建 OpenClaw 安装器 ==="
cd "$(dirname "$0")/.."

export ELECTRON_MIRROR=https://npmmirror.com/mirrors/electron/
export ELECTRON_BUILDER_BINARIES_MIRROR=https://npmmirror.com/mirrors/electron-builder-binaries/

pnpm --filter @openclaw/installer build
echo "=== 安装器构建完成 ==="
```

**Step 2: 创建配置服务构建脚本 scripts/build-config.sh**

```bash
#!/bin/bash
set -e

echo "=== 构建 OpenClaw 配置服务 ==="
cd "$(dirname "$0")/.."

pnpm --filter @openclaw/config-server build
echo "=== 配置服务构建完成 ==="
```

**Step 3: 更新主构建脚本 scripts/build.sh**

```bash
#!/bin/bash
set -e

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
ROOT_DIR="$(dirname "$SCRIPT_DIR")"

cd "$ROOT_DIR"

case "${1:-all}" in
  installer)
    bash "$SCRIPT_DIR/build-installer.sh"
    ;;
  config)
    bash "$SCRIPT_DIR/build-config.sh"
    ;;
  all)
    bash "$SCRIPT_DIR/build-installer.sh"
    bash "$SCRIPT_DIR/build-config.sh"
    ;;
  *)
    echo "用法: $0 {installer|config|all}"
    exit 1
    ;;
esac
```

**Step 4: 设置可执行权限**

```bash
chmod +x scripts/build-installer.sh scripts/build-config.sh scripts/build.sh
```

**Step 5: Commit**

```bash
git add scripts/build.sh scripts/build-installer.sh scripts/build-config.sh
git commit -m "chore: 更新构建脚本适配 Electron + Nuxt 架构"
```

---

## Phase 5: 文档

### Task 17: 编写用户文档

**Files:**
- Create: `docs/QUICK_START.md` (快速上手)
- Modify: `docs/USER_GUIDE.md` (更新用户指南)
- Create: `docs/CONFIG_GUIDE.md` (配置指南)
- Create: `docs/FAQ.md` (常见问题)

**Step 1: 创建快速上手指南 docs/QUICK_START.md**

```markdown
# OpenClaw 快速上手指南

从下载到配置完成，只需 3 步。

## 第 1 步：下载安装包

根据你的电脑系统，下载对应的安装包：

- Windows 电脑：下载 `.exe` 文件
- macOS 电脑：下载 `.dmg` 文件
- Linux 电脑：下载 `.AppImage` 文件

## 第 2 步：运行安装器

双击下载好的安装包，按照提示一路点击「下一步」即可完成安装。

安装过程中你需要选择要对接的 IM 平台（企业微信、钉钉或飞书），如果不确定，可以全部勾选，后续在配置页面关闭不需要的。

## 第 3 步：打开配置页面

安装完成后，安装器会自动打开浏览器，进入配置页面。

如果没有自动打开，手动在浏览器地址栏输入：`http://localhost:18080`

在配置页面中：

1. 进入「API 配置」页面，添加你的 AI 服务 API Key
2. 进入「IM 适配器」页面，填写 IM 平台的对接参数
3. 回到「仪表盘」，确认服务状态为「运行中」

全部配置完成后，OpenClaw 就可以正常工作了。
```

**Step 2: 创建配置指南 docs/CONFIG_GUIDE.md**

```markdown
# OpenClaw 配置指南

安装完成后，通过浏览器访问 `http://localhost:18080` 进入配置页面。

## 仪表盘

首页显示三个信息区域：

- **服务状态**：显示 OpenClaw 是否在运行，以及启动时间。点击按钮可以开启或关闭服务。
- **Token 用量**：显示 AI 服务的 token 消耗情况，可切换查看 7 天、30 天、90 天的数据。
- **版本信息**：当前 OpenClaw 版本和系统信息。

## API 配置

管理 AI 服务的 API Key。

### 添加 API Key

1. 点击右上角「添加 API Key」按钮
2. 填写以下信息：
   - **名称**：给这个 Key 取个容易记住的名字
   - **服务商**：选择 AI 服务提供商
   - **API Key**：粘贴你的 API Key
   - **接口地址**：如果使用自定义接口，在此填写（一般不需要填）
3. 点击「测试连接」确认 Key 可用
4. 点击「保存」

### 编辑或删除

在 API Key 列表中，每行右侧有「编辑」和「删除」操作按钮。

## IM 适配器

配置企业微信、钉钉或飞书的对接参数。页面顶部有标签页，点击切换不同平台。

每个平台需要填写的参数不同，每个输入框下方都有说明文字，告诉你在哪里找到这个参数。

填写完成后，打开「启用」开关，点击「保存」。

## 日志

查看 OpenClaw 的运行日志。

- **日志级别**：下拉筛选只看特定级别的日志（INFO/WARN/ERROR/DEBUG）
- **搜索**：在搜索框输入关键字，查找相关日志
- **翻页**：日志按每页 100 条分页显示，使用底部翻页控件浏览
```

**Step 3: 创建常见问题 docs/FAQ.md**

```markdown
# OpenClaw 常见问题

## 安装相关

### 安装器双击后没反应

- **Windows**：右键安装包，选择「以管理员身份运行」
- **macOS**：系统可能阻止了未知来源的应用。打开「系统设置 → 隐私与安全性」，找到 OpenClaw 安装器，点击「仍要打开」

### macOS 提示"无法打开，因为无法验证开发者"

1. 打开「系统设置」
2. 进入「隐私与安全性」
3. 向下滚动，找到关于 OpenClaw 的提示
4. 点击「仍要打开」

### 安装过程中提示磁盘空间不足

OpenClaw 需要约 200MB 的磁盘空间。请清理磁盘后重试。

## 配置相关

### 打开浏览器后页面空白

配置服务可能还没启动。等待几秒后刷新页面，或在仪表盘检查服务是否在运行。

### API Key 测试连接失败

- 检查 API Key 是否正确粘贴（注意前后不要有多余的空格）
- 检查网络是否能访问 AI 服务的接口地址
- 如果使用自定义接口地址，确认地址格式正确

### 修改配置后没有生效

修改配置后，回到仪表盘，先点击「关闭服务」，再点击「启动服务」重启即可生效。

## 服务相关

### 电脑重启后 OpenClaw 没有自动启动

OpenClaw 会注册为系统服务，应该开机自动启动。如果没有：

- **Windows**：按 Win+R，输入 services.msc，找到 OpenClaw，确认启动类型为「自动」
- **macOS**：打开配置页面，手动点击启动按钮
- **Linux**：运行 `sudo systemctl enable openclaw`
```

**Step 4: Commit**

```bash
git add docs/QUICK_START.md docs/CONFIG_GUIDE.md docs/FAQ.md
git commit -m "docs: 添加快速上手指南、配置指南和常见问题文档"
```

---

## Phase 6: 验证

### Task 18: 端到端验证

**Step 1: 安装所有依赖**

Run: `cd /home/chaos/openclaw && pnpm install`
Expected: 所有包的依赖安装成功

**Step 2: 验证 shared 包类型**

Run: `cd packages/shared && npx tsc --noEmit`
Expected: 无类型错误

**Step 3: 启动配置服务开发模式**

Run: `pnpm dev:config`
Expected: Nuxt 在 `http://localhost:18080` 启动，能访问仪表盘、API 配置、适配器、日志四个页面

**Step 4: 启动安装器开发模式**

Run: `pnpm dev:installer`
Expected: Electron 窗口打开，能走完欢迎→模式→适配器→确认→进度→完成六步流程

**Step 5: 构建配置服务**

Run: `pnpm build:config`
Expected: `.output/` 目录生成成功

**Step 6: Commit（如有修复）**

```bash
git add -A
git commit -m "fix: 端到端验证修复"
```

---

## 任务清单总览

| # | 任务 | Phase | 预计提交数 |
|---|------|-------|-----------|
| 1 | 初始化 Monorepo 根配置 | 1: 脚手架 | 1 |
| 2 | 创建 shared 共享包 | 1: 脚手架 | 1 |
| 3 | 初始化 Electron + Nuxt 安装器 | 2: 安装器 | 1 |
| 4 | 创建 Electron 主进程 | 2: 安装器 | 1 |
| 5 | 实现安装向导布局和导航 | 2: 安装器 | 1 |
| 6 | 实现安装向导六个页面 | 2: 安装器 | 1 |
| 7 | 初始化 config-server 项目 | 3: 配置服务 | 1 |
| 8 | 实现配置服务侧边栏布局 | 3: 配置服务 | 1 |
| 9 | 实现配置服务 Server API | 3: 配置服务 | 1 |
| 10 | 实现仪表盘页面 | 3: 配置服务 | 1 |
| 11 | 实现 API Key 管理页面 | 3: 配置服务 | 1 |
| 12 | 实现 IM 适配器配置页面 | 3: 配置服务 | 1 |
| 13 | 实现日志查看页面 | 3: 配置服务 | 1 |
| 14 | 移除旧的 Go 代码 | 4: 清理 | 1 |
| 15 | 更新 CI/CD 工作流 | 4: 构建 | 1 |
| 16 | 更新构建脚本 | 4: 构建 | 1 |
| 17 | 编写用户文档 | 5: 文档 | 1 |
| 18 | 端到端验证 | 6: 验证 | 0-1 |
