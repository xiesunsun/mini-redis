# mini-redis

用 Go 实现的 Redis 服务器，支持三种数据类型（String/List/Hash）、16 条核心命令和 AOF 持久化，能被真实的 redis-cli 连接使用。

## 文档目录

详细文档在 docs/ 目录下，按需查阅：
- 架构设计与分层规则 → docs/architecture.md
- RESP 协议格式      → docs/resp-protocol.md
- 支持的命令列表     → docs/commands.md
- 测试规范 → docs/testing.md
- 开发进度跟踪 → docs/progress.json

## 架构分层

项目分为六层，依赖方向严格单向，禁止跨层引用：

```
types → store → expiry      → command → network → cmd/server
              → persistence ↗
```

各层职责：
- types：共享数据结构定义，不依赖任何层
- store：内存数据存储，只依赖 types
- expiry：过期键清理，只依赖 store
- persistence：AOF 持久化，只依赖 store
- command：命令解析与执行，依赖 store / expiry / persistence
- network：TCP 连接与 RESP 协议，只依赖 command
- cmd/server：程序入口，负责组装所有层启动服务器

## 禁止规则

以下依赖由 scripts/check_deps.sh 在 CI 中自动检测，违反则构建失败，PR 无法合并：

- types 禁止引用任何内部包
- store 禁止引用 command / network
- expiry 禁止引用 command / network
- persistence 禁止引用 command / network
- command 禁止引用 network

## 开发工作流

0. 收到需求，先读 docs/progress.json 了解当前项目进度
1. 确认当前在 main 分支，拉取最新代码：
   `git checkout main && git pull origin main`
2. 创建新分支：
   - 新功能：`git checkout -b feat/XXX`
   - 修复错误：`git checkout -b fix/XXX`

3. 实现具体功能与对应测试函数编写，每个文件第一行必须是正确的 package 声明，文件只能放在对应层的目录下
   - 测试规范参考 docs/testing.md
   - 涉及代码行为变化时，同时编写或更新对应的测试
   - 仅修改文档、注释，不需要测试

测试规范参考 docs/testing.md
4. 执行以下检查，全部通过才能继续，否则携带报错信息回到 3：
   - `go build ./...`              → 编译必须通过
   - `bash scripts/check_deps.sh` → 依赖层级必须合法
   - `go test ./...`              → 所有测试必须通过
5. 按格式提交代码并推送，等待人类合并 PR：
   "PR 已创建：[链接]，等待审核。""
   - commit message 格式：`类型: 描述`
   - 例如：`feat: 实现 SET/GET 命令`、`fix: 修复 RESP 解析错误`
6. 若 PR 合并通过，转到 7；否则携带人类反馈回到 3
7. 更新 docs/progress.json 中对应任务状态为 done，切回 main，拉取最新代码，删除功能分支，，输出：任务 XXX 已完成，docs/progress.json 已更新。：
   `git checkout main && git pull origin main && git branch -d XXX`

## 遇到不确定的情况

- 不要大范围猜测性修改代码
- 先在 PR 描述里写清楚方案，等人类确认后再动手
- 命令行为不确定时，查 docs/commands.md

## 完成定义

一个功能算完成，必须同时满足：
- `go build ./...` 通过
- `bash scripts/check_deps.sh` 通过
- `go test ./...` 通过
- 用 `redis-cli -p 6379` 连接本地服务器，相关命令行为与官方 Redis 完全一致