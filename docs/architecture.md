# 架构设计

## 项目目标

实现一个能被真实 redis-cli 连接的 Redis 服务器，支持：
- 三种数据类型：String、List、Hash
- 核心命令：SET/GET/DEL/EXPIRE/TTL/LPUSH/RPUSH/LRANGE/HSET/HGET
- AOF 持久化：重启后数据不丢失
- 多客户端并发连接

## 整体架构
```
请求流向：
redis-cli → network → command → store
                             → expiry
                             → persistence
```

数据流向：
- 入：redis-cli 发送 RESP 格式字节流 → network 层解析 → command 层执行
- 出：command 层返回结果 → network 层序列化为 RESP 格式 → redis-cli 收到响应


## 各层详细说明

### types
- 职责：定义所有层共享的数据结构
- 不依赖任何层
- 核心结构：
  - `Command`：表示一条解析后的命令，包含命令名和参数
  - `Value`：表示存储的值，包含数据类型、实际数据、过期时间

### store
- 职责：内存数据存储，管理所有 key-value 数据
- 只依赖 types
- 核心数据结构：`map[string]*types.Value`
- 并发安全：使用 `sync.RWMutex` 保护读写操作

### expiry
- 职责：管理 key 的过期清理
- 只依赖 store
- 两种清理策略：
  - 惰性删除：访问 key 时检查是否过期，过期则删除后返回 nil
  - 定期删除：每隔 100ms 扫描一批 key，清理已过期的

### persistence
- 职责：AOF 日志的写入与恢复
- 只依赖 store
- 写入：每条命令执行后，把命令原文追加到 aof 文件
- 恢复：服务器启动时，重新执行 aof 文件里的所有命令

### command
- 职责：解析命令、路由到对应处理函数、返回结果
- 依赖 store / expiry / persistence
- 每条命令对应一个 handler 函数

### network
- 职责：TCP 连接管理、RESP 协议解析与序列化
- 只依赖 command
- 每个客户端连接开一个 goroutine 处理

### cmd/server
- 职责：程序入口，初始化所有层，启动服务器
- 可以依赖所有层

## 依赖规则

### 合法依赖方向
```
types → store → expiry     → command → network → cmd/server
              → persistence ↗
```

箭头含义：被依赖 → 依赖方（右边可以引用左边，左边不能引用右边）

### 禁止规则

| 层 | 禁止引用 |
|---|---|
| types | 任何内部包 |
| store | command、network |
| expiry | command、network |
| persistence | command、network |
| command | network |

### 违反后果

`scripts/check_deps.sh` 在 CI 中自动检测，违反则构建失败，PR 无法合并。

## 架构决策记录

**为什么选 AOF 而不是 RDB**
AOF 记录的是命令，逻辑直观，适合教学目的；RDB 是二进制快照，实现复杂度更高。

**为什么 expiry 独立成一层而不是放在 store 里**
过期清理有自己的定时逻辑，独立成层后 store 职责更单一，也更容易测试。

**为什么用 sync.RWMutex 而不是 channel**
store 是纯数据操作，读多写少，RWMutex 在这个场景下更直接，性能也更好。