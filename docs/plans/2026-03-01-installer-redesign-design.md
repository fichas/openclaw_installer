# OpenClaw 安装器重构设计

> 日期: 2026-03-01
> 状态: 已批准

## 背景

当前安装器是纯 CLI 工具，不符合最初规划的图形化安装体验。安装后也缺少 Web 配置界面来管理 API Key 和其他配置。目标用户是计算机小白，需要"双击即装、打开即配"的体验。

## 决策记录

| 决策项 | 选择 | 理由 |
|--------|------|------|
| 安装器形态 | Electron | 跨平台桌面应用，用户双击即可运行 |
| 前端框架 | Nuxt 3 (Vue 3) | 安装器和配置服务统一技术栈 |
| UI 组件库 | Naive UI | Vue 3 原生、TypeScript 友好、中文文档完善 |
| 配置服务 | Nuxt 全栈 (server routes) | 复用 Nuxt 能力，无需额外后端框架 |
| 项目结构 | Monorepo (pnpm workspace) | 共享组件和类型，独立打包部署 |
| 旧代码 | 完全移除 installer/ 和 wails-installer/ | 全面转向 Electron + Nuxt |

## 架构概览

两个独立应用，通过 shared 包共享代码：

```
用户双击安装包
       ↓
┌──────────────────┐
│  Electron 安装器   │  ← 安装向导，装完即退出
│  (Nuxt 渲染)      │
└────────┬─────────┘
         ↓ 安装完成，启动后台服务
┌──────────────────┐
│  Nuxt 配置服务     │  ← 常驻后台，浏览器访问 localhost:18080
│  (Node.js)        │
└──────────────────┘
```

## 项目结构

```
openclaw/
├── packages/
│   ├── installer/              # Electron + Nuxt 安装向导
│   │   ├── electron/           # Electron 主进程
│   │   │   ├── main.ts         # 主进程入口
│   │   │   └── preload.ts      # 预加载脚本
│   │   ├── pages/              # Nuxt 页面（安装步骤）
│   │   │   ├── index.vue       # 欢迎页
│   │   │   ├── mode.vue        # 安装模式选择
│   │   │   ├── adapter.vue     # IM 适配器选择
│   │   │   ├── confirm.vue     # 安装确认
│   │   │   ├── progress.vue    # 安装进度
│   │   │   └── done.vue        # 完成页
│   │   ├── nuxt.config.ts
│   │   └── package.json
│   │
│   ├── config-server/          # Nuxt 全栈 Web 配置服务
│   │   ├── pages/
│   │   │   ├── index.vue       # 仪表盘（状态 + Token 用量 + 服务开关）
│   │   │   ├── apikeys.vue     # API Key 管理
│   │   │   ├── adapters.vue    # IM 适配器配置
│   │   │   └── logs.vue        # 日志查看
│   │   ├── server/
│   │   │   ├── api/config.ts   # 配置读写 API
│   │   │   ├── api/service.ts  # 服务启停 API
│   │   │   ├── api/tokens.ts   # Token 用量查询 API
│   │   │   └── api/logs.ts     # 日志查询 API（分页）
│   │   ├── nuxt.config.ts
│   │   └── package.json
│   │
│   └── shared/                 # 共享代码
│       ├── components/         # 共享 Vue 组件
│       ├── types/              # TypeScript 类型定义
│       ├── utils/              # 工具函数（平台检测等）
│       └── package.json
│
├── adapters/                   # IM 适配器配置（保留现有）
├── docs/                       # 文档
├── scripts/                    # 构建与发布脚本
├── pnpm-workspace.yaml
├── package.json
└── tsconfig.json
```

## 技术栈

- **运行时**: Node.js 20 LTS
- **包管理**: pnpm workspace
- **框架**: Nuxt 3 + Vue 3
- **语言**: TypeScript
- **UI 组件库**: Naive UI
- **桌面应用**: Electron + electron-builder
- **Nuxt-Electron 集成**: nuxt-electron 模块

## 安装器交互流程

线性向导，面向小白用户，默认选项覆盖 90% 场景：

```
步骤1: 欢迎页
  "欢迎安装 OpenClaw" + 版本号
  [开始安装]
       ↓
步骤2: 安装模式
  ○ 标准安装（推荐）— 安装到默认目录，自动配置
  ○ 自定义安装 — 选择安装目录
  [上一步] [下一步]
       ↓
步骤3: IM 适配器选择
  ☑ 企业微信  ☐ 钉钉  ☐ 飞书
  （可多选，至少选一个）
  [上一步] [下一步]
       ↓
步骤4: 安装确认
  安装目录 / 已选适配器 / 所需空间
  [上一步] [开始安装]
       ↓
步骤5: 安装进度
  进度条 + 当前步骤说明（不可返回）
       ↓
步骤6: 完成
  ✓ 安装成功！
  ☑ 立即打开配置页面（浏览器）
  [完成]
```

**交互原则：**
- 连点"下一步"就能装好（默认选项覆盖主流场景）
- 每步可回退（安装进行中除外）
- 安装完成自动打开浏览器进入配置页
- 出错时用大白话提示，告诉用户该怎么做

## Web 配置服务

安装后常驻后台，浏览器访问 `http://localhost:18080`。

### 页面设计

**仪表盘（首页）：**
- 服务状态：运行中/已停止，启动时间，开启/关闭按钮
- Token 用量：按时间段统计，图表展示趋势
- 版本信息，快捷操作入口

**API 配置：**
- 添加/编辑/删除 API Key
- 连接测试（点击验证 Key 是否有效）

**IM 适配器：**
- 企业微信/钉钉/飞书参数配置
- 表单式填写，每个字段带说明和示例

**日志：**
- 转发 OpenClaw 日志
- 分页加载（每页 100 条），不全量拉取
- 按日志级别筛选，支持搜索

### 后台服务托管

| 平台 | 机制 | 说明 |
|------|------|------|
| Windows | Windows Service | 开机自启 |
| macOS | launchd plist | 开机自启 |
| Linux | systemd service | 开机自启 |

### 数据存储

配置以 JSON 文件存储：
- Windows: `%APPDATA%\OpenClaw\config.json`
- macOS: `~/Library/Application Support/OpenClaw/config.json`
- Linux: `~/.config/openclaw/config.json`

## 跨平台打包

| 平台 | 安装器格式 | 配置服务 |
|------|-----------|---------|
| Windows amd64/arm64 | `.exe` (NSIS) | Node.js 服务 + Windows Service |
| macOS Intel/Apple Silicon | `.dmg` | Node.js 服务 + launchd |
| Linux amd64/arm64 | `.AppImage` / `.deb` | Node.js 服务 + systemd |

## 文档规划

面向计算机小白，以清晰文字为主，关键步骤配图。

| 文档 | 内容 |
|------|------|
| **快速上手指南** | 从下载到配置完成的完整流程 |
| **Windows 安装 SOP** | Windows 下安装全流程 |
| **macOS 安装 SOP** | macOS 下安装全流程，含系统安全提示处理（配图） |
| **Linux 安装 SOP** | Linux 下安装全流程 |
| **配置指南** | Web 配置界面各页面操作说明 |
| **常见问题 FAQ** | 安装失败、配置不生效等常见问题，问答式 |

**文档原则：**
- 以文字为主，描述清晰准确
- 只在关键易混淆步骤配图（如 macOS 安全提示弹窗、首次打开配置页面）
- 零术语，用大白话描述
- 覆盖常见错误场景（防火墙弹窗、权限不足等）

## 网络连通性策略

目标用户在中国大陆，需确保安装和使用过程中不依赖境内无法访问的服务。

### 安装过程

- **完全离线安装**：安装包内嵌所有依赖（Node.js runtime、npm 包、前端资源），安装过程零网络请求
- **不使用 CDN 外链**：所有前端资源（字体、图标、CSS）打包在本地，不引用 Google Fonts、unpkg、cdnjs 等境外 CDN
- **npm 镜像**：开发和 CI 构建时使用 npmmirror（淘宝镜像），`.npmrc` 中预配置 `registry=https://registry.npmmirror.com`

### 配置服务运行时

- **API Key 验证**：连接测试时直接请求用户配置的 AI 服务地址，不经过第三方中转
- **更新检查**：检查更新使用国内可达的服务器（如自建或 Gitee），不依赖 GitHub API
- **日志和 Token 用量**：纯本地数据，不涉及外部请求

### 开发和构建

- **CI/CD**：GitHub Actions 构建时，依赖安装使用镜像源
- **Electron 下载**：构建时配置 `ELECTRON_MIRROR=https://npmmirror.com/mirrors/electron/`
- **electron-builder 下载**：配置 `ELECTRON_BUILDER_BINARIES_MIRROR=https://npmmirror.com/mirrors/electron-builder-binaries/`

## 需要移除的旧代码

- `installer/` — Go CLI 安装器
- `wails-installer/` — Wails GUI 安装器
- 相关构建脚本中的 Go 编译部分
