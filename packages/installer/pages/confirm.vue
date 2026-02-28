<template>
  <div class="confirm-page">
    <h2 class="page-title">确认安装信息</h2>
    <p class="page-desc">请确认以下安装配置，确认无误后点击"开始安装"。</p>

    <n-descriptions
      label-placement="left"
      bordered
      :column="1"
      class="confirm-details"
      size="medium"
    >
      <n-descriptions-item label="安装模式">
        {{ installOptions.mode === 'standard' ? '标准安装' : '自定义安装' }}
      </n-descriptions-item>
      <n-descriptions-item label="安装目录">
        {{ installOptions.installDir }}
      </n-descriptions-item>
      <n-descriptions-item label="已选适配器">
        {{ adapterNames }}
      </n-descriptions-item>
    </n-descriptions>

    <div class="nav-buttons">
      <n-button @click="goBack">上一步</n-button>
      <n-button type="primary" @click="handleInstall">开始安装</n-button>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { NDescriptions, NDescriptionsItem, NButton } from 'naive-ui'
import { useWizard } from '~/composables/useWizard'

const { installOptions, goNext, goBack } = useWizard()

const adapterNameMap: Record<string, string> = {
  'wechat-work': '企业微信',
  dingtalk: '钉钉',
  feishu: '飞书',
}

const adapterNames = computed(() => {
  return installOptions.value.adapters
    .map((a) => adapterNameMap[a] || a)
    .join('、')
})

function handleInstall() {
  goNext()
}
</script>

<style scoped>
.confirm-page {
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

.confirm-details {
  width: 100%;
}

.nav-buttons {
  display: flex;
  gap: 12px;
  margin-top: 32px;
}
</style>
