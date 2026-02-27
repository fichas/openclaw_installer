# OpenClaw 用户使用指南

> 本文档面向非技术用户，帮助您快速上手 OpenClaw 企业 IM 集成工具。

---

## 目录

1. [快速开始（1分钟上手）](#快速开始1分钟上手)
2. [详细安装指南](#详细安装指南)
3. [首次配置](#首次配置)
4. [日常使用](#日常使用)
5. [常见问题](#常见问题)
6. [故障排除](#故障排除)

---

## 快速开始（1分钟上手）

### 什么是 OpenClaw？

OpenClaw 是一款企业 IM 集成工具，可以将 AI 助手接入您的企业微信、钉钉或飞书，让员工通过熟悉的聊天工具与 AI 对话。

### 三步完成安装

**第一步：下载对应版本**

| 您的系统 | 下载文件 |
|---------|---------|
| Windows 10/11 | `OpenClaw-Installer-windows.exe` |
| Mac 电脑（Intel） | `OpenClaw-Installer-mac-intel` |
| Mac 电脑（M1/M2/M3） | `OpenClaw-Installer-mac-apple` |
| Linux | `OpenClaw-Installer-linux` |

**第二步：运行安装器**

- **Windows**：双击下载的安装文件
- **Mac**：双击安装文件，按提示允许运行
- **Linux**：右键安装文件，选择"作为程序运行"

**第三步：完成配置**

1. 浏览器会自动打开配置页面
2. 按页面提示填写信息（约需 5 分钟）
3. 点击"启动服务"完成安装

---

## 详细安装指南

### 系统要求

在安装前，请确认您的电脑满足以下条件：

| 项目 | 最低要求 |
|-----|---------|
| 操作系统 | Windows 10 / macOS 11 / Ubuntu 20.04 |
| 可用空间 | 500 MB |
| 网络 | 能访问互联网（用于验证配置） |
| 权限 | 管理员权限（Windows）或 root 权限（Linux） |

---

### Windows 安装步骤

#### 方法一：图形界面安装（推荐）

1. **插入 U 盘**（如果使用 U 盘安装）
2. **打开安装文件**
   - 双击 `OpenClaw-Installer-windows.exe`
   - 如果弹出"Windows 已保护你的电脑"，点击**"更多信息"** -> **"仍要运行"**

3. **允许管理员权限**
   - 弹出 UAC 提示时，点击**"是"**

4. **等待浏览器打开**
   - 安装器会自动打开浏览器进入配置页面
   - 如果没有自动打开，手动访问 `http://localhost:18080`

#### 方法二：从 U 盘安装

1. 将 U 盘插入电脑
2. 打开文件资源管理器，进入 U 盘
3. 双击 `installers/OpenClaw-Installer-windows.exe`
4. 按提示完成安装

#### 安装位置

安装完成后，文件会存放在以下位置：

| 类型 | 路径 |
|-----|------|
| 程序文件 | `C:\Program Files\OpenClaw` |
| 配置文件 | `C:\ProgramData\OpenClaw\config` |
| 日志文件 | `C:\ProgramData\OpenClaw\logs` |

---

### macOS 安装步骤

#### 适用于 Intel Mac

1. **插入 U 盘**（如果使用 U 盘安装）
2. **运行安装器**
   - 双击 `OpenClaw-Installer-mac-intel`
   - 如果提示"无法打开"，按住 **Control 键**再点击，选择**"打开"**

3. **绕过安全限制（如需要）**
   - 打开"系统设置" -> "隐私与安全性"
   - 在"安全性"部分点击**"仍要打开"**
   - 或在终端执行：
   ```bash
   xattr -d com.apple.quarantine /path/to/installer
   ```

#### 适用于 Apple Silicon Mac（M1/M2/M3）

1. **插入 U 盘**（如果使用 U 盘安装）
2. **运行安装器**
   - 双击 `OpenClaw-Installer-mac-apple`
   - 如果提示"无法打开"，按住 **Control 键**再点击，选择**"打开"**

3. **允许运行**
   - 在弹出的终端窗口中按**回车键**继续

#### 安装位置

| 类型 | 路径 |
|-----|------|
| 程序文件 | `/usr/local/bin` |
| 配置文件 | `/usr/local/etc/openclaw` |
| 日志文件 | `/var/log/openclaw` |

---

### Linux 安装步骤

#### 支持的发行版

- Ubuntu 20.04 或更高版本
- Debian 11 或更高版本
- CentOS 8 / RHEL 8 或更高版本
- Fedora 34 或更高版本

#### 安装步骤

1. **插入 U 盘**（如果使用 U 盘安装）
   ```bash
   # 查看 U 盘挂载位置
   lsblk
   ```

2. **进入 U 盘目录**
   ```bash
   cd /media/$USER/OpenClaw
   # 或
   cd /mnt/usb/OpenClaw
   ```

3. **运行安装器**
   ```bash
   sudo ./installers/OpenClaw-Installer-linux
   ```

4. **输入密码**
   - 输入您的用户密码（输入时不显示）
   - 按回车确认

#### 安装位置

| 类型 | 路径 |
|-----|------|
| 程序文件 | `/usr/local/bin` |
| 配置文件 | `/etc/openclaw` |
| 日志文件 | `/var/log/openclaw` |

---

## 首次配置

### 配置流程概览

```
欢迎页 → 基础配置 → 企业微信（可选） → 钉钉（可选） → 飞书（可选） → 完成
```

---

### 第一步：基础配置（AI 模型）

**配置项说明**

| 配置项 | 必填 | 说明 |
|--------|------|------|
| AI 模型提供商 | 是 | 选择使用的 AI 服务 |
| API 密钥 | 是 | 对应 AI 平台的密钥 |
| 服务端口 | 是 | 默认 8080 |

**如何选择 AI 模型**

| 您的需求 | 推荐选择 |
|---------|---------|
| 追求最佳回答质量 | GPT-4 |
| 追求性价比 | GPT-3.5 |
| 数据敏感，不能出境 | 本地模型 |

**获取 API 密钥**

**OpenAI（ChatGPT）：**
1. 访问 [OpenAI 官网](https://platform.openai.com)
2. 注册或登录账号
3. 点击右上角头像 -> "View API keys"
4. 点击 "Create new secret key"
5. 复制生成的密钥（以 `sk-` 开头）

**Anthropic（Claude）：**
1. 访问 [Anthropic 控制台](https://console.anthropic.com)
2. 登录后进入 API keys 页面
3. 创建新的 API key

---

### 第二步：企业微信配置（可选）

如果您不需要接入企业微信，可以点击"跳过"。

**需要准备的信息**

| 信息项 | 获取位置 |
|--------|---------|
| 企业 ID | 企业微信管理后台 -> 我的企业 |
| 应用 ID | 应用管理 -> 自建应用 |
| 应用密钥 | 应用详情页 |

**详细获取步骤**

**1. 获取企业 ID**

1. 用管理员账号登录 [企业微信管理后台](https://work.weixin.qq.com/wework_admin)
2. 点击顶部菜单**"我的企业"**
3. 在页面底部找到**"企业ID"**，点击复制

**2. 创建应用**

1. 进入**"应用管理"**页面
2. 向下滚动到**"自建"**区域
3. 点击**"创建应用"**
4. 填写应用信息：
   - 应用名称：**OpenClaw**
   - 应用介绍：AI 智能助手
   - 应用图标：可以上传公司 Logo
5. 点击**"创建应用"**

**3. 获取应用 ID 和密钥**

1. 在应用列表中找到刚创建的"OpenClaw"
2. 点击进入应用详情
3. 记录**"AgentId"**（这就是应用 ID）
4. 点击**"Secret"**旁边的**"查看"**
5. 按提示扫码确认后，复制密钥

**4. 配置接收消息（可选）**

如需让 AI 主动回复消息：

1. 在应用详情页，找到**"接收消息"**
2. 点击**"设置API接收"**
3. 填写以下信息：
   - URL：`http://您的服务器地址:8080/wecom/webhook`
   - Token：随机填写，如 `openclaw123`
   - EncodingAESKey：点击**"随机生成"**
4. 将 Token 和 EncodingAESKey 填入 OpenClaw 配置

---

### 第三步：钉钉配置（可选）

如果您不需要接入钉钉，可以点击"跳过"。

**需要准备的信息**

| 信息项 | 获取位置 |
|--------|---------|
| AppKey | 钉钉开放平台 -> 应用详情 |
| AppSecret | 应用详情页 |
| Webhook 地址 | 群机器人设置（可选） |

**详细获取步骤**

**1. 创建钉钉应用**

1. 登录 [钉钉开放平台](https://open.dingtalk.com)
2. 点击**"应用开发"** -> **"企业内部开发"**
3. 点击**"创建应用"**
4. 填写应用信息：
   - 应用名称：**OpenClaw**
   - 应用类型：H5微应用
   - 开发方式：企业自助开发

**2. 获取 AppKey 和 AppSecret**

1. 进入应用详情页的**"基础信息"**
2. 记录**"AppKey"**
3. 点击**"AppSecret"**旁边的**"查看"**，获取密钥

**3. 配置权限**

1. 进入**"权限管理"**
2. 添加以下权限：
   - 通讯录管理
   - 群会话管理
   - 机器人管理

**4. 创建群机器人（可选）**

1. 在钉钉群中，点击右上角**"群设置"**
2. 选择**"智能群助手"**
3. 点击**"添加机器人"** -> **"自定义"**
4. 设置机器人名称和头像
5. 选择安全设置（推荐选择**"加签"**）
6. 复制 Webhook 地址和加签密钥

---

### 第四步：飞书配置（可选）

如果您不需要接入飞书，可以点击"跳过"。

**需要准备的信息**

| 信息项 | 获取位置 |
|--------|---------|
| App ID | 飞书开发者平台 -> 应用凭证 |
| App Secret | 应用凭证 |
| Encrypt Key | 事件订阅设置（可选） |

**详细获取步骤**

**1. 创建飞书应用**

1. 登录 [飞书开发者平台](https://open.feishu.cn/app)
2. 点击**"创建企业自建应用"**
3. 填写应用信息：
   - 应用名称：**OpenClaw**
   - 应用描述：AI 智能助手

**2. 获取 App ID 和 App Secret**

1. 进入应用详情页的**"凭证与基础信息"**
2. 记录**"App ID"**（以 `cli_` 开头）
3. 点击**"App Secret"**旁边的**"查看"**，获取密钥

**3. 配置事件订阅（可选）**

1. 进入**"事件与回调"**页面
2. 点击**"启用事件"**
3. 配置加密方式：
   - 点击**"随机生成"**生成 Encrypt Key
   - 记录 Verification Token
4. 配置请求网址：
   - 将 OpenClaw 配置页面显示的地址填入
   - 格式：`http://您的服务器地址:8080/feishu/webhook`

**4. 发布应用**

1. 进入**"版本管理与发布"**
2. 点击**"创建版本"**
3. 填写版本信息并发布
4. 在飞书管理后台审核通过

---

### 第五步：完成配置

1. **查看配置摘要**
   - 确认 AI 模型配置
   - 确认已启用的 IM 渠道

2. **选择启动选项**
   - **开机自动启动**：系统启动时自动运行 OpenClaw
   - **启动后最小化**：启动后隐藏到系统托盘

3. **点击"启动服务"**
   - 等待服务启动完成
   - 显示"启动成功"即完成安装

---

## 日常使用

### 如何启动 OpenClaw

**Windows：**
- 方法 1：点击开始菜单中的"OpenClaw"
- 方法 2：双击桌面快捷方式（如果创建）
- 方法 3：按 `Win + R`，输入 `openclaw` 回车

**macOS：**
- 打开终端，输入 `openclaw` 回车
- 或在启动台中找到 OpenClaw

**Linux：**
```bash
# 启动服务
openclaw

# 或作为系统服务启动
sudo systemctl start openclaw
```

### 如何修改配置

**方法一：通过 Web 界面**

1. 打开浏览器访问 `http://localhost:8080`
2. 点击"设置"或"配置"选项
3. 修改需要更改的配置项
4. 点击"保存"

**方法二：直接编辑配置文件**

| 系统 | 配置文件路径 |
|-----|-------------|
| Windows | `C:\ProgramData\OpenClaw\config\config.yaml` |
| macOS | `/usr/local/etc/openclaw/config.yaml` |
| Linux | `/etc/openclaw/config.yaml` |

**注意**：修改配置文件后需要重启服务才能生效。

### 如何查看日志

**日志文件位置**

| 系统 | 日志路径 |
|-----|---------|
| Windows | `C:\ProgramData\OpenClaw\logs\openclaw.log` |
| macOS | `/var/log/openclaw/openclaw.log` |
| Linux | `/var/log/openclaw/openclaw.log` |

**实时查看日志**

```bash
# macOS / Linux
tail -f /var/log/openclaw/openclaw.log

# Windows（PowerShell）
Get-Content C:\ProgramData\OpenClaw\logs\openclaw.log -Wait
```

### 如何更新版本

**自动更新**

1. OpenClaw 会定期检查更新
2. 有新版本时会在系统托盘显示提示
3. 点击"立即更新"即可自动完成

**手动更新**

1. 下载最新版本的安装包
2. 运行安装器
3. 选择"升级现有安装"
4. 按提示完成更新

**命令行更新**

```bash
# 检查更新
openclaw-update check

# 执行更新
openclaw-update -yes

# 回滚到上一版本（如果更新后出现问题）
openclaw-update -rollback
```

---

## 常见问题

### 安装相关问题

**Q: 运行安装器时提示"权限不足"怎么办？**

A:
- **Windows**：右键点击安装器，选择**"以管理员身份运行"**
- **macOS/Linux**：在命令前加 `sudo`，如 `sudo ./OpenClaw-Installer`

**Q: Mac 提示"无法打开，因为无法验证开发者"？**

A: 打开**"系统设置"** -> **"隐私与安全性"**，在安全性部分点击**"仍要打开"**。或者在终端执行：
```bash
xattr -d com.apple.quarantine /path/to/installer
```

**Q: 浏览器没有自动打开怎么办？**

A:
1. 手动打开浏览器
2. 访问 `http://localhost:18080`
3. 如果页面无法访问，可能是端口被占用，尝试使用其他端口启动安装器：
   ```bash
   ./OpenClaw-Installer --port 18081
   ```

### 配置相关问题

**Q: 如何获取 OpenAI API Key？**

A:
1. 访问 [OpenAI Platform](https://platform.openai.com)
2. 注册/登录账号
3. 点击右上角头像 -> "View API keys"
4. 点击 "Create new secret key"
5. 复制生成的密钥（以 `sk-` 开头）

**Q: 企业微信的 Secret 在哪里找？**

A: 在企业微信管理后台的应用详情页，点击 Secret 旁边的**"查看"**，需要管理员扫码确认后才能显示。

**Q: 配置完成后如何测试是否成功？**

A:
1. 在配置完成页点击**"测试连接"**按钮
2. 或在对应 IM 平台中 @机器人 发送消息测试
3. 查看日志文件确认消息是否正常处理

### 使用相关问题

**Q: 如何修改已完成的配置？**

A:
- **方式一**：重新运行安装器，选择"修改配置"
- **方式二**：直接编辑配置文件（参见"如何修改配置"章节）

**Q: 服务启动失败怎么办？**

A: 检查以下几点：
1. 端口是否被其他程序占用
2. API Key 是否正确有效
3. 配置文件格式是否正确
4. 查看日志文件获取详细错误信息

**Q: 如何卸载 OpenClaw？**

A:
- **Windows**：打开"控制面板" -> "程序和功能"，找到 OpenClaw 卸载
- **macOS**：运行卸载脚本 `sudo /usr/local/bin/openclaw-uninstall`
- **Linux**：运行卸载脚本 `sudo /usr/local/bin/openclaw-uninstall`

**Q: 如何重置所有配置？**

A:
1. 停止 OpenClaw 服务
2. 删除配置文件目录
3. 重新运行安装器

---

## 故障排除

### 故障排查流程图

```
遇到问题
    |
    v
+-------------------+
| 1. 查看日志文件    |
|    定位错误信息    |
+-------------------+
    |
    v
+-------------------+
| 2. 对照下表查找    |
|    对应解决方案    |
+-------------------+
    |
    v
+-------------------+
| 3. 尝试解决方案    |
|    是否解决？      |
+-------------------+
    |           |
   是          否
    |           |
    v           v
  完成    +-------------------+
          | 4. 联系技术支持    |
          |    提供日志文件    |
          +-------------------+
```

### 各平台常见问题

#### Windows 常见问题

| 问题 | 原因 | 解决方案 |
|-----|------|---------|
| 安装器无法运行 | 缺少 .NET Runtime | 安装 [.NET 6.0 Runtime](https://dotnet.microsoft.com/download) |
| 服务无法启动 | 端口被占用 | 修改配置文件中的端口号 |
| 防火墙拦截 | Windows 防火墙阻止 | 允许 OpenClaw 通过防火墙 |
| 杀毒软件拦截 | 误报为病毒 | 将 OpenClaw 添加到白名单 |

#### macOS 常见问题

| 问题 | 原因 | 解决方案 |
|-----|------|---------|
| 无法打开应用 | Gatekeeper 拦截 | 系统设置 -> 隐私与安全性 -> 仍要打开 |
| 命令找不到 | 未添加到 PATH | 重新安装或手动添加 `/usr/local/bin` 到 PATH |
| 权限不足 | 需要 root 权限 | 使用 `sudo` 运行命令 |

#### Linux 常见问题

| 问题 | 原因 | 解决方案 |
|-----|------|---------|
| 依赖缺失 | 缺少必要的库 | 安装依赖：`sudo apt-get install -f` |
| 服务启动失败 | systemd 配置错误 | 检查服务状态：`sudo systemctl status openclaw` |
| 端口权限 | 低端口需要 root | 使用 1024 以上的端口号 |

### 配置验证失败

**API Key 验证失败**

| 错误信息 | 原因 | 解决方案 |
|----------|------|----------|
| Invalid API Key | API Key 错误或已失效 | 重新获取有效的 API Key |
| Insufficient quota | API 账户余额不足 | 检查 OpenAI 账户余额并充值 |
| Rate limit exceeded | 请求频率过高 | 稍后再试或升级账户 |
| Network error | 网络连接问题 | 检查网络连接，确认能访问 api.openai.com |

**企业微信连接测试失败**

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

**钉钉连接测试失败**

1. **检查 AppKey 和 AppSecret**
   - 确认来自同一应用
   - 确认应用已发布

2. **检查 IP 白名单**
   - 在钉钉开放平台，将服务器公网 IP 添加到白名单

3. **检查权限**
   - 确认应用具有必要的权限
   - 确认权限已申请并通过

**飞书连接测试失败**

1. **检查 AppID 格式**
   - 应以 `cli_` 开头

2. **检查应用发布状态**
   - 确认应用已创建版本并发布
   - 确认管理员已审核通过

3. **检查事件订阅配置**
   - 确认事件订阅地址正确
   - 确认服务器可被飞书访问

### 服务启动失败

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

### 日志文件位置速查

| 系统 | 日志路径 |
|-----|---------|
| Windows | `C:\ProgramData\OpenClaw\logs\openclaw.log` |
| macOS | `/var/log/openclaw/openclaw.log` |
| Linux | `/var/log/openclaw/openclaw.log` |

### 获取帮助的渠道

如果您无法自行解决问题，可以通过以下方式获取帮助：

1. **查看日志**：日志文件包含详细的错误信息
2. **开启调试模式**：使用 `--debug` 参数运行，获取更详细的输出
3. **提交 Issue**：在 GitHub 仓库提交问题，附上日志和错误信息

---

## 附录

### 配置文件示例

```yaml
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

### 命令行参数参考

| 参数 | 简写 | 说明 | 默认值 |
|------|------|------|--------|
| `--port` | `-p` | Web 配置服务端口 | `18080` |
| `--no-browser` | `-n` | 不自动打开浏览器 | `false` |
| `--debug` | `-d` | 启用调试日志 | `false` |
| `--config` | `-c` | 指定配置文件路径 | 自动检测 |
| `--help` | `-h` | 显示帮助信息 | - |
| `--version` | `-v` | 显示版本信息 | - |

### 版本历史

| 版本 | 发布日期 | 主要更新 |
|-----|---------|---------|
| v1.0.0 | 2026-02-28 | 初始版本，支持三平台安装 |

---

**文档版本**: 1.0.0
**最后更新**: 2026-02-28
**适用 OpenClaw 版本**: v1.0.0 及以上
