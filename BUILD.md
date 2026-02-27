# OpenClaw 安装器 - 构建文档

## 环境要求

- Go 1.22+
- Make
- Git

## 快速构建

```bash
# 克隆仓库
git clone <repo-url>
cd openclaw

# 构建所有平台
make build-all

# 创建 U 盘部署包
make usb-deploy
```

## 详细构建指南

### 1. 安装依赖

```bash
# 检查 Go 版本
go version  # 需要 >= 1.22

# 整理模块依赖
cd installer
go mod tidy
cd ..
```

### 2. 构建目标平台

#### 构建所有平台
```bash
make build-all
```

输出位置：`dist/`

#### 构建特定平台
```bash
# Linux
make build-linux
# 输出：
#   dist/openclaw-installer-linux-amd64
#   dist/openclaw-installer-linux-arm64

# macOS
make build-darwin
# 输出：
#   dist/openclaw-installer-darwin-amd64
#   dist/openclaw-installer-darwin-arm64

# Windows
make build-windows
# 输出：
#   dist/openclaw-installer-windows-amd64.exe
#   dist/openclaw-installer-windows-arm64.exe
```

#### 构建单个平台
```bash
make build-single PLATFORM=linux-amd64
```

### 3. 构建更新程序

```bash
make build-updater
```

### 4. 创建 U 盘部署包

```bash
make usb-deploy
```

输出位置：`usb-deploy/OpenClaw/`

目录结构：
```
usb-deploy/OpenClaw/
├── README.txt                    # 使用说明
├── installers/                   # 6个平台安装器
│   ├── openclaw-installer-darwin-amd64
│   ├── openclaw-installer-darwin-arm64
│   ├── openclaw-installer-linux-amd64
│   ├── openclaw-installer-linux-arm64
│   ├── openclaw-installer-windows-amd64.exe
│   └── openclaw-installer-windows-arm64.exe
├── packages/config-templates/    # 配置文件模板
│   ├── openclaw.yaml.template
│   ├── wecom-adapter.yaml.template
│   ├── dingtalk-adapter.yaml.template
│   └── feishu-adapter.yaml.template
└── autorun/                      # 自动运行脚本
    ├── autorun.inf              # Windows
    ├── install-mac.command      # macOS
    └── install-linux.sh         # Linux
```

### 5. 创建发布包

```bash
make package
```

输出位置：`release/`

包含：
- 各平台压缩包（zip/tar.gz）
- 校验和文件（SHA256）

### 6. 完整发布流程

```bash
make release-all
```

执行：
1. clean - 清理旧构建
2. build-all - 构建所有平台
3. package - 创建发布包
4. usb-deploy - 创建 U 盘部署包

## 开发模式

```bash
# 开发运行（当前平台）
make dev

# 或
make run
```

## 测试

```bash
# 运行所有测试
make test

# 仅测试安装器
cd installer && go test -v ./...

# 编译测试（不运行）
make test-compile
```

## 代码质量

```bash
# 格式化代码
make fmt

# 静态检查
make vet

# 代码检查
make lint
```

## CI/CD 集成

```bash
# CI 构建流程
make ci-build

# CI 测试流程
make ci-test
```

## Docker 构建

```bash
make docker-build
```

## 交叉编译说明

本项目使用 Go 的交叉编译功能：

```bash
# Linux AMD64
GOOS=linux GOARCH=amd64 go build

# Linux ARM64
GOOS=linux GOARCH=arm64 go build

# macOS AMD64
GOOS=darwin GOARCH=amd64 go build

# macOS ARM64 (Apple Silicon)
GOOS=darwin GOARCH=arm64 go build

# Windows AMD64
GOOS=windows GOARCH=amd64 go build

# Windows ARM64
GOOS=windows GOARCH=arm64 go build
```

## 构建产物大小

| 平台 | 架构 | 大小 |
|------|------|------|
| macOS | amd64 | ~6.6M |
| macOS | arm64 | ~6.5M |
| Linux | amd64 | ~6.2M |
| Linux | arm64 | ~6.0M |
| Windows | amd64 | ~6.4M |
| Windows | arm64 | ~6.0M |

## 故障排除

### 构建失败：找不到模块
```bash
cd installer
go mod init openclaw/installer  # 如果 go.mod 不存在
go mod tidy
```

### 交叉编译失败
确保使用 Go 1.22+，支持所有目标平台的交叉编译。

### 权限问题（Linux/macOS）
```bash
chmod +x dist/openclaw-installer-*
```

## Makefile 目标参考

| 目标 | 说明 |
|------|------|
| `all` | 完整构建流程 |
| `build-all` | 构建所有平台 |
| `build-linux` | 构建 Linux 版本 |
| `build-darwin` | 构建 macOS 版本 |
| `build-windows` | 构建 Windows 版本 |
| `build-updater` | 构建更新程序 |
| `usb-deploy` | 创建 U 盘部署包 |
| `package` | 创建发布包 |
| `release-all` | 完整发布流程 |
| `test` | 运行测试 |
| `clean` | 清理构建产物 |
| `dev` / `run` | 开发模式运行 |
