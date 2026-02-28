<template>
  <div class="adapters-page">
    <n-card title="IM 适配器配置">
      <n-tabs v-model:value="activeTab" type="line">
        <n-tab-pane name="wechat-work" tab="企业微信">
          <n-form label-placement="left" label-width="120" style="max-width: 500px; margin-top: 16px;">
            <n-form-item label="启用">
              <n-switch v-model:value="adapters.wechatWork.enabled" />
            </n-form-item>
            <n-form-item label="企业ID">
              <n-input v-model:value="adapters.wechatWork.options.corp_id" placeholder="请输入企业微信的企业ID" />
            </n-form-item>
            <n-form-item label="应用密钥">
              <n-input v-model:value="adapters.wechatWork.options.corp_secret" type="password" show-password-on="click" placeholder="请输入应用的 Secret" />
            </n-form-item>
            <n-form-item label="AgentID">
              <n-input v-model:value="adapters.wechatWork.options.agent_id" placeholder="请输入应用的 AgentID" />
            </n-form-item>
            <n-form-item label="">
              <n-button type="primary" :loading="saving === 'wechat-work'" @click="saveAdapter('wechat-work')">保存</n-button>
            </n-form-item>
          </n-form>
        </n-tab-pane>

        <n-tab-pane name="dingtalk" tab="钉钉">
          <n-form label-placement="left" label-width="120" style="max-width: 500px; margin-top: 16px;">
            <n-form-item label="启用">
              <n-switch v-model:value="adapters.dingtalk.enabled" />
            </n-form-item>
            <n-form-item label="AppKey">
              <n-input v-model:value="adapters.dingtalk.options.app_key" placeholder="请输入钉钉应用的 AppKey" />
            </n-form-item>
            <n-form-item label="AppSecret">
              <n-input v-model:value="adapters.dingtalk.options.app_secret" type="password" show-password-on="click" placeholder="请输入钉钉应用的 AppSecret" />
            </n-form-item>
            <n-form-item label="机器人编码">
              <n-input v-model:value="adapters.dingtalk.options.robot_code" placeholder="可选，请输入机器人编码" />
            </n-form-item>
            <n-form-item label="">
              <n-button type="primary" :loading="saving === 'dingtalk'" @click="saveAdapter('dingtalk')">保存</n-button>
            </n-form-item>
          </n-form>
        </n-tab-pane>

        <n-tab-pane name="feishu" tab="飞书">
          <n-form label-placement="left" label-width="120" style="max-width: 500px; margin-top: 16px;">
            <n-form-item label="启用">
              <n-switch v-model:value="adapters.feishu.enabled" />
            </n-form-item>
            <n-form-item label="App ID">
              <n-input v-model:value="adapters.feishu.options.app_id" placeholder="请输入飞书应用的 App ID" />
            </n-form-item>
            <n-form-item label="App Secret">
              <n-input v-model:value="adapters.feishu.options.app_secret" type="password" show-password-on="click" placeholder="请输入飞书应用的 App Secret" />
            </n-form-item>
            <n-form-item label="加密密钥">
              <n-input v-model:value="adapters.feishu.options.encrypt_key" type="password" show-password-on="click" placeholder="可选，请输入事件加密密钥" />
            </n-form-item>
            <n-form-item label="">
              <n-button type="primary" :loading="saving === 'feishu'" @click="saveAdapter('feishu')">保存</n-button>
            </n-form-item>
          </n-form>
        </n-tab-pane>
      </n-tabs>
    </n-card>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted } from 'vue'
import {
  NCard,
  NTabs,
  NTabPane,
  NForm,
  NFormItem,
  NInput,
  NSwitch,
  NButton,
  useMessage,
} from 'naive-ui'
import type { OpenClawConfig, AdapterConfig, ApiKeyConfig } from '@openclaw/shared'

const message = useMessage()
const activeTab = ref('wechat-work')
const saving = ref<string | null>(null)

type EditableAdapter = {
  enabled: boolean
  options: Record<string, string>
}

const adapters = reactive<Record<'wechatWork' | 'dingtalk' | 'feishu', EditableAdapter>>({
  wechatWork: {
    enabled: false,
    options: {
      corp_id: '',
      corp_secret: '',
      agent_id: '',
    },
  },
  dingtalk: {
    enabled: false,
    options: {
      app_key: '',
      app_secret: '',
      robot_code: '',
    },
  },
  feishu: {
    enabled: false,
    options: {
      app_id: '',
      app_secret: '',
      encrypt_key: '',
    },
  },
})

const adapterKeyMap: Record<string, 'wechatWork' | 'dingtalk' | 'feishu'> = {
  'wechat-work': 'wechatWork',
  dingtalk: 'dingtalk',
  feishu: 'feishu',
}

const displayNameMap: Record<string, string> = {
  'wechat-work': '企业微信',
  dingtalk: '钉钉',
  feishu: '飞书',
}

function detectClientPlatform(): string {
  const ua = navigator.userAgent.toLowerCase()
  if (ua.includes('windows')) return 'windows'
  if (ua.includes('mac')) return 'darwin'
  return 'linux'
}

function normalizeAdapters(raw: any): AdapterConfig[] {
  if (Array.isArray(raw)) {
    return raw.map((item) => ({
      name: item?.name || '',
      type: item?.type || 'messaging',
      displayName: item?.displayName || item?.name || '',
      enabled: Boolean(item?.enabled),
      options: { ...(item?.options || item?.config || {}) },
    })).filter((item) => item.name)
  }

  if (raw && typeof raw === 'object') {
    return Object.entries(raw).map(([name, value]: [string, any]) => ({
      name,
      type: value?.type || 'messaging',
      displayName: value?.displayName || name,
      enabled: Boolean(value?.enabled),
      options: { ...(value?.options || value?.config || {}) },
    }))
  }

  return []
}

function normalizeApiKeys(raw: any): ApiKeyConfig[] {
  if (!Array.isArray(raw)) return []
  return raw
    .map((item: any): ApiKeyConfig | null => {
      const id = item?.id || ''
      const name = item?.name || ''
      const provider = item?.provider || ''
      const key = item?.key || item?.apiKey || ''
      if (!id || !name || !provider || !key) return null
      return {
        id,
        name,
        provider,
        key,
        endpoint: item?.endpoint || undefined,
        createdAt: item?.createdAt || new Date().toISOString(),
      }
    })
    .filter((item): item is ApiKeyConfig => item !== null)
}

function normalizeConfig(raw: any): OpenClawConfig {
  return {
    version: raw?.version || '2.0.0',
    platform: raw?.platform || detectClientPlatform(),
    server: {
      host: raw?.server?.host || '0.0.0.0',
      port: Number(raw?.server?.port) || 18080,
      tls: Boolean(raw?.server?.tls),
    },
    adapters: normalizeAdapters(raw?.adapters),
    apiKeys: normalizeApiKeys(raw?.apiKeys),
    settings: raw?.settings && typeof raw.settings === 'object' ? raw.settings : {},
  }
}

onMounted(() => {
  fetchConfig()
})

async function fetchConfig() {
  try {
    const res = await $fetch<{ success: boolean; data?: any }>('/api/config')
    const config = normalizeConfig(res.success ? res.data : null)

    for (const adapter of config.adapters) {
      const reactiveKey = adapterKeyMap[adapter.name]
      if (!reactiveKey) continue
      adapters[reactiveKey].enabled = Boolean(adapter.enabled)
      adapters[reactiveKey].options = {
        ...adapters[reactiveKey].options,
        ...adapter.options,
      }
    }
  } catch (err: any) {
    message.error('加载配置失败: ' + (err.message || '请检查网络连接'))
  }
}

function buildAdapterArray(): AdapterConfig[] {
  return [
    {
      name: 'wechat-work',
      type: 'messaging',
      displayName: '企业微信',
      enabled: adapters.wechatWork.enabled,
      options: { ...adapters.wechatWork.options },
    },
    {
      name: 'dingtalk',
      type: 'messaging',
      displayName: '钉钉',
      enabled: adapters.dingtalk.enabled,
      options: { ...adapters.dingtalk.options },
    },
    {
      name: 'feishu',
      type: 'messaging',
      displayName: '飞书',
      enabled: adapters.feishu.enabled,
      options: { ...adapters.feishu.options },
    },
  ]
}

async function saveAdapter(adapterName: string) {
  saving.value = adapterName
  try {
    const res = await $fetch<{ success: boolean; data?: any }>('/api/config')
    const config = normalizeConfig(res.success ? res.data : null)
    const merged = new Map<string, AdapterConfig>()
    for (const adapter of config.adapters) {
      merged.set(adapter.name, adapter)
    }

    const currentAdapters = buildAdapterArray()
    const updated = currentAdapters.find((item) => item.name === adapterName)
    if (updated) {
      merged.set(adapterName, updated)
    }

    config.adapters = Array.from(merged.values())

    await $fetch('/api/config', {
      method: 'POST',
      body: config,
    })

    message.success(`${displayNameMap[adapterName]}配置已保存`)
  } catch (err: any) {
    message.error('保存失败: ' + (err.message || '未知错误'))
  } finally {
    saving.value = null
  }
}
</script>

<style scoped>
.adapters-page {
  max-width: 800px;
}
</style>
