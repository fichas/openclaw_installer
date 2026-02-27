# OpenClaw 安装器

跨平台 OpenClaw AI 助手安装器，支持企业微信、钉钉、飞书适配器配置。

## 快速开始

### 1. 准备 U 盘

将 `usb-deploy/OpenClaw/` 目录复制到 U 盘根目录：
```bash
cp -r usb-deploy/OpenClaw/* /mnt/usb/
```

### 2. 运行安装器

**Windows:**
- 插入 U 盘，自动弹出安装界面
- 或手动运行 `install.bat`

**macOS:**
```bash
/Volumes/USB/OpenClaw/autorun/install-mac.command
```

**Linux:**
```bash
sudo /media/user/USB/OpenClaw/installers/openclaw-installer-linux-amd64
```

### 3. 配置界面

安装器启动后自动打开浏览器访问 `http://localhost:18080`

## 配置流程

### Step 1: 基础配置
| 字段 | 说明 | 示例 |
|------|------|------|
| AI 模型 | 选择模型提供商 | `anthropic/claude-opus-4-6` |
| API Key | 模型提供商的 API Key | `sk-ant-...` |
| Gateway 端口 | 服务监听端口 | `18080` |

### Step 2: 企业微信配置
| 字段 | 获取位置 |
|------|----------|
| CorpID | 企业微信管理后台 > 我的企业 > 企业ID |
| AgentID | 应用管理 > 自建应用 > AgentId |
| Secret | 应用详情 > Secret |
| Token | 接收消息 > Token |
| EncodingAESKey | 接收消息 > EncodingAESKey |

### Step 3: 钉钉配置
| 字段 | 获取位置 |
|------|----------|
| AppKey | 开发者后台 > 应用详情 |
| AppSecret | 应用详情 |
| Webhook | 群机器人设置 |

### Step 4: 飞书配置
| 字段 | 获取位置 |
|------|----------|
| AppID | 开发者平台 > 应用凭证 |
| AppSecret | 应用凭证 |
| Encrypt Key | 事件订阅 > Encrypt Key |

## 启动 OpenClaw

```bash
# 启动 Gateway
openclaw gateway --port 18789

# 发送测试消息
openclaw agent --message "你好"

# 查看状态
openclaw doctor
```

## 命令行参数

| 参数 | 说明 | 示例 |
|------|------|------|
| `--port` | 指定端口 | `--port 18080` |
| `--no-browser` | 不自动打开浏览器 | `--no-browser` |
| `--debug` | 调试模式 | `--debug` |

## 故障排除

### 浏览器未自动打开
手动访问：`http://localhost:18080`

### 端口被占用
```bash
./openclaw-installer --port 18081
```

### 配置验证失败
- 检查 API Key 是否正确
- 确认 Secret 未过期
- 查看日志 `~/.openclaw/logs/`

## 更新系统

```bash
# 检查更新
openclaw-update check

# 执行更新
openclaw-update -yes

# 回滚版本
openclaw-update -rollback
```

## 系统支持

| 平台 | 架构 | 状态 |
|------|------|------|
| Windows | x64, ARM64 | ✅ |
| macOS | x64, ARM64 | ✅ |
| Linux | x64, ARM64 | ✅ |

## 文档

- [详细构建文档](BUILD.md)
- [架构设计](docs/architecture.md)
- [使用文档](docs/USAGE.md)
- [测试报告](docs/TEST_REPORT.md)

## 许可证

MIT License
