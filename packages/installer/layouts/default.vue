<template>
  <n-config-provider :theme="darkTheme" :locale="zhCN" :date-locale="dateZhCN">
    <n-global-style />
    <div class="installer-layout">
      <header class="installer-header">
        <h1 class="installer-title">OpenClaw 安装器</h1>
        <div class="installer-steps">
          <n-steps :current="currentStep + 1" size="small">
            <n-step
              v-for="(step, index) in steps"
              :key="index"
              :title="step.title"
            />
          </n-steps>
        </div>
      </header>
      <main class="installer-content">
        <slot />
      </main>
    </div>
  </n-config-provider>
</template>

<script setup lang="ts">
import { darkTheme } from 'naive-ui'
import { zhCN, dateZhCN } from 'naive-ui'
import { NConfigProvider, NGlobalStyle, NSteps, NStep } from 'naive-ui'
import { useWizard } from '~/composables/useWizard'

const { currentStep, steps } = useWizard()
</script>

<style>
* {
  margin: 0;
  padding: 0;
  box-sizing: border-box;
}

body {
  background-color: #0a0c10;
  color: #e0e0e0;
  font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', 'PingFang SC',
    'Microsoft YaHei', sans-serif;
  overflow: hidden;
  user-select: none;
}

.installer-layout {
  display: flex;
  flex-direction: column;
  height: 100vh;
  background-color: #0a0c10;
}

.installer-header {
  padding: 20px 32px 0;
  flex-shrink: 0;
}

.installer-title {
  font-size: 18px;
  font-weight: 600;
  color: #fff;
  margin-bottom: 16px;
  text-align: center;
}

.installer-steps {
  padding: 0 20px 16px;
  border-bottom: 1px solid rgba(255, 255, 255, 0.08);
}

.installer-content {
  flex: 1;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  padding: 24px 32px;
  overflow-y: auto;
}
</style>
