# OpenClaw 版本管理策略

## 概述

本文档定义 OpenClaw 项目及其组件的版本管理策略，包括版本号格式、发布流程、兼容性规则和更新通道。

## 版本号格式

采用**语义化版本控制 2.0.0**（Semantic Versioning 2.0.0）：

```
MAJOR.MINOR.PATCH[-PRERELEASE][+BUILD]

示例:
1.2.3           # 正式版本
1.2.3-beta.1    # 预发布版本
1.2.3+20240115  # 带构建元数据
```

### 版本号规则

| 部分 | 说明 | 递增条件 |
|------|------|----------|
| MAJOR | 主版本号 | 不兼容的 API 变更 |
| MINOR | 次版本号 | 向后兼容的功能添加 |
| PATCH | 修订号 | 向后兼容的问题修复 |
| PRERELEASE | 预发布标识 | alpha, beta, rc 等 |
| BUILD | 构建元数据 | 构建时间、commit hash 等 |

### 版本比较规则

1. 主版本号不同：数值大的版本新
2. 主版本相同，比较次版本号
3. 主次版本相同，比较修订号
4. 预发布版本 **小于** 正式版本
5. 预发布版本比较：按点分隔的标识符逐个比较

```
1.0.0 > 1.0.0-rc.1 > 1.0.0-beta.2 > 1.0.0-beta.1 > 1.0.0-alpha
```

## 组件版本管理

### 核心组件 (OpenClaw Core)

- **版本范围**: 1.x.x - 主版本
- **发布周期**: 每月功能更新，每周问题修复
- **支持周期**: 当前主版本 + 前两个主版本

### 适配器版本

每个适配器独立版本管理：

| 适配器 | 版本范围 | 说明 |
|--------|----------|------|
| 企业微信 (wecom) | 1.x.x | 跟随企业微信 API 更新 |
| 钉钉 (dingtalk) | 1.x.x | 跟随钉钉 API 更新 |
| 飞书 (feishu) | 1.x.x | 跟随飞书 API 更新 |

### 版本兼容性矩阵

```
核心版本    适配器最低版本    适配器最高版本
1.0.x       1.0.0            1.x.x
1.1.x       1.0.0            1.x.x
1.2.x       1.1.0            2.x.x
2.0.x       2.0.0            2.x.x
```

## 发布流程

### 发布通道

| 通道 | 说明 | 适用场景 |
|------|------|----------|
| stable | 稳定版本，经过完整测试 | 生产环境 |
| rc | 发布候选版，功能冻结 | 预发布验证 |
| beta | 公测版本，功能基本完整 | 早期测试 |
| alpha | 内测版本，可能不稳定 | 开发测试 |
| dev | 开发版本，每日构建 | 开发调试 |

### 发布流程图

```
功能开发 ──> 功能完成 ──> 合并到 develop 分支
                              │
                              ▼
                    ┌─────────────────┐
                    │  每日构建        │──> dev 通道
                    │  (CI/CD)        │
                    └────────┬────────┘
                             │
                             ▼
                    ┌─────────────────┐
                    │  功能冻结        │
                    │  代码审查        │
                    └────────┬────────┘
                             │
                             ▼
                    ┌─────────────────┐
                    │  内部测试        │──> alpha 通道
                    │  (QA Team)      │
                    └────────┬────────┘
                             │
                             ▼
                    ┌─────────────────┐
                    │  公测           │──> beta 通道
                    │  (Beta Users)   │
                    └────────┬────────┘
                             │
                             ▼
                    ┌─────────────────┐
                    │  发布候选       │──> rc 通道
                    │  (RC Testing)   │
                    └────────┬────────┘
                             │
                             ▼
                    ┌─────────────────┐
                    │  正式发布       │──> stable 通道
                    │  (Production)   │
                    └─────────────────┘
```

### 版本发布检查清单

#### 正式版本 (stable)

- [ ] 所有单元测试通过
- [ ] 所有集成测试通过
- [ ] 性能测试达标
- [ ] 安全审计通过
- [ ] 文档已更新
- [ ] 变更日志已编写
- [ ] 安装包已签名
- [ ] 回滚测试通过

#### 预发布版本 (rc/beta/alpha)

- [ ] 核心功能测试通过
- [ ] 无阻塞性 bug
- [ ] 基本文档可用

## 版本清单格式

### 远程版本清单 (manifest.json)

```json
{
  "version": "1.2.3",
  "channel": "stable",
  "release_date": "2024-01-15T10:30:00Z",
  "eol_date": "2025-01-15T00:00:00Z",
  "core": {
    "version": "1.2.3",
    "download_url": "https://releases.openclaw.io/v1.2.3/openclaw-{os}-{arch}.{ext}",
    "checksum": "sha256:abc123...",
    "size": 12345678,
    "release_notes": "https://releases.openclaw.io/v1.2.3/notes.md",
    "min_updater_version": "1.0.0"
  },
  "adapters": {
    "wecom": {
      "version": "1.1.0",
      "download_url": "https://releases.openclaw.io/adapters/wecom/v1.1.0/wecom-adapter-{os}-{arch}.{ext}",
      "checksum": "sha256:def456...",
      "size": 2345678,
      "min_core_version": "1.0.0",
      "release_notes": "https://releases.openclaw.io/adapters/wecom/v1.1.0/notes.md"
    },
    "dingtalk": {
      "version": "1.0.5",
      "download_url": "...",
      "checksum": "sha256:...",
      "min_core_version": "1.0.0"
    },
    "feishu": {
      "version": "1.0.3",
      "download_url": "...",
      "checksum": "sha256:...",
      "min_core_version": "1.1.0"
    }
  },
  "metadata": {
    "signature": "...",
    "published_by": "release-bot@openclaw.io",
    "git_commit": "abc123def456"
  }
}
```

### 本地版本记录 (.version)

```json
{
  "installation_id": "uuid-generated-at-install",
  "core": {
    "version": "1.2.3",
    "install_date": "2024-01-15T10:35:00Z",
    "install_path": "/usr/local/bin/openclaw",
    "channel": "stable"
  },
  "adapters": {
    "wecom": {
      "version": "1.1.0",
      "install_date": "2024-01-15T10:35:00Z",
      "install_path": "/usr/local/bin/wecom-adapter"
    },
    "dingtalk": {
      "version": "1.0.5",
      "install_date": "2024-01-15T10:35:00Z",
      "install_path": "/usr/local/bin/dingtalk-adapter"
    },
    "feishu": {
      "version": "1.0.3",
      "install_date": "2024-01-15T10:35:00Z",
      "install_path": "/usr/local/bin/feishu-adapter"
    }
  }
}
```

## 版本兼容性规则

### 核心向后兼容性

- **PATCH 更新**: 完全向后兼容，无破坏性变更
- **MINOR 更新**: 向后兼容，新增功能，废弃旧功能
- **MAJOR 更新**: 可能有不兼容变更，需要迁移指南

### 适配器兼容性

```
适配器版本必须满足:
min_core_version <= 当前核心版本 <= max_core_version
```

### 更新程序兼容性

- 更新程序版本必须 >= min_updater_version
- 旧版更新程序可能无法处理新格式的清单

## 版本号递增策略

### 日常开发

```
1.2.3-dev.20240115+abc123
1.2.3-dev.20240116+def456
```

### 功能开发完成

```
1.3.0-alpha.1  # 第一个 alpha
1.3.0-alpha.2  # 修复问题
1.3.0-beta.1   # 进入 beta
1.3.0-rc.1     # 发布候选
1.3.0-rc.2     # 修复阻塞性问题
1.3.0          # 正式发布
```

### 问题修复

```
1.2.3  # 当前版本
1.2.4  # 修复 bug
```

### 紧急安全修复

```
1.2.3       # 当前版本
1.2.4       # 安全修复版本
```

## 版本支持周期

```
当前版本 (Current):     1.2.x ────────────────────────────────
维护版本 (Maintenance): 1.1.x ──────────────── EOL
旧版本 (Legacy):        1.0.x ─────── EOL

时间线:
─────────────────────────────────────────────────────────────
    1.0    1.1    1.2    1.3    1.4
    EOL    EOL    Curr   Next   Future
```

### 支持政策

| 版本类型 | 支持周期 | 更新类型 |
|----------|----------|----------|
| Current | 直到下个 MAJOR 发布 | 功能 + 修复 |
| Maintenance | 6 个月 | 仅修复 |
| Legacy | 已停止支持 | 无 |

## 版本检测与报告

### 版本检测 API

```go
// 获取当前版本
func GetCurrentVersion() (*VersionInfo, error)

// 检查更新
func CheckForUpdates(channel string) (*UpdateInfo, error)

// 验证版本兼容性
func ValidateCompatibility(coreVer, adapterVer string) error
```

### 版本报告

更新程序定期向服务器报告版本信息（可选，可禁用）：

```json
{
  "installation_id": "uuid",
  "core_version": "1.2.3",
  "adapter_versions": {
    "wecom": "1.1.0",
    "dingtalk": "1.0.5",
    "feishu": "1.0.3"
  },
  "platform": "linux/amd64",
  "last_check": "2024-01-15T10:30:00Z"
}
```

## 版本迁移指南

### 从 1.x 迁移到 2.x

当发布 MAJOR 版本时，需要提供详细的迁移指南：

```markdown
## 从 1.x 迁移到 2.x

### 破坏性变更
1. 配置文件格式变更
2. API 端点变更
3. 命令行参数变更

### 迁移步骤
1. 备份当前配置
2. 运行迁移脚本: openclaw-migrate --from=1.x --to=2.x
3. 验证配置
4. 启动新版本

### 回滚方案
如果迁移失败，可以回滚到 1.x 版本...
```

## 版本管理工具

### 版本号管理脚本

```bash
# 生成新版本号
scripts/version.sh bump patch   # 1.2.3 -> 1.2.4
scripts/version.sh bump minor   # 1.2.3 -> 1.3.0
scripts/version.sh bump major   # 1.2.3 -> 2.0.0
scripts/version.sh bump rc      # 1.2.3 -> 1.2.4-rc.1

# 生成版本清单
scripts/version.sh manifest --channel=stable

# 验证版本号
scripts/version.sh validate 1.2.3
```

### CI/CD 集成

```yaml
# .github/workflows/release.yml
name: Release
on:
  push:
    tags:
      - 'v*'

jobs:
  release:
    steps:
      - name: Parse version
        run: |
          VERSION=${GITHUB_REF#refs/tags/v}
          echo "VERSION=$VERSION" >> $GITHUB_ENV

      - name: Validate version
        run: scripts/version.sh validate $VERSION

      - name: Build release
        run: make release VERSION=$VERSION

      - name: Generate manifest
        run: scripts/version.sh manifest --version=$VERSION

      - name: Publish release
        run: scripts/publish.sh --version=$VERSION
```
