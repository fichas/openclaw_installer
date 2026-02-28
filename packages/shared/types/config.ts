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
