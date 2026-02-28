<template>
  <n-layout has-sider style="height: 100vh">
    <n-layout-sider
      bordered
      :width="200"
      :collapsed-width="64"
      collapse-mode="width"
      :collapsed="collapsed"
      show-trigger
      @collapse="collapsed = true"
      @expand="collapsed = false"
      :native-scrollbar="false"
      style="height: 100vh"
    >
      <div class="logo-area">
        <span v-if="!collapsed" class="logo-text">OpenClaw</span>
        <span v-else class="logo-icon">OC</span>
      </div>
      <n-menu
        :value="activeKey"
        :options="menuOptions"
        :collapsed="collapsed"
        :collapsed-width="64"
        :collapsed-icon-size="22"
        @update:value="handleMenuSelect"
      />
    </n-layout-sider>
    <n-layout>
      <n-layout-header bordered style="height: 56px; padding: 0 24px; display: flex; align-items: center;">
        <h2 style="margin: 0; font-size: 18px; font-weight: 600;">{{ pageTitle }}</h2>
      </n-layout-header>
      <n-layout-content content-style="padding: 24px;" :native-scrollbar="false">
        <slot />
      </n-layout-content>
    </n-layout>
  </n-layout>
</template>

<script setup lang="ts">
import { ref, computed, h } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import {
  NLayout,
  NLayoutSider,
  NLayoutHeader,
  NLayoutContent,
  NMenu,
  NIcon,
} from 'naive-ui'
import type { MenuOption } from 'naive-ui'

const route = useRoute()
const router = useRouter()
const collapsed = ref(false)

const activeKey = computed(() => route.path)

const pageTitleMap: Record<string, string> = {
  '/': '仪表盘',
  '/apikeys': 'API 配置',
  '/adapters': 'IM 适配器',
  '/logs': '日志',
}

const pageTitle = computed(() => pageTitleMap[route.path] || '仪表盘')

const menuOptions: MenuOption[] = [
  {
    label: '仪表盘',
    key: '/',
    icon: () => h('span', { style: 'font-size: 18px' }, '📊'),
  },
  {
    label: 'API 配置',
    key: '/apikeys',
    icon: () => h('span', { style: 'font-size: 18px' }, '🔑'),
  },
  {
    label: 'IM 适配器',
    key: '/adapters',
    icon: () => h('span', { style: 'font-size: 18px' }, '💬'),
  },
  {
    label: '日志',
    key: '/logs',
    icon: () => h('span', { style: 'font-size: 18px' }, '📋'),
  },
]

function handleMenuSelect(key: string) {
  router.push(key)
}
</script>

<style scoped>
.logo-area {
  height: 56px;
  display: flex;
  align-items: center;
  justify-content: center;
  border-bottom: 1px solid rgba(255, 255, 255, 0.09);
}

.logo-text {
  font-size: 20px;
  font-weight: 700;
  color: #63e2b7;
  letter-spacing: 1px;
}

.logo-icon {
  font-size: 18px;
  font-weight: 700;
  color: #63e2b7;
}
</style>
