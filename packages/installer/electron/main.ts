import { app, BrowserWindow, ipcMain } from 'electron'
import path from 'path'

let mainWindow: BrowserWindow | null = null

function createWindow() {
  mainWindow = new BrowserWindow({
    width: 800,
    height: 600,
    resizable: false,
    title: 'OpenClaw 安装器',
    webPreferences: {
      // 安全最佳实践：启用上下文隔离
      contextIsolation: true,
      // 禁用 Node 集成，通过 preload 暴露必要 API
      nodeIntegration: false,
      // 指定 preload 脚本
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

// IPC 处理器
ipcMain.on('app-quit', () => {
  app.quit()
})
