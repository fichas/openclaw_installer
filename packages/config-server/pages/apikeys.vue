<template>
  <div class="apikeys-page">
    <n-card title="API 密钥管理">
      <template #header-extra>
        <n-button type="primary" @click="openAddModal">添加密钥</n-button>
      </template>

      <n-data-table
        :columns="columns"
        :data="apiKeys"
        :bordered="true"
        :single-line="false"
      />
    </n-card>

    <n-modal
      v-model:show="showModal"
      preset="dialog"
      :title="editingId ? '编辑密钥' : '添加密钥'"
      :positive-text="editingId ? '保存' : '添加'"
      negative-text="取消"
      @positive-click="handleSave"
      @negative-click="showModal = false"
      style="width: 500px;"
    >
      <n-form
        ref="formRef"
        :model="formData"
        :rules="formRules"
        label-placement="left"
        label-width="100"
        style="margin-top: 16px;"
      >
        <n-form-item label="名称" path="name">
          <n-input v-model:value="formData.name" placeholder="例如: 我的 OpenAI 密钥" />
        </n-form-item>
        <n-form-item label="提供商" path="provider">
          <n-select
            v-model:value="formData.provider"
            :options="providerOptions"
            placeholder="请选择 API 提供商"
          />
        </n-form-item>
        <n-form-item label="API 密钥" path="key">
          <n-input
            v-model:value="formData.key"
            type="password"
            show-password-on="click"
            placeholder="请输入 API 密钥"
          />
        </n-form-item>
        <n-form-item label="接入点" path="endpoint">
          <n-input v-model:value="formData.endpoint" placeholder="可选，自定义 API 地址" />
        </n-form-item>
        <n-form-item label="">
          <n-button :loading="testLoading" @click="handleTestConnection">测试连接</n-button>
        </n-form-item>
      </n-form>
    </n-modal>
  </div>
</template>

<script setup lang="ts">
import { ref, h, onMounted } from 'vue'
import {
  NCard,
  NButton,
  NDataTable,
  NModal,
  NForm,
  NFormItem,
  NInput,
  NSelect,
  NSpace,
  NPopconfirm,
  useMessage,
} from 'naive-ui'
import type { DataTableColumns, FormRules } from 'naive-ui'
import type { ApiKeyConfig, OpenClawConfig, AdapterConfig } from '@openclaw/shared'

const message = useMessage()

const apiKeys = ref<ApiKeyConfig[]>([])
const showModal = ref(false)
const editingId = ref<string | null>(null)
const testLoading = ref(false)
const formRef = ref()

const formData = ref({
  name: '',
  provider: '',
  key: '',
  endpoint: '',
})

const formRules: FormRules = {
  name: { required: true, message: '请输入名称', trigger: 'blur' },
  provider: { required: true, message: '请选择提供商', trigger: 'change' },
  key: { required: true, message: '请输入 API 密钥', trigger: 'blur' },
}

const providerOptions = [
  { label: 'OpenAI', value: 'openai' },
  { label: 'Anthropic', value: 'anthropic' },
  { label: '其他', value: 'other' },
]

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

function maskKey(key: string): string {
  if (key.length <= 8) return '****'
  return key.substring(0, 4) + '****' + key.substring(key.length - 4)
}

const columns: DataTableColumns<ApiKeyConfig> = [
  { title: '名称', key: 'name', width: 160 },
  {
    title: '提供商',
    key: 'provider',
    width: 120,
    render(row) {
      const map: Record<string, string> = {
        openai: 'OpenAI',
        anthropic: 'Anthropic',
        other: '其他',
      }
      return map[row.provider] || row.provider
    },
  },
  {
    title: '密钥',
    key: 'key',
    render(row) {
      return maskKey(row.key)
    },
  },
  {
    title: '操作',
    key: 'actions',
    width: 160,
    render(row) {
      return h(NSpace, null, {
        default: () => [
          h(
            NButton,
            {
              size: 'small',
              onClick: () => openEditModal(row),
            },
            { default: () => '编辑' }
          ),
          h(
            NPopconfirm,
            {
              onPositiveClick: () => handleDelete(row.id),
            },
            {
              trigger: () =>
                h(NButton, { size: 'small', type: 'error' }, { default: () => '删除' }),
              default: () => '确定删除该密钥？',
            }
          ),
        ],
      })
    },
  },
]

onMounted(() => {
  fetchConfig()
})

async function fetchConfig() {
  try {
    const res = await $fetch<{ success: boolean; data?: any }>('/api/config')
    const config = normalizeConfig(res.success ? res.data : null)
    apiKeys.value = config.apiKeys
  } catch (err: any) {
    message.error('加载配置失败: ' + (err.message || '请检查网络连接'))
  }
}

async function saveToConfig(nextKeys: ApiKeyConfig[]) {
  try {
    const res = await $fetch<{ success: boolean; data?: any }>('/api/config')
    const config = normalizeConfig(res.success ? res.data : null)
    config.apiKeys = nextKeys

    await $fetch('/api/config', {
      method: 'POST',
      body: config,
    })
  } catch (err: any) {
    message.error('保存失败: ' + (err.message || '未知错误'))
    throw err
  }
}

function openAddModal() {
  editingId.value = null
  formData.value = { name: '', provider: '', key: '', endpoint: '' }
  showModal.value = true
}

function openEditModal(row: ApiKeyConfig) {
  editingId.value = row.id
  formData.value = {
    name: row.name,
    provider: row.provider,
    key: row.key,
    endpoint: row.endpoint || '',
  }
  showModal.value = true
}

async function handleSave(): Promise<boolean> {
  try {
    await formRef.value?.validate()
  } catch {
    return false
  }

  const nextKeys = [...apiKeys.value]

  if (editingId.value) {
    const idx = apiKeys.value.findIndex((k) => k.id === editingId.value)
    if (idx >= 0) {
      nextKeys[idx] = {
        ...nextKeys[idx],
        name: formData.value.name,
        provider: formData.value.provider,
        key: formData.value.key,
        endpoint: formData.value.endpoint || undefined,
      }
    }
  } else {
    nextKeys.push({
      id: crypto.randomUUID ? crypto.randomUUID() : Date.now().toString(36),
      name: formData.value.name,
      provider: formData.value.provider,
      key: formData.value.key,
      endpoint: formData.value.endpoint || undefined,
      createdAt: new Date().toISOString(),
    })
  }

  await saveToConfig(nextKeys)
  apiKeys.value = nextKeys
  message.success(editingId.value ? '密钥已更新' : '密钥已添加')
  showModal.value = false
  return true
}

async function handleDelete(id: string) {
  const nextKeys = apiKeys.value.filter((k) => k.id !== id)
  await saveToConfig(nextKeys)
  apiKeys.value = nextKeys
  message.success('密钥已删除')
}

async function handleTestConnection() {
  testLoading.value = true
  await new Promise((resolve) => setTimeout(resolve, 1500))
  testLoading.value = false
  message.success('连接测试成功（模拟）')
}
</script>

<style scoped>
.apikeys-page {
  max-width: 900px;
}
</style>
