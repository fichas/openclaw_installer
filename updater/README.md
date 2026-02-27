# OpenClaw 更新程序

## 概述

OpenClaw Updater 是一个独立的 Go 程序，用于检查、下载和安装 OpenClaw 核心及适配器的更新。

## 技术方案选择

选择 **独立的 Go 更新程序**（方案 A），原因如下：

1. **跨平台一致性**: 单一代码库支持 macOS/Windows/Linux
2. **可靠性**: 编译后的二进制文件，无脚本依赖问题
3. **原子性更新**: 支持事务性更新和回滚
4. **版本管理**: 内置版本比较和依赖检查
5. **可测试性**: 易于编写单元测试和集成测试

## 功能特性

- 自动检测当前安装的版本
- 检查远程版本更新
- 并发下载多个组件
- 保留用户配置
- 自动回滚失败更新
- 支持离线更新（从本地目录）
- 详细的日志记录

## 目录结构

```
updater/
├── main.go                      # 程序入口
├── updater.go                   # 核心更新逻辑
├── platform.go                  # 平台检测
├── config.go                    # 配置管理
├── go.mod                       # Go 模块定义
├── scripts/
│   ├── update.sh                # Linux/macOS 包装脚本
│   └── update.bat               # Windows 包装脚本
├── internal/                    # 内部包（扩展用）
│   ├── version/                 # 版本检查
│   ├── download/                # 文件下载
│   ├── backup/                  # 备份管理
│   ├── install/                 # 安装逻辑
│   ├── rollback/                # 回滚机制
│   └── config/                  # 配置保留
├── pkg/
│   └── types/                   # 共享类型定义
└── README.md
```

## 版本管理策略

### 版本号格式

采用语义化版本控制（SemVer）：`MAJOR.MINOR.PATCH`

- MAJOR: 不兼容的 API 变更
- MINOR: 向后兼容的功能添加
- PATCH: 向后兼容的问题修复

### 版本文件格式

```json
{
  "core": {
    "version": "1.2.3",
    "download_url": "https://releases.openclaw.io/v1.2.3/openclaw-{os}-{arch}.{ext}",
    "checksum": "sha256:abc123...",
    "release_date": "2024-01-15T10:30:00Z",
    "release_notes": "https://releases.openclaw.io/v1.2.3/notes.md"
  },
  "adapters": {
    "wecom": {
      "version": "1.1.0",
      "download_url": "...",
      "checksum": "sha256:def456...",
      "min_core_version": "1.0.0"
    },
    "dingtalk": { ... },
    "feishu": { ... }
  }
}
```

### 更新策略

1. **自动更新**: 定期检查（可配置，默认每周）
2. **手动更新**: 用户触发
3. **离线更新**: 从本地目录/U盘更新

## 使用方式

### 命令行用法

```bash
# 检查更新
openclaw-updater -check

# 执行更新（带确认）
openclaw-updater -yes

# 更新特定适配器
openclaw-updater -adapter=wecom -yes

# 模拟更新（不实际执行）
openclaw-updater -dry-run -yes

# 强制更新（即使版本相同）
openclaw-updater -force -yes

# 回滚到上一版本
openclaw-updater -rollback

# 列出可用备份
openclaw-updater -list-backups

# 显示帮助
openclaw-updater -help
```

### 命令行选项

| 选项 | 说明 | 默认值 |
|------|------|--------|
| `-check` | 仅检查更新，不安装 | false |
| `-force` | 强制更新 | false |
| `-dry-run` | 模拟运行 | false |
| `-config` | 配置文件路径 | 自动检测 |
| `-install-dir` | 安装目录 | 自动检测 |
| `-backup-dir` | 备份目录 | 自动检测 |
| `-version-url` | 版本检查 URL | https://api.openclaw.io/versions/latest |
| `-adapter` | 指定适配器 (wecom/dingtalk/feishu/all) | all |
| `-yes` | 自动确认 | false |
| `-rollback` | 回滚到上一版本 | false |
| `-list-backups` | 列出备份 | false |
| `-v` | 详细输出 | false |

### 配置文件

```yaml
# /etc/openclaw/updater.yaml (Linux)
# /usr/local/etc/openclaw/updater.yaml (macOS)
# %ProgramData%\OpenClaw\updater.yaml (Windows)

update:
  check_interval: "7d"           # 检查间隔
  auto_update: false             # 是否自动更新
  channel: "stable"              # 更新通道: stable, beta, dev
  max_backups: 5                 # 最大备份数量

source:
  type: "remote"                 # remote 或 local
  url: "https://releases.openclaw.io"
  local_path: ""                 # 本地源路径

components:
  core: true                     # 更新核心
  adapters:
    - wecom
    - dingtalk
    - feishu

backup:
  keep_count: 3                  # 保留备份数量
  backup_dir: "/var/backups/openclaw"

proxy:
  enabled: false
  http_proxy: ""
  https_proxy: ""
```

## 更新流程

```
1. 初始化
   ├── 检测平台 (OS/Arch)
   ├── 解析命令行参数
   ├── 加载配置文件
   └── 确定安装/备份路径

2. 检查版本
   ├── 获取远程版本清单
   ├── 读取本地已安装版本
   └── 比较确定可更新组件

3. 下载更新包
   ├── 创建临时目录
   ├── 并发下载组件包
   └── 验证 SHA256 校验和

4. 备份当前版本
   ├── 停止运行中的服务
   ├── 备份二进制文件
   ├── 备份配置文件（标记保留）
   └── 记录备份元数据

5. 安装更新
   ├── 解压更新包到临时位置
   ├── 替换二进制文件
   ├── 合并配置（保留用户设置）
   └── 设置文件权限

6. 验证安装
   ├── 检查文件完整性
   ├── 测试启动
   └── 验证配置加载

7. 完成或回滚
   ├── 成功: 清理临时文件，记录新版本
   └── 失败: 自动回滚到备份版本
```

## 安全考虑

1. **校验和验证**: 所有下载的文件必须验证 SHA256 校验和
2. **签名验证**: 支持 GPG 签名验证（可选）
3. **HTTPS 强制**: 远程更新必须使用 HTTPS
4. **权限控制**: 更新程序需要管理员/root 权限
5. **备份保护**: 备份文件设置适当权限 (0750)，防止篡改
6. **路径安全**: 验证所有路径在目标目录内，防止路径遍历

## 回滚机制

更新程序维护一个备份链，支持多级回滚：

```
/var/backups/openclaw/  (Linux)
/usr/local/var/backups/openclaw/  (macOS)
%ProgramData%\OpenClaw\backups\  (Windows)

├── backup-20240115-103022/     # 最新备份
│   ├── manifest.json           # 备份清单
│   ├── core/                   # 核心备份
│   ├── adapters/               # 适配器备份
│   └── config/                 # 配置备份
├── backup-20240108-091500/     # 上一版本备份
└── backup-20240101-143000/     # 更早备份
```

回滚时：
1. 读取备份清单
2. 停止当前服务
3. 恢复备份文件
4. 验证恢复成功
5. 更新版本记录

## 平台特定说明

### Linux
- 默认安装路径: `/usr/local/bin`
- 默认配置路径: `/etc/openclaw`
- 默认备份路径: `/var/backups/openclaw`
- 需要 root 权限进行系统目录更新

### macOS
- 默认安装路径: `/usr/local/bin`
- 默认配置路径: `/usr/local/etc/openclaw`
- 默认备份路径: `/usr/local/var/backups/openclaw`
- 可能需要绕过 Gatekeeper

### Windows
- 默认安装路径: `C:\Program Files\OpenClaw`
- 默认配置路径: `C:\ProgramData\OpenClaw\config`
- 默认备份路径: `C:\ProgramData\OpenClaw\backups`
- 需要管理员权限

## 集成到安装器

更新程序可以作为安装器的一个子命令：

```bash
# 安装器内置更新命令
openclaw-installer update
```

或者作为独立工具分发，由安装器安装到系统 PATH。

## 构建

```bash
# 构建所有平台
make build-all

# 构建特定平台
GOOS=linux GOARCH=amd64 go build -o openclaw-updater-linux-amd64
GOOS=darwin GOARCH=arm64 go build -o openclaw-updater-darwin-arm64
GOOS=windows GOARCH=amd64 go build -o openclaw-updater-windows-amd64.exe
```

## 测试

```bash
# 运行单元测试
go test ./...

# 运行集成测试（模拟模式）
go test -tags=integration ./...

# 检查更新（不安装）
./openclaw-updater -check -v

# 模拟更新
./openclaw-updater -dry-run -yes -v
```
