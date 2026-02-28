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
import { ref, reactive, onMounted, computed } from 'vue'
import { NProgress, NButton } from 'naive-ui'
import { useWizard } from '~/composables/useWizard'
import type { InstallProgress } from '@openclaw/shared'

const { goToStep } = useWizard()

const progress = reactive<InstallProgress>({
  step: 0,
  totalSteps: 5,
  message: '准备安装...',
  percent: 0,
  error: undefined,
})

const progressStatus = computed(() => {
  if (progress.error) return 'error'
  if (progress.percent >= 100) return 'success'
  return 'default'
})

interface InstallStep {
  message: string
  percent: number
  duration: number
}

const installSteps: InstallStep[] = [
  { message: '正在创建安装目录...', percent: 10, duration: 800 },
  { message: '正在复制程序文件...', percent: 35, duration: 1500 },
  { message: '正在安装 IM 适配器...', percent: 60, duration: 1200 },
  { message: '正在生成配置文件...', percent: 85, duration: 1000 },
  { message: '正在启动服务...', percent: 100, duration: 800 },
]

async function runInstall() {
  progress.error = undefined
  progress.step = 0
  progress.percent = 0
  progress.message = '准备安装...'

  for (let i = 0; i < installSteps.length; i++) {
    const step = installSteps[i]
    progress.step = i + 1
    progress.message = step.message

    // Simulate gradual progress
    const startPercent = i === 0 ? 0 : installSteps[i - 1].percent
    const endPercent = step.percent
    const duration = step.duration
    const interval = 50
    const increments = duration / interval
    const perIncrement = (endPercent - startPercent) / increments

    for (let j = 0; j < increments; j++) {
      await new Promise((resolve) => setTimeout(resolve, interval))
      progress.percent = Math.min(
        endPercent,
        Math.round(startPercent + perIncrement * (j + 1))
      )
    }

    // Simulate a possible error (in real implementation, actual installation logic goes here)
    // For now, we simulate success for all steps
  }

  progress.message = '安装完成！'

  // Wait a moment before navigating to done page
  await new Promise((resolve) => setTimeout(resolve, 600))
  goToStep(5)
}

function retryInstall() {
  runInstall()
}

onMounted(() => {
  runInstall()
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
