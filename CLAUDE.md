# OpenClaw 安装器

OpenClaw AI 助手的跨平台安装器，提供基于 Web 的配置界面。支持 Windows、macOS 和 Linux，具备离线 U 盘部署能力。

## 项目概述

OpenClaw 安装器是一个基于 Go 的安装工具，提供以下功能：
- **跨平台支持**: Windows (amd64/arm64)、macOS (Intel/Apple Silicon)、Linux (amd64/arm64)
- **离线安装**: 无需网络连接，直接从 U 盘部署
- **Web 配置**: 内置 HTTP 服务器 (端口 18080) 进行配置
- **IM 适配器**: 预配置支持企业微信、钉钉、飞书
- **图形安装器**: 基于 Wails 的桌面应用，一键安装
- **自动更新**: 内置更新机制，无缝升级

## 快速开始

```bash
# 构建所有组件
./scripts/build.sh all

# 构建指定平台
./scripts/build.sh windows
./scripts/build.sh macos
./scripts/build.sh linux

# 构建 Wails 图形安装器
cd wails-installer && ./build.sh

# 创建发布包
./scripts/create-release.sh
```

## 项目结构

```
├── adapters/           # IM 适配器配置 (企业微信/钉钉/飞书)
├── build/              # 构建产物 (git 忽略)
├── dist/               # 分发包 (git 忽略)
├── docs/               # 文档
│   ├── architecture.md # 系统架构
│   ├── USER_GUIDE.md   # 用户文档 (中文)
│   └── *.md           # 设计文档、测试计划等
├── frontend/           # Web 配置界面 (静态文件)
├── installer/          # 命令行安装器 (Go)
│   ├── main.go        # CLI 入口
│   ├── server.go      # Web UI 的 HTTP 服务器
│   ├── config.go      # 配置管理
│   └── *_test.go      # 单元测试
├── release/            # 发布包模板
│   └── OpenClaw-v1.0.0/ # 各平台安装脚本
├── scripts/            # 构建和工具脚本
│   ├── build.sh       # 主构建脚本
│   ├── build-*.sh     # 各平台专用构建
│   └── create-release.sh # 发布打包
├── updater/            # 自动更新系统
├── usb-template/       # U 盘部署模板文件
└── wails-installer/    # 基于 Wails 的图形安装器
    ├── main.go        # Wails 入口
    ├── frontend/      # GUI 前端 (HTML/CSS/JS)
    └── internal/      # 安装器逻辑模块
```

## 常用命令

### 开发

```bash
# 运行测试
cd installer && go test ./...

# 运行并生成覆盖率报告
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out -o coverage.html

# 格式化代码
go fmt ./...

# 代码检查
golangci-lint run
```

### 构建

```bash
# 构建所有平台
./scripts/build.sh all

# 指定版本构建
VERSION=1.1.0 ./scripts/build.sh macos

# 构建 Wails GUI
cd wails-installer && ./build.sh

# 交叉编译 (需要 Docker)
./scripts/build-linux.sh --docker
```

### 测试

```bash
# 运行所有测试
cd installer && go test -v ./...

# 运行集成测试
go test -tags=integration -v ./...

# 测试指定组件
go test -v -run TestInstaller ./...
```

### 发布

```bash
# 创建发布包
./scripts/create-release.sh

# 版本升级
./scripts/create-release.sh --version 1.1.0

# 清理并重建
./scripts/create-release.sh --clean --build
```

## 架构

### 组件

1. **命令行安装器** (`installer/`)
   - 基于 Go 的 CLI 工具
   - 使用 `//go:embed` 嵌入静态 Web UI
   - HTTP 服务器提供配置界面 (端口 18080)
   - 跨平台编译支持

2. **Wails 图形安装器** (`wails-installer/`)
   - 使用 Wails v2 的桌面应用
   - 原生窗口，无控制台
   - 四步向导：欢迎 → 模式 → 适配器 → 进度
   - 平台检测和自动架构选择

3. **Web 配置界面** (`frontend/`)
   - 单文件 HTML/CSS/JS 应用
   - 无需构建步骤
   - 由嵌入式 HTTP 服务器提供
   - 配置验证和预览

4. **自动更新器** (`updater/`)
   - 后台更新检查
   - 增量更新减少下载
   - 失败时支持回滚
   - 各平台专用更新机制

### 构建系统

- **主构建**: `scripts/build.sh` - 主要协调
- **平台专用**: `scripts/build-{windows,macos,linux}.sh`
- **Wails**: `wails-installer/build.sh` - GUI 应用构建
- **发布**: `scripts/create-release.sh` - 打包创建

### 配置流程

1. 用户运行安装器 (CLI 或 GUI)
2. 自动检测平台和架构
3. 选择安装模式 (系统/用户/便携)
4. 自定义适配器配置
5. 文件安装到目标目录
6. 更新 PATH (系统/用户)
7. 创建快捷方式 (平台专用)
8. 注册服务 (Linux/macOS 可选)

## 开发工作流

### 添加功能

1. **先写测试** - 遵循 TDD 方法
2. **实现功能** - 平台专用代码保持隔离
3. **本地测试** - 使用 `./scripts/build.sh` 快速迭代
4. **更新文档** - 用户可见更改更新 USER_GUIDE.md
5. **创建发布** - 用 `./scripts/create-release.sh --build` 验证

### 添加 IM 适配器

1. 在 `adapters/<name>/adapter.json` 创建适配器配置
2. 在 `adapters/<name>/` 添加配置模板
3. 如需自定义逻辑，更新 `installer/config.go`
4. 在 `installer/config_test.go` 添加测试
5. 更新文档

### 平台专用更改

- 使用 `internal/platform/` 进行平台抽象
- 使用构建标签分离平台代码
- 在目标平台测试或使用 Docker 构建 Linux

## 测试策略

### 单元测试

- 位置: 源文件旁的 `*_test.go`
- 运行: `go test ./...`
- 覆盖率目标: 70%+

### 集成测试

- 位置: `integration_test.go`
- 运行: `go test -tags=integration`
- 测试端到端安装流程

### 手动测试

1. 为目标平台构建
2. 复制到测试机或虚拟机
3. 测试全新安装
4. 测试升级路径
5. 测试卸载
6. 验证服务状态 (Linux/macOS)

### 平台测试矩阵

| 平台 | 架构 | 全新安装 | 升级 | 卸载 |
|------|------|----------|------|------|
| Windows 11 | amd64 | ✓ | ✓ | ✓ |
| Windows 10 | amd64 | ✓ | ✓ | ✓ |
| macOS 14 | arm64 | ✓ | ✓ | ✓ |
| macOS 13 | amd64 | ✓ | ✓ | ✓ |
| Ubuntu 22.04 | amd64 | ✓ | ✓ | ✓ |
| Ubuntu 22.04 | arm64 | ✓ | ✓ | ✓ |

## 部署

### U 盘部署

1. 运行 `./scripts/create-release.sh`
2. 将 `release/OpenClaw-v1.0.0/` 复制到 U 盘
3. U 盘结构：
   ```
   /OpenClaw/
   ├── windows/
   ├── macos/
   ├── linux/
   ├── shared/
   └── README.txt
   ```
4. 用户运行对应平台的安装脚本

### 网络部署

1. 上传发布包到分发服务器
2. 更新自动更新器的版本清单
3. 通知用户新版本

### 代码签名

**macOS** (见 `docs/macos-signing.md`)：
```bash
./scripts/sign-macos.sh --cert "Developer ID" \
  --input dist/openclaw-installer-darwin-amd64 \
  --output dist/openclaw-installer-darwin-amd64-signed
```

**Windows**: 使用 `signtool.exe` 和代码签名证书

## 关键依赖

- **Go 1.22+**: 主要语言
- **Wails v2**: 桌面 GUI 框架
- **WebView2**: Windows 网页引擎 (需要运行时)

## CI/CD

GitHub Actions 工作流在 `.github/workflows/`：
- 推送/PR 时构建
- 跨平台测试
- 标签推送时自动创建发布

## 资源

- **用户指南**: `docs/USER_GUIDE.md` (中文)
- **架构文档**: `docs/architecture.md`
- **构建指南**: `BUILD.md`
- **测试计划**: `docs/TEST_PLAN.md`

## 注意事项

- 所有面向用户的文档均为中文
- 构建脚本支持 `VERSION` 环境变量
- `dist/` 和 `release/` 包含大文件 - 使用 `.gitignore`
- U 盘模板文件应保持精简以加快复制速度
