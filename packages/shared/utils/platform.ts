import { platform, arch } from 'os'
import { join } from 'path'
import type { PlatformInfo, PlatformPaths, OS, Arch } from '../types'

export function detectPlatform(): PlatformInfo {
  const os = platform()
  const cpuArch = arch()

  const normalizedOS: OS = os === 'win32' ? 'windows' : os as OS
  const normalizedArch: Arch = cpuArch === 'x64' ? 'amd64' : 'arm64'

  return {
    os: normalizedOS,
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
