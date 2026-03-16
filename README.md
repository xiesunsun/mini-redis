# mini-redis

一个用 Go 实现的 Redis 服务器，支持真实 `redis-cli` 连接，具备基础数据结构能力、RESP 协议解析和 AOF 持久化。

## 功能特性

- 三种数据类型：`String`、`List`、`Hash`
- 16 条核心命令：
  - String: `SET` `GET` `DEL` `EXPIRE` `TTL`
  - List: `LPUSH` `RPUSH` `LRANGE` `LLEN` `LPOP` `RPOP`
  - Hash: `HSET` `HGET` `HDEL` `HGETALL` `HEXISTS`
- RESP 协议支持（请求/响应）
- AOF 持久化（默认文件：`appendonly.aof`）
- 多客户端并发连接
- 严格参数校验（参数数量不匹配时返回 `wrong number of arguments`）

## 快速开始

### 1. 环境要求

- Go `1.25.0`（见 `go.mod`）
- `redis-cli`（用于手工验证）

### 2. 启动服务

在项目根目录执行：

```bash
go run ./cmd/server
```

默认监听地址：`127.0.0.1:6379`（服务端绑定 `:6379`）。

### 3. 使用 redis-cli 验证

```bash
redis-cli -p 6379
```

示例命令：

```redis
SET name alice
GET name
LPUSH nums 1
RPUSH nums 2
LRANGE nums 0 -1
HSET user:1 name alice
HGETALL user:1
EXPIRE name 10
TTL name
```

## AOF 持久化

- 启动时会自动打开/创建 `appendonly.aof`
- 每次写命令成功执行后追加写入 AOF
- 服务重启时会回放 AOF 恢复数据

## 项目结构与分层

```text
types → store → expiry      → command → network → cmd/server
              → persistence ↗
```

- `internal/types`: 通用结构定义
- `internal/store`: 内存存储与数据结构操作
- `internal/expiry`: 过期键清理（惰性删除 + 定期清理）
- `internal/persistence`: AOF 写入与回放
- `internal/command`: 命令处理与路由
- `internal/network`: TCP 服务与 RESP 解析/序列化
- `cmd/server`: 程序入口与组件组装

## 开发与验证

```bash
go build ./...
bash scripts/check_deps.sh
go test ./...
```

说明：

- `scripts/check_deps.sh` 会检查分层依赖是否合法，并检查实现文件是否有对应测试文件（`types` 层除外）
- `go test ./...` 包含单元测试与网络层/端到端测试

## Release 发布（GitHub Actions）

本项目通过 `.github/workflows/release.yml` 自动构建并发布 Release。

### 1. 准备发布版本号

示例：`v1.0.0`。

### 2. 在远端 main 对应提交打 tag 并推送

```bash
git fetch origin
git tag -d v1.0.0 2>/dev/null || true
git tag -a v1.0.0 origin/main -m "release v1.0.0"
git push origin v1.0.0
```

说明：
- `origin/main` 可避免本地分支落后导致 tag 打到错误提交
- 推送 tag 后会自动触发 `Release` workflow

### 3. 查看构建与发布结果

可在 GitHub 仓库的 `Actions` 页面查看 `Release` 工作流状态，成功后会自动创建同名 Release 并上传附件。

### 4. 产物命名规则

- macOS Apple Silicon: `mini-redis_darwin_arm64`
- macOS Intel: `mini-redis_darwin_amd64`
- Linux AMD64: `mini-redis_linux_amd64`
- Linux ARM64: `mini-redis_linux_arm64`
- Windows AMD64: `mini-redis_windows_amd64.exe`
- Windows ARM64: `mini-redis_windows_arm64.exe`
- 校验文件: `SHA256SUMS`

### 5. 重新发布同一版本（可选）

如果同版本 tag/release 已存在，需要先删除远端 Release 与 tag，再重新执行第 2 步。

## 当前限制

- `DEL` 暂不支持一次删除多个 key
- `LPUSH` / `RPUSH` 暂不支持一次插入多个 value
- `HSET` 暂不支持一次设置多个 field/value

## 文档

- 架构设计：`docs/architecture.md`
- RESP 协议：`docs/resp-protocol.md`
- 命令说明：`docs/commands.md`
- 测试规范：`docs/testing.md`
- 任务进度：`docs/progress.json`
