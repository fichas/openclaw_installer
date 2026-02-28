<template>
  <div class="progress-page">
    <h2 class="page-title">正在安装</h2>
    <p class="page-desc">请耐心等待，安装过程中请勿关闭窗口。</p>

    <div class="progress-area">
      <n-progress
        type="line"
        :percentage="progress.percent"
        :status="progressStatus"
        :show-indicator="true"
      />
      <p class="progress-message">{{ progress.message }}</p>
    </div>

    <div v-if="progress.error" class="error-area">
      <p class="error-title">安装出错了</p>
      <p class="error-message">{{ progress.error }}</p>
      <n-button type="warning" @click="retryInstall">重试</n-button>
    </div>
  </div>
</template>

<script setup lang="ts">
import { reactive, onMounted, computed, onBeforeUnmount } from 'vue'
import { NProgress, NButton } from 'naive-ui'
import { useWizard } from '~/composables/useWizard'
import type { InstallProgress } from '@openclaw/shared'

const { goToStep, installOptions } = useWizard()

const progress = reactive<InstallProgress>({
  step: 0,
  totalSteps: 5,
  message: '准备安装...',
  percent: 0,
  error: undefined,
})

let disposeProgressListener: (() => void) | null = null

const progressStatus = computed(() => {
  if (progress.error) return 'error'
  if (progress.percent >= 100) return 'success'
  return 'default'
})

async function runInstall() {
  progress.error = undefined
  progress.step = 0
  progress.percent = 0
  progress.message = '准备安装...'

  if (!window.electronAPI?.startInstall) {
    progress.error = '当前环境不支持安装执行，请在安装器中运行。'
    return
  }

  const result = await window.electronAPI.startInstall(installOptions.value)
  if (result.success) {
    await new Promise((resolve) => setTimeout(resolve, 500))
    goToStep(5)
  } else {
    progress.error = result.error || '安装失败，请稍后重试。'
  }
}

function retryInstall() {
  runInstall()
}

onMounted(() => {
  if (window.electronAPI?.onInstallProgress) {
    disposeProgressListener = window.electronAPI.onInstallProgress((next) => {
      progress.step = next.step
      progress.totalSteps = next.totalSteps
      progress.message = next.message
      progress.percent = next.percent
      progress.error = next.error
    })
  }

  runInstall()
})

onBeforeUnmount(() => {
  if (disposeProgressListener) {
    disposeProgressListener()
    disposeProgressListener = null
  }
})
</script>

<style scoped>
.progress-page {
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
  margin-bottom: 32px;
}

.progress-area {
  width: 100%;
}

.progress-message {
  font-size: 14px;
  color: rgba(255, 255, 255, 0.65);
  margin-top: 16px;
  text-align: center;
}

.error-area {
  margin-top: 24px;
  text-align: center;
}

.error-title {
  font-size: 16px;
  font-weight: 500;
  color: #e88080;
  margin-bottom: 8px;
}

.error-message {
  font-size: 14px;
  color: rgba(255, 255, 255, 0.55);
  margin-bottom: 16px;
  max-width: 400px;
  line-height: 1.6;
}
</style>
