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

    <!-- 添加/编辑弹窗 -->
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
        <n-form-item label="API 密钥" path="apiKey">
          <n-input
            v-model:value="formData.apiKey"
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

const message = useMessage()

interface ApiKey {
  id: string
  name: string
  provider: string
  apiKey: string
  endpoint?: string
  createdAt: string
  updatedAt: string
}

const apiKeys = ref<ApiKey[]>([])
const showModal = ref(false)
const editingId = ref<string | null>(null)
const testLoading = ref(false)
const formRef = ref()

const formData = ref({
  name: '',
  provider: '',
  apiKey: '',
  endpoint: '',
})

const formRules: FormRules = {
  name: { required: true, message: '请输入名称', trigger: 'blur' },
  provider: { required: true, message: '请选择提供商', trigger: 'change' },
  apiKey: { required: true, message: '请输入 API 密钥', trigger: 'blur' },
}

const providerOptions = [
  { label: 'OpenAI', value: 'openai' },
  { label: 'Anthropic', value: 'anthropic' },
  { label: '其他', value: 'other' },
]

function maskKey(key: string): string {
  if (key.length <= 8) return '****'
  return key.substring(0, 4) + '****' + key.substring(key.length - 4)
}

const columns: DataTableColumns<ApiKey> = [
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
    key: 'apiKey',
    render(row) {
      return maskKey(row.apiKey)
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
    if (res.success && res.data && Array.isArray(res.data.apiKeys)) {
      apiKeys.value = res.data.apiKeys
    }
  } catch (err: any) {
    message.error('加载配置失败: ' + (err.message || '请检查网络连接'))
  }
}

async function saveToConfig() {
  try {
    const res = await $fetch<{ success: boolean; data?: any }>('/api/config')
    const config = res.success && res.data ? res.data : { version: '2.0.0', apiKeys: [], adapters: {} }
    config.apiKeys = apiKeys.value

    await $fetch('/api/config', {
      method: 'POST',
      body: config,
    })
  } catch (err: any) {
    message.error('保存失败: ' + (err.message || '未知错误'))
  }
}

function openAddModal() {
  editingId.value = null
  formData.value = { name: '', provider: '', apiKey: '', endpoint: '' }
  showModal.value = true
}

function openEditModal(row: ApiKey) {
  editingId.value = row.id
  formData.value = {
    name: row.name,
    provider: row.provider,
    apiKey: row.apiKey,
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

  const now = new Date().toISOString()

  if (editingId.value) {
    const idx = apiKeys.value.findIndex((k) => k.id === editingId.value)
    if (idx >= 0) {
      apiKeys.value[idx] = {
        ...apiKeys.value[idx],
        name: formData.value.name,
        provider: formData.value.provider,
        apiKey: formData.value.apiKey,
        endpoint: formData.value.endpoint || undefined,
        updatedAt: now,
      }
    }
  } else {
    apiKeys.value.push({
      id: crypto.randomUUID ? crypto.randomUUID() : Date.now().toString(36),
      name: formData.value.name,
      provider: formData.value.provider,
      apiKey: formData.value.apiKey,
      endpoint: formData.value.endpoint || undefined,
      createdAt: now,
      updatedAt: now,
    })
  }

  await saveToConfig()
  message.success(editingId.value ? '密钥已更新' : '密钥已添加')
  showModal.value = false
  return true
}

async function handleDelete(id: string) {
  apiKeys.value = apiKeys.value.filter((k) => k.id !== id)
  await saveToConfig()
  message.success('密钥已删除')
}

async function handleTestConnection() {
  testLoading.value = true
  // 模拟测试连接
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
