# OpenClaw 安装器

OpenClaw AI 助手的跨平台安装器与配置管理系统。基于 Electron + Nuxt 3 的图形化安装向导，以及安装后常驻的 Web 配置服务。支持 Windows、macOS 和 Linux，具备离线安装能力。

## 项目概述

- **图形化安装器**: Electron + Nuxt 3 桌面应用，向导式点击安装
- **Web 配置服务**: Nuxt 全栈应用，安装后常驻后台，浏览器访问 `localhost:18080` 管理配置
- **跨平台支持**: Windows (amd64/arm64)、macOS (Intel/Apple Silicon)、Linux (amd64/arm64)
- **离线安装**: 安装包内嵌所有依赖，安装过程零网络请求
- **IM 适配器**: 预配置支持企业微信、钉钉、飞书
- **中国大陆友好**: 不依赖境外服务，构建使用 npmmirror

## 技术栈

- **运行时**: Node.js 20 LTS
- **包管理**: pnpm workspace (Monorepo)
- **框架**: Nuxt 3 + Vue 3
- **语言**: TypeScript
- **UI 组件库**: Naive UI
- **桌面应用**: Electron + electron-builder
- **Nuxt-Electron 集成**: nuxt-electron 模块

## 项目结构

```
openclaw/
├── packages/
│   ├── installer/              # Electron + Nuxt 安装向导
│   │   ├── electron/           # Electron 主进程 (main.ts, preload.ts)
│   │   ├── pages/              # 安装步骤页面
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
│   │   │   └── logs.vue        # 日志查看（分页加载）
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
│       └── utils/              # 工具函数（平台检测等）
│
├── adapters/                   # IM 适配器配置 (企业微信/钉钉/飞书)
├── docs/                       # 文档
│   ├── plans/                  # 设计文档
│   └── *.md                    # 用户指南、架构、测试计划等
├── scripts/                    # 构建与发布脚本
├── pnpm-workspace.yaml
├── package.json
└── tsconfig.json
```

## 常用命令

### 开发

```bash
# 安装依赖
pnpm install

# 启动安装器开发模式
pnpm --filter @openclaw/installer dev

# 启动配置服务开发模式
pnpm --filter @openclaw/config-server dev

# 代码检查
pnpm lint

# 运行测试
pnpm test
```

### 构建

```bash
# 构建所有包
pnpm build

# 构建安装器（Electron 打包）
pnpm --filter @openclaw/installer build

# 构建配置服务
pnpm --filter @openclaw/config-server build
```

### 发布

```bash
# 创建发布包（各平台安装器）
pnpm --filter @openclaw/installer package
```

## 架构

### 组件

1. **Electron 安装器** (`packages/installer/`)
   - Electron + Nuxt 3 桌面应用
   - 六步线性向导：欢迎 → 模式 → 适配器 → 确认 → 进度 → 完成
   - 默认选项覆盖 90% 场景，连点"下一步"即可安装
   - 安装完成自动打开浏览器进入配置页

2. **Web 配置服务** (`packages/config-server/`)
   - Nuxt 全栈应用，常驻后台运行
   - 浏览器访问 `http://localhost:18080`
   - 四个页面：仪表盘、API 配置、IM 适配器、日志
   - 仪表盘整合：服务状态/启停、Token 用量统计、版本信息

3. **共享包** (`packages/shared/`)
   - 共享 Vue 组件、TypeScript 类型、工具函数
   - 被安装器和配置服务共同引用

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

### 跨平台打包

| 平台 | 安装器格式 | 配置服务 |
|------|-----------|---------|
| Windows amd64/arm64 | `.exe` (NSIS) | Node.js 服务 + Windows Service |
| macOS Intel/Apple Silicon | `.dmg` | Node.js 服务 + launchd |
| Linux amd64/arm64 | `.AppImage` / `.deb` | Node.js 服务 + systemd |

## 网络连通性策略

目标用户在中国大陆，安装和使用过程中不依赖境内无法访问的服务。

- **安装过程完全离线**: 安装包内嵌所有依赖，零网络请求
- **不使用境外 CDN**: 字体、图标、CSS 全部打包在本地
- **npm 镜像**: `.npmrc` 配置 `registry=https://registry.npmmirror.com`
- **Electron 镜像**: `ELECTRON_MIRROR=https://npmmirror.com/mirrors/electron/`
- **更新检查**: 使用国内可达的服务器，不依赖 GitHub API

## 开发工作流

### 添加功能

1. **先写测试** - 遵循 TDD 方法
2. **实现功能** - 平台专用代码保持隔离
3. **本地测试** - 开发模式快速迭代
4. **更新文档** - 用户可见更改同步更新文档
5. **构建验证** - 确保各平台打包正常

### 添加 IM 适配器

1. 在 `adapters/<name>/adapter.json` 创建适配器配置
2. 在 `adapters/<name>/` 添加配置模板
3. 在 `packages/config-server/` 添加适配器配置页面逻辑
4. 添加测试
5. 更新文档

## 测试策略

### 单元测试

- 位置: 源文件旁的 `*.test.ts` / `*.spec.ts`
- 运行: `pnpm test`
- 覆盖率目标: 70%+

### 平台测试矩阵

| 平台 | 架构 | 全新安装 | 升级 | 卸载 |
|------|------|----------|------|------|
| Windows 11 | amd64 | ✓ | ✓ | ✓ |
| Windows 10 | amd64 | ✓ | ✓ | ✓ |
| macOS 14 | arm64 | ✓ | ✓ | ✓ |
| macOS 13 | amd64 | ✓ | ✓ | ✓ |
| Ubuntu 22.04 | amd64 | ✓ | ✓ | ✓ |
| Ubuntu 22.04 | arm64 | ✓ | ✓ | ✓ |

## 文档规划

面向计算机小白，以清晰文字为主，关键步骤配图。

| 文档 | 内容 |
|------|------|
| **快速上手指南** | 从下载到配置完成的完整流程 |
| **Windows 安装 SOP** | Windows 下安装全流程 |
| **macOS 安装 SOP** | macOS 下安装全流程，含系统安全提示处理（配图） |
| **Linux 安装 SOP** | Linux 下安装全流程 |
| **配置指南** | Web 配置界面各页面操作说明 |
| **常见问题 FAQ** | 安装失败、配置不生效等常见问题 |

**文档原则：**
- 以文字为主，描述清晰准确
- 只在关键易混淆步骤配图
- 零术语，用大白话描述
- 覆盖常见错误场景

## CI/CD

GitHub Actions 工作流在 `.github/workflows/`：
- 推送/PR 时构建和测试
- 跨平台打包
- 标签推送时自动创建发布

## 设计文档

- **安装器重构设计**: `docs/plans/2026-03-01-installer-redesign-design.md`

## 注意事项

- 所有面向用户的文档均为中文
- 目标用户是计算机小白，UI 和文档需极简易懂
- `dist/` 和 `build/` 包含构建产物，已 gitignore
- 安装过程零网络请求，所有依赖内嵌
- 构建时使用 npmmirror 镜像源
