import { exec } from 'node:child_process'
import { promisify } from 'node:util'
import { detectPlatform } from '@openclaw/shared'

const execAsync = promisify(exec)

export default defineEventHandler(async (event) => {
  try {
    const body = await readBody(event)
    const action = body.action as 'start' | 'stop'

    if (!action || !['start', 'stop'].includes(action)) {
      return { success: false, error: '无效的操作，必须为 start 或 stop' }
    }

    const { platform } = detectPlatform()

    if (platform === 'linux') {
      await execAsync(`sudo systemctl ${action} openclaw`)
    } else if (platform === 'darwin') {
      if (action === 'start') {
        await execAsync('launchctl load ~/Library/LaunchAgents/com.openclaw.plist')
      } else {
        await execAsync('launchctl unload ~/Library/LaunchAgents/com.openclaw.plist')
      }
    } else if (platform === 'windows') {
      if (action === 'start') {
        await execAsync('sc start OpenClaw')
      } else {
        await execAsync('sc stop OpenClaw')
      }
    }

    return { success: true }
  } catch (error: any) {
    return { success: false, error: error.message }
  }
})
