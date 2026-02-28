# 发布检查清单

## 版本号格式
遵循语义化版本控制 (SemVer)：
- `v1.0.0` - 主版本（重大变更）
- `v1.1.0` - 次版本（新功能）
- `v1.1.1` - 补丁版本（Bug修复）

## 发布前检查

### 1. 代码检查
- [ ] 所有测试通过 (`make test`)
- [ ] 代码格式化 (`make fmt`)
- [ ] 静态检查通过 (`make vet`)
- [ ] 无未提交的更改 (`git status`)

### 2. 构建检查
- [ ] 所有平台构建成功 (`make build-all`)
- [ ] 二进制文件大小正常 (~6MB)
- [ ] U盘部署包生成成功 (`make usb-deploy`)

### 3. 功能检查
- [ ] Web界面正常 (`http://localhost:18080`)
- [ ] 浏览器自动打开正常
- [ ] 配置保存正常
- [ ] 更新程序工作正常

### 4. 文档检查
- [ ] README.md 已更新
- [ ] CHANGELOG.md 已更新
- [ ] 版本号一致

## 发布步骤

### 1. 更新版本号
```bash
# 更新 main.go 中的版本
# 更新 README.md 中的版本
# 更新 CHANGELOG.md
```

### 2. 提交更改
```bash
git add -A
git commit -m "Release v1.0.0"
git tag v1.0.0
git push origin main --tags
```

### 3. CI/CD 自动发布
GitHub Actions 会自动：
- 运行测试
- 构建所有平台
- 创建 GitHub Release
- 上传构建产物

### 4. 手动验证
- [ ] 下载各平台二进制
- [ ] 在目标系统测试安装
- [ ] 验证配置流程

## 发布后

### 1. 通知
- [ ] 更新项目首页
- [ ] 通知用户（Discord/微信群）
- [ ] 更新文档网站

### 2. 监控
- [ ] 监控 Issue 反馈
- [ ] 收集用户反馈
- [ ] 准备补丁版本（如需要）

## 回滚流程

如发布出现严重问题：

1. 在 GitHub Release 标记为预发布
2. 通知用户暂停更新
3. 修复问题，发布补丁版本
4. 如需要，删除有问题的 Release
