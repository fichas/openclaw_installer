import { readFile, readdir } from 'node:fs/promises'
import { existsSync } from 'node:fs'
import { join } from 'node:path'
import { getPlatformPaths } from '@openclaw/shared'

export default defineEventHandler(async (event) => {
  try {
    const query = getQuery(event)
    const page = Math.max(1, Number(query.page) || 1)
    const pageSize = Math.max(1, Math.min(500, Number(query.pageSize) || 100))
    const level = (query.level as string) || ''
    const search = (query.search as string) || ''

    const paths = getPlatformPaths()
    const logDir = paths.logDir

    if (!existsSync(logDir)) {
      return { success: true, data: { logs: [], total: 0, page, pageSize } }
    }

    // 找到最新的 .log 文件
    const files = await readdir(logDir)
    const logFiles = files
      .filter((f) => f.endsWith('.log'))
      .sort()
      .reverse()

    if (logFiles.length === 0) {
      return { success: true, data: { logs: [], total: 0, page, pageSize } }
    }

    const latestLog = join(logDir, logFiles[0])
    const content = await readFile(latestLog, 'utf-8')
    let lines = content.split('\n').filter((line) => line.trim().length > 0)

    // 按日志级别过滤
    if (level) {
      lines = lines.filter((line) => line.toUpperCase().includes(level.toUpperCase()))
    }

    // 按搜索关键词过滤
    if (search) {
      lines = lines.filter((line) => line.toLowerCase().includes(search.toLowerCase()))
    }

    const total = lines.length

    // 分页（从最新开始）
    const start = (page - 1) * pageSize
    const paginatedLogs = lines.slice(start, start + pageSize)

    return {
      success: true,
      data: {
        logs: paginatedLogs,
        total,
        page,
        pageSize,
      },
    }
  } catch (error: any) {
    return { success: false, error: error.message }
  }
})
