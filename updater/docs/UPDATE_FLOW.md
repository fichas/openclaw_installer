# OpenClaw 更新流程文档

本文档详细描述 OpenClaw 更新程序的完整工作流程。

## 目录

1. [概述](#概述)
2. [更新流程图](#更新流程图)
3. [详细步骤说明](#详细步骤说明)
4. [错误处理与回滚](#错误处理与回滚)
5. [配置管理](#配置管理)
6. [安全机制](#安全机制)

## 概述

OpenClaw 更新程序采用事务性更新机制，确保系统在更新失败时能够自动回滚到可用状态。更新过程分为多个阶段，每个阶段都有明确的检查点和回滚能力。

## 更新流程图

```
┌─────────────────────────────────────────────────────────────────────────┐
│                              更新流程                                    │
└─────────────────────────────────────────────────────────────────────────┘

    ┌─────────┐
    │  Start  │
    └────┬────┘
         │
         ▼
┌─────────────────┐
│  1. 初始化阶段   │
│                 │
│ • 检测平台      │
│ • 解析参数      │
│ • 加载配置      │
│ • 初始化日志    │
└────────┬────────┘
         │
         ▼
┌─────────────────┐     ┌─────────────────┐
│  2. 版本检查     │────▶│  无可用更新      │
│                 │     │  退出程序        │
│ • 获取远程版本   │     └─────────────────┘
│ • 读取本地版本   │
│ • 比较版本差异   │
└────────┬────────┘
         │ 有更新
         ▼
┌─────────────────┐
│  3. 预更新检查   │
│                 │
│ • 检查磁盘空间   │
│ • 检查网络连接   │
│ • 检查权限      │
│ • 获取锁文件    │
└────────┬────────┘
         │
         ▼
┌─────────────────┐
│  4. 创建备份     │
│                 │
│ • 停止服务      │
│ • 备份二进制    │
│ • 备份配置      │
│ • 记录元数据    │
└────────┬────────┘
         │
         ▼
┌─────────────────┐     ┌─────────────────┐
│  5. 下载更新包   │────▶│  下载失败       │
│                 │     │  自动回滚       │
│ • 创建临时目录   │     └─────────────────┘
│ • 下载组件包    │
│ • 验证校验和    │
└────────┬────────┘
         │
         ▼
┌─────────────────┐     ┌─────────────────┐
│  6. 安装更新     │────▶│  安装失败       │
│                 │     │  自动回滚       │
│ • 解压更新包    │     └─────────────────┘
│ • 替换二进制    │
│ • 更新配置      │
│ • 设置权限      │
└────────┬────────┘
         │
         ▼
┌─────────────────┐     ┌─────────────────┐
│  7. 验证安装     │────▶│  验证失败       │
│                 │     │  自动回滚       │
│ • 检查文件完整性 │     └─────────────────┘
│ • 测试启动      │
│ • 验证配置      │
└────────┬────────┘
         │
         ▼
┌─────────────────┐
│  8. 完成更新     │
│                 │
│ • 启动服务      │
│ • 清理临时文件   │
│ • 记录新版本    │
│ • 发送通知      │
└────────┬────────┘
         │
         ▼
    ┌─────────┐
    │   End   │
    └─────────┘
```

## 详细步骤说明

### 阶段 1: 初始化

**目标**: 准备更新环境，加载必要配置

**步骤**:
1. **平台检测**
   - 检测操作系统类型 (Windows/macOS/Linux)
   - 检测系统架构 (AMD64/ARM64)
   - 确定默认路径

2. **参数解析**
   - 解析命令行参数
   - 验证参数有效性
   - 设置运行模式 (check/dry-run/normal)

3. **配置加载**
   ```go
   config, err := LoadUpdateConfig(configPath)
   if err != nil {
       // 使用默认配置
       config = DefaultUpdateConfig()
   }
   ```

4. **日志初始化**
   - 创建日志目录
   - 初始化日志文件
   - 设置日志级别

**输出**:
- 初始化的 Updater 实例
- 加载的配置
- 日志记录器

### 阶段 2: 版本检查

**目标**: 确定是否有可用更新

**步骤**:
1. **获取本地版本**
   ```go
   currentVersion, err := getCurrentVersion()
   // 读取安装目录下的 version.txt
   ```

2. **获取远程版本**
   ```go
   versionInfo, err := fetchVersionInfo()
   // HTTP GET 请求到版本服务器
   ```

3. **适配器版本检查**
   - 遍历配置的适配器
   - 检查每个适配器的远程版本
   - 比较本地与远程版本

4. **版本比较**
   ```go
   // 语义化版本比较
   if latestVersion > currentVersion {
       hasUpdate = true
   }
   ```

**输出**:
- UpdateInfo 结构体，包含:
  - 当前版本
  - 最新版本
  - 可用的适配器更新列表
  - 下载 URL 和校验和

### 阶段 3: 预更新检查

**目标**: 确保更新可以安全进行

**检查项**:

| 检查项 | 要求 | 失败处理 |
|--------|------|----------|
| 磁盘空间 | 至少 500MB 可用 | 报错退出 |
| 网络连接 | 能访问版本服务器 | 报错退出 |
| 文件权限 | 能写入安装目录 | 报错退出 |
| 并发锁 | 无其他更新进程 | 等待或退出 |
| 服务状态 | 记录当前状态 | 用于恢复 |

**锁机制**:
```go
// 创建锁文件防止并发更新
lockFile := "/var/run/openclaw-update.lock"
if err := acquireLock(lockFile); err != nil {
    return fmt.Errorf("another update is in progress")
}
defer releaseLock(lockFile)
```

### 阶段 4: 创建备份

**目标**: 创建可恢复的备份点

**步骤**:
1. **停止服务** (如果运行中)
   ```bash
   systemctl stop openclaw  # Linux
   # 或发送信号优雅关闭
   ```

2. **创建备份目录**
   ```
   /var/lib/openclaw/backups/
   └── backup-20240115-103022/
       ├── manifest.json      # 备份清单
       ├── core/              # 核心文件备份
       ├── adapters/          # 适配器备份
       └── config/            # 配置备份
   ```

3. **备份文件**
   ```go
   // 创建 tar.gz 归档
   backupPath := createBackup(installDir, backupDir)
   ```

4. **记录元数据**
   ```json
   {
     "version": "1.0.0",
     "timestamp": "2024-01-15T10:30:22Z",
     "components": ["core", "wecom-adapter"],
     "files": [...]
   }
   ```

### 阶段 5: 下载更新包

**目标**: 获取更新文件

**步骤**:
1. **创建临时目录**
   ```go
   tempDir, err := os.MkdirTemp("", "openclaw-update-*")
   defer os.RemoveAll(tempDir)
   ```

2. **并发下载**
   ```go
   var wg sync.WaitGroup
   for _, component := range components {
       wg.Add(1)
       go func(c Component) {
           defer wg.Done()
           downloadComponent(c, tempDir)
       }(component)
   }
   wg.Wait()
   ```

3. **校验和验证**
   ```go
   actualChecksum := sha256Sum(downloadedFile)
   if actualChecksum != expectedChecksum {
       return fmt.Errorf("checksum mismatch")
   }
   ```

### 阶段 6: 安装更新

**目标**: 将更新应用到系统

**步骤**:
1. **保留用户配置**
   ```go
   // 备份当前配置
   userConfig := backupUserConfig(configDir)
   ```

2. **解压更新包**
   ```go
   // 根据平台选择解压方式
   if platform.IsWindows() {
       extractZip(packagePath, tempExtractDir)
   } else {
       extractTarGz(packagePath, tempExtractDir)
   }
   ```

3. **原子替换**
   ```go
   // 使用临时文件 + 重命名实现原子操作
   tempInstall := installDir + ".new"
   if err := os.Rename(tempExtractDir, tempInstall); err != nil {
       return err
   }
   if err := os.Rename(tempInstall, installDir); err != nil {
       return err
   }
   ```

4. **恢复用户配置**
   ```go
   restoreUserConfig(userConfig, configDir)
   ```

5. **设置权限**
   ```go
   os.Chmod(binaryPath, 0755)
   ```

### 阶段 7: 验证安装

**目标**: 确保更新成功

**验证项**:

| 验证项 | 方法 | 失败处理 |
|--------|------|----------|
| 文件存在 | os.Stat() | 回滚 |
| 文件权限 | 检查可执行 | 修复或回滚 |
| 版本信息 | 执行 --version | 回滚 |
| 配置加载 | 解析配置文件 | 恢复默认 |
| 服务启动 | 尝试启动 | 回滚 |

### 阶段 8: 完成更新

**目标**: 清理并记录更新

**步骤**:
1. **启动服务**
   ```bash
   systemctl start openclaw
   ```

2. **清理临时文件**
   ```go
   os.RemoveAll(tempDir)
   ```

3. **记录新版本**
   ```go
   saveVersion(latestVersion)
   ```

4. **清理旧备份**
   ```go
   cleanupOldBackups(maxBackups)
   ```

5. **发送通知** (如果配置)
   ```go
   if config.Notifications.OnSuccess {
       sendNotification("Update successful", ...)
   }
   ```

## 错误处理与回滚

### 回滚触发条件

- 下载失败
- 校验和验证失败
- 安装失败
- 验证失败
- 用户取消

### 回滚流程

```
┌─────────────────┐
│   检测到错误    │
└────────┬────────┘
         │
         ▼
┌─────────────────┐
│  停止当前操作   │
│  (如果可能)    │
└────────┬────────┘
         │
         ▼
┌─────────────────┐
│  检查备份存在   │
└────────┬────────┘
         │
    ┌────┴────┐
    │         │
    ▼         ▼
┌───────┐ ┌───────────┐
│ 存在   │ │ 不存在    │
└───┬───┘ └─────┬─────┘
    │           │
    ▼           ▼
┌─────────────────┐ ┌─────────────────┐
│  从备份恢复     │ │  报错退出       │
│                 │ │  需要手动修复   │
│ • 停止服务      │ └─────────────────┘
│ • 恢复文件      │
│ • 恢复配置      │
│ • 启动服务      │
└────────┬────────┘
         │
         ▼
┌─────────────────┐
│  验证回滚成功   │
└────────┬────────┘
    ┌────┴────┐
    │         │
    ▼         ▼
┌───────┐ ┌───────────┐
│ 成功   │ │ 失败      │
└───┬───┘ └─────┬─────┘
    │           │
    ▼           ▼
┌─────────────────┐ ┌─────────────────┐
│  记录回滚事件   │ │  报错退出       │
│  发送通知       │ │  系统状态不一致 │
└─────────────────┘ └─────────────────┘
```

### 回滚实现

```go
func (u *Updater) Rollback() error {
    // 1. 找到最新备份
    backup, err := u.findLatestBackup()
    if err != nil {
        return fmt.Errorf("no backup available: %w", err)
    }

    // 2. 停止服务
    if err := u.stopService(); err != nil {
        u.logger.Warn("Failed to stop service: %v", err)
    }

    // 3. 恢复备份
    if err := u.extractBackup(backup.Path, u.installDir); err != nil {
        return fmt.Errorf("failed to extract backup: %w", err)
    }

    // 4. 启动服务
    if err := u.startService(); err != nil {
        return fmt.Errorf("failed to start service: %w", err)
    }

    // 5. 记录回滚
    u.logger.Info("Rollback completed successfully")

    return nil
}
```

## 配置管理

### 配置保护策略

1. **白名单机制**
   - 只保留特定配置文件
   - 其他文件被新配置覆盖

2. **合并策略**
   ```go
   // 保留用户修改的配置项
   newConfig.Merge(userConfig)
   ```

3. **备份恢复**
   - 更新前备份用户配置
   - 更新后恢复用户配置

### 配置文件优先级

```
高优先级 ──────────────────────────> 低优先级

命令行参数 > 环境变量 > 配置文件 > 默认值
```

## 安全机制

### 1. 校验和验证

所有下载的文件必须经过 SHA256 校验：

```go
func verifyChecksum(filePath, expectedChecksum string) error {
    file, err := os.Open(filePath)
    if err != nil {
        return err
    }
    defer file.Close()

    hasher := sha256.New()
    if _, err := io.Copy(hasher, file); err != nil {
        return err
    }

    actualChecksum := hex.EncodeToString(hasher.Sum(nil))
    if actualChecksum != expectedChecksum {
        return fmt.Errorf("checksum mismatch")
    }

    return nil
}
```

### 2. 路径遍历防护

```go
func safeExtractPath(basePath, filePath string) (string, error) {
    targetPath := filepath.Join(basePath, filePath)
    absTarget, _ := filepath.Abs(targetPath)
    absBase, _ := filepath.Abs(basePath)

    if !strings.HasPrefix(absTarget, absBase+string(os.PathSeparator)) {
        return "", fmt.Errorf("path traversal detected: %s", filePath)
    }

    return targetPath, nil
}
```

### 3. 权限控制

```go
// 设置安全的文件权限
const (
    BinaryPermission = 0755  // rwxr-xr-x
    ConfigPermission = 0640  // rw-r-----
    DirPermission    = 0750  // rwxr-x---
)
```

### 4. 锁机制

防止并发更新：

```go
func acquireLock(lockFile string) error {
    // 尝试创建锁文件
    file, err := os.OpenFile(lockFile, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0644)
    if err != nil {
        return fmt.Errorf("failed to acquire lock: %w", err)
    }

    // 写入 PID
    fmt.Fprintf(file, "%d", os.Getpid())
    file.Close()

    return nil
}
```

## 日志记录

### 日志级别

- **INFO**: 正常流程信息
- **SUCCESS**: 成功完成的操作
- **WARN**: 警告信息，不影响流程
- **ERROR**: 错误信息，可能导致回滚
- **FATAL**: 致命错误，程序退出
- **DEBUG**: 调试信息 (verbose 模式)

### 日志格式

```
[2024-01-15 10:30:22] [INFO] Starting update check
[2024-01-15 10:30:23] [INFO] Current version: 1.0.0
[2024-01-15 10:30:23] [INFO] Latest version: 1.1.0
[2024-01-15 10:30:23] [INFO] Creating backup...
[2024-01-15 10:30:25] [SUCCESS] Backup created: /var/lib/openclaw/backups/backup-20240115-103023.tar.gz
[2024-01-15 10:30:25] [INFO] Downloading update packages...
[2024-01-15 10:30:30] [SUCCESS] Download completed
[2024-01-15 10:30:30] [INFO] Installing update...
[2024-01-15 10:30:35] [SUCCESS] Update installed successfully
[2024-01-15 10:30:35] [INFO] Verifying installation...
[2024-01-15 10:30:36] [SUCCESS] Verification passed
[2024-01-15 10:30:36] [SUCCESS] Update completed successfully
```

## 附录

### A. 错误代码

| 代码 | 含义 | 处理建议 |
|------|------|----------|
| 0 | 成功 | - |
| 1 | 通用错误 | 查看日志 |
| 2 | 网络错误 | 检查网络连接 |
| 3 | 权限错误 | 使用管理员权限运行 |
| 4 | 磁盘空间不足 | 清理磁盘空间 |
| 5 | 校验和错误 | 重新下载 |
| 6 | 回滚失败 | 手动恢复备份 |
| 7 | 验证失败 | 检查系统状态 |

### B. 环境变量

| 变量名 | 说明 | 默认值 |
|--------|------|--------|
| `OPENCLAW_INSTALL_DIR` | 安装目录 | 平台特定 |
| `OPENCLAW_CONFIG_DIR` | 配置目录 | 平台特定 |
| `OPENCLAW_BACKUP_DIR` | 备份目录 | 平台特定 |
| `OPENCLAW_VERSION_URL` | 版本检查 URL | api.openclaw.io |
| `HTTP_PROXY` | HTTP 代理 | - |
| `HTTPS_PROXY` | HTTPS 代理 | - |
