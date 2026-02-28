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
