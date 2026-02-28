<template>
  <div class="welcome-page">
    <div class="welcome-icon">
      <svg width="80" height="80" viewBox="0 0 80 80" fill="none" xmlns="http://www.w3.org/2000/svg">
        <circle cx="40" cy="40" r="38" stroke="#18a058" stroke-width="3" fill="none" />
        <path
          d="M25 40 C25 28 33 20 40 20 C47 20 55 28 55 40 C55 52 47 60 40 60 C33 60 25 52 25 40Z"
          fill="#18a058"
          opacity="0.15"
        />
        <path
          d="M32 35 C32 30 35 27 40 27 C45 27 48 30 48 35"
          stroke="#18a058"
          stroke-width="2.5"
          stroke-linecap="round"
          fill="none"
        />
        <circle cx="34" cy="40" r="2.5" fill="#18a058" />
        <circle cx="46" cy="40" r="2.5" fill="#18a058" />
        <path
          d="M33 47 C35 50 38 52 40 52 C42 52 45 50 47 47"
          stroke="#18a058"
          stroke-width="2.5"
          stroke-linecap="round"
          fill="none"
        />
      </svg>
    </div>
    <h2 class="welcome-title">欢迎安装 OpenClaw</h2>
    <p class="welcome-version">版本 2.0.0</p>
    <p class="welcome-desc">
      OpenClaw 是您的 AI 智能助手，支持企业微信、钉钉、飞书等多个平台。
      <br />
      安装向导将引导您完成全部配置，只需几步即可开始使用。
    </p>
    <n-button type="primary" size="large" @click="handleStart">
      开始安装
    </n-button>
  </div>
</template>

<script setup lang="ts">
import { NButton } from 'naive-ui'
import { useWizard } from '~/composables/useWizard'

const { goToStep } = useWizard()

function handleStart() {
  goToStep(1)
}
</script>

<style scoped>
.welcome-page {
  display: flex;
  flex-direction: column;
  align-items: center;
  text-align: center;
  gap: 12px;
}

.welcome-icon {
  margin-bottom: 8px;
}

.welcome-title {
  font-size: 28px;
  font-weight: 700;
  color: #fff;
}

.welcome-version {
  font-size: 14px;
  color: rgba(255, 255, 255, 0.45);
}

.welcome-desc {
  font-size: 15px;
  color: rgba(255, 255, 255, 0.65);
  line-height: 1.8;
  max-width: 480px;
  margin-bottom: 12px;
}
</style>
