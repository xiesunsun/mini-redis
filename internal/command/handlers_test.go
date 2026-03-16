package command

import (
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/xiesunsun/mini-redis/internal/persistence"
	"github.com/xiesunsun/mini-redis/internal/store"
	"github.com/xiesunsun/mini-redis/internal/types"
)

// newTestCtx creates a Context with a fresh Store and no AOF (nil is safe).
func newTestCtx() *Context {
	return &Context{
		Store: store.New(),
		AOF:   nil,
	}
}

func cmd(name string, args ...string) types.Command {
	return types.Command{Name: name, Args: args}
}

// --- SET ---

func TestHandleSet_ValidArgs(t *testing.T) {
	ctx := newTestCtx()
	got := HandleSet(cmd("SET", "k", "v"), ctx)
	if got != respOK() {
		t.Fatalf("expected +OK, got %q", got)
	}
}

func TestHandleSet_MissingArgs(t *testing.T) {
	ctx := newTestCtx()
	got := HandleSet(cmd("SET", "k"), ctx)
	if !strings.HasPrefix(got, "-ERR") {
		t.Fatalf("expected ERR, got %q", got)
	}
}

// --- GET ---

func TestHandleGet_ExistingKey(t *testing.T) {
	ctx := newTestCtx()
	HandleSet(cmd("SET", "k", "hello"), ctx)
	got := HandleGet(cmd("GET", "k"), ctx)
	if got != respBulk("hello") {
		t.Fatalf("expected bulk hello, got %q", got)
	}
}

func TestHandleGet_NonExistentKey(t *testing.T) {
	ctx := newTestCtx()
	got := HandleGet(cmd("GET", "missing"), ctx)
	if got != respNil() {
		t.Fatalf("expected nil, got %q", got)
	}
}

func TestHandleGet_WrongType(t *testing.T) {
	ctx := newTestCtx()
	ctx.Store.LPush("mylist", "a")
	got := HandleGet(cmd("GET", "mylist"), ctx)
	if !strings.HasPrefix(got, "-WRONGTYPE") {
		t.Fatalf("expected WRONGTYPE, got %q", got)
	}
}

func TestHandleGet_MissingArgs(t *testing.T) {
	ctx := newTestCtx()
	got := HandleGet(cmd("GET"), ctx)
	if !strings.HasPrefix(got, "-ERR") {
		t.Fatalf("expected ERR, got %q", got)
	}
}

// --- DEL ---

func TestHandleDel_ExistingKey(t *testing.T) {
	ctx := newTestCtx()
	HandleSet(cmd("SET", "k", "v"), ctx)
	got := HandleDel(cmd("DEL", "k"), ctx)
	if got != respInt(1) {
		t.Fatalf("expected :1, got %q", got)
	}
}

func TestHandleDel_NonExistentKey(t *testing.T) {
	ctx := newTestCtx()
	got := HandleDel(cmd("DEL", "missing"), ctx)
	if got != respInt(0) {
		t.Fatalf("expected :0, got %q", got)
	}
}

func TestHandleDel_MissingArgs(t *testing.T) {
	ctx := newTestCtx()
	got := HandleDel(cmd("DEL"), ctx)
	if !strings.HasPrefix(got, "-ERR") {
		t.Fatalf("expected ERR, got %q", got)
	}
}

// --- EXPIRE ---

func TestHandleExpire_ExistingKey(t *testing.T) {
	ctx := newTestCtx()
	HandleSet(cmd("SET", "k", "v"), ctx)
	got := HandleExpire(cmd("EXPIRE", "k", "100"), ctx)
	if got != respInt(1) {
		t.Fatalf("expected :1, got %q", got)
	}
}

func TestHandleExpire_NonExistentKey(t *testing.T) {
	ctx := newTestCtx()
	got := HandleExpire(cmd("EXPIRE", "missing", "100"), ctx)
	if got != respInt(0) {
		t.Fatalf("expected :0, got %q", got)
	}
}

func TestHandleExpire_InvalidSeconds(t *testing.T) {
	ctx := newTestCtx()
	HandleSet(cmd("SET", "k", "v"), ctx)
	got := HandleExpire(cmd("EXPIRE", "k", "notanumber"), ctx)
	if !strings.HasPrefix(got, "-ERR") {
		t.Fatalf("expected ERR, got %q", got)
	}
}

// --- TTL ---

func TestHandleTTL_NonExistentKey(t *testing.T) {
	ctx := newTestCtx()
	got := HandleTTL(cmd("TTL", "missing"), ctx)
	if got != respInt(-2) {
		t.Fatalf("expected :-2, got %q", got)
	}
}

func TestHandleTTL_NoExpiry(t *testing.T) {
	ctx := newTestCtx()
	HandleSet(cmd("SET", "k", "v"), ctx)
	got := HandleTTL(cmd("TTL", "k"), ctx)
	if got != respInt(-1) {
		t.Fatalf("expected :-1, got %q", got)
	}
}

func TestHandleTTL_WithExpiry(t *testing.T) {
	ctx := newTestCtx()
	HandleSet(cmd("SET", "k", "v"), ctx)
	HandleExpire(cmd("EXPIRE", "k", "100"), ctx)
	got := HandleTTL(cmd("TTL", "k"), ctx)
	// TTL should be close to 100
	if got == respInt(-1) || got == respInt(-2) {
		t.Fatalf("expected positive TTL, got %q", got)
	}
}

func TestHandleTTL_ExpiredKey(t *testing.T) {
	ctx := newTestCtx()
	HandleSet(cmd("SET", "k", "v"), ctx)
	// Manually set expiry in the past
	v := ctx.Store.Get("k")
	v.Expiry = time.Now().Add(-1 * time.Second)
	got := HandleTTL(cmd("TTL", "k"), ctx)
	if got != respInt(-2) {
		t.Fatalf("expected :-2 for expired key, got %q", got)
	}
}

// --- LPUSH ---

func TestHandleLPush_ValidArgs(t *testing.T) {
	ctx := newTestCtx()
	got := HandleLPush(cmd("LPUSH", "mylist", "a"), ctx)
	if got != respInt(1) {
		t.Fatalf("expected :1, got %q", got)
	}
	got = HandleLPush(cmd("LPUSH", "mylist", "b"), ctx)
	if got != respInt(2) {
		t.Fatalf("expected :2, got %q", got)
	}
}

func TestHandleLPush_WrongType(t *testing.T) {
	ctx := newTestCtx()
	HandleSet(cmd("SET", "k", "v"), ctx)
	got := HandleLPush(cmd("LPUSH", "k", "a"), ctx)
	if !strings.HasPrefix(got, "-WRONGTYPE") {
		t.Fatalf("expected WRONGTYPE, got %q", got)
	}
}

func TestHandleLPush_MissingArgs(t *testing.T) {
	ctx := newTestCtx()
	got := HandleLPush(cmd("LPUSH", "mylist"), ctx)
	if !strings.HasPrefix(got, "-ERR") {
		t.Fatalf("expected ERR, got %q", got)
	}
}

// --- RPUSH ---

func TestHandleRPush_ValidArgs(t *testing.T) {
	ctx := newTestCtx()
	got := HandleRPush(cmd("RPUSH", "mylist", "a"), ctx)
	if got != respInt(1) {
		t.Fatalf("expected :1, got %q", got)
	}
	got = HandleRPush(cmd("RPUSH", "mylist", "b"), ctx)
	if got != respInt(2) {
		t.Fatalf("expected :2, got %q", got)
	}
}

func TestHandleRPush_WrongType(t *testing.T) {
	ctx := newTestCtx()
	HandleSet(cmd("SET", "k", "v"), ctx)
	got := HandleRPush(cmd("RPUSH", "k", "a"), ctx)
	if !strings.HasPrefix(got, "-WRONGTYPE") {
		t.Fatalf("expected WRONGTYPE, got %q", got)
	}
}

// --- LRANGE ---

func TestHandleLRange_ValidRange(t *testing.T) {
	ctx := newTestCtx()
	HandleRPush(cmd("RPUSH", "mylist", "a"), ctx)
	HandleRPush(cmd("RPUSH", "mylist", "b"), ctx)
	HandleRPush(cmd("RPUSH", "mylist", "c"), ctx)
	got := HandleLRange(cmd("LRANGE", "mylist", "0", "-1"), ctx)
	want := respArray([]string{"a", "b", "c"})
	if got != want {
		t.Fatalf("expected %q, got %q", want, got)
	}
}

func TestHandleLRange_NonExistentKey(t *testing.T) {
	ctx := newTestCtx()
	got := HandleLRange(cmd("LRANGE", "missing", "0", "-1"), ctx)
	if got != "*0\r\n" {
		t.Fatalf("expected empty array, got %q", got)
	}
}

func TestHandleLRange_WrongType(t *testing.T) {
	ctx := newTestCtx()
	HandleSet(cmd("SET", "k", "v"), ctx)
	got := HandleLRange(cmd("LRANGE", "k", "0", "-1"), ctx)
	if !strings.HasPrefix(got, "-WRONGTYPE") {
		t.Fatalf("expected WRONGTYPE, got %q", got)
	}
}

func TestHandleLRange_InvalidIndex(t *testing.T) {
	ctx := newTestCtx()
	got := HandleLRange(cmd("LRANGE", "mylist", "x", "1"), ctx)
	if !strings.HasPrefix(got, "-ERR") {
		t.Fatalf("expected ERR, got %q", got)
	}
}

// --- LLEN ---

func TestHandleLLen_ValidKey(t *testing.T) {
	ctx := newTestCtx()
	HandleRPush(cmd("RPUSH", "mylist", "a"), ctx)
	HandleRPush(cmd("RPUSH", "mylist", "b"), ctx)
	got := HandleLLen(cmd("LLEN", "mylist"), ctx)
	if got != respInt(2) {
		t.Fatalf("expected :2, got %q", got)
	}
}

func TestHandleLLen_NonExistentKey(t *testing.T) {
	ctx := newTestCtx()
	got := HandleLLen(cmd("LLEN", "missing"), ctx)
	if got != respInt(0) {
		t.Fatalf("expected :0, got %q", got)
	}
}

func TestHandleLLen_WrongType(t *testing.T) {
	ctx := newTestCtx()
	HandleSet(cmd("SET", "k", "v"), ctx)
	got := HandleLLen(cmd("LLEN", "k"), ctx)
	if !strings.HasPrefix(got, "-WRONGTYPE") {
		t.Fatalf("expected WRONGTYPE, got %q", got)
	}
}

// --- LPOP ---

func TestHandleLPop_ExistingKey(t *testing.T) {
	ctx := newTestCtx()
	HandleLPush(cmd("LPUSH", "mylist", "b"), ctx)
	HandleLPush(cmd("LPUSH", "mylist", "a"), ctx)
	got := HandleLPop(cmd("LPOP", "mylist"), ctx)
	if got != respBulk("a") {
		t.Fatalf("expected bulk 'a', got %q", got)
	}
}

func TestHandleLPop_NonExistentKey(t *testing.T) {
	ctx := newTestCtx()
	got := HandleLPop(cmd("LPOP", "missing"), ctx)
	if got != respNil() {
		t.Fatalf("expected nil, got %q", got)
	}
}

func TestHandleLPop_WrongType(t *testing.T) {
	ctx := newTestCtx()
	HandleSet(cmd("SET", "k", "v"), ctx)
	got := HandleLPop(cmd("LPOP", "k"), ctx)
	if !strings.HasPrefix(got, "-WRONGTYPE") {
		t.Fatalf("expected WRONGTYPE, got %q", got)
	}
}

// --- RPOP ---

func TestHandleRPop_ExistingKey(t *testing.T) {
	ctx := newTestCtx()
	HandleRPush(cmd("RPUSH", "mylist", "a"), ctx)
	HandleRPush(cmd("RPUSH", "mylist", "b"), ctx)
	got := HandleRPop(cmd("RPOP", "mylist"), ctx)
	if got != respBulk("b") {
		t.Fatalf("expected bulk 'b', got %q", got)
	}
}

func TestHandleRPop_NonExistentKey(t *testing.T) {
	ctx := newTestCtx()
	got := HandleRPop(cmd("RPOP", "missing"), ctx)
	if got != respNil() {
		t.Fatalf("expected nil, got %q", got)
	}
}

func TestHandleRPop_WrongType(t *testing.T) {
	ctx := newTestCtx()
	HandleSet(cmd("SET", "k", "v"), ctx)
	got := HandleRPop(cmd("RPOP", "k"), ctx)
	if !strings.HasPrefix(got, "-WRONGTYPE") {
		t.Fatalf("expected WRONGTYPE, got %q", got)
	}
}

// --- HSET ---

func TestHandleHSet_NewField(t *testing.T) {
	ctx := newTestCtx()
	got := HandleHSet(cmd("HSET", "h", "f", "v"), ctx)
	if got != respInt(1) {
		t.Fatalf("expected :1, got %q", got)
	}
}

func TestHandleHSet_ExistingField(t *testing.T) {
	ctx := newTestCtx()
	HandleHSet(cmd("HSET", "h", "f", "v1"), ctx)
	got := HandleHSet(cmd("HSET", "h", "f", "v2"), ctx)
	if got != respInt(0) {
		t.Fatalf("expected :0, got %q", got)
	}
}

func TestHandleHSet_WrongType(t *testing.T) {
	ctx := newTestCtx()
	HandleSet(cmd("SET", "k", "v"), ctx)
	got := HandleHSet(cmd("HSET", "k", "f", "v"), ctx)
	if !strings.HasPrefix(got, "-WRONGTYPE") {
		t.Fatalf("expected WRONGTYPE, got %q", got)
	}
}

func TestHandleHSet_MissingArgs(t *testing.T) {
	ctx := newTestCtx()
	got := HandleHSet(cmd("HSET", "h", "f"), ctx)
	if !strings.HasPrefix(got, "-ERR") {
		t.Fatalf("expected ERR, got %q", got)
	}
}

// --- HGET ---

func TestHandleHGet_ExistingField(t *testing.T) {
	ctx := newTestCtx()
	HandleHSet(cmd("HSET", "h", "f", "hello"), ctx)
	got := HandleHGet(cmd("HGET", "h", "f"), ctx)
	if got != respBulk("hello") {
		t.Fatalf("expected bulk hello, got %q", got)
	}
}

func TestHandleHGet_NonExistentKey(t *testing.T) {
	ctx := newTestCtx()
	got := HandleHGet(cmd("HGET", "missing", "f"), ctx)
	if got != respNil() {
		t.Fatalf("expected nil, got %q", got)
	}
}

func TestHandleHGet_NonExistentField(t *testing.T) {
	ctx := newTestCtx()
	HandleHSet(cmd("HSET", "h", "f1", "v"), ctx)
	got := HandleHGet(cmd("HGET", "h", "f2"), ctx)
	if got != respNil() {
		t.Fatalf("expected nil, got %q", got)
	}
}

func TestHandleHGet_WrongType(t *testing.T) {
	ctx := newTestCtx()
	HandleSet(cmd("SET", "k", "v"), ctx)
	got := HandleHGet(cmd("HGET", "k", "f"), ctx)
	if !strings.HasPrefix(got, "-WRONGTYPE") {
		t.Fatalf("expected WRONGTYPE, got %q", got)
	}
}

// --- HDEL ---

func TestHandleHDel_ExistingField(t *testing.T) {
	ctx := newTestCtx()
	HandleHSet(cmd("HSET", "h", "f", "v"), ctx)
	got := HandleHDel(cmd("HDEL", "h", "f"), ctx)
	if got != respInt(1) {
		t.Fatalf("expected :1, got %q", got)
	}
}

func TestHandleHDel_NonExistentField(t *testing.T) {
	ctx := newTestCtx()
	HandleHSet(cmd("HSET", "h", "f1", "v"), ctx)
	got := HandleHDel(cmd("HDEL", "h", "f2"), ctx)
	if got != respInt(0) {
		t.Fatalf("expected :0, got %q", got)
	}
}

func TestHandleHDel_WrongType(t *testing.T) {
	ctx := newTestCtx()
	HandleSet(cmd("SET", "k", "v"), ctx)
	got := HandleHDel(cmd("HDEL", "k", "f"), ctx)
	if !strings.HasPrefix(got, "-WRONGTYPE") {
		t.Fatalf("expected WRONGTYPE, got %q", got)
	}
}

// --- HGETALL ---

func TestHandleHGetAll_ValidKey(t *testing.T) {
	ctx := newTestCtx()
	HandleHSet(cmd("HSET", "h", "f1", "v1"), ctx)
	HandleHSet(cmd("HSET", "h", "f2", "v2"), ctx)
	got := HandleHGetAll(cmd("HGETALL", "h"), ctx)
	// Should be a 4-element array: f1 v1 f2 v2 (order may vary)
	if !strings.HasPrefix(got, "*4\r\n") {
		t.Fatalf("expected 4-element array, got %q", got)
	}
	if !strings.Contains(got, "f1") || !strings.Contains(got, "v1") {
		t.Fatalf("missing f1/v1 in response: %q", got)
	}
	if !strings.Contains(got, "f2") || !strings.Contains(got, "v2") {
		t.Fatalf("missing f2/v2 in response: %q", got)
	}
}

func TestHandleHGetAll_NonExistentKey(t *testing.T) {
	ctx := newTestCtx()
	got := HandleHGetAll(cmd("HGETALL", "missing"), ctx)
	if got != "*0\r\n" {
		t.Fatalf("expected empty array, got %q", got)
	}
}

func TestHandleHGetAll_WrongType(t *testing.T) {
	ctx := newTestCtx()
	HandleSet(cmd("SET", "k", "v"), ctx)
	got := HandleHGetAll(cmd("HGETALL", "k"), ctx)
	if !strings.HasPrefix(got, "-WRONGTYPE") {
		t.Fatalf("expected WRONGTYPE, got %q", got)
	}
}

// --- HEXISTS ---

func TestHandleHExists_ExistingField(t *testing.T) {
	ctx := newTestCtx()
	HandleHSet(cmd("HSET", "h", "f", "v"), ctx)
	got := HandleHExists(cmd("HEXISTS", "h", "f"), ctx)
	if got != respInt(1) {
		t.Fatalf("expected :1, got %q", got)
	}
}

func TestHandleHExists_NonExistentKey(t *testing.T) {
	ctx := newTestCtx()
	got := HandleHExists(cmd("HEXISTS", "missing", "f"), ctx)
	if got != respInt(0) {
		t.Fatalf("expected :0, got %q", got)
	}
}

func TestHandleHExists_NonExistentField(t *testing.T) {
	ctx := newTestCtx()
	HandleHSet(cmd("HSET", "h", "f1", "v"), ctx)
	got := HandleHExists(cmd("HEXISTS", "h", "f2"), ctx)
	if got != respInt(0) {
		t.Fatalf("expected :0, got %q", got)
	}
}

func TestHandleHExists_WrongType(t *testing.T) {
	ctx := newTestCtx()
	HandleSet(cmd("SET", "k", "v"), ctx)
	got := HandleHExists(cmd("HEXISTS", "k", "f"), ctx)
	if !strings.HasPrefix(got, "-WRONGTYPE") {
		t.Fatalf("expected WRONGTYPE, got %q", got)
	}
}

func TestHandlers_ExtraArgsStrictArity(t *testing.T) {
	cases := []struct {
		name    string
		handler HandlerFunc
		cmd     types.Command
		want    string
	}{
		{"SET", HandleSet, cmd("SET", "k", "v", "x"), respErr("wrong number of arguments for 'SET' command")},
		{"GET", HandleGet, cmd("GET", "k", "x"), respErr("wrong number of arguments for 'GET' command")},
		{"DEL", HandleDel, cmd("DEL", "k", "x"), respErr("wrong number of arguments for 'DEL' command")},
		{"EXPIRE", HandleExpire, cmd("EXPIRE", "k", "100", "x"), respErr("wrong number of arguments for 'EXPIRE' command")},
		{"TTL", HandleTTL, cmd("TTL", "k", "x"), respErr("wrong number of arguments for 'TTL' command")},
		{"LPUSH", HandleLPush, cmd("LPUSH", "list", "v", "x"), respErr("wrong number of arguments for 'LPUSH' command")},
		{"RPUSH", HandleRPush, cmd("RPUSH", "list", "v", "x"), respErr("wrong number of arguments for 'RPUSH' command")},
		{"LRANGE", HandleLRange, cmd("LRANGE", "list", "0", "-1", "x"), respErr("wrong number of arguments for 'LRANGE' command")},
		{"LLEN", HandleLLen, cmd("LLEN", "list", "x"), respErr("wrong number of arguments for 'LLEN' command")},
		{"LPOP", HandleLPop, cmd("LPOP", "list", "x"), respErr("wrong number of arguments for 'LPOP' command")},
		{"RPOP", HandleRPop, cmd("RPOP", "list", "x"), respErr("wrong number of arguments for 'RPOP' command")},
		{"HSET", HandleHSet, cmd("HSET", "h", "f", "v", "x"), respErr("wrong number of arguments for 'HSET' command")},
		{"HGET", HandleHGet, cmd("HGET", "h", "f", "x"), respErr("wrong number of arguments for 'HGET' command")},
		{"HDEL", HandleHDel, cmd("HDEL", "h", "f", "x"), respErr("wrong number of arguments for 'HDEL' command")},
		{"HGETALL", HandleHGetAll, cmd("HGETALL", "h", "x"), respErr("wrong number of arguments for 'HGETALL' command")},
		{"HEXISTS", HandleHExists, cmd("HEXISTS", "h", "f", "x"), respErr("wrong number of arguments for 'HEXISTS' command")},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := tc.handler(tc.cmd, newTestCtx())
			if got != tc.want {
				t.Fatalf("expected %q, got %q", tc.want, got)
			}
		})
	}
}

// --- Additional edge cases ---

func TestHandleGet_ExpiredKey(t *testing.T) {
	ctx := newTestCtx()
	HandleSet(cmd("SET", "k", "v"), ctx)
	v := ctx.Store.Get("k")
	v.Expiry = time.Now().Add(-1 * time.Second)
	got := HandleGet(cmd("GET", "k"), ctx)
	if got != respNil() {
		t.Fatalf("expected nil for expired key, got %q", got)
	}
}

func TestHandleDel_ExpiredKey(t *testing.T) {
	ctx := newTestCtx()
	HandleSet(cmd("SET", "k", "v"), ctx)
	v := ctx.Store.Get("k")
	v.Expiry = time.Now().Add(-1 * time.Second)
	got := HandleDel(cmd("DEL", "k"), ctx)
	if got != respInt(0) {
		t.Fatalf("expected :0 for expired key, got %q", got)
	}
}

func TestHandleRespBulk_Format(t *testing.T) {
	s := "hello"
	want := "$5\r\nhello\r\n"
	if got := respBulk(s); got != want {
		t.Fatalf("respBulk(%q) = %q, want %q", s, got, want)
	}
}

func newClosedAOFCtx(t *testing.T) *Context {
	t.Helper()

	aof, err := persistence.New(filepath.Join(t.TempDir(), "appendonly.aof"))
	if err != nil {
		t.Fatalf("create AOF: %v", err)
	}
	if err := aof.Close(); err != nil {
		t.Fatalf("close AOF: %v", err)
	}

	return &Context{
		Store: store.New(),
		AOF:   aof,
	}
}

func TestWriteCommands_AOFWriteError(t *testing.T) {
	cases := []struct {
		name    string
		prepare func(ctx *Context)
		invoke  func(ctx *Context) string
	}{
		{
			name: "SET",
			invoke: func(ctx *Context) string {
				return HandleSet(cmd("SET", "k", "v"), ctx)
			},
		},
		{
			name: "DEL",
			prepare: func(ctx *Context) {
				ctx.Store.SetString("k", "v")
			},
			invoke: func(ctx *Context) string {
				return HandleDel(cmd("DEL", "k"), ctx)
			},
		},
		{
			name: "EXPIRE",
			prepare: func(ctx *Context) {
				ctx.Store.SetString("k", "v")
			},
			invoke: func(ctx *Context) string {
				return HandleExpire(cmd("EXPIRE", "k", "10"), ctx)
			},
		},
		{
			name: "LPUSH",
			invoke: func(ctx *Context) string {
				return HandleLPush(cmd("LPUSH", "list", "v"), ctx)
			},
		},
		{
			name: "RPUSH",
			invoke: func(ctx *Context) string {
				return HandleRPush(cmd("RPUSH", "list", "v"), ctx)
			},
		},
		{
			name: "LPOP",
			prepare: func(ctx *Context) {
				if _, err := ctx.Store.RPush("list", "v"); err != nil {
					t.Fatalf("prepare list for LPOP: %v", err)
				}
			},
			invoke: func(ctx *Context) string {
				return HandleLPop(cmd("LPOP", "list"), ctx)
			},
		},
		{
			name: "RPOP",
			prepare: func(ctx *Context) {
				if _, err := ctx.Store.RPush("list", "v"); err != nil {
					t.Fatalf("prepare list for RPOP: %v", err)
				}
			},
			invoke: func(ctx *Context) string {
				return HandleRPop(cmd("RPOP", "list"), ctx)
			},
		},
		{
			name: "HSET",
			invoke: func(ctx *Context) string {
				return HandleHSet(cmd("HSET", "h", "f", "v"), ctx)
			},
		},
		{
			name: "HDEL",
			prepare: func(ctx *Context) {
				if _, err := ctx.Store.HSet("h", "f", "v"); err != nil {
					t.Fatalf("prepare hash for HDEL: %v", err)
				}
			},
			invoke: func(ctx *Context) string {
				return HandleHDel(cmd("HDEL", "h", "f"), ctx)
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			ctx := newClosedAOFCtx(t)
			if tc.prepare != nil {
				tc.prepare(ctx)
			}

			got := tc.invoke(ctx)
			want := respErr("AOF write failed")
			if got != want {
				t.Fatalf("expected %q, got %q", want, got)
			}
		})
	}
}
