import { app, BrowserWindow, ipcMain, dialog } from 'electron'
import { existsSync } from 'node:fs'
import { mkdir, writeFile } from 'node:fs/promises'
import { join } from 'node:path'
import path from 'path'
import { spawn } from 'node:child_process'
import { getPlatformPaths } from '@openclaw/shared'

let mainWindow: BrowserWindow | null = null

interface InstallStartPayload {
  installDir: string
  configDir: string
  adapters: string[]
  mode: 'standard' | 'custom'
}

interface InstallProgressPayload {
  step: number
  totalSteps: number
  message: string
  percent: number
  error?: string
}

function emitInstallProgress(payload: InstallProgressPayload) {
  if (mainWindow && !mainWindow.isDestroyed()) {
    mainWindow.webContents.send('install-progress', payload)
  }
}

async function runInstall(payload: InstallStartPayload): Promise<void> {
  const totalSteps = 5

  emitInstallProgress({ step: 1, totalSteps, message: '正在创建安装目录...', percent: 10 })
  await mkdir(payload.installDir, { recursive: true })

  emitInstallProgress({ step: 2, totalSteps, message: '正在写入程序文件...', percent: 35 })
  await writeFile(
    join(payload.installDir, 'openclaw-installed.txt'),
    `OpenClaw installed at ${new Date().toISOString()}\nmode=${payload.mode}\nadapters=${payload.adapters.join(',')}`,
    'utf-8'
  )

  emitInstallProgress({ step: 3, totalSteps, message: '正在生成配置文件...', percent: 65 })
  await mkdir(payload.configDir, { recursive: true })
  const configFile = join(payload.configDir, 'config.json')
  await writeFile(
    configFile,
    JSON.stringify(
      {
        version: '2.0.0',
        platform: process.platform,
        server: { host: '0.0.0.0', port: 18080, tls: false },
        adapters: payload.adapters.map((name) => ({
          name,
          type: 'messaging',
          displayName: name,
          enabled: true,
          options: {},
        })),
        apiKeys: [],
        settings: {},
      },
      null,
      2
    ),
    { encoding: 'utf-8', mode: 0o600 }
  )

  emitInstallProgress({ step: 4, totalSteps, message: '正在准备配置服务...', percent: 85 })
  const configServerEntry = join(process.cwd(), 'packages', 'config-server', '.output', 'server', 'index.mjs')
  if (existsSync(configServerEntry)) {
    const child = spawn(process.execPath, [configServerEntry], {
      detached: true,
      stdio: 'ignore',
    })
    child.unref()
  }

  emitInstallProgress({ step: 5, totalSteps, message: '安装完成！', percent: 100 })
}

function createWindow() {
  mainWindow = new BrowserWindow({
    width: 800,
    height: 600,
    resizable: false,
    title: 'OpenClaw 安装器',
    webPreferences: {
      contextIsolation: true,
      nodeIntegration: false,
      preload: path.join(__dirname, 'preload.js'),
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

ipcMain.on('app-quit', () => {
  app.quit()
})

ipcMain.handle('get-platform-paths', () => {
  return getPlatformPaths()
})

ipcMain.handle('select-directory', async () => {
  const result = await dialog.showOpenDialog({
    properties: ['openDirectory'],
  })
  if (result.canceled || result.filePaths.length === 0) {
    return null
  }
  return result.filePaths[0]
})

ipcMain.handle('install-start', async (_event, payload: InstallStartPayload) => {
  try {
    await runInstall(payload)
    return { success: true }
  } catch (error: any) {
    emitInstallProgress({
      step: 0,
      totalSteps: 5,
      message: '安装失败',
      percent: 0,
      error: error?.message || '未知错误',
    })
    return { success: false, error: error?.message || '安装失败' }
  }
})
