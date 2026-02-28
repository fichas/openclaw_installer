# OpenClaw 2.0

OpenClaw 安装与配置系统（Electron 安装器 + Nuxt 配置服务 + Go updater）。

## 架构

- `packages/installer`: Electron + Nuxt 安装向导（一次性运行）
- `packages/config-server`: Nuxt 全栈配置服务（默认 `http://localhost:18080`）
- `packages/shared`: 共享类型与工具函数
- `updater/`: Go 更新器与更新链路

## 开发环境

- Node.js >= 20
- pnpm >= 9
- Go >= 1.22（用于 updater）

## 常用命令

```bash
# 安装依赖
pnpm install

# 安装器开发模式
pnpm dev:installer

# 配置服务开发模式
pnpm dev:config

# Node 工作区测试（类型检查门禁）
pnpm test

# 构建配置服务
pnpm build:config

# 构建安装器
pnpm build:installer

# Go updater 测试
cd updater && go test ./... && go test -race ./...
```

## 目录结构

```text
openclaw/
├── packages/
│   ├── installer/
│   ├── config-server/
│   └── shared/
├── updater/
├── adapters/
├── scripts/
├── docs/
└── .github/workflows/
```

## 发布流程

- CI workflow: `.github/workflows/build.yml`
- 发布前需通过：
  - `pnpm test`
  - `cd updater && go test ./... && go test -race ./...`
  - `pnpm build:config`
  - 安装器关键流程手动回归（模式页 -> 进度页 -> 完成页）

## 文档

- [快速上手](docs/QUICK_START.md)
- [配置指南](docs/CONFIG_GUIDE.md)
- [常见问题](docs/FAQ.md)
- [重构设计](docs/plans/2026-03-01-installer-redesign-design.md)
- [实施计划](docs/plans/2026-03-01-installer-redesign-plan.md)
