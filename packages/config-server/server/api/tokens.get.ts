import { readFile } from 'node:fs/promises'
import { existsSync } from 'node:fs'
import { join } from 'node:path'
import { getPlatformPaths } from '@openclaw/shared'

export default defineEventHandler(async (event) => {
  try {
    const query = getQuery(event)
    const days = Number(query.days) || 7

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
    return { success: false, error: error.message }
  }
})
