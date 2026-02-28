import { exec } from 'node:child_process'
import { promisify } from 'node:util'
import { detectPlatform } from '@openclaw/shared'

const execAsync = promisify(exec)

export default defineEventHandler(async () => {
  try {
    const { platform } = detectPlatform()
    let running = false
    let startTime: string | undefined

    if (platform === 'linux') {
      try {
        const { stdout } = await execAsync('systemctl is-active openclaw')
        running = stdout.trim() === 'active'
        if (running) {
          const { stdout: showOutput } = await execAsync(
            'systemctl show openclaw --property=ActiveEnterTimestamp --value'
          )
          startTime = showOutput.trim()
        }
      } catch {
        running = false
      }
    } else if (platform === 'darwin') {
      try {
        const { stdout } = await execAsync('launchctl list | grep openclaw')
        running = stdout.trim().length > 0
        // macOS launchctl does not directly expose start time
      } catch {
        running = false
      }
    } else if (platform === 'windows') {
      try {
        const { stdout } = await execAsync('sc query OpenClaw')
        running = stdout.includes('RUNNING')
      } catch {
        running = false
      }
    }

    return { success: true, data: { running, startTime } }
  } catch (error: any) {
    return { success: false, error: error.message }
  }
})
