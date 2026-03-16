# 测试规范

## 测试文件位置

单元测试与被测文件放在同一目录下，文件名为被测文件名加 `_test.go`：
- `internal/store/store.go` → `internal/store/store_test.go`

端到端测试放在 `tests/` 目录下。

## 测试粒度

以行为为单位，每个行为对应一个独立的测试函数。
行为的定义来自 docs/commands.md 里的规范，规范里有几种情况，测试里就有几个函数。

测试函数命名格式：`Test被测功能_场景描述`
例如：`TestGet_ExistingKey`、`TestGet_NonExistentKey`、`TestGet_WrongType`

每个功能至少覆盖：
- 正常路径
- key 不存在
- 类型不匹配（适用于 store/command 层）
- 边界情况（空值、负数索引等，按实际情况判断）

## 各层测试要求

**types 层**
无需测试。

**store 层**
每个实现文件对应一个 `_test.go`，放在同一目录下。
行为定义参考 docs/commands.md。

**expiry 层**
`internal/expiry/cleaner_test.go`。
过期时间使用极短时间（1ms）模拟，不使用真实秒数。

**persistence 层**
`internal/persistence/aof_test.go`。
使用临时文件，测试结束后清理。

**command 层**
`internal/command/handlers_test.go`。
初始化真实的 store 实例，不使用 mock。

**network 层**
`tests/` 下的端到端测试。
建立真实 TCP 连接，发送 RESP 格式命令，验证响应与官方 Redis 一致。

## 运行测试
```bash
go test ./...
```

成功时无输出，失败时携带完整错误信息回到实现步骤。