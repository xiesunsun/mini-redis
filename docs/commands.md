# 支持的命令列表

## string类型相关命令
SET key value         → 存入，返回 OK
GET key               → 取出，不存在返回 nil，类型不是string返回WRONGTYPE Operation against a key  holding wrong kind of value 
DEL key               → 删除，返回删除数量（0或1）
EXPIRE key seconds    → 设置过期时间，返回 0或1
TTL key               → 查询剩余时间，返回秒数/-1/-2
> 暂不支持DEL 同时删除多个Key

## list类型相关

LPUSH key value     → 从左插入，返回列表长度
RPUSH key value     → 从右插入，返回列表长度
LRANGE key start stop → 取范围元素，不存在返回空列表
LLEN key            → 返回列表长度，不存在返回 0
LPOP key            → 弹出最左元素，不存在返回 nil
RPOP key            → 弹出最右元素，不存在返回 nil
> 暂不支持LPUSH/RPUSH 一次插入多个值

## Hash类型相关命令

HSET key field value   → 设置 field，返回新增数量（0或1）
HGET key field         → 取出 field，不存在返回 nil
HDEL key field         → 删除 field，返回 0或1
HGETALL key            → 取出所有 field 和 value
HEXISTS key field      → field 存在返回 1，不存在返回 0
> 暂不支持：SET 一次设置多个field
