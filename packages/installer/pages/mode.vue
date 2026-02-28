<template>
  <div class="mode-page">
    <h2 class="page-title">选择安装模式</h2>
    <p class="page-desc">请选择安装方式，推荐使用标准安装。</p>

    <n-radio-group v-model:value="installOptions.mode" class="mode-group">
      <div class="mode-card" :class="{ active: installOptions.mode === 'standard' }">
        <n-radio value="standard" size="large">
          <span class="mode-label">标准安装</span>
          <span class="mode-recommend">（推荐）</span>
        </n-radio>
        <p class="mode-detail">安装到默认目录，适合大多数用户。</p>
        <p class="mode-path">安装目录: {{ defaultInstallDir }}</p>
      </div>
      <div class="mode-card" :class="{ active: installOptions.mode === 'custom' }">
        <n-radio value="custom" size="large">
          <span class="mode-label">自定义安装</span>
        </n-radio>
        <p class="mode-detail">选择自定义的安装目录。</p>
        <n-input
          v-if="installOptions.mode === 'custom'"
          v-model:value="installOptions.installDir"
          placeholder="请输入安装目录路径"
          class="custom-dir-input"
        />
      </div>
    </n-radio-group>

    <div class="nav-buttons">
      <n-button @click="goBack">上一步</n-button>
      <n-button type="primary" @click="handleNext">下一步</n-button>
    </div>
  </div>
</template>

<script setup lang="ts">
import { onMounted } from 'vue'
import { NRadioGroup, NRadio, NButton, NInput } from 'naive-ui'
import { useWizard } from '~/composables/useWizard'

const { installOptions, goNext, goBack } = useWizard()

const defaultInstallDir = ref('')

onMounted(() => {
  try {
    const { getPlatformPaths } = require('@openclaw/shared')
    const paths = getPlatformPaths()
    defaultInstallDir.value = paths.installDir
    if (installOptions.value.installDir === '') {
      installOptions.value.installDir = paths.installDir
      installOptions.value.configDir = paths.configDir
    }
  } catch {
    // Fallback for browser environment during dev
    defaultInstallDir.value = '/usr/local/bin'
    if (installOptions.value.installDir === '') {
      installOptions.value.installDir = '/usr/local/bin'
      installOptions.value.configDir = '~/.config/openclaw'
    }
  }
})

function handleNext() {
  if (installOptions.value.mode === 'standard') {
    installOptions.value.installDir = defaultInstallDir.value
  }
  goNext()
}
</script>

<style scoped>
.mode-page {
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

.mode-group {
  display: flex;
  flex-direction: column;
  gap: 16px;
  width: 100%;
}

.mode-card {
  border: 1px solid rgba(255, 255, 255, 0.12);
  border-radius: 8px;
  padding: 16px 20px;
  transition: border-color 0.2s;
}

.mode-card.active {
  border-color: #18a058;
}

.mode-label {
  font-size: 16px;
  font-weight: 500;
  color: #fff;
}

.mode-recommend {
  font-size: 13px;
  color: #18a058;
  margin-left: 4px;
}

.mode-detail {
  font-size: 13px;
  color: rgba(255, 255, 255, 0.5);
  margin: 6px 0 0 26px;
}

.mode-path {
  font-size: 12px;
  color: rgba(255, 255, 255, 0.35);
  margin: 4px 0 0 26px;
  font-family: monospace;
}

.custom-dir-input {
  margin: 10px 0 0 26px;
  width: calc(100% - 26px);
}

.nav-buttons {
  display: flex;
  gap: 12px;
  margin-top: 32px;
}
</style>
