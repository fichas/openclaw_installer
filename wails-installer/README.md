# OpenClaw Wails 安装器

基于 Wails 框架的跨平台图形化安装器，提供现代化的安装向导体验。

## 特性

- **纯图形界面**: 无命令行窗口，原生桌面应用体验
- **跨平台支持**: Windows、macOS、Linux 三平台
- **单文件分发**: 所有资源嵌入，单个可执行文件
- **现代化 UI**: 深色主题、流畅动画、响应式设计
- **安装向导**: 分步骤引导，配置 IM 适配器
- **实时进度**: 安装过程可视化，状态实时更新

## 项目结构

```
wails-installer/
├── main.go                      # 应用入口
├── wails.json                   # Wails 配置文件
├── go.mod                       # Go 模块依赖
├── build.sh                     # 构建脚本
├── README.md                    # 项目说明
│
├── frontend/
│   └── dist/
│       └── index.html          # 前端界面（单文件）
│
├── internal/
│   ├── platform/               # 平台检测
│   │   └── platform.go
│   ├── installer/              # 安装逻辑
│   │   └── installer.go
│   ├── config/                 # 配置管理
│   │   └── config.go
│   └── updater/                # 更新检查
│
└── build/
    ├── windows/                # Windows 资源
    │   ├── README.md
    │   └── installer.nsi       # NSIS 安装脚本
    ├── darwin/                 # macOS 资源
    │   └── README.md
    └── linux/                  # Linux 资源
        └── README.md
```

## 技术栈

- **后端**: Go 1.21+ with Wails v2
- **前端**: HTML5 + CSS3 + Vanilla JavaScript
- **UI 设计**: 深色主题，毛玻璃效果，渐变动画
- **构建**: Wails CLI + 交叉编译

## 安装向导流程

1. **欢迎页**: 显示平台信息，开始安装
2. **安装方式**: 系统安装/用户安装/便携模式
3. **适配器配置**: 企业微信、钉钉、飞书配置
4. **安装进度**: 实时显示进度和状态
5. **完成页**: 安装成功，启动应用

## 构建

### 前置要求

- Go 1.21 或更高版本
- Wails CLI v2.8.0+
- Node.js 18+ (可选，用于前端开发)

安装 Wails:
```bash
go install github.com/wailsapp/wails/v2/cmd/wails@latest
```

### 开发模式

```bash
# 进入项目目录
cd wails-installer

# 开发模式（带热重载）
wails dev
```

### 生产构建

```bash
# 构建当前平台
wails build

# 使用脚本构建所有平台
./build.sh --all

# 构建特定平台
./build.sh --windows
./build.sh --macos
./build.sh --linux
```

### 交叉编译

```bash
# Windows (from macOS/Linux)
GOOS=windows GOARCH=amd64 CGO_ENABLED=0 \
  go build -ldflags "-H=windowsgui -s -w" -o dist/OpenClaw-Installer.exe .

# macOS (from Linux - 需要 osxcross)
GOOS=darwin GOARCH=amd64 CGO_ENABLED=0 \
  go build -ldflags "-s -w" -o dist/OpenClaw-Installer .

# Linux (from macOS/Windows)
GOOS=linux GOARCH=amd64 CGO_ENABLED=0 \
  go build -ldflags "-s -w" -o dist/OpenClaw-Installer .
```

## 配置说明

### wails.json

```json
{
  "name": "openclaw-installer",
  "outputfilename": "OpenClaw-Installer",
  "frontend": {
    "dir": "./frontend/dist"
  },
  "info": {
    "companyName": "OpenClaw",
    "productName": "OpenClaw Installer",
    "productVersion": "1.0.0"
  }
}
```

### 前端绑定

Go 函数通过 Wails 绑定到前端:

```go
// Go 后端
func (a *App) GetPlatformInfo() platform.PlatformInfo {
    return a.platform.GetInfo()
}
```

```javascript
// JavaScript 前端
const info = await window.go.main.App.GetPlatformInfo();
console.log(info.os, info.arch);
```

## 平台特定说明

### Windows

- 使用 `-H=windowsgui` 链接器标志隐藏控制台窗口
- 支持 Windows 10/11 (amd64, arm64)
- 可选 NSIS 安装包生成

### macOS

- 支持 Intel 和 Apple Silicon (universal binary)
- 深色模式自动适配
- 可生成 .app 和 .dmg

### Linux

- 支持主流发行版 (Ubuntu, Fedora, etc.)
- 依赖: libgtk-3-0, libwebkit2gtk-4.0-37
- 可生成 .deb, .rpm, AppImage

## 前端界面

### 主题颜色

- **主背景**: `#0f172a` (深蓝灰)
- **卡片背景**: `rgba(255, 255, 255, 0.05)`
- **主色调**: `#00d4ff` (青色)
- **次色调**: `#7b2ff7` (紫色)
- **成功色**: `#10b981` (绿色)
- **文字主色**: `#e8e8e8`
- **文字次色**: `#8892a0`

### 动画效果

- 页面切换: fadeIn 0.3s ease
- 按钮悬停: translateY(-2px) + shadow
- 进度条: width transition 0.3s
- 加载动画: spin 0.8s linear infinite

## 开发指南

### 添加新的 Go 绑定函数

1. 在 `main.go` 中添加方法:
```go
func (a *App) MyNewFunction() string {
    return "Hello from Go!"
}
```

2. 在前端调用:
```javascript
const result = await window.go.main.App.MyNewFunction();
```

### 修改前端界面

直接编辑 `frontend/dist/index.html`，包含所有 HTML、CSS 和 JavaScript。

### 调试

```bash
# 开发模式（带浏览器开发者工具）
wails dev

# 日志输出
wails dev -loglevel debug
```

## 分发

### 单文件分发

构建后的单个可执行文件包含所有资源，可直接分发：

```
dist/
├── OpenClaw-Installer-windows-amd64.exe
├── OpenClaw-Installer-darwin-amd64
├── OpenClaw-Installer-darwin-arm64
├── OpenClaw-Installer-linux-amd64
└── OpenClaw-Installer-linux-arm64
```

### U盘部署

将安装器复制到 U盘，与 packages 目录一起分发:

```
/OpenClaw/
├── installers/
│   ├── OpenClaw-Installer-windows-amd64.exe
│   ├── OpenClaw-Installer-darwin-arm64
│   └── OpenClaw-Installer-linux-amd64
└── packages/
    ├── openclaw-core/
    └── adapters/
```

## 许可证

Copyright © 2026 OpenClaw Team
