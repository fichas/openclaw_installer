<template>
  <div class="adapters-page">
    <n-card title="IM 适配器配置">
      <n-tabs v-model:value="activeTab" type="line">
        <!-- 企业微信 -->
        <n-tab-pane name="wechat-work" tab="企业微信">
          <n-form
            label-placement="left"
            label-width="120"
            style="max-width: 500px; margin-top: 16px;"
          >
            <n-form-item label="启用">
              <n-switch v-model:value="adapters.wechatWork.enabled" />
            </n-form-item>
            <n-form-item label="企业ID">
              <n-input
                v-model:value="adapters.wechatWork.config.corp_id"
                placeholder="请输入企业微信的企业ID"
              />
              <template #feedback>
                在企业微信管理后台 "我的企业" 页面获取
              </template>
            </n-form-item>
            <n-form-item label="应用密钥">
              <n-input
                v-model:value="adapters.wechatWork.config.corp_secret"
                type="password"
                show-password-on="click"
                placeholder="请输入应用的 Secret"
              />
              <template #feedback>
                在企业微信管理后台 "应用管理" 中对应应用详情页获取
              </template>
            </n-form-item>
            <n-form-item label="AgentID">
              <n-input
                v-model:value="adapters.wechatWork.config.agent_id"
                placeholder="请输入应用的 AgentID"
              />
              <template #feedback>
                在企业微信管理后台 "应用管理" 中对应应用详情页获取
              </template>
            </n-form-item>
            <n-form-item label="">
              <n-button type="primary" :loading="saving === 'wechat-work'" @click="saveAdapter('wechat-work')">
                保存
              </n-button>
            </n-form-item>
          </n-form>
        </n-tab-pane>

        <!-- 钉钉 -->
        <n-tab-pane name="dingtalk" tab="钉钉">
          <n-form
            label-placement="left"
            label-width="120"
            style="max-width: 500px; margin-top: 16px;"
          >
            <n-form-item label="启用">
              <n-switch v-model:value="adapters.dingtalk.enabled" />
            </n-form-item>
            <n-form-item label="AppKey">
              <n-input
                v-model:value="adapters.dingtalk.config.app_key"
                placeholder="请输入钉钉应用的 AppKey"
              />
              <template #feedback>
                在钉钉开放平台 "应用开发" 中获取
              </template>
            </n-form-item>
            <n-form-item label="AppSecret">
              <n-input
                v-model:value="adapters.dingtalk.config.app_secret"
                type="password"
                show-password-on="click"
                placeholder="请输入钉钉应用的 AppSecret"
              />
              <template #feedback>
                在钉钉开放平台 "应用开发" 中获取
              </template>
            </n-form-item>
            <n-form-item label="机器人编码">
              <n-input
                v-model:value="adapters.dingtalk.config.robot_code"
                placeholder="可选，请输入机器人编码"
              />
              <template #feedback>
                在钉钉开放平台 "机器人" 配置中获取（可选）
              </template>
            </n-form-item>
            <n-form-item label="">
              <n-button type="primary" :loading="saving === 'dingtalk'" @click="saveAdapter('dingtalk')">
                保存
              </n-button>
            </n-form-item>
          </n-form>
        </n-tab-pane>

        <!-- 飞书 -->
        <n-tab-pane name="feishu" tab="飞书">
          <n-form
            label-placement="left"
            label-width="120"
            style="max-width: 500px; margin-top: 16px;"
          >
            <n-form-item label="启用">
              <n-switch v-model:value="adapters.feishu.enabled" />
            </n-form-item>
            <n-form-item label="App ID">
              <n-input
                v-model:value="adapters.feishu.config.app_id"
                placeholder="请输入飞书应用的 App ID"
              />
              <template #feedback>
                在飞书开放平台 "凭证与基础信息" 中获取
              </template>
            </n-form-item>
            <n-form-item label="App Secret">
              <n-input
                v-model:value="adapters.feishu.config.app_secret"
                type="password"
                show-password-on="click"
                placeholder="请输入飞书应用的 App Secret"
              />
              <template #feedback>
                在飞书开放平台 "凭证与基础信息" 中获取
              </template>
            </n-form-item>
            <n-form-item label="加密密钥">
              <n-input
                v-model:value="adapters.feishu.config.encrypt_key"
                type="password"
                show-password-on="click"
                placeholder="可选，请输入事件加密密钥"
              />
              <template #feedback>
                在飞书开放平台 "事件订阅" 中获取（可选）
              </template>
            </n-form-item>
            <n-form-item label="">
              <n-button type="primary" :loading="saving === 'feishu'" @click="saveAdapter('feishu')">
                保存
              </n-button>
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

const message = useMessage()
const activeTab = ref('wechat-work')
const saving = ref<string | null>(null)

const adapters = reactive({
  wechatWork: {
    enabled: false,
    config: {
      corp_id: '',
      corp_secret: '',
      agent_id: '',
    },
  },
  dingtalk: {
    enabled: false,
    config: {
      app_key: '',
      app_secret: '',
      robot_code: '',
    },
  },
  feishu: {
    enabled: false,
    config: {
      app_id: '',
      app_secret: '',
      encrypt_key: '',
    },
  },
})

// 适配器名称映射
const adapterKeyMap: Record<string, keyof typeof adapters> = {
  'wechat-work': 'wechatWork',
  dingtalk: 'dingtalk',
  feishu: 'feishu',
}

const displayNameMap: Record<string, string> = {
  'wechat-work': '企业微信',
  dingtalk: '钉钉',
  feishu: '飞书',
}

onMounted(() => {
  fetchConfig()
})

async function fetchConfig() {
  try {
    const res = await $fetch<{ success: boolean; data?: any }>('/api/config')
    if (res.success && res.data && res.data.adapters) {
      const savedAdapters = res.data.adapters

      for (const [adapterName, reactiveKey] of Object.entries(adapterKeyMap)) {
        if (savedAdapters[adapterName]) {
          const saved = savedAdapters[adapterName]
          const target = adapters[reactiveKey]
          target.enabled = saved.enabled || false
          if (saved.config) {
            Object.assign(target.config, saved.config)
          }
        }
      }
    }
  } catch {
    // 加载配置失败
  }
}

async function saveAdapter(adapterName: string) {
  saving.value = adapterName
  try {
    // 先读取当前配置
    const res = await $fetch<{ success: boolean; data?: any }>('/api/config')
    const config = res.success && res.data ? res.data : { version: '2.0.0', apiKeys: [], adapters: {} }

    if (!config.adapters) {
      config.adapters = {}
    }

    const reactiveKey = adapterKeyMap[adapterName]
    const adapterData = adapters[reactiveKey]

    config.adapters[adapterName] = {
      name: adapterName,
      displayName: displayNameMap[adapterName],
      enabled: adapterData.enabled,
      config: { ...adapterData.config },
    }

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
