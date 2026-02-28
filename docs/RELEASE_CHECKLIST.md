# 发布检查清单（OpenClaw 2.0）

## 1. 代码与测试

- [ ] `pnpm test` 通过
- [ ] `cd updater && go test ./...` 通过
- [ ] `cd updater && go test -race ./...` 通过
- [ ] `git status` 无非预期变更

## 2. 构建验证

- [ ] `pnpm build:config` 通过
- [ ] `pnpm --filter @openclaw/installer test` 通过
- [ ] 产物目录生成正常（`packages/config-server/.output`、`packages/installer/release`）

## 3. 关键流程验收

- [ ] 安装器流程可走通：欢迎 -> 模式 -> 适配器 -> 确认 -> 进度 -> 完成
- [ ] 配置服务可保存 API Key 与适配器配置
- [ ] 服务状态接口可正常返回（`/api/service`）

## 4. 文档一致性

- [ ] README 与当前架构一致（Electron + Nuxt + Go updater）
- [ ] `docs/QUICK_START.md`、`docs/CONFIG_GUIDE.md`、`docs/FAQ.md` 内容可用
- [ ] 变更点写入发布说明

## 5. 发布步骤

1. 提交代码并打标签
2. 推送到 `main/master` 或 `v*` tag
3. 等待 GitHub Actions `Build & Release` 完成
4. 抽检下载产物并做最小安装验证
