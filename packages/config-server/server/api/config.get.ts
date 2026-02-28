import { readFile } from 'node:fs/promises'
import { existsSync } from 'node:fs'
import { getPlatformPaths, decryptSensitiveData } from '@openclaw/shared'

export default defineEventHandler(async () => {
  try {
    const paths = getPlatformPaths()
    const configFile = paths.configFile

    if (!existsSync(configFile)) {
      return { success: true, data: null }
    }

    const content = await readFile(configFile, 'utf-8')
    const data = JSON.parse(content)

    // 解密敏感数据
    const decryptedData = decryptSensitiveData(data)

    return { success: true, data: decryptedData }
  } catch (error: any) {
    throw createError({
      statusCode: 500,
      statusMessage: error.message || '读取配置失败',
    })
  }
})
