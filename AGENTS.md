# mini-redis
一个用Go 实现的Redis 服务器，目标是支持核心命令、三种数据类型和AOF持久化。能够被真实的redis-ci连接和使用。

## 文档目录
- 架构设计与分层规则 -> docs/architechture.md
- RESP 协议格式 -> docs/resp-protocol.md
- 支持的命令列表 -> docs/commands.md

## 架构分层
项目分为六层，依赖方向严格单向，禁止跨层引用
```
types -> store -> expiry -> command->network-> cmd/server
              
               -> persistence  ↗
```
各层职责
- types：所有共享数据结构定义、不依赖于任何层
- stores：内存数据存储，只依赖types
- expiry：过期键清理。只依赖stores
- persistency：AOF持久化，只依赖store
- command：命令解析与执行，依赖store/expiry/persistency
- network：TCP 连接与RESP 协议，只依赖command
- cmd/server：程序入口，负责组装所有层启动服务器

## 禁止规则
以下依赖在CI中会被check-deps.sh自动检测，违反则构建失败：
- store禁止依赖command / network
- expiry 禁止引用 command / network
- persistency 禁止引用 command / network
- command 禁止引用 network
- types 禁止引用任何内部包

## 开发规范
**新增文件**
- 每个文件第一行必须是正确的package 声明
- 文件只能放在对应层的目录下，不能跨目录放置
**新增依赖**
- 引用新的内部包之前，先确认依赖方向是否合法
- 禁止引用与当前层无关的外部库
**提交代码**
- 禁止直接push到main，所有改动必须要通过PR
- PR合并前必须要通过CI(编译检查+依赖层级检查)
- commit message 格式： `类型：描述`
    - 例如：`feat:实现SET/GET命令`、`fix：修复RESP解析错误` 
**验证标志**
- 每条命令必须与真实的Redis一致
- 用redis-cli 连接本地服务器进行验证，命令行为不一致视为bug



