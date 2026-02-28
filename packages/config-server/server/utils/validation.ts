import { z } from 'zod'

// API Key 验证 Schema
export const apiKeySchema = z.object({
  id: z.string().min(1, 'ID 不能为空'),
  name: z.string().min(1, '名称不能为空').max(100, '名称过长'),
  provider: z.string().min(1, '服务商不能为空'),
  key: z.string().min(1, 'API Key 不能为空'),
  endpoint: z.string().url('无效的 URL 格式').optional().or(z.literal('')),
  createdAt: z.string().datetime('无效的时间格式'),
})

// 适配器配置验证 Schema
export const adapterConfigSchema = z.object({
  name: z.string().min(1, '适配器名称不能为空'),
  type: z.string().min(1, '类型不能为空'),
  displayName: z.string().min(1, '显示名称不能为空'),
  enabled: z.boolean(),
  options: z.record(z.string()).refine(
    (opts) => Object.keys(opts).length > 0,
    '配置选项不能为空'
  ),
})

// 服务控制验证 Schema
export const serviceActionSchema = z.object({
  action: z.enum(['start', 'stop'], {
    errorMap: () => ({ message: '操作必须是 start 或 stop' }),
  }),
})

// 日志查询参数验证 Schema
export const logsQuerySchema = z.object({
  page: z.coerce.number().int().min(1).default(1),
  pageSize: z.coerce.number().int().min(1).max(1000).default(100),
  level: z.enum(['INFO', 'WARN', 'ERROR', 'DEBUG']).optional(),
  search: z.string().max(500).optional(),
})

// Token 用量查询参数验证 Schema
export const tokensQuerySchema = z.object({
  days: z.coerce.number().int().min(1).max(365).default(7),
})

// 完整配置验证 Schema
export const configSchema = z.object({
  version: z.string().default('2.0.0'),
  platform: z.string().default('unknown'),
  server: z.object({
    host: z.string().default('0.0.0.0'),
    port: z.number().int().min(1).max(65535).default(18080),
    tls: z.boolean().default(false),
  }).default({
    host: '0.0.0.0',
    port: 18080,
    tls: false,
  }),
  adapters: z.array(adapterConfigSchema).default([]),
  apiKeys: z.array(apiKeySchema).default([]),
  settings: z.record(z.string()).default({}),
})

// 验证函数辅助工具
export function validateBody<T>(schema: z.ZodSchema<T>, body: unknown): T {
  const result = schema.safeParse(body)
  if (!result.success) {
    const issues = result.error.issues.map(i => `${i.path.join('.')}: ${i.message}`).join(', ')
    throw createError({
      statusCode: 400,
      statusMessage: `验证失败: ${issues}`,
    })
  }
  return result.data
}

export function validateQuery<T>(schema: z.ZodSchema<T>, query: unknown): T {
  const result = schema.safeParse(query)
  if (!result.success) {
    const issues = result.error.issues.map(i => `${i.path.join('.')}: ${i.message}`).join(', ')
    throw createError({
      statusCode: 400,
      statusMessage: `查询参数验证失败: ${issues}`,
    })
  }
  return result.data
}
