<template>
  <div class="dashboard">
    <!-- 服务状态 -->
    <n-card title="服务状态" style="margin-bottom: 16px;">
      <div class="service-row">
        <div class="service-info">
          <n-tag :type="serviceRunning ? 'success' : 'error'" size="large">
            {{ serviceRunning ? '运行中' : '已停止' }}
          </n-tag>
          <span v-if="serviceStartTime" class="start-time">
            启动时间: {{ serviceStartTime }}
          </span>
        </div>
        <n-button
          :type="serviceRunning ? 'error' : 'success'"
          :loading="serviceLoading"
          @click="toggleService"
        >
          {{ serviceRunning ? '关闭' : '开启' }}
        </n-button>
      </div>
    </n-card>

    <!-- Token 用量 -->
    <n-card title="Token 用量" style="margin-bottom: 16px;">
      <template #header-extra>
        <n-radio-group v-model:value="tokenDays" size="small" @update:value="fetchTokens">
          <n-radio-button :value="7">7 天</n-radio-button>
          <n-radio-button :value="30">30 天</n-radio-button>
          <n-radio-button :value="90">90 天</n-radio-button>
        </n-radio-group>
      </template>

      <n-statistic label="总用量" :value="tokenTotal" style="margin-bottom: 16px;" />

      <div class="chart-container" v-if="tokenDaily.length > 0">
        <div class="chart-bars">
          <div
            v-for="item in tokenDaily"
            :key="item.date"
            class="chart-bar-wrapper"
          >
            <div
              class="chart-bar"
              :style="{ height: getBarHeight(item.tokens) + 'px' }"
              :title="`${item.date}: ${item.tokens} tokens`"
            />
            <div class="chart-label">{{ formatDate(item.date) }}</div>
          </div>
        </div>
      </div>
      <n-empty v-else description="暂无用量数据" />
    </n-card>

    <!-- 版本信息 -->
    <n-card title="版本信息">
      <n-descriptions bordered :column="1" label-placement="left">
        <n-descriptions-item label="版本">2.0.0</n-descriptions-item>
        <n-descriptions-item label="平台">{{ platformString }}</n-descriptions-item>
      </n-descriptions>
    </n-card>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import {
  NCard,
  NTag,
  NButton,
  NStatistic,
  NRadioGroup,
  NRadioButton,
  NDescriptions,
  NDescriptionsItem,
  NEmpty,
  useMessage,
} from 'naive-ui'

const message = useMessage()

const serviceRunning = ref(false)
const serviceStartTime = ref('')
const serviceLoading = ref(false)

const tokenDays = ref(7)
const tokenTotal = ref(0)
const tokenDaily = ref<Array<{ date: string; tokens: number }>>([])

const platformString = ref('')

onMounted(() => {
  // 检测平台
  const ua = navigator.userAgent
  if (ua.includes('Win')) platformString.value = 'Windows'
  else if (ua.includes('Mac')) platformString.value = 'macOS'
  else if (ua.includes('Linux')) platformString.value = 'Linux'
  else platformString.value = navigator.platform

  fetchServiceStatus()
  fetchTokens()
})

async function fetchServiceStatus() {
  try {
    const res = await $fetch<{ success: boolean; data?: { running: boolean; startTime?: string } }>('/api/service')
    if (res.success && res.data) {
      serviceRunning.value = res.data.running
      serviceStartTime.value = res.data.startTime || ''
    }
  } catch {
    // 获取服务状态失败，保持默认值
  }
}

async function toggleService() {
  serviceLoading.value = true
  try {
    const action = serviceRunning.value ? 'stop' : 'start'
    const res = await $fetch<{ success: boolean; error?: string }>('/api/service', {
      method: 'POST',
      body: { action },
    })
    if (res.success) {
      message.success(action === 'start' ? '服务已启动' : '服务已停止')
      await fetchServiceStatus()
    } else {
      message.error(res.error || '操作失败')
    }
  } catch (err: any) {
    message.error('操作失败: ' + (err.message || '未知错误'))
  } finally {
    serviceLoading.value = false
  }
}

async function fetchTokens() {
  try {
    const res = await $fetch<{
      success: boolean
      data?: { total: number; daily: Array<{ date: string; tokens: number }> }
    }>('/api/tokens', {
      params: { days: tokenDays.value },
    })
    if (res.success && res.data) {
      tokenTotal.value = res.data.total
      tokenDaily.value = res.data.daily
    }
  } catch {
    // 获取 token 数据失败
  }
}

function getBarHeight(tokens: number): number {
  if (tokenDaily.value.length === 0) return 0
  const max = Math.max(...tokenDaily.value.map((d) => d.tokens), 1)
  return Math.max(4, (tokens / max) * 120)
}

function formatDate(dateStr: string): string {
  const parts = dateStr.split('-')
  if (parts.length >= 3) {
    return `${parts[1]}/${parts[2]}`
  }
  return dateStr
}
</script>

<style scoped>
.dashboard {
  max-width: 800px;
}

.service-row {
  display: flex;
  align-items: center;
  justify-content: space-between;
}

.service-info {
  display: flex;
  align-items: center;
  gap: 16px;
}

.start-time {
  color: #999;
  font-size: 13px;
}

.chart-container {
  overflow-x: auto;
}

.chart-bars {
  display: flex;
  align-items: flex-end;
  gap: 4px;
  min-height: 140px;
  padding-top: 8px;
}

.chart-bar-wrapper {
  display: flex;
  flex-direction: column;
  align-items: center;
  flex: 1;
  min-width: 20px;
}

.chart-bar {
  width: 100%;
  max-width: 32px;
  border-radius: 4px 4px 0 0;
  background: linear-gradient(180deg, #63e2b7 0%, #18a058 100%);
  transition: height 0.3s ease;
}

.chart-label {
  font-size: 10px;
  color: #888;
  margin-top: 4px;
  white-space: nowrap;
}
</style>
