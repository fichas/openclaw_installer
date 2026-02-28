import { writeFile, mkdir } from 'node:fs/promises'
import { dirname } from 'node:path'
import { getPlatformPaths } from '@openclaw/shared'

export default defineEventHandler(async (event) => {
  try {
    const body = await readBody(event)
    const paths = getPlatformPaths()
    const configFile = paths.configFile

    await mkdir(dirname(configFile), { recursive: true })
    await writeFile(configFile, JSON.stringify(body, null, 2), 'utf-8')

    return { success: true }
  } catch (error: any) {
    return { success: false, error: error.message }
  }
})
