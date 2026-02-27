# OpenClaw 安装器测试计划

## 概述

本文档描述了 OpenClaw 安装器的完整测试计划，包括单元测试、集成测试和跨平台测试。

## 测试文件列表

| 测试文件 | 描述 |
|---------|------|
| `platform_test.go` | 平台检测相关测试 |
| `config_test.go` | 配置生成和验证测试 |
| `installer_test.go` | 文件安装操作测试 |
| `server_test.go` | Web 服务器 API 测试 |
| `validation_test.go` | 表单验证测试 |
| `integration_test.go` | 集成测试 |

## 单元测试

### 1. 平台检测测试 (platform_test.go)

#### 测试用例

| 测试名称 | 描述 | 预期结果 |
|---------|------|---------|
| `TestDetectPlatform` | 检测当前平台 | 返回正确的 OS 和 Arch |
| `TestNormalizeArch` | 架构名称标准化 | amd64/x86_64 统一为 amd64 |
| `TestPlatformIsWindows` | Windows 检测 | 正确识别 Windows 平台 |
| `TestPlatformIsMacOS` | macOS 检测 | 正确识别 Darwin 平台 |
| `TestPlatformIsLinux` | Linux 检测 | 正确识别 Linux 平台 |
| `TestPlatformGetInstallDir` | 安装目录获取 | 返回平台特定的安装路径 |
| `TestPlatformGetConfigDir` | 配置目录获取 | 返回平台特定的配置路径 |
| `TestPlatformGetBinaryName` | 二进制文件名 | Windows 添加 .exe 后缀 |
| `TestCrossPlatformSupport` | 跨平台支持 | 支持所有目标平台组合 |

#### 支持的平台组合

- Windows: amd64, arm64
- macOS: amd64, arm64
- Linux: amd64, arm64

### 2. 配置生成测试 (config_test.go)

#### 测试用例

| 测试名称 | 描述 | 预期结果 |
|---------|------|---------|
| `TestGenerateConfig` | 配置生成 | 生成有效的配置对象 |
| `TestConfigValidate` | 配置验证 | 验证所有必填字段 |
| `TestConfigSaveAndLoad` | 保存和加载 | 配置持久化正确 |
| `TestConfigToJSON` | JSON 转换 | 生成有效的 JSON |
| `TestLoadConfigNotExist` | 加载不存在的配置 | 返回错误 |
| `TestLoadConfigInvalidJSON` | 加载无效 JSON | 返回错误 |
| `TestConfigWithMultipleAdapters` | 多适配器配置 | 支持多个适配器 |

#### 配置验证规则

- Version: 必填
- Server.Port: 1-65535
- Adapters: 至少一个，每个必须有 name 和 type

### 3. 安装器测试 (installer_test.go)

#### 测试用例

| 测试名称 | 描述 | 预期结果 |
|---------|------|---------|
| `TestNewInstaller` | 安装器创建 | 正确初始化安装器 |
| `TestCopyFile` | 文件复制 | 正确复制文件内容和权限 |
| `TestInstallWithOptions` | 带选项安装 | 正确安装二进制文件 |
| `TestInstallWithAdapter` | 带适配器安装 | 同时安装主程序和适配器 |
| `TestUninstall` | 卸载 | 正确删除文件 |
| `TestVerifyInstallation` | 安装验证 | 验证文件存在性 |
| `TestCopyFilePermissions` | 权限保留 | 复制后权限不变 |

### 4. Web 服务器测试 (server_test.go)

#### API 端点测试

| 端点 | 方法 | 测试名称 | 描述 |
|-----|------|---------|------|
| `/api/platform` | GET | `TestHandlePlatform` | 返回平台信息 |
| `/api/status` | GET | `TestHandleStatus` | 返回安装状态 |
| `/api/install` | POST | `TestHandleInstall` | 执行安装 |
| `/api/config` | GET/POST | `TestHandleConfigGet/Post` | 配置管理 |
| `/api/verify` | GET | `TestHandleVerify` | 验证安装 |
| `/` | GET | `TestHandleIndex` | 返回安装页面 |

#### 错误处理测试

| 测试名称 | 描述 |
|---------|------|
| `TestHandlePlatformMethodNotAllowed` | 错误方法返回 405 |
| `TestHandleInstallInvalidJSON` | 无效 JSON 返回错误 |
| `TestHandleIndexNotFound` | 404 处理 |

### 5. 表单验证测试 (validation_test.go)

#### 验证规则测试

| 测试名称 | 验证内容 | 有效值 | 无效值 |
|---------|---------|--------|--------|
| `TestValidatePort` | 端口号 | 1-65535 | 0, -1, 65536 |
| `TestValidateVersion` | 版本号 | 1.0.0, v1.0.0 | 空字符串 |
| `TestValidateAdapterName` | 适配器名称 | 非空字符串 | 空字符串 |
| `TestValidateAdapterType` | 适配器类型 | ollama, openai, anthropic | 空字符串 |
| `TestValidateMultipleAdapters` | 多适配器 | 至少一个有效适配器 | 空数组 |

## 集成测试 (integration_test.go)

### 完整安装流程测试

```
TestFullInstallationFlow
├── 创建临时目录
├── 创建源二进制文件
├── 生成配置
├── 执行安装
├── 验证安装
├── 保存配置
├── 加载并验证配置
└── 检查最终状态
```

### Web API 集成测试

```
TestWebAPIIntegration
├── GetPlatform - 获取平台信息
├── GetStatus - 获取安装状态
└── PostConfig - 提交配置
```

### 跨平台安装测试

测试平台：
- Linux amd64/arm64
- macOS amd64/arm64
- Windows amd64

### 错误处理测试

| 测试场景 | 预期行为 |
|---------|---------|
| 无效源目录 | 返回错误 |
| 权限不足 | 返回错误 |
| 无效配置 | 验证失败 |

### 并发测试

- `TestConcurrentOperations`: 测试并发状态读取

## 跨平台测试清单

### Windows 测试

| 版本 | 架构 | 测试项目 |
|-----|------|---------|
| Windows 10 | amd64 | 安装、卸载、配置生成 |
| Windows 11 | amd64 | 安装、卸载、配置生成 |
| Windows Server 2019 | amd64 | 安装、服务配置 |

#### Windows 特定测试

- [ ] 路径包含空格处理
- [ ] 管理员权限检查
- [ ] .exe 后缀添加
- [ ] Windows 服务注册
- [ ] 注册表写入（如需要）

### macOS 测试

| 版本 | 架构 | 测试项目 |
|-----|------|---------|
| macOS 12 (Monterey) | amd64/arm64 | 安装、卸载 |
| macOS 13 (Ventura) | amd64/arm64 | 安装、卸载 |
| macOS 14 (Sonoma) | amd64/arm64 | 安装、卸载 |

#### macOS 特定测试

- [ ] 应用程序签名检查
- [ ] Gatekeeper 兼容性
- [ ] 权限请求（Accessibility、Files and Folders）
- [ ] Rosetta 2 兼容性（ARM64）

### Linux 测试

| 发行版 | 版本 | 架构 | 测试项目 |
|-------|------|------|---------|
| Ubuntu | 20.04, 22.04, 24.04 | amd64/arm64 | 完整测试 |
| Debian | 11, 12 | amd64/arm64 | 完整测试 |
| CentOS/RHEL | 8, 9 | amd64/arm64 | 完整测试 |
| Fedora | 38, 39, 40 | amd64/arm64 | 完整测试 |
| Alpine | 3.18, 3.19 | amd64/arm64 | 完整测试 |

#### Linux 特定测试

- [ ] systemd 服务文件生成
- [ ] 文件权限设置 (0755)
- [ ] 用户/组权限
- [ ] SELinux 兼容性

## 异常情况测试

### 端口占用

```go
TestPortOccupied
├── 端口已被占用时安装
└── 应提供清晰的错误信息
```

### 权限不足

```go
TestPermissionDenied
├── 非管理员/Root 用户安装到系统目录
└── 应返回权限错误
```

### 磁盘空间不足

```go
TestInsufficientDiskSpace
├── 磁盘空间不足时安装
└── 应提前检查并返回错误
```

### 网络问题（如需要下载）

```go
TestNetworkFailure
├── 网络不可用时
└── 应返回网络错误
```

## 运行测试

### 运行所有测试

```bash
cd /home/chaos/openclaw/installer
go test -v ./...
```

### 运行特定测试文件

```bash
go test -v -run TestDetectPlatform
go test -v -run TestConfig
go test -v -run TestInstall
```

### 运行集成测试

```bash
go test -v -run Integration
```

### 生成覆盖率报告

```bash
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out -o coverage.html
```

## 测试覆盖率目标

| 模块 | 目标覆盖率 |
|-----|-----------|
| platform.go | 90% |
| config.go | 85% |
| installer.go | 80% |
| server.go | 75% |
| 整体 | 80% |

## 持续集成

建议在 CI 中运行以下测试矩阵：

```yaml
strategy:
  matrix:
    os: [ubuntu-latest, macos-latest, windows-latest]
    go: ['1.21', '1.22']
    arch: [amd64, arm64]
    exclude:
      - os: ubuntu-latest
        arch: arm64  # GitHub Actions 不支持 ARM64 Linux
```

## 测试数据

测试使用临时目录，无需清理：

```go
tmpDir := t.TempDir()  // 自动清理
```

## 备注

1. 所有测试文件使用 `_test.go` 后缀
2. 测试函数使用 `TestXxx` 命名规范
3. 子测试使用 `t.Run()` 组织
4. 表格驱动测试用于多场景验证
