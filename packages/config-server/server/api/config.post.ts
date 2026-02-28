import { writeFile, mkdir, chmod } from 'node:fs/promises'
import { dirname } from 'node:path'
import { getPlatformPaths, encryptSensitiveData } from '@openclaw/shared'
import { configSchema, validateBody } from '../utils/validation'

export default defineEventHandler(async (event) => {
  try {
    const body = await readBody(event)

    // 使用 Zod 验证输入
    const validatedConfig = validateBody(configSchema, body)

    const paths = getPlatformPaths()
    const configFile = paths.configFile

    const encryptedConfig = encryptSensitiveData(validatedConfig)

    await mkdir(dirname(configFile), { recursive: true })
    await writeFile(configFile, JSON.stringify(encryptedConfig, null, 2), {
      encoding: 'utf-8',
      mode: 0o600,
    })
    await chmod(configFile, 0o600).catch(() => {})

    return { success: true }
  } catch (error: any) {
    // 如果是验证错误，直接抛出（会被 Nuxt 错误处理捕获）
    if (error.statusCode) {
      throw error
    }
    // 其他错误
    throw createError({
      statusCode: 500,
      statusMessage: error.message || '保存配置失败',
    })
  }
})
