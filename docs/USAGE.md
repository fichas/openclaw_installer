# OpenClaw 安装器使用文档

> 本文档面向中国用户，提供 OpenClaw 安装和配置的完整指南。

---

## 目录

1. [快速开始](#快速开始)
2. [安装步骤详解](#安装步骤详解)
3. [Web 配置界面使用说明](#web-配置界面使用说明)
4. [命令行参数](#命令行参数)
5. [常见问题解答](#常见问题解答)
6. [故障排除指南](#故障排除指南)

---

## 快速开始

### 系统要求

| 平台 | 最低版本 | 架构 |
|------|----------|------|
| Windows | Windows 10 | x64 / ARM64 |
| macOS | macOS 11 (Big Sur) | Intel / Apple Silicon |
| Linux | Ubuntu 20.04 / CentOS 8 | x64 / ARM64 |

### 5 分钟快速安装

1. **准备 U 盘**
   - 将 OpenClaw 安装包解压到 U 盘根目录
   - 确保 U 盘目录结构完整（包含 `installers/` 和 `packages/` 目录）

2. **运行安装器**

   **Windows:**
   ```powershell
   # 双击运行或使用 PowerShell
   .\installers\openclaw-installer-windows-amd64.exe
   ```

   **macOS:**
   ```bash
   # 双击运行或在终端执行
   ./installers/openclaw-installer-darwin-arm64
   ```

   **Linux:**
   ```bash
   # 在终端执行
   sudo ./installers/openclaw-installer-linux-amd64
   ```

3. **完成配置**
   - 浏览器会自动打开 `http://localhost:18080`
   - 按照 Web 界面指引完成配置（约 5-10 分钟）
   - 点击"启动服务"完成安装

---

## 安装步骤详解

### Windows 安装

#### 方法一：图形界面安装（推荐）

1. 插入包含 OpenClaw 安装文件的 U 盘
2. 打开文件资源管理器，进入 U 盘目录
3. 双击 `installers/openclaw-installer-windows-amd64.exe`
4. 如果弹出"Windows 已保护你的电脑"，点击"更多信息" -> "仍要运行"
5. 等待浏览器自动打开配置界面

#### 方法二：命令行安装

```powershell
# 以管理员身份运行 PowerShell
# 进入 U 盘目录
cd D:\OpenClaw  # 根据实际盘符调整

# 运行安装器
.\installers\openclaw-installer-windows-amd64.exe

# 或使用参数指定端口
.\installers\openclaw-installer-windows-amd64.exe --port 18081
```

#### 安装路径

| 类型 | 默认路径 |
|------|----------|
| 程序文件 | `C:\Program Files\OpenClaw` |
| 配置文件 | `C:\ProgramData\OpenClaw\config` |
| 数据文件 | `C:\ProgramData\OpenClaw\data` |
| 日志文件 | `C:\ProgramData\OpenClaw\logs` |

---

### macOS 安装

#### 方法一：双击运行

1. 插入 U 盘，在 Finder 中打开
2. 进入 `installers` 目录
3. 双击对应的安装器文件（Intel 选 amd64，Apple Silicon 选 arm64）
4. 如果提示"无法打开"，请按住 Control 键再点击，选择"打开"
5. 在弹出的终端窗口中按回车键继续

#### 方法二：终端安装

```bash
# 插入 U 盘，进入目录
cd /Volumes/OpenClaw

# 运行安装器（Apple Silicon）
./installers/openclaw-installer-darwin-arm64

# 或 Intel Mac
./installers/openclaw-installer-darwin-amd64

# 可能需要绕过 Gatekeeper
xattr -d com.apple.quarantine ./installers/openclaw-installer-darwin-arm64
```

#### 安装路径

| 类型 | 默认路径 |
|------|----------|
| 程序文件 | `/usr/local/bin` |
| 配置文件 | `/usr/local/etc/openclaw` |
| 数据文件 | `/usr/local/share/openclaw` |
| 日志文件 | `/var/log/openclaw` |

---

### Linux 安装

#### 支持的发行版

- Ubuntu 20.04+
- Debian 11+
- CentOS 8 / RHEL 8+
- Fedora 34+

#### 安装步骤

```bash
# 1. 插入 U 盘并挂载（通常会自动挂载）
# 查看挂载点
lsblk

# 2. 进入 U 盘目录
cd /media/$USER/OpenClaw

# 3. 以 root 权限运行安装器
sudo ./installers/openclaw-installer-linux-amd64

# 或使用 sudo -E 保留环境变量
sudo -E ./installers/openclaw-installer-linux-amd64
```

#### systemd 服务（可选）

安装完成后，可以创建 systemd 服务：

```bash
# 创建服务文件
sudo tee /etc/systemd/system/openclaw.service > /dev/null <<EOF
[Unit]
Description=OpenClaw AI Assistant
After=network.target

[Service]
Type=simple
User=openclaw
ExecStart=/usr/local/bin/openclaw
Restart=always
RestartSec=5

[Install]
WantedBy=multi-user.target
EOF

# 启用并启动服务
sudo systemctl enable openclaw
sudo systemctl start openclaw
```

#### 安装路径

| 类型 | 默认路径 |
|------|----------|
| 程序文件 | `/usr/local/bin` |
| 配置文件 | `/etc/openclaw` |
| 数据文件 | `/var/lib/openclaw` |
| 日志文件 | `/var/log/openclaw` |

---

## Web 配置界面使用说明

### 配置流程概览

```
欢迎页 → 基础配置 → 企业微信 → 钉钉 → 飞书 → 完成页
```

每个步骤都可以保存进度，支持随时返回修改。

---

### 第一步：欢迎页

**功能**：介绍 OpenClaw 并开始配置流程

**操作**：
- 查看支持的 IM 平台（企业微信、钉钉、飞书）
- 点击"开始配置"进入基础配置

**提示**：配置预计需要 5-10 分钟

---

### 第二步：基础配置

**功能**：配置 AI 模型和 Gateway 服务

#### 配置项说明

| 配置项 | 必填 | 说明 |
|--------|------|------|
| AI 模型提供商 | 是 | 选择使用的 AI 服务：OpenAI、Anthropic、Azure OpenAI、本地模型 |
| API 密钥 | 是 | 对应 AI 平台的 API Key |
| Gateway 端口 | 是 | OpenClaw 服务监听端口，默认 8080 |

#### AI 模型选择建议

| 场景 | 推荐模型 | 说明 |
|------|----------|------|
| 追求最佳效果 | GPT-4 | 智能程度最高，适合复杂对话 |
| 追求性价比 | GPT-3.5 | 成本较低，响应速度快 |
| 数据敏感场景 | 本地模型 | 数据不出境，适合机密环境 |

#### API 密钥获取指南

**OpenAI:**
1. 访问 [OpenAI Platform](https://platform.openai.com)
2. 登录后进入 API keys 页面
3. 点击 "Create new secret key"
4. 复制生成的密钥（以 `sk-` 开头）

**Anthropic (Claude):**
1. 访问 [Anthropic Console](https://console.anthropic.com)
2. 进入 API keys 页面
3. 创建新的 API key

---

### 第三步：企业微信配置

**功能**：配置企业微信应用，实现与企业微信的对接

#### 配置项说明

| 配置项 | 必填 | 说明 | 示例 |
|--------|------|------|------|
| CorpID | 是 | 企业微信的企业 ID | `wwxxxxxxxxxxxxxxxx` |
| AgentID | 是 | 自建应用的 ID | `1000002` |
| Secret | 是 | 应用的密钥 | 32位随机字符串 |
| Token | 否 | 消息加密令牌 | 用于回调验证 |
| EncodingAESKey | 否 | 消息加密密钥 | 43位字符串 |

#### 如何获取企业微信参数

**1. 获取 CorpID（企业 ID）**

1. 登录 [企业微信管理后台](https://work.weixin.qq.com/wework_admin)
2. 点击顶部菜单「我的企业」
3. 在页面底部找到「企业ID」，点击复制

**2. 创建自建应用**

1. 进入「应用管理」页面
2. 点击「自建」->「创建应用」
3. 填写应用信息：
   - 应用名称：OpenClaw
   - 应用介绍：AI 智能助手
   - 应用图标：上传 Logo
4. 点击「创建应用」

**3. 获取 AgentID 和 Secret**

1. 在应用列表中找到刚创建的「OpenClaw」应用
2. 点击进入应用详情
3. 记录「AgentId」
4. 点击「Secret」旁边的「查看」，按提示操作获取 Secret

**4. 配置接收消息（可选）**

如需让 AI 主动回复消息：

1. 在应用详情页，找到「接收消息」
2. 点击「设置API接收」
3. 填写以下信息：
   - URL：`http://your-server:8080/wecom/webhook`
   - Token：随机生成或自定义
   - EncodingAESKey：点击「随机生成」
4. 将 Token 和 EncodingAESKey 填入 OpenClaw 配置

---

### 第四步：钉钉配置

**功能**：配置钉钉应用和群机器人

#### 配置项说明

| 配置项 | 必填 | 说明 | 示例 |
|--------|------|------|------|
| AppKey | 是 | 钉钉应用标识 | `dingxxxxxxxxxxxxxx` |
| AppSecret | 是 | 应用密钥 | 随机字符串 |
| Webhook 地址 | 否 | 群机器人 Webhook | `https://oapi.dingtalk.com/...` |
| Webhook Secret | 否 | 群机器人加签密钥 | 用于安全验证 |
| 机器人名称 | 否 | 群机器人名称 | `OpenClaw AI助手` |

#### 如何获取钉钉参数

**1. 创建钉钉应用**

1. 登录 [钉钉开放平台](https://open.dingtalk.com)
2. 进入「应用开发」->「企业内部开发」
3. 点击「创建应用」
4. 填写应用信息：
   - 应用名称：OpenClaw
   - 应用类型：H5微应用
   - 开发方式：企业自助开发

**2. 获取 AppKey 和 AppSecret**

1. 进入应用详情页的「基础信息」
2. 记录「AppKey」
3. 点击「AppSecret」旁边的「查看」，获取密钥

**3. 配置权限**

1. 进入「权限管理」
2. 添加以下权限：
   - 通讯录管理
   - 群会话管理
   - 机器人管理

**4. 创建群机器人（可选）**

1. 在钉钉群中，点击「群设置」->「智能群助手」
2. 点击「添加机器人」->「自定义」
3. 设置机器人名称和头像
4. 选择安全设置：
   - 推荐选择「加签」，复制密钥
   - 或设置「IP地址（段）」白名单
5. 复制 Webhook 地址

---

### 第五步：飞书配置

**功能**：配置飞书应用，实现与飞书的对接

#### 配置项说明

| 配置项 | 必填 | 说明 | 示例 |
|--------|------|------|------|
| AppID | 是 | 飞书应用 ID | `cli_xxxxxxxxxxxxxx` |
| AppSecret | 是 | 应用密钥 | 随机字符串 |
| Encrypt Key | 否 | 事件加密密钥 | 用于消息加密 |
| Verification Token | 否 | 事件验证令牌 | 用于验证请求 |
| 事件订阅地址 | - | 自动生成 | 用于配置飞书后台 |

#### 如何获取飞书参数

**1. 创建飞书应用**

1. 登录 [飞书开发者平台](https://open.feishu.cn/app)
2. 点击「创建企业自建应用」
3. 填写应用信息：
   - 应用名称：OpenClaw
   - 应用描述：AI 智能助手

**2. 获取 AppID 和 AppSecret**

1. 进入应用详情页的「凭证与基础信息」
2. 记录「App ID」
3. 点击「App Secret」旁边的「查看」，获取密钥

**3. 配置事件订阅（可选）**

1. 进入「事件与回调」页面
2. 点击「启用事件」
3. 配置加密方式：
   - 点击「随机生成」生成 Encrypt Key
   - 记录 Verification Token
4. 配置请求网址：
   - 将 OpenClaw 配置页面显示的「事件订阅地址」填入
   - 格式：`http://your-server:8080/feishu/webhook`

**4. 发布应用**

1. 进入「版本管理与发布」
2. 点击「创建版本」
3. 填写版本信息并发布
4. 在飞书管理后台审核通过

---

### 第六步：完成配置

**功能**：查看配置摘要并启动服务

#### 配置摘要

页面会显示：
- AI 模型配置
- Gateway 地址
- 已启用的 IM 渠道

#### 启动选项

| 选项 | 说明 |
|------|------|
| 开机自动启动 | 系统启动时自动运行 OpenClaw |
| 启动后最小化到托盘 | 启动后隐藏到系统托盘 |

#### 启动服务

点击「启动服务」按钮，等待服务启动完成。

启动成功后，您可以通过以下方式访问：
- Web 界面：`http://localhost:8080`
- API 接口：`http://localhost:8080/api`

---

## 命令行参数

### 基本用法

```bash
openclaw-installer [选项]
```

### 参数列表

| 参数 | 简写 | 说明 | 默认值 |
|------|------|------|--------|
| `--port` | `-p` | Web 配置服务端口 | `18080` |
| `--no-browser` | `-n` | 不自动打开浏览器 | `false` |
| `--debug` | `-d` | 启用调试日志 | `false` |
| `--config` | `-c` | 指定配置文件路径 | 自动检测 |
| `--help` | `-h` | 显示帮助信息 | - |
| `--version` | `-v` | 显示版本信息 | - |

### 使用示例

```bash
# 使用自定义端口
openclaw-installer --port 18081

# 不自动打开浏览器（适用于远程服务器）
openclaw-installer --no-browser

# 启用调试模式
openclaw-installer --debug

# 指定配置文件
openclaw-installer --config /path/to/config.yaml

# 组合使用
openclaw-installer -p 18081 -n -d
```

---

## 常见问题解答

### 一般问题

**Q: OpenClaw 是免费的吗？**

A: OpenClaw 安装器本身是开源免费的，但使用 AI 服务（如 OpenAI）需要支付相应的 API 费用。

**Q: 支持哪些 AI 模型？**

A: 目前支持 OpenAI GPT 系列、Anthropic Claude、Azure OpenAI 以及本地部署的开源模型（如 Llama、ChatGLM 等）。

**Q: 可以同时配置多个 IM 平台吗？**

A: 可以。OpenClaw 支持同时接入企业微信、钉钉、飞书，用户可以在不同平台同时使用 AI 助手。

---

### 安装问题

**Q: 运行安装器时提示"权限不足"怎么办？**

A:
- **Windows**: 右键点击安装器，选择"以管理员身份运行"
- **macOS/Linux**: 在命令前加 `sudo`，如 `sudo ./openclaw-installer`

**Q: macOS 提示"无法打开，因为无法验证开发者"？**

A: 打开「系统设置」->「隐私与安全性」，在安全性部分点击「仍要打开」。或者在终端执行：
```bash
xattr -d com.apple.quarantine /path/to/installer
```

**Q: 浏览器没有自动打开怎么办？**

A: 手动访问 `http://localhost:18080`。如果端口被占用，可以使用 `--port` 参数指定其他端口。

---

### 配置问题

**Q: 如何获取 OpenAI API Key？**

A:
1. 访问 [OpenAI Platform](https://platform.openai.com)
2. 注册/登录账号
3. 进入 API keys 页面
4. 创建新的 secret key

**Q: 企业微信的 Secret 在哪里找？**

A: 在企业微信管理后台的应用详情页，点击 Secret 旁边的「查看」，需要管理员扫码确认后才能显示。

**Q: 配置完成后如何测试是否成功？**

A:
1. 在配置完成页点击「测试连接」按钮
2. 或在对应 IM 平台中 @机器人 发送消息测试
3. 查看日志文件确认消息是否正常处理

---

### 使用问题

**Q: 如何修改已完成的配置？**

A:
- 方式一：重新运行安装器，选择「修改配置」
- 方式二：直接编辑配置文件（位于 `/etc/openclaw/config.yaml` 或 `C:\ProgramData\OpenClaw\config\config.yaml`）

**Q: 服务启动失败怎么办？**

A: 检查以下几点：
1. 端口是否被其他程序占用
2. API Key 是否正确有效
3. 配置文件格式是否正确
4. 查看日志文件获取详细错误信息

**Q: 如何查看运行日志？**

A:
- **Windows**: `C:\ProgramData\OpenClaw\logs\openclaw.log`
- **macOS**: `/var/log/openclaw/openclaw.log`
- **Linux**: `/var/log/openclaw/openclaw.log`

---

## 故障排除指南

### 安装器无法启动

#### 症状：双击安装器无反应

**排查步骤：**

1. **检查系统架构**
   ```bash
   # macOS
   uname -m  # arm64 或 x86_64

   # Linux
   uname -m  # x86_64 或 aarch64
   ```
   确保下载的安装器与系统架构匹配。

2. **检查文件完整性**
   - 确认 U 盘目录结构完整
   - 检查 `packages/` 目录是否存在

3. **检查权限**
   ```bash
   # macOS/Linux
   ls -la installers/
   # 确保文件有执行权限
   chmod +x installers/openclaw-installer-*
   ```

#### 症状：提示"找不到 U 盘"

**解决方案：**

1. 确认 U 盘已正确插入
2. 检查 U 盘目录结构：
   ```
   /OpenClaw/
   ├── installers/
   └── packages/
   ```
3. 手动指定 U 盘路径（如支持）：
   ```bash
   openclaw-installer --usb-path /path/to/usb
   ```

---

### Web 界面无法访问

#### 症状：浏览器显示"无法访问此网站"

**排查步骤：**

1. **检查安装器是否运行**
   ```bash
   # 查看进程
   # Windows
   Get-Process | Where-Object {$_.ProcessName -like "*openclaw*"}

   # macOS/Linux
   ps aux | grep openclaw
   ```

2. **检查端口占用**
   ```bash
   # 查看 18080 端口是否被占用
   # Windows
   netstat -ano | findstr :18080

   # macOS/Linux
   lsof -i :18080
   ```

3. **更换端口启动**
   ```bash
   openclaw-installer --port 18081
   ```

4. **检查防火墙**
   - Windows: 允许安装器通过防火墙
   - macOS: 在「系统设置」->「网络」->「防火墙」中添加例外
   - Linux: `sudo ufw allow 18080/tcp`

---

### 配置验证失败

#### 症状：API Key 验证失败

**可能原因及解决方案：**

| 错误信息 | 原因 | 解决方案 |
|----------|------|----------|
| Invalid API Key | API Key 错误或已失效 | 重新获取有效的 API Key |
| Insufficient quota | API 账户余额不足 | 检查 OpenAI 账户余额并充值 |
| Rate limit exceeded | 请求频率过高 | 稍后再试或升级账户 |
| Network error | 网络连接问题 | 检查网络连接，确认能访问 api.openai.com |

#### 症状：企业微信连接测试失败

**排查步骤：**

1. **检查 CorpID 格式**
   - 应以 `ww` 开头
   - 长度应为 18 位

2. **检查 AgentID**
   - 应为纯数字
   - 确认与应用对应

3. **检查 Secret**
   - 确认与应用匹配
   - Secret 区分大小写

4. **检查应用状态**
   - 登录企业微信后台
   - 确认应用已启用
   - 确认应用可见范围包含测试用户

#### 症状：钉钉连接测试失败

**排查步骤：**

1. **检查 AppKey 和 AppSecret**
   - 确认来自同一应用
   - 确认应用已发布

2. **检查 IP 白名单**
   - 在钉钉开放平台，将服务器公网 IP 添加到白名单

3. **检查权限**
   - 确认应用具有必要的权限
   - 确认权限已申请并通过

#### 症状：飞书连接测试失败

**排查步骤：**

1. **检查 AppID 格式**
   - 应以 `cli_` 开头

2. **检查应用发布状态**
   - 确认应用已创建版本并发布
   - 确认管理员已审核通过

3. **检查事件订阅配置**
   - 确认事件订阅地址正确
   - 确认服务器可被飞书访问

---

### 服务启动失败

#### 症状：点击"启动服务"后报错

**常见错误及解决方案：**

**错误：端口被占用**
```
Error: listen tcp :8080: bind: address already in use
```
**解决：** 修改 Gateway 端口为其他未被占用的端口（如 8081、9090 等）

**错误：配置文件权限不足**
```
Error: permission denied: /etc/openclaw/config.yaml
```
**解决：**
```bash
# macOS/Linux
sudo chown -R $USER /etc/openclaw

# 或修改权限
sudo chmod 644 /etc/openclaw/config.yaml
```

**错误：无法写入日志**
```
Error: failed to open log file: permission denied
```
**解决：**
```bash
# 创建日志目录并设置权限
sudo mkdir -p /var/log/openclaw
sudo chown -R $USER /var/log/openclaw
```

---

### 获取帮助

如果以上方法无法解决问题，可以通过以下方式获取帮助：

1. **查看日志**：日志文件包含详细的错误信息
2. **开启调试模式**：使用 `--debug` 参数运行，获取更详细的输出
3. **提交 Issue**：在 GitHub 仓库提交问题，附上日志和错误信息

---

## 附录

### 配置文件示例

```yaml
# /etc/openclaw/config.yaml

# AI 配置
ai:
  provider: openai
  api_key: sk-your-api-key-here
  model: gpt-4

# Gateway 配置
gateway:
  port: 8080
  host: 0.0.0.0

# 日志配置
log:
  level: info
  file: /var/log/openclaw/openclaw.log

# 适配器配置
adapters:
  wecom:
    enabled: true
    corp_id: wwxxxxxxxxxxxxxxxx
    agent_id: 1000002
    secret: your-secret-here
    token: your-token
    encoding_aes_key: your-aes-key

  dingtalk:
    enabled: true
    app_key: dingxxxxxxxxxxxxxx
    app_secret: your-secret-here
    webhook_url: https://oapi.dingtalk.com/robot/send?access_token=xxx
    webhook_secret: your-webhook-secret

  feishu:
    enabled: true
    app_id: cli_xxxxxxxxxxxxxx
    app_secret: your-secret-here
    encrypt_key: your-encrypt-key
    verification_token: your-verification-token
```

### 目录结构参考

```
/OpenClaw/                          # U 盘根目录
├── installers/                     # 各平台安装器
│   ├── openclaw-installer-darwin-amd64
│   ├── openclaw-installer-darwin-arm64
│   ├── openclaw-installer-windows-amd64.exe
│   ├── openclaw-installer-windows-arm64.exe
│   ├── openclaw-installer-linux-amd64
│   └── openclaw-installer-linux-arm64
│
├── packages/                       # 预打包的安装包
│   ├── openclaw-core/             # OpenClaw 核心
│   │   ├── openclaw-darwin-amd64.tar.gz
│   │   ├── openclaw-darwin-arm64.tar.gz
│   │   ├── openclaw-windows-amd64.zip
│   │   ├── openclaw-windows-arm64.zip
│   │   ├── openclaw-linux-amd64.tar.gz
│   │   └── openclaw-linux-arm64.tar.gz
│   │
│   └── adapters/                  # 适配器包
│       ├── wecom-adapter/
│       ├── dingtalk-adapter/
│       └── feishu-adapter/
│
└── resources/                     # 附加资源
    ├── icons/
    ├── licenses/
    └── docs/
```

---

**文档版本**: 1.0.0
**最后更新**: 2026-02-28
