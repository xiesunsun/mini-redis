package command

import (
	"fmt"
	"strings"

	"github.com/xiesunsun/mini-redis/internal/types"
)

// Router stores command-to-handler mappings and executes commands.
type Router struct {
	ctx    *Context
	routes map[string]HandlerFunc
}

// NewRouter creates a command router with all supported core commands.
func NewRouter(ctx *Context) *Router {
	return &Router{
		ctx:    ctx,
		routes: defaultRoutes(),
	}
}

func defaultRoutes() map[string]HandlerFunc {
	return map[string]HandlerFunc{
		"SET":     HandleSet,
		"GET":     HandleGet,
		"DEL":     HandleDel,
		"EXPIRE":  HandleExpire,
		"TTL":     HandleTTL,
		"LPUSH":   HandleLPush,
		"RPUSH":   HandleRPush,
		"LRANGE":  HandleLRange,
		"LLEN":    HandleLLen,
		"LPOP":    HandleLPop,
		"RPOP":    HandleRPop,
		"HSET":    HandleHSet,
		"HGET":    HandleHGet,
		"HDEL":    HandleHDel,
		"HGETALL": HandleHGetAll,
		"HEXISTS": HandleHExists,
	}
}

// Dispatch routes cmd to a command handler and returns a RESP-formatted response.
func (r *Router) Dispatch(cmd types.Command) string {
	if r == nil || r.ctx == nil || r.ctx.Store == nil {
		return respErr("router is not initialized")
	}

	name := strings.ToUpper(strings.TrimSpace(cmd.Name))
	handler, ok := r.routes[name]
	if !ok {
		return respErr(fmt.Sprintf("unknown command '%s'", strings.ToLower(name)))
	}

	cmd.Name = name
	return handler(cmd, r.ctx)
}

// DispatchParts is a helper that dispatches command name and args directly.
func (r *Router) DispatchParts(name string, args []string) string {
	return r.Dispatch(types.Command{Name: name, Args: args})
}

// Dispatch is a convenience entry point that creates a default router and executes cmd.
func Dispatch(cmd types.Command, ctx *Context) string {
	return NewRouter(ctx).Dispatch(cmd)
}

// DispatchParts is a convenience entry point for name/args inputs.
func DispatchParts(name string, args []string, ctx *Context) string {
	return Dispatch(types.Command{Name: name, Args: args}, ctx)
}
