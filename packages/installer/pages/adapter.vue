<template>
  <div class="adapter-page">
    <h2 class="page-title">选择 IM 适配器</h2>
    <p class="page-desc">请选择要安装的即时通讯平台适配器，至少选择一个。</p>

    <n-checkbox-group v-model:value="installOptions.adapters" class="adapter-group">
      <div
        v-for="adapter in adapterList"
        :key="adapter.value"
        class="adapter-card"
        :class="{ active: installOptions.adapters.includes(adapter.value) }"
      >
        <n-checkbox :value="adapter.value" size="large">
          <span class="adapter-label">{{ adapter.label }}</span>
        </n-checkbox>
        <p class="adapter-desc">{{ adapter.desc }}</p>
      </div>
    </n-checkbox-group>

    <p v-if="showError" class="adapter-error">请至少选择一个适配器</p>

    <div class="nav-buttons">
      <n-button @click="goBack">上一步</n-button>
      <n-button type="primary" @click="handleNext">下一步</n-button>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import { NCheckboxGroup, NCheckbox, NButton } from 'naive-ui'
import { useWizard } from '~/composables/useWizard'

const { installOptions, goNext, goBack } = useWizard()

const showError = ref(false)

const adapterList = [
  {
    value: 'wechat-work',
    label: '企业微信',
    desc: '接入企业微信，在工作群中使用 AI 助手。',
  },
  {
    value: 'dingtalk',
    label: '钉钉',
    desc: '接入钉钉，在钉钉群聊中使用 AI 助手。',
  },
  {
    value: 'feishu',
    label: '飞书',
    desc: '接入飞书，在飞书群组中使用 AI 助手。',
  },
]

function handleNext() {
  if (installOptions.value.adapters.length === 0) {
    showError.value = true
    return
  }
  showError.value = false
  goNext()
}
</script>

<style scoped>
.adapter-page {
  display: flex;
  flex-direction: column;
  align-items: center;
  width: 100%;
  max-width: 520px;
}

.page-title {
  font-size: 22px;
  font-weight: 600;
  color: #fff;
  margin-bottom: 8px;
}

.page-desc {
  font-size: 14px;
  color: rgba(255, 255, 255, 0.55);
  margin-bottom: 24px;
}

.adapter-group {
  display: flex;
  flex-direction: column;
  gap: 12px;
  width: 100%;
}

.adapter-card {
  border: 1px solid rgba(255, 255, 255, 0.12);
  border-radius: 8px;
  padding: 14px 20px;
  transition: border-color 0.2s;
}

.adapter-card.active {
  border-color: #18a058;
}

.adapter-label {
  font-size: 16px;
  font-weight: 500;
  color: #fff;
}

.adapter-desc {
  font-size: 13px;
  color: rgba(255, 255, 255, 0.5);
  margin: 6px 0 0 26px;
}

.adapter-error {
  color: #e88080;
  font-size: 13px;
  margin-top: 8px;
}

.nav-buttons {
  display: flex;
  gap: 12px;
  margin-top: 32px;
}
</style>
