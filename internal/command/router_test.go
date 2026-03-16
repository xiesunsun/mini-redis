package command

import (
	"testing"

	"github.com/xiesunsun/mini-redis/internal/types"
)

func TestNewRouter_RegisterAllCommands(t *testing.T) {
	router := NewRouter(newTestCtx())

	expected := []string{
		"SET", "GET", "DEL", "EXPIRE", "TTL",
		"LPUSH", "RPUSH", "LRANGE", "LLEN", "LPOP", "RPOP",
		"HSET", "HGET", "HDEL", "HGETALL", "HEXISTS",
	}

	if len(router.routes) != len(expected) {
		t.Fatalf("expected %d routes, got %d", len(expected), len(router.routes))
	}

	for _, name := range expected {
		if _, ok := router.routes[name]; !ok {
			t.Fatalf("missing route %q", name)
		}
	}
}

func TestRouterDispatch_KnownCommand(t *testing.T) {
	ctx := newTestCtx()
	router := NewRouter(ctx)

	if got := router.Dispatch(cmd("SET", "k", "v")); got != respOK() {
		t.Fatalf("expected +OK, got %q", got)
	}
	if got := router.Dispatch(cmd("GET", "k")); got != respBulk("v") {
		t.Fatalf("expected bulk v, got %q", got)
	}
}

func TestRouterDispatch_CaseInsensitiveCommand(t *testing.T) {
	ctx := newTestCtx()
	router := NewRouter(ctx)

	if got := router.Dispatch(cmd("set", "k", "v")); got != respOK() {
		t.Fatalf("expected +OK, got %q", got)
	}
	if got := router.Dispatch(cmd(" gEt ", "k")); got != respBulk("v") {
		t.Fatalf("expected bulk v, got %q", got)
	}
}

func TestRouterDispatch_UnknownCommand(t *testing.T) {
	router := NewRouter(newTestCtx())

	got := router.Dispatch(types.Command{Name: "UNKNOWN", Args: []string{"k"}})
	want := respErr("unknown command 'unknown'")
	if got != want {
		t.Fatalf("expected %q, got %q", want, got)
	}
}

func TestRouterDispatch_NilRouter(t *testing.T) {
	var router *Router

	got := router.Dispatch(cmd("GET", "k"))
	want := respErr("router is not initialized")
	if got != want {
		t.Fatalf("expected %q, got %q", want, got)
	}
}

func TestDispatch_FunctionEntry(t *testing.T) {
	ctx := newTestCtx()

	if got := Dispatch(cmd("SET", "k", "v"), ctx); got != respOK() {
		t.Fatalf("expected +OK, got %q", got)
	}
	if got := Dispatch(cmd("GET", "k"), ctx); got != respBulk("v") {
		t.Fatalf("expected bulk v, got %q", got)
	}
}
