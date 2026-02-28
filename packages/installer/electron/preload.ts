import { contextBridge, ipcRenderer, shell } from 'electron'

// 暴露给渲染进程的安全 API
contextBridge.exposeInMainWorld('electronAPI', {
  // 应用控制
  quitApp: () => ipcRenderer.send('app-quit'),

  // 打开外部链接
  openExternal: (url: string) => shell.openExternal(url),

  // 平台信息
  platform: process.platform,
})

// 类型声明，供渲染进程使用
declare global {
  interface Window {
    electronAPI: {
      quitApp: () => void
      openExternal: (url: string) => void
      platform: string
    }
  }
}
