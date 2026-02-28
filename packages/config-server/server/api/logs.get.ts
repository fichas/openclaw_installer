import { readFile, readdir } from 'node:fs/promises'
import { existsSync } from 'node:fs'
import { join } from 'node:path'
import { getPlatformPaths } from '@openclaw/shared'
import { logsQuerySchema, validateQuery } from '../utils/validation'

export default defineEventHandler(async (event) => {
  try {
    const rawQuery = getQuery(event)

    // 使用 Zod 验证查询参数
    const query = validateQuery(logsQuerySchema, rawQuery)

    const paths = getPlatformPaths()
    const logDir = paths.logDir

    if (!existsSync(logDir)) {
      return { success: true, data: { logs: [], total: 0, page: query.page, pageSize: query.pageSize } }
    }

    // 找到最新的 .log 文件
    const files = await readdir(logDir)
    const logFiles = files
      .filter((f) => f.endsWith('.log'))
      .sort()
      .reverse()

    if (logFiles.length === 0) {
      return { success: true, data: { logs: [], total: 0, page: query.page, pageSize: query.pageSize } }
    }

    const latestLog = join(logDir, logFiles[0])
    const content = await readFile(latestLog, 'utf-8')
    let lines = content.split('\n').filter((line) => line.trim().length > 0)

    // 按日志级别过滤
    if (query.level) {
      lines = lines.filter((line) => line.toUpperCase().includes(query.level!.toUpperCase()))
    }

    // 按搜索关键词过滤
    if (query.search) {
      lines = lines.filter((line) => line.toLowerCase().includes(query.search!.toLowerCase()))
    }

    const total = lines.length

    // 分页（从最新开始）
    const start = (query.page - 1) * query.pageSize
    const paginatedLogs = lines.slice(start, start + query.pageSize)

    return {
      success: true,
      data: {
        logs: paginatedLogs,
        total,
        page: query.page,
        pageSize: query.pageSize,
      },
    }
  } catch (error: any) {
    // 如果是验证错误，直接抛出
    if (error.statusCode) {
      throw error
    }
    throw createError({
      statusCode: 500,
      statusMessage: error.message || '读取日志失败',
    })
  }
})
