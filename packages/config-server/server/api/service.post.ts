import { spawn } from 'node:child_process'
import { detectPlatform } from '@openclaw/shared'
import { serviceActionSchema, validateBody } from '../utils/validation'

// 使用 spawn 执行命令，避免 shell 注入
function spawnPromise(command: string, args: string[]): Promise<void> {
  return new Promise((resolve, reject) => {
    const child = spawn(command, args, { stdio: 'pipe' })

    let stderr = ''
    child.stderr?.on('data', (data) => {
      stderr += data.toString()
    })

    child.on('close', (code) => {
      if (code === 0) {
        resolve()
      } else {
        reject(new Error(stderr || `Command failed with exit code ${code}`))
      }
    })

    child.on('error', (err) => {
      reject(err)
    })
  })
}

export default defineEventHandler(async (event) => {
  try {
    const body = await readBody(event)

    // 使用 Zod 严格验证输入
    const { action } = validateBody(serviceActionSchema, body)

    const { platform } = detectPlatform()

    if (platform === 'linux') {
      // 使用 spawn 传递数组参数，避免 shell 注入
      await spawnPromise('sudo', ['systemctl', action, 'openclaw'])
    } else if (platform === 'darwin') {
      const plistPath = `${process.env.HOME}/Library/LaunchAgents/com.openclaw.plist`
      if (action === 'start') {
        await spawnPromise('launchctl', ['load', plistPath])
      } else {
        await spawnPromise('launchctl', ['unload', plistPath])
      }
    } else if (platform === 'windows') {
      await spawnPromise('sc', [action, 'OpenClaw'])
    }

    return { success: true }
  } catch (error: any) {
    // 如果是验证错误，直接抛出
    if (error.statusCode) {
      throw error
    }
    throw createError({
      statusCode: 500,
      statusMessage: error.message || '服务操作失败',
    })
  }
})