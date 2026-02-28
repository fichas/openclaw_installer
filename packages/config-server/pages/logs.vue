<template>
  <n-space vertical :size="16">
    <!-- 筛选栏 -->
    <n-space>
      <n-select
        v-model:value="level"
        :options="levelOptions"
        placeholder="日志级别"
        clearable
        style="width: 140px;"
        @update:value="fetchLogs"
      />
      <n-input
        v-model:value="search"
        placeholder="搜索日志内容"
        clearable
        style="width: 240px;"
        @clear="fetchLogs"
        @keyup.enter="fetchLogs"
      >
        <template #suffix>
          <n-button text size="small" @click="fetchLogs">搜索</n-button>
        </template>
      </n-input>
    </n-space>

    <!-- 日志列表 -->
    <n-card>
      <n-spin :show="loading">
        <div v-if="logs.length > 0" class="log-container">
          <pre v-for="(line, i) in logs" :key="i" class="log-line" :class="getLogLevel(line)">{{ line }}</pre>
        </div>
        <n-empty v-else description="暂无日志" />
      </n-spin>
    </n-card>

    <!-- 分页 -->
    <n-space justify="end">
      <n-pagination
        v-model:page="page"
        :page-count="pageCount"
        :page-size="pageSize"
        @update:page="fetchLogs"
      />
    </n-space>
  </n-space>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'

const loading = ref(false)
const logs = ref<string[]>([])
const total = ref(0)
const page = ref(1)
const pageSize = 100
const level = ref<string | null>(null)
const search = ref('')

const pageCount = computed(() => Math.ceil(total.value / pageSize))

const levelOptions = [
  { label: 'INFO', value: 'INFO' },
  { label: 'WARN', value: 'WARN' },
  { label: 'ERROR', value: 'ERROR' },
  { label: 'DEBUG', value: 'DEBUG' },
]

function getLogLevel(line: string): string {
  if (line.includes('[ERROR]')) return 'log-error'
  if (line.includes('[WARN]')) return 'log-warn'
  if (line.includes('[INFO]')) return 'log-info'
  if (line.includes('[DEBUG]')) return 'log-debug'
  return ''
}

onMounted(() => {
  fetchLogs()
})

async function fetchLogs() {
  loading.value = true
  try {
    const params = new URLSearchParams({
      page: page.value.toString(),
      pageSize: pageSize.toString(),
    })
    if (level.value) params.set('level', level.value)
    if (search.value) params.set('search', search.value)

    const res = await $fetch(`/api/logs?${params}`)
    logs.value = res.data.logs
    total.value = res.data.total
  } catch {}
  loading.value = false
}
</script>

<style scoped>
.log-container {
  max-height: 500px;
  overflow-y: auto;
  font-family: 'Courier New', monospace;
  font-size: 13px;
}
.log-line {
  margin: 0;
  padding: 2px 8px;
  white-space: pre-wrap;
  word-break: break-all;
  border-bottom: 1px solid #1a1a1a;
}
.log-error { color: #ef4444; }
.log-warn { color: #f59e0b; }
.log-info { color: #22c55e; }
.log-debug { color: #888; }
</style>
