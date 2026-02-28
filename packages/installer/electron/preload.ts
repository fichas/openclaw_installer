import { contextBridge, ipcRenderer, shell } from 'electron'

type InstallProgress = {
  step: number
  totalSteps: number
  message: string
  percent: number
  error?: string
}

type InstallOptions = {
  installDir: string
  configDir: string
  sourceDir: string
  adapters: string[]
  mode: 'standard' | 'custom'
}

type PlatformPaths = {
  installDir: string
  configDir: string
  configFile: string
  logDir: string
}

contextBridge.exposeInMainWorld('electronAPI', {
  quitApp: () => ipcRenderer.send('app-quit'),
  openExternal: (url: string) => shell.openExternal(url),
  platform: process.platform,
  getPlatformPaths: (): Promise<PlatformPaths> => ipcRenderer.invoke('get-platform-paths'),
  selectDirectory: (): Promise<string | null> => ipcRenderer.invoke('select-directory'),
  startInstall: (payload: InstallOptions): Promise<{ success: boolean; error?: string }> =>
    ipcRenderer.invoke('install-start', payload),
  onInstallProgress: (callback: (progress: InstallProgress) => void) => {
    const handler = (_event: unknown, data: InstallProgress) => callback(data)
    ipcRenderer.on('install-progress', handler)
    return () => ipcRenderer.removeListener('install-progress', handler)
  },
})

declare global {
  interface Window {
    electronAPI: {
      quitApp: () => void
      openExternal: (url: string) => void
      platform: string
      getPlatformPaths: () => Promise<PlatformPaths>
      selectDirectory: () => Promise<string | null>
      startInstall: (payload: InstallOptions) => Promise<{ success: boolean; error?: string }>
      onInstallProgress: (callback: (progress: InstallProgress) => void) => () => void
    }
  }
}
