<template>
  <div class="done-page">
    <div class="done-icon">
      <svg width="72" height="72" viewBox="0 0 72 72" fill="none" xmlns="http://www.w3.org/2000/svg">
        <circle cx="36" cy="36" r="34" stroke="#18a058" stroke-width="3" fill="none" />
        <path
          d="M22 36L32 46L50 28"
          stroke="#18a058"
          stroke-width="4"
          stroke-linecap="round"
          stroke-linejoin="round"
          fill="none"
        />
      </svg>
    </div>
    <h2 class="done-title">安装成功</h2>
    <p class="done-desc">OpenClaw 已成功安装到您的电脑上。</p>

    <n-checkbox v-model:checked="openConfig" class="done-checkbox">
      立即打开配置页面
    </n-checkbox>

    <n-button type="primary" size="large" @click="handleFinish">
      完成
    </n-button>
  </div>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import { NCheckbox, NButton } from 'naive-ui'

const openConfig = ref(true)

function handleFinish() {
  if (openConfig.value) {
    // Open the config page in system browser
    const url = 'http://localhost:18080'
    try {
      const { shell } = require('electron')
      shell.openExternal(url)
    } catch {
      // Fallback for non-Electron env (dev mode in browser)
      window.open(url, '_blank')
    }
  }

  // Quit the Electron app
  try {
    const { ipcRenderer } = require('electron')
    // Give browser a moment to open before quitting
    setTimeout(() => {
      ipcRenderer.send('app-quit')
    }, 500)
  } catch {
    // Not in Electron, just close the window
    window.close()
  }
}
</script>

<style scoped>
.done-page {
  display: flex;
  flex-direction: column;
  align-items: center;
  text-align: center;
  gap: 12px;
}

.done-icon {
  margin-bottom: 8px;
}

.done-title {
  font-size: 28px;
  font-weight: 700;
  color: #18a058;
}

.done-desc {
  font-size: 15px;
  color: rgba(255, 255, 255, 0.65);
  margin-bottom: 8px;
}

.done-checkbox {
  margin-bottom: 12px;
}
</style>
