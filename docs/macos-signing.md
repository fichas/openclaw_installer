# macOS 应用签名和公证指南

本文档介绍如何为 OpenClaw macOS 应用进行代码签名和公证。

## 概述

macOS 10.15+ 要求所有从互联网下载的应用必须经过 Apple 公证才能在默认 Gatekeeper 设置下运行。签名和公证流程包括：

1. **代码签名** - 使用 Apple Developer 证书对应用进行数字签名
2. **公证** - 将应用提交给 Apple 进行安全扫描
3. **票证附加** - 将公证票证附加到应用，允许离线验证

## 前置要求

### 1. Apple Developer 账号

- 有效的 Apple Developer Program 会员资格（$99/年）
- 团队管理员或开发者权限

### 2. 证书

需要以下证书（可在 Apple Developer Portal 下载）：

- **Developer ID Application** - 用于签名应用
- **Developer ID Installer** - 用于签名安装包

### 3. 工具

- macOS 10.15 或更高版本
- Xcode Command Line Tools
- 有效的 Apple ID（启用双重认证）

## 安装证书

### 从 Apple Developer Portal 下载

1. 访问 [Apple Developer Portal](https://developer.apple.com/account/resources/certificates/list)
2. 下载 Developer ID Application 证书
3. 双击下载的 `.cer` 文件安装到钥匙串

### 验证证书安装

```bash
security find-identity -v -p codesigning
```

应看到类似输出：
```
1) ABCDEF1234567890 "Developer ID Application: Your Name (TEAMID)"
   1 identities found
```

## 配置 App 专用密码

1. 访问 [appleid.apple.com](https://appleid.apple.com)
2. 登录并进入"安全"部分
3. 点击"生成密码..."
4. 标签填写 "OpenClaw Notarization"
5. 复制生成的密码（格式：xxxx-xxxx-xxxx-xxxx）

## 签名流程

### 快速签名

```bash
./scripts/sign-macos.sh \
    -c "Developer ID Application: Your Name (TEAMID)" \
    -a dist/OpenClaw.app
```

### 完整流程（签名 + 公证）

```bash
./scripts/sign-macos.sh \
    -c "Developer ID Application: Your Name (TEAMID)" \
    -u "your.email@example.com" \
    -p "xxxx-xxxx-xxxx-xxxx" \
    -t "TEAMID" \
    -a dist/OpenClaw.app \
    -n \
    -s
```

### 参数说明

| 参数 | 说明 | 必需 |
|------|------|------|
| `-c, --cert` | 证书名称 | 是 |
| `-a, --app-path` | 应用路径 | 否（默认: dist/OpenClaw.app） |
| `-u, --username` | Apple ID 邮箱 | 仅公证时需要 |
| `-p, --password` | App 专用密码 | 仅公证时需要 |
| `-t, --team-id` | 团队 ID | 建议提供 |
| `-n, --notarize` | 执行公证 | 否 |
| `-s, --staple` | 附加票证 | 否 |
| `-v, --verbose` | 详细输出 | 否 |

## Entitlements 配置

Entitlements 文件定义应用所需的权限。默认配置位于 `build/entitlements.plist`。

### 关键权限说明

```xml
<!-- 沙盒 - 必须启用 -->
<key>com.apple.security.app-sandbox</key>
<true/>

<!-- 网络访问 -->
<key>com.apple.security.network.client</key>
<true/>
<key>com.apple.security.network.server</key>
<true/>

<!-- 执行外部命令（更新器必需） -->
<key>com.apple.security.cs.allow-jit</key>
<true/>
<key>com.apple.security.cs.allow-unsigned-executable-memory</key>
<true/>
<key>com.apple.security.cs.disable-library-validation</key>
<true/>
```

### 自定义 Entitlements

创建自定义 `custom-entitlements.plist`：

```bash
./scripts/sign-macos.sh \
    -c "Developer ID Application: Your Name" \
    -e custom-entitlements.plist \
    -a dist/OpenClaw.app
```

## 验证签名

### 基础验证

```bash
codesign --verify --deep --strict dist/OpenClaw.app
echo $?  # 0 表示成功
```

### 详细验证

```bash
codesign -d --deep -vv dist/OpenClaw.app
```

### Gatekeeper 测试

```bash
# 检查 Gatekeeper 兼容性
spctl --assess --type exec dist/OpenClaw.app

# 查看评估结果
spctl -a -vv dist/OpenClaw.app
```

## 公证流程详解

### 1. 创建提交包

应用需要打包为 DMG 或 ZIP 进行公证：

```bash
# 创建 DMG
hdiutil create -size 100m -volname "OpenClaw" \
    -srcfolder dist/OpenClaw.app -ov -format UDZO \
    dist/OpenClaw.dmg
```

### 2. 提交公证

```bash
xcrun notarytool submit dist/OpenClaw.dmg \
    --apple-id "your.email@example.com" \
    --password "xxxx-xxxx-xxxx-xxxx" \
    --team-id "TEAMID" \
    --wait
```

### 3. 查看公证状态

```bash
# 使用 submission ID 查询状态
xcrun notarytool history \
    --apple-id "your.email@example.com" \
    --password "xxxx-xxxx-xxxx-xxxx"

# 获取详细日志
xcrun notarytool log SUBMISSION_ID \
    --apple-id "your.email@example.com" \
    --password "xxxx-xxxx-xxxx-xxxx"
```

### 4. 附加票证

```bash
# 附加到 DMG
xcrun stapler staple dist/OpenClaw.dmg

# 附加到应用（需要先解压）
xcrun stapler staple dist/OpenClaw.app
```

### 5. 验证票证

```bash
xcrun stapler validate dist/OpenClaw.app
```

## 故障排除

### 证书问题

**问题**: `errSecInternalComponent`

**解决**:
```bash
# 重启 securityd
sudo killall -9 securityd

# 或解锁钥匙串
security unlock-keychain ~/Library/Keychains/login.keychain-db
```

### 签名验证失败

**问题**: `code object is not signed at all`

**解决**:
```bash
# 强制重新签名所有内容
codesign --force --deep --sign "Developer ID Application: Your Name" \
    --entitlements build/entitlements.plist dist/OpenClaw.app
```

### 公证失败

**问题**: `The signature of the binary is invalid`

**解决**:
1. 确保证书未过期
2. 检查 entitlements 格式正确
3. 使用 `--options=runtime` 进行签名

### Gatekeeper 阻止

**问题**: 应用被 Gatekeeper 阻止

**解决**:
```bash
# 临时允许（仅开发测试）
sudo spctl --master-disable

# 或移除隔离属性
xattr -rd com.apple.quarantine dist/OpenClaw.app
```

## CI/CD 集成

### GitHub Actions 示例

```yaml
name: Sign and Notarize

on:
  release:
    types: [created]

jobs:
  sign:
    runs-on: macos-latest
    steps:
      - uses: actions/checkout@v3

      - name: Import certificates
        uses: apple-actions/import-codesign-certs@v1
        with:
          p12-file-base64: ${{ secrets.CERTIFICATES_P12 }}
          p12-password: ${{ secrets.CERTIFICATES_P12_PASSWORD }}

      - name: Sign and Notarize
        env:
          APPLE_ID: ${{ secrets.APPLE_ID }}
          APPLE_PASSWORD: ${{ secrets.APPLE_PASSWORD }}
          TEAM_ID: ${{ secrets.TEAM_ID }}
        run: |
          ./scripts/sign-macos.sh \
            -c "Developer ID Application: OpenClaw Inc ($TEAM_ID)" \
            -u "$APPLE_ID" \
            -p "$APPLE_PASSWORD" \
            -t "$TEAM_ID" \
            -n -s
```

### 必需 Secrets

- `CERTIFICATES_P12`: Base64 编码的 .p12 证书文件
- `CERTIFICATES_P12_PASSWORD`: .p12 文件密码
- `APPLE_ID`: Apple ID 邮箱
- `APPLE_PASSWORD`: App 专用密码
- `TEAM_ID`: Apple Developer Team ID

## 最佳实践

1. **使用 Developer ID 证书** - 不要仅使用 Mac Development 证书
2. **启用 Hardened Runtime** - 使用 `--options=runtime`
3. **包含时间戳** - 使用 `--timestamp`
4. **测试干净系统** - 在全新 macOS 安装上测试
5. **保留构建日志** - 保存所有签名和公证日志
6. **定期更新证书** - 证书有效期通常为 1 年

## 参考资源

- [Apple Code Signing Guide](https://developer.apple.com/library/archive/documentation/Security/Conceptual/CodeSigningGuide/)
- [Notarizing macOS Software](https://developer.apple.com/documentation/security/notarizing_macos_software_before_distribution)
- [Hardened Runtime](https://developer.apple.com/documentation/security/hardened_runtime)
- [Entitlements Key Reference](https://developer.apple.com/library/archive/documentation/Miscellaneous/Reference/EntitlementKeyReference/Chapters/AboutEntitlements.html)
