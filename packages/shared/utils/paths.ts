import { readFileSync, writeFileSync, existsSync, mkdirSync } from 'fs'
import { dirname } from 'path'
import type { OpenClawConfig } from '../types'
import { encryptSensitiveData, decryptSensitiveData } from './crypto'

export function loadConfig(configPath: string): OpenClawConfig | null {
  if (!existsSync(configPath)) return null
  const raw = readFileSync(configPath, 'utf-8')
  const config = JSON.parse(raw) as OpenClawConfig
  // 解密敏感数据
  return decryptSensitiveData(config)
}

export function saveConfig(configPath: string, config: OpenClawConfig): void {
  const dir = dirname(configPath)
  if (!existsSync(dir)) {
    mkdirSync(dir, { recursive: true })
  }
  // 加密敏感数据
  const encryptedConfig = encryptSensitiveData(config)
  writeFileSync(configPath, JSON.stringify(encryptedConfig, null, 2), 'utf-8')
}
