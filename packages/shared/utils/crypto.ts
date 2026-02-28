import { createCipheriv, createDecipheriv, randomBytes, scryptSync } from 'crypto'
import { homedir } from 'os'

// 敏感字段列表
export const SENSITIVE_FIELDS = ['key', 'corp_secret', 'app_secret', 'app_secret', 'encrypt_key']

// 检测字段名是否敏感
export function isSensitiveField(fieldName: string): boolean {
  const lowerField = fieldName.toLowerCase()
  return SENSITIVE_FIELDS.some(sensitive => lowerField.includes(sensitive))
}

// 生成派生密钥（基于机器和用户特定信息）
function getDerivedKey(): Buffer {
  // 使用用户主目录路径 + 固定 salt 作为密钥材料
  // 这样同一用户在同一台机器上可以解密，换用户或换机器则无法解密
  const keyMaterial = `${homedir()}:openclaw-secret-v1`
  return scryptSync(keyMaterial, 'openclaw-salt-fixed-v1', 32)
}

// 加密单个值
export function encryptValue(value: string): string {
  try {
    const key = getDerivedKey()
    const iv = randomBytes(16)
    const cipher = createCipheriv('aes-256-gcm', key, iv)

    let encrypted = cipher.update(value, 'utf8', 'base64')
    encrypted += cipher.final('base64')

    const authTag = cipher.getAuthTag()

    // 返回格式: iv:authTag:encryptedData (都是 base64)
    return `enc:${iv.toString('base64')}:${authTag.toString('base64')}:${encrypted}`
  } catch {
    // 加密失败时返回原值（降级处理）
    return value
  }
}

// 解密单个值
export function decryptValue(encryptedValue: string): string {
  // 如果不是加密格式，直接返回
  if (!encryptedValue.startsWith('enc:')) {
    return encryptedValue
  }

  try {
    const key = getDerivedKey()
    const parts = encryptedValue.split(':')

    if (parts.length !== 4) {
      return encryptedValue
    }

    const iv = Buffer.from(parts[1], 'base64')
    const authTag = Buffer.from(parts[2], 'base64')
    const encrypted = parts[3]

    const decipher = createDecipheriv('aes-256-gcm', key, iv)
    decipher.setAuthTag(authTag)

    let decrypted = decipher.update(encrypted, 'base64', 'utf8')
    decrypted += decipher.final('utf8')

    return decrypted
  } catch {
    // 解密失败时返回原值
    return encryptedValue
  }
}

// 深度加密对象中的敏感字段
export function encryptSensitiveData(obj: any): any {
  if (typeof obj !== 'object' || obj === null) {
    return obj
  }

  if (Array.isArray(obj)) {
    return obj.map(item => encryptSensitiveData(item))
  }

  const result: any = {}
  for (const [key, value] of Object.entries(obj)) {
    if (typeof value === 'string' && isSensitiveField(key) && value && !value.startsWith('enc:')) {
      // 加密敏感字段
      result[key] = encryptValue(value)
    } else if (typeof value === 'object' && value !== null) {
      // 递归处理嵌套对象
      result[key] = encryptSensitiveData(value)
    } else {
      result[key] = value
    }
  }

  return result
}

// 深度解密对象中的敏感字段
export function decryptSensitiveData(obj: any): any {
  if (typeof obj !== 'object' || obj === null) {
    return obj
  }

  if (Array.isArray(obj)) {
    return obj.map(item => decryptSensitiveData(item))
  }

  const result: any = {}
  for (const [key, value] of Object.entries(obj)) {
    if (typeof value === 'string' && isSensitiveField(key) && value.startsWith('enc:')) {
      // 解密敏感字段
      result[key] = decryptValue(value)
    } else if (typeof value === 'object' && value !== null) {
      // 递归处理嵌套对象
      result[key] = decryptSensitiveData(value)
    } else {
      result[key] = value
    }
  }

  return result
}
