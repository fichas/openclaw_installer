import { readFile } from 'node:fs/promises'
import { existsSync } from 'node:fs'
import { join } from 'node:path'
import { getPlatformPaths } from '@openclaw/shared'
import { tokensQuerySchema, validateQuery } from '../utils/validation'

export default defineEventHandler(async (event) => {
  try {
    const rawQuery = getQuery(event)

    // 使用 Zod 验证查询参数
    const query = validateQuery(tokensQuerySchema, rawQuery)
    const days = query.days ?? 7

    const paths = getPlatformPaths()
    const tokenFile = join(paths.configDir, 'token-usage.json')

    if (!existsSync(tokenFile)) {
      return { success: true, data: { total: 0, daily: [] } }
    }

    const content = await readFile(tokenFile, 'utf-8')
    const records: Array<{ date: string; tokens: number }> = JSON.parse(content)

    // 计算截止日期
    const cutoff = new Date()
    cutoff.setDate(cutoff.getDate() - days)
    const cutoffStr = cutoff.toISOString().split('T')[0]

    // 过滤并汇总
    const filtered = records.filter((r) => r.date >= cutoffStr)
    const total = filtered.reduce((sum, r) => sum + r.tokens, 0)

    return {
      success: true,
      data: {
        total,
        daily: filtered,
      },
    }
  } catch (error: any) {
    // 如果是验证错误，直接抛出
    if (error.statusCode) {
      throw error
    }
    throw createError({
      statusCode: 500,
      statusMessage: error.message || '读取 Token 用量失败',
    })
  }
})
