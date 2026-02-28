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
