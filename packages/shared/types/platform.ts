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
