import { readFile } from 'node:fs/promises'
import { existsSync } from 'node:fs'
import { getPlatformPaths } from '@openclaw/shared'

export default defineEventHandler(async () => {
  try {
    const paths = getPlatformPaths()
    const configFile = paths.configFile

    if (!existsSync(configFile)) {
      return { success: true, data: null }
    }

    const content = await readFile(configFile, 'utf-8')
    const data = JSON.parse(content)
    return { success: true, data }
  } catch (error: any) {
    return { success: false, error: error.message }
  }
})
