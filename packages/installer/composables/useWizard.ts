import { ref, computed } from 'vue'
import type { InstallOptions } from '@openclaw/shared'

export interface WizardStep {
  name: string
  title: string
  path: string
}

const steps: WizardStep[] = [
  { name: 'welcome', title: '欢迎', path: '/' },
  { name: 'mode', title: '安装模式', path: '/mode' },
  { name: 'adapter', title: '适配器', path: '/adapter' },
  { name: 'confirm', title: '确认', path: '/confirm' },
  { name: 'progress', title: '安装中', path: '/progress' },
  { name: 'done', title: '完成', path: '/done' },
]

const currentStep = ref(0)

const installOptions = ref<InstallOptions>({
  installDir: '',
  configDir: '',
  sourceDir: '',
  adapters: [],
  mode: 'standard',
})

export function useWizard() {
  const canGoBack = computed(() => currentStep.value > 0 && currentStep.value < 4)
  const isInstalling = computed(() => currentStep.value === 4)

  function goNext() {
    if (currentStep.value < steps.length - 1) {
      currentStep.value++
      navigateTo(steps[currentStep.value].path)
    }
  }

  function goBack() {
    if (canGoBack.value) {
      currentStep.value--
      navigateTo(steps[currentStep.value].path)
    }
  }

  function goToStep(index: number) {
    if (index >= 0 && index < steps.length) {
      currentStep.value = index
      navigateTo(steps[index].path)
    }
  }

  return {
    currentStep,
    steps,
    installOptions,
    goNext,
    goBack,
    goToStep,
    canGoBack,
    isInstalling,
  }
}
