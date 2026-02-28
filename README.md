# OpenClaw 安装器

跨平台 OpenClaw AI 助手安装器，支持企业微信、钉钉、飞书适配器配置。

## 功能特性

- **跨平台支持**: Windows、macOS、Linux (x64 / ARM64)
- **双模式安装**:
  - 🖥️ **Wails GUI** - 图形化一键安装
  - ⌨️ **命令行** - 轻量级 Web 配置界面
- **离线部署**: 从 U 盘直接安装，无需网络
- **自动更新**: 内置更新系统
- **预配置适配器**: 企业微信、钉钉、飞书

## 快速开始

### 方式一：Wails 图形安装器（推荐）

下载对应平台的安装器，双击运行：

```bash
# Windows
./OpenClaw-Installer.exe

# macOS
./OpenClaw-Installer.app

# Linux
./openclaw-installer
```

### 方式二：U 盘离线安装

1. 从 [Releases](../../releases) 下载 `OpenClaw-v1.0.0.zip`
2. 解压到 U 盘根目录
3. 插入目标电脑，运行对应脚本：

```bash
# Windows - 右键以管理员运行
.\install.ps1

# macOS
./install-mac.command

# Linux
sudo ./install-linux.sh
```

### 方式三：从源码构建

```bash
# 克隆仓库
git clone https://github.com/yourusername/openclaw-installer.git
cd openclaw-installer

# 构建命令行安装器
./scripts/build.sh all

# 或构建 Wails GUI
cd wails-installer && ./build.sh
```

## 配置界面

安装器启动后会自动打开配置界面：

- **Web 配置**: `http://localhost:18080`
- **配置项**:
  - AI 模型设置 (API Key、模型选择)
  - 企业微信 (CorpID、AgentID、Secret)
  - 钉钉 (AppKey、AppSecret)
  - 飞书 (AppID、AppSecret)

## 项目结构

```
openclaw-installer/
├── installer/          # 命令行安装器 (Go)
├── wails-installer/    # Wails 图形安装器
├── updater/            # 自动更新系统
├── usb-template/       # U 盘部署模板
├── adapters/           # IM 适配器配置
├── scripts/            # 构建脚本
├── docs/               # 文档
└── release/            # 发布包模板
```

## 开发

### 构建

```bash
# 构建所有平台
./scripts/build.sh all

# 构建指定平台
./scripts/build.sh windows    # 或 macos / linux

# 构建 Wails GUI
cd wails-installer && ./build.sh
```

### 测试

```bash
cd installer
go test -v ./...
```

### 创建发布

```bash
./scripts/create-release.sh
```

## 系统要求

| 平台 | 最低版本 | 架构 |
|------|---------|------|
| Windows | Windows 10 | x64, ARM64 |
| macOS | macOS 10.15 (Catalina) | Intel, Apple Silicon |
| Linux | Ubuntu 18.04 / CentOS 7 | x64, ARM64 |

## 文档

- [用户指南](docs/USER_GUIDE.md) - 详细安装和配置说明
- [构建指南](BUILD.md) - 开发和构建文档
- [架构设计](docs/architecture.md) - 系统设计文档
- [CLAUDE.md](CLAUDE.md) - 项目上下文和开发规范

## 许可证

MIT License
