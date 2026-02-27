# OpenClaw 安装器系统架构设计文档

## 1. 项目概述

OpenClaw 安装器是一个跨平台的安装工具，支持从 U 盘运行，提供离线安装能力。安装器内置 Web 配置界面，用于配置 OpenClaw 及相关适配器。

### 1.1 核心特性
- **跨平台支持**: macOS / Windows / Linux (x64 / ARM64)
- **离线安装**: 从 U 盘直接运行，无需网络连接
- **单文件分发**: 每个平台一个独立的可执行文件
- **Web 配置界面**: 内置 HTTP 服务 (端口 18080)
- **预打包适配器**: 企业微信、钉钉、飞书适配器

### 1.2 技术栈
- **语言**: Go 1.22+
- **静态资源**: Go embed 内嵌
- **Web 框架**: 标准库 net/http + 内嵌静态文件
- **构建**: 交叉编译生成各平台二进制

---

## 2. U 盘目录结构

```
/OpenClaw/
├── installers/                    # 各平台安装器二进制
│   ├── openclaw-installer-darwin-amd64
│   ├── openclaw-installer-darwin-arm64
│   ├── openclaw-installer-windows-amd64.exe
│   ├── openclaw-installer-windows-arm64.exe
│   ├── openclaw-installer-linux-amd64
│   └── openclaw-installer-linux-arm64
│
├── packages/                      # 预打包的安装包
│   ├── openclaw-core/            # OpenClaw 核心
│   │   ├── openclaw-darwin-amd64.tar.gz
│   │   ├── openclaw-darwin-arm64.tar.gz
│   │   ├── openclaw-windows-amd64.zip
│   │   ├── openclaw-windows-arm64.zip
│   │   ├── openclaw-linux-amd64.tar.gz
│   │   └── openclaw-linux-arm64.tar.gz
│   │
│   ├── adapters/                 # 适配器包
│   │   ├── wecom-adapter/      # 企业微信适配器
│   │   │   ├── wecom-adapter-darwin-amd64.tar.gz
│   │   │   ├── wecom-adapter-darwin-arm64.tar.gz
│   │   │   ├── wecom-adapter-windows-amd64.zip
│   │   │   ├── wecom-adapter-windows-arm64.zip
│   │   │   ├── wecom-adapter-linux-amd64.tar.gz
│   │   │   └── wecom-adapter-linux-arm64.tar.gz
│   │   │
│   │   ├── dingtalk-adapter/     # 钉钉适配器
│   │   │   └── ...
│   │   │
│   │   └── feishu-adapter/       # 飞书适配器
│   │       └── ...
│   │
│   └── config-templates/         # 配置文件模板
│       ├── openclaw.yaml.template
│       ├── wecom-adapter.yaml.template
│       ├── dingtalk-adapter.yaml.template
│       └── feishu-adapter.yaml.template
│
├── resources/                    # 附加资源
│   ├── icons/
│   ├── licenses/
│   └── docs/
│
└── autorun/                      # 自动运行脚本（可选）
    ├── autorun.inf              # Windows 自动运行
    ├── install-mac.command      # macOS 双击运行
    └── install-linux.sh         # Linux 运行脚本
```

---

## 3. 系统架构

### 3.1 组件架构图

```
┌─────────────────────────────────────────────────────────────────┐
│                        OpenClaw Installer                        │
│                                                                  │
│  ┌─────────────────┐  ┌─────────────────┐  ┌─────────────────┐  │
│  │   Platform      │  │   Package       │  │   Config        │  │
│  │   Detector      │  │   Manager       │  │   Generator     │  │
│  │                 │  │                 │  │                 │  │
│  │ - OS detection  │  │ - Extract       │  │ - Template      │  │
│  │ - Arch detection│  │ - Copy files    │  │   rendering     │  │
│  │ - Path resolver │  │ - Permission    │  │ - Validation    │  │
│  │                 │  │   setup         │  │                 │  │
│  └────────┬────────┘  └────────┬────────┘  └────────┬────────┘  │
│           │                    │                    │           │
│           └────────────────────┼────────────────────┘           │
│                                │                                │
│                                ▼                                │
│  ┌─────────────────────────────────────────────────────────┐   │
│  │                 Web Configuration Server                 │   │
│  │                      (Port 18080)                        │   │
│  │                                                          │   │
│  │  ┌─────────────┐  ┌─────────────┐  ┌─────────────────┐  │   │
│  │  │  HTTP API   │  │  Static     │  │  WebSocket      │  │   │
│  │  │  Handler    │  │  File       │  │  (optional)     │  │   │
│  │  │             │  │  Server     │  │                 │  │   │
│  │  │ - /api/     │  │             │  │ - Real-time     │  │   │
│  │  │   config    │  │ - HTML/CSS  │  │   progress      │  │   │
│  │  │ - /api/     │  │ - JS        │  │                 │  │   │
│  │  │   install   │  │ - Assets    │  │                 │  │   │
│  │  │ - /api/     │  │             │  │                 │  │   │
│  │  │   status    │  │             │  │                 │  │   │
│  │  └─────────────┘  └─────────────┘  └─────────────────┘  │   │
│  └─────────────────────────────────────────────────────────┘   │
│                                                                │
└────────────────────────────────────────────────────────────────┘
```

### 3.2 模块职责

| 模块 | 职责 | 关键接口 |
|------|------|----------|
| Platform Detector | 检测操作系统类型、架构、安装路径 | `DetectOS()`, `DetectArch()`, `GetInstallPaths()` |
| Package Manager | 解压、复制、设置权限 | `ExtractPackage()`, `InstallFiles()`, `SetPermissions()` |
| Config Generator | 渲染配置模板，验证配置 | `GenerateConfig()`, `ValidateConfig()` |
| Web Server | 提供 HTTP API 和静态文件服务 | `Start()`, `Stop()`, `HandleRoutes()` |

---

## 4. 安装流程

### 4.1 主流程图

```
┌─────────┐
│  Start  │
└────┬────┘
     │
     ▼
┌─────────────────┐     ┌─────────────────┐
│  Detect Platform│────▶│  Error:         │
│  - OS Type      │Fail │  Unsupported    │
│  - Architecture │     │  Platform       │
│  - U盘路径       │     └─────────────────┘
└────────┬────────┘
         │ Success
         ▼
┌─────────────────┐     ┌─────────────────┐
│  Check U盘      │────▶│  Error:         │
│  Directory      │Fail │  Invalid U盘    │
│  Structure      │     │  Structure      │
└────────┬────────┘     └─────────────────┘
         │ Success
         ▼
┌─────────────────┐
│  Start Web      │
│  Server on      │
│  Port 18080     │
└────────┬────────┘
         │
         ▼
┌─────────────────┐
│  Open Browser   │
│  (系统默认浏览器) │
└────────┬────────┘
         │
         ▼
┌─────────────────┐
│  Wait for User  │
│  Configuration  │
└────────┬────────┘
         │
         ▼
┌─────────────────┐     ┌─────────────────┐
│  User Submit    │────▶│  Validate       │
│  Configuration  │     │  Configuration  │
└────────┬────────┘     └────────┬────────┘
         │◄───────────────────────┘ Fail
         │ Success
         ▼
┌─────────────────┐
│  Install Files  │
│  - Copy Core    │
│  - Copy Adapters│
│  - Set Config   │
└────────┬────────┘
         │
         ▼
┌─────────────────┐     ┌─────────────────┐
│  Installation   │────▶│  Show Error     │
│  Success?       │Fail │  in Web UI      │
└────────┬────────┘     └─────────────────┘
         │ Success
         ▼
┌─────────────────┐
│  Show Success   │
│  Page           │
└────────┬────────┘
         │
         ▼
┌─────────────────┐
│  Stop Web       │
│  Server         │
└────────┬────────┘
         │
         ▼
┌─────────┐
│   End   │
└─────────┘
```

### 4.2 平台检测逻辑

```go
// 平台检测流程
type PlatformInfo struct {
    OS           string // "darwin", "windows", "linux"
    Arch         string // "amd64", "arm64"
    U盘Path      string // U盘挂载路径
    InstallPaths InstallPaths
}

type InstallPaths struct {
    BinaryDir    string // 二进制安装目录
    ConfigDir    string // 配置文件目录
    DataDir      string // 数据目录
    LogDir       string // 日志目录
}

// 各平台默认安装路径
// macOS:
//   - BinaryDir: /usr/local/bin
//   - ConfigDir: /usr/local/etc/openclaw
//   - DataDir:   /usr/local/share/openclaw
//   - LogDir:    /var/log/openclaw
//
// Windows:
//   - BinaryDir: C:\Program Files\OpenClaw
//   - ConfigDir: C:\ProgramData\OpenClaw\config
//   - DataDir:   C:\ProgramData\OpenClaw\data
//   - LogDir:    C:\ProgramData\OpenClaw\logs
//
// Linux:
//   - BinaryDir: /usr/local/bin
//   - ConfigDir: /etc/openclaw
//   - DataDir:   /var/lib/openclaw
//   - LogDir:    /var/log/openclaw
```

### 4.3 安装步骤详解

1. **启动阶段**
   - 解析命令行参数（可选：--port, --no-browser, --debug）
   - 初始化日志系统
   - 检测运行环境（是否从 U 盘运行）

2. **平台检测**
   - 使用 `runtime.GOOS` 和 `runtime.GOARCH` 检测平台
   - 遍历常见挂载点查找 U 盘目录结构
   - 验证 U 盘包含必要的 packages 目录

3. **Web 服务启动**
   - 绑定到 0.0.0.0:18080
   - 注册 API 路由
   - 提供内嵌的静态文件

4. **浏览器打开**
   - 使用系统命令打开默认浏览器
   - macOS: `open http://localhost:18080`
   - Windows: `start http://localhost:18080`
   - Linux: `xdg-open http://localhost:18080`

5. **用户配置**
   - 展示配置表单
   - 实时验证输入
   - 保存配置到临时存储

6. **执行安装**
   - 根据配置选择要安装的包
   - 解压/复制文件到目标目录
   - 设置正确的文件权限
   - 生成配置文件

7. **完成**
   - 显示安装结果
   - 可选：启动 OpenClaw 服务
   - 关闭 Web 服务器

---

## 5. Web 配置服务设计

### 5.1 API 设计

```yaml
# RESTful API 规范

# 获取系统信息
GET /api/system/info
Response:
  {
    "platform": "darwin",
    "arch": "arm64",
    "version": "1.0.0",
    "u盘_detected": true,
    "u盘_path": "/Volumes/OpenClaw",
    "install_paths": {
      "binary_dir": "/usr/local/bin",
      "config_dir": "/usr/local/etc/openclaw",
      "data_dir": "/usr/local/share/openclaw",
      "log_dir": "/var/log/openclaw"
    }
  }

# 获取可用包列表
GET /api/packages
Response:
  {
    "core": {
      "name": "OpenClaw Core",
      "version": "1.0.0",
      "available": true,
      "path": "packages/openclaw-core/openclaw-darwin-arm64.tar.gz"
    },
    "adapters": [
      {
        "id": "wecom",
        "name": "企业微信适配器",
        "version": "1.0.0",
        "available": true,
        "selected": true
      },
      {
        "id": "dingtalk",
        "name": "钉钉适配器",
        "version": "1.0.0",
        "available": true,
        "selected": false
      },
      {
        "id": "feishu",
        "name": "飞书适配器",
        "version": "1.0.0",
        "available": true,
        "selected": false
      }
    ]
  }

# 获取配置模板
GET /api/config/template?adapter=wecom
Response:
  {
    "template": {
      "webhook_url": "",
      "corp_id": "",
      "corp_secret": ""
    },
    "fields": [
      {
        "name": "webhook_url",
        "type": "string",
        "required": true,
        "label": "Webhook URL",
        "description": "企业微信机器人的 Webhook 地址"
      }
    ]
  }

# 提交配置
POST /api/config
Body:
  {
    "core": {
      "port": 8080,
      "log_level": "info"
    },
    "adapters": {
      "wecom": {
        "enabled": true,
        "webhook_url": "https://qyapi.weixin.qq.com/..."
      }
    }
  }
Response:
  {
    "valid": true,
    "errors": []
  }

# 开始安装
POST /api/install
Body:
  {
    "packages": ["core", "wecom-adapter"],
    "config": { ... }
  }
Response (Stream or Poll):
  {
    "status": "installing",  # "pending", "installing", "completed", "failed"
    "progress": 45,          # 0-100
    "current_step": "Extracting core package...",
    "error": null
  }

# 获取安装状态
GET /api/install/status
Response: (同上)

# 取消安装
POST /api/install/cancel
Response:
  {
    "success": true
  }
```

### 5.2 Web UI 页面结构

```
Web UI 单页应用 (SPA) 结构:

/
├── / (首页 - 欢迎页面)
│   └── 显示平台信息、U盘状态、开始安装按钮
│
├── /select (选择包页面)
│   └── 显示可安装的包列表（核心 + 适配器）
│   └── 复选框选择要安装的组件
│
├── /config (配置页面)
│   └── 动态表单，根据选择的适配器显示相应配置项
│   └── 配置文件预览
│
├── /install (安装进度页面)
│   └── 进度条显示
│   └── 实时日志输出
│   └── 取消按钮
│
└── /complete (完成页面)
    └── 安装结果摘要
    └── 启动服务按钮
    └── 查看日志按钮
```

### 5.3 静态资源内嵌方案

```go
// 使用 Go embed 内嵌 Web 资源
// 项目结构:
// internal/
//   web/
//     static/          # 静态文件目录
//       index.html
//       css/
//       js/
//       assets/
//     embed.go         # embed 定义
//     server.go        # HTTP 服务器

// embed.go
package web

import "embed"

//go:embed static/*
var StaticFS embed.FS

// server.go
func (s *Server) setupRoutes() {
    // API 路由
    s.router.HandleFunc("/api/system/info", s.handleSystemInfo)
    s.router.HandleFunc("/api/packages", s.handlePackages)
    s.router.HandleFunc("/api/config", s.handleConfig)
    s.router.HandleFunc("/api/install", s.handleInstall)
    s.router.HandleFunc("/api/install/status", s.handleInstallStatus)

    // 静态文件服务
    staticFS, _ := fs.Sub(StaticFS, "static")
    s.router.Handle("/", http.FileServer(http.FS(staticFS)))
}
```

---

## 6. 数据流设计

### 6.1 配置数据流

```
┌─────────────┐     ┌─────────────┐     ┌─────────────┐
│   U盘       │────▶│   Web       │────▶│   Memory    │
│   Templates │     │   UI Form   │     │   Store     │
└─────────────┘     └─────────────┘     └──────┬──────┘
                                               │
                                               ▼
┌─────────────┐     ┌─────────────┐     ┌─────────────┐
│   Target    │◀────│   Config    │◀────│   Validate  │
│   Config    │     │   Generator │     │   & Merge   │
│   Files     │     │             │     │             │
└─────────────┘     └─────────────┘     └─────────────┘
```

### 6.2 安装数据流

```
┌─────────────┐     ┌─────────────┐     ┌─────────────┐
│   U盘       │────▶│   Package   │────▶│   Temp      │
│   Packages  │     │   Extractor │     │   Directory │
└─────────────┘     └─────────────┘     └──────┬──────┘
                                               │
                                               ▼
┌─────────────┐     ┌─────────────┐     ┌─────────────┐
│   Target    │◀────│   File      │◀────│   Install   │
│   System    │     │   Installer │     │   Planner   │
│   Locations │     │             │     │             │
└─────────────┘     └─────────────┘     └─────────────┘
        │
        ▼
┌─────────────┐
│   Config    │
│   Files     │
└─────────────┘
```

---

## 7. 错误处理策略

### 7.1 错误分类

| 类别 | 示例 | 处理方式 |
|------|------|----------|
| 平台错误 | 不支持的操作系统 | 显示错误页面，提供手动安装指南 |
| U盘错误 | 找不到 packages 目录 | 提示用户检查 U 盘结构 |
| 权限错误 | 无法写入系统目录 | 提示使用管理员权限运行 |
| 安装错误 | 文件复制失败 | 回滚已安装文件，显示详细错误 |
| 配置错误 | 配置验证失败 | 在 Web UI 中高亮显示错误字段 |

### 7.2 回滚机制

```go
// 安装事务管理
type InstallTransaction struct {
    Steps []InstallStep
}

type InstallStep struct {
    Action   string // "copy", "mkdir", "chmod"
    Source   string
    Target   string
    Backup   string // 备份路径（用于回滚）
}

// 如果安装失败，按相反顺序执行回滚
func (t *InstallTransaction) Rollback() error {
    for i := len(t.Steps) - 1; i >= 0; i-- {
        step := t.Steps[i]
        // 根据 Action 类型执行回滚
        // - copy: 删除 Target，恢复 Backup
        // - mkdir: 删除目录（如果为空）
        // - chmod: 恢复原始权限
    }
}
```

---

## 8. 安全考虑

1. **文件权限**
   - 二进制文件: 0755 (rwxr-xr-x)
   - 配置文件: 0640 (rw-r-----)
   - 数据目录: 0750 (rwxr-x---)

2. **路径安全**
   - 验证所有路径在目标目录内（防止路径遍历攻击）
   - 使用绝对路径，避免相对路径解析问题

3. **输入验证**
   - 所有用户输入在服务端验证
   - 配置文件使用 YAML/JSON Schema 验证

4. **权限提升**
   - Windows: 请求管理员权限（manifest）
   - macOS/Linux: 检测 root/sudo，必要时提示用户

---

## 9. 构建配置

### 9.1 Makefile 示例

```makefile
# 构建所有平台
.PHONY: build-all
build-all: build-darwin build-windows build-linux

# macOS
.PHONY: build-darwin
build-darwin:
	GOOS=darwin GOARCH=amd64 go build -o dist/openclaw-installer-darwin-amd64 ./cmd/installer
	GOOS=darwin GOARCH=arm64 go build -o dist/openclaw-installer-darwin-arm64 ./cmd/installer

# Windows
.PHONY: build-windows
build-windows:
	GOOS=windows GOARCH=amd64 go build -o dist/openclaw-installer-windows-amd64.exe ./cmd/installer
	GOOS=windows GOARCH=arm64 go build -o dist/openclaw-installer-windows-arm64.exe ./cmd/installer

# Linux
.PHONY: build-linux
build-linux:
	GOOS=linux GOARCH=amd64 go build -o dist/openclaw-installer-linux-amd64 ./cmd/installer
	GOOS=linux GOARCH=arm64 go build -o dist/openclaw-installer-linux-arm64 ./cmd/installer
```

### 9.2 版本信息注入

```go
// 通过 ldflags 注入版本信息
// go build -ldflags "-X main.Version=1.0.0 -X main.BuildTime=..."

var (
    Version   = "dev"
    BuildTime = "unknown"
    GitCommit = "unknown"
)
```

---

## 10. 项目代码结构

```
openclaw-installer/
├── cmd/
│   └── installer/
│       └── main.go              # 入口点
│
├── internal/
│   ├── platform/
│   │   ├── detector.go          # 平台检测
│   │   ├── paths.go             # 路径解析
│   │   └── permissions.go       # 权限管理
│   │
│   ├── package/
│   │   ├── extractor.go         # 包解压
│   │   ├── installer.go         # 文件安装
│   │   └── rollback.go          # 回滚机制
│   │
│   ├── config/
│   │   ├── loader.go            # 配置加载
│   │   ├── generator.go         # 配置生成
│   │   └── validator.go         # 配置验证
│   │
│   ├── web/
│   │   ├── server.go            # HTTP 服务器
│   │   ├── handlers.go          # API 处理器
│   │   ├── static/              # 静态文件
│   │   │   ├── index.html
│   │   │   ├── css/
│   │   │   └── js/
│   │   └── embed.go             # embed 定义
│   │
│   └── usb/
│       └── finder.go            # U盘检测
│
├── pkg/
│   └── ...                      # 可复用包
│
├── web/
│   └── src/                     # Web UI 源码（可选，用于开发）
│
├── docs/
│   └── architecture.md          # 本文档
│
├── Makefile
├── go.mod
└── README.md
```

---

## 11. 附录

### 11.1 平台特定说明

**macOS:**
- 可能需要绕过 Gatekeeper: `xattr -d com.apple.quarantine <binary>`
- 建议使用 .pkg 格式进行正式分发

**Windows:**
- 需要管理员权限进行系统目录安装
- 建议使用资源文件嵌入图标和版本信息

**Linux:**
- 支持 systemd 服务文件生成
- 考虑 AppImage 作为替代分发方式

### 11.2 参考文档

- [Go embed 文档](https://pkg.go.dev/embed)
- [Go 交叉编译](https://go.dev/doc/install/source#environment)
- [Windows 安装程序最佳实践](https://docs.microsoft.com/en-us/windows/win32/msi/windows-installer-best-practices)
