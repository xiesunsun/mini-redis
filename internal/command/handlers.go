package command

import (
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/xiesunsun/mini-redis/internal/expiry"
	"github.com/xiesunsun/mini-redis/internal/persistence"
	"github.com/xiesunsun/mini-redis/internal/store"
	"github.com/xiesunsun/mini-redis/internal/types"
)

// Context carries the dependencies needed by all command handlers.
type Context struct {
	Store *store.Store
	AOF   *persistence.AOF
}

// HandlerFunc is the function type for all command handlers.
type HandlerFunc func(cmd types.Command, ctx *Context) string

// writeAOF appends cmd to the AOF log. No-op when ctx.AOF is nil.
func writeAOF(ctx *Context, cmd types.Command) {
	if ctx.AOF != nil {
		_ = ctx.AOF.WriteCommand(cmd)
	}
}

// RESP encoding helpers.
func respOK() string            { return "+OK\r\n" }
func respNil() string           { return "$-1\r\n" }
func respInt(n int64) string    { return fmt.Sprintf(":%d\r\n", n) }
func respErr(msg string) string { return fmt.Sprintf("-ERR %s\r\n", msg) }
func respBulk(s string) string  { return fmt.Sprintf("$%d\r\n%s\r\n", len(s), s) }
func respWrongType() string {
	return fmt.Sprintf("-%s\r\n", store.ErrWrongType.Error())
}

func respArray(items []string) string {
	if len(items) == 0 {
		return "*0\r\n"
	}
	result := fmt.Sprintf("*%d\r\n", len(items))
	for _, item := range items {
		result += respBulk(item)
	}
	return result
}

// isWrongType reports whether err is a type-mismatch error.
func isWrongType(err error) bool {
	return errors.Is(err, store.ErrWrongType)
}

// --- String commands ---

// HandleSet handles: SET key value → +OK
func HandleSet(cmd types.Command, ctx *Context) string {
	if len(cmd.Args) != 2 {
		return respErr("wrong number of arguments for 'SET' command")
	}
	ctx.Store.SetString(cmd.Args[0], cmd.Args[1])
	writeAOF(ctx, cmd)
	return respOK()
}

// HandleGet handles: GET key → bulk string or nil
func HandleGet(cmd types.Command, ctx *Context) string {
	if len(cmd.Args) != 1 {
		return respErr("wrong number of arguments for 'GET' command")
	}
	v := expiry.GetOrExpire(ctx.Store, cmd.Args[0])
	if v == nil {
		return respNil()
	}
	if v.DataType != types.StringType {
		return respWrongType()
	}
	return respBulk(v.Data.(string))
}

// HandleDel handles: DEL key → integer (0 or 1)
func HandleDel(cmd types.Command, ctx *Context) string {
	if len(cmd.Args) != 1 {
		return respErr("wrong number of arguments for 'DEL' command")
	}
	key := cmd.Args[0]
	v := expiry.GetOrExpire(ctx.Store, key)
	if v == nil {
		return respInt(0)
	}
	ctx.Store.Delete(key)
	writeAOF(ctx, cmd)
	return respInt(1)
}

// HandleExpire handles: EXPIRE key seconds → 0 or 1
func HandleExpire(cmd types.Command, ctx *Context) string {
	if len(cmd.Args) != 2 {
		return respErr("wrong number of arguments for 'EXPIRE' command")
	}
	secs, err := strconv.ParseInt(cmd.Args[1], 10, 64)
	if err != nil {
		return respErr("value is not an integer or out of range")
	}
	v := ctx.Store.Get(cmd.Args[0])
	if v == nil {
		return respInt(0)
	}
	v.Expiry = time.Now().Add(time.Duration(secs) * time.Second)
	writeAOF(ctx, cmd)
	return respInt(1)
}

// HandleTTL handles: TTL key → seconds / -1 (no expiry) / -2 (not found)
func HandleTTL(cmd types.Command, ctx *Context) string {
	if len(cmd.Args) != 1 {
		return respErr("wrong number of arguments for 'TTL' command")
	}
	v := ctx.Store.Get(cmd.Args[0])
	if v == nil {
		return respInt(-2)
	}
	if !v.Expiry.IsZero() && time.Now().After(v.Expiry) {
		ctx.Store.Delete(cmd.Args[0])
		return respInt(-2)
	}
	if v.Expiry.IsZero() {
		return respInt(-1)
	}
	ttl := int64(time.Until(v.Expiry).Seconds())
	if ttl < 0 {
		ttl = 0
	}
	return respInt(ttl)
}

// --- List commands ---

// HandleLPush handles: LPUSH key value → list length
func HandleLPush(cmd types.Command, ctx *Context) string {
	if len(cmd.Args) != 2 {
		return respErr("wrong number of arguments for 'LPUSH' command")
	}
	n, err := ctx.Store.LPush(cmd.Args[0], cmd.Args[1])
	if err != nil {
		if isWrongType(err) {
			return respWrongType()
		}
		return respErr(err.Error())
	}
	writeAOF(ctx, cmd)
	return respInt(int64(n))
}

// HandleRPush handles: RPUSH key value → list length
func HandleRPush(cmd types.Command, ctx *Context) string {
	if len(cmd.Args) != 2 {
		return respErr("wrong number of arguments for 'RPUSH' command")
	}
	n, err := ctx.Store.RPush(cmd.Args[0], cmd.Args[1])
	if err != nil {
		if isWrongType(err) {
			return respWrongType()
		}
		return respErr(err.Error())
	}
	writeAOF(ctx, cmd)
	return respInt(int64(n))
}

// HandleLRange handles: LRANGE key start stop → array
func HandleLRange(cmd types.Command, ctx *Context) string {
	if len(cmd.Args) != 3 {
		return respErr("wrong number of arguments for 'LRANGE' command")
	}
	start, err := strconv.Atoi(cmd.Args[1])
	if err != nil {
		return respErr("value is not an integer or out of range")
	}
	stop, err := strconv.Atoi(cmd.Args[2])
	if err != nil {
		return respErr("value is not an integer or out of range")
	}
	items, err := ctx.Store.LRange(cmd.Args[0], start, stop)
	if err != nil {
		if isWrongType(err) {
			return respWrongType()
		}
		return respErr(err.Error())
	}
	return respArray(items)
}

// HandleLLen handles: LLEN key → list length (0 if key does not exist)
func HandleLLen(cmd types.Command, ctx *Context) string {
	if len(cmd.Args) != 1 {
		return respErr("wrong number of arguments for 'LLEN' command")
	}
	n, err := ctx.Store.LLen(cmd.Args[0])
	if err != nil {
		if isWrongType(err) {
			return respWrongType()
		}
		return respErr(err.Error())
	}
	return respInt(int64(n))
}

// HandleLPop handles: LPOP key → bulk string or nil
func HandleLPop(cmd types.Command, ctx *Context) string {
	if len(cmd.Args) != 1 {
		return respErr("wrong number of arguments for 'LPOP' command")
	}
	val, err := ctx.Store.LPop(cmd.Args[0])
	if err != nil {
		if errors.Is(err, store.ErrKeyNotFound) {
			return respNil()
		}
		if isWrongType(err) {
			return respWrongType()
		}
		return respErr(err.Error())
	}
	writeAOF(ctx, cmd)
	return respBulk(val)
}

// HandleRPop handles: RPOP key → bulk string or nil
func HandleRPop(cmd types.Command, ctx *Context) string {
	if len(cmd.Args) != 1 {
		return respErr("wrong number of arguments for 'RPOP' command")
	}
	val, err := ctx.Store.RPop(cmd.Args[0])
	if err != nil {
		if errors.Is(err, store.ErrKeyNotFound) {
			return respNil()
		}
		if isWrongType(err) {
			return respWrongType()
		}
		return respErr(err.Error())
	}
	writeAOF(ctx, cmd)
	return respBulk(val)
}

// --- Hash commands ---

// HandleHSet handles: HSET key field value → 1 (new field) or 0 (updated)
func HandleHSet(cmd types.Command, ctx *Context) string {
	if len(cmd.Args) != 3 {
		return respErr("wrong number of arguments for 'HSET' command")
	}
	n, err := ctx.Store.HSet(cmd.Args[0], cmd.Args[1], cmd.Args[2])
	if err != nil {
		if isWrongType(err) {
			return respWrongType()
		}
		return respErr(err.Error())
	}
	writeAOF(ctx, cmd)
	return respInt(int64(n))
}

// HandleHGet handles: HGET key field → bulk string or nil
func HandleHGet(cmd types.Command, ctx *Context) string {
	if len(cmd.Args) != 2 {
		return respErr("wrong number of arguments for 'HGET' command")
	}
	val, err := ctx.Store.HGet(cmd.Args[0], cmd.Args[1])
	if err != nil {
		if errors.Is(err, store.ErrKeyNotFound) {
			return respNil()
		}
		if isWrongType(err) {
			return respWrongType()
		}
		return respErr(err.Error())
	}
	return respBulk(val)
}

// HandleHDel handles: HDEL key field → 0 or 1
func HandleHDel(cmd types.Command, ctx *Context) string {
	if len(cmd.Args) != 2 {
		return respErr("wrong number of arguments for 'HDEL' command")
	}
	n, err := ctx.Store.HDel(cmd.Args[0], cmd.Args[1])
	if err != nil {
		if isWrongType(err) {
			return respWrongType()
		}
		return respErr(err.Error())
	}
	if n > 0 {
		writeAOF(ctx, cmd)
	}
	return respInt(int64(n))
}

// HandleHGetAll handles: HGETALL key → flat array of field-value pairs (empty if not found)
func HandleHGetAll(cmd types.Command, ctx *Context) string {
	if len(cmd.Args) != 1 {
		return respErr("wrong number of arguments for 'HGETALL' command")
	}
	items, err := ctx.Store.HGetAll(cmd.Args[0])
	if err != nil {
		if isWrongType(err) {
			return respWrongType()
		}
		return respErr(err.Error())
	}
	return respArray(items)
}

// HandleHExists handles: HEXISTS key field → 1 or 0
func HandleHExists(cmd types.Command, ctx *Context) string {
	if len(cmd.Args) != 2 {
		return respErr("wrong number of arguments for 'HEXISTS' command")
	}
	n, err := ctx.Store.HExists(cmd.Args[0], cmd.Args[1])
	if err != nil {
		if isWrongType(err) {
			return respWrongType()
		}
		return respErr(err.Error())
	}
	return respInt(int64(n))
}
