# OpenClaw 2.0 测试计划（当前基线）

## 目标

建立可复现、可执行的最小质量门禁，覆盖 Node 工作区与 Go updater。

## 测试分层

1. Node 工作区（`packages/*`）
- 命令：`pnpm test`
- 作用：Nuxt 类型生成 + TypeScript 类型检查

2. Go updater
- 命令：`cd updater && go test ./...`
- 命令：`cd updater && go test -race ./...`
- 作用：基础行为验证 + 并发竞态检查

3. 构建层
- 命令：`pnpm build:config`
- 命令：`pnpm --filter @openclaw/installer test`
- 作用：确保配置服务可构建、安装器类型与 Nuxt 入口正常

4. 手动关键路径
- 安装器向导全流程
- 配置服务 API Key 与适配器保存
- 服务状态接口读取

## CI 门禁（build.yml）

- `test-node`
- `test-go`
- `build-installer`（依赖 test-node + test-go）
- `build-config-server`（依赖 test-node + test-go）

## 后续增强（待排期）

- config-server API 单元测试
- installer IPC 集成测试
- updater 下载/回滚链路集成测试
