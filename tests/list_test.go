package tests

import (
	"fmt"
	"strings"
	"testing"

	"github.com/xiesunsun/mini-redis/internal/network"
)

func requireArrayStrings(t *testing.T, resp network.Value, want []string) {
	t.Helper()
	if resp.Type != network.RespArray || resp.IsNull {
		t.Fatalf("unexpected array response: %+v", resp)
	}
	if len(resp.Array) != len(want) {
		t.Fatalf("unexpected array length: got %d, want %d", len(resp.Array), len(want))
	}
	for i, wantItem := range want {
		item := resp.Array[i]
		if item.Type != network.RespBulkString || item.IsNull || item.String != wantItem {
			t.Fatalf("unexpected array item at %d: got %+v, want %q", i, item, wantItem)
		}
	}
}

func TestList_LPush_RPush_LRange_LLen_NormalPath(t *testing.T) {
	addr, cleanup := startStringTestServer(t)
	defer cleanup()

	rw := connectStringClient(t, addr)

	requireInteger(t, sendCommand(t, rw, "LPUSH", "nums", "b"), 1)
	requireInteger(t, sendCommand(t, rw, "LPUSH", "nums", "a"), 2)
	requireInteger(t, sendCommand(t, rw, "RPUSH", "nums", "c"), 3)
	requireInteger(t, sendCommand(t, rw, "LLEN", "nums"), 3)
	requireArrayStrings(t, sendCommand(t, rw, "LRANGE", "nums", "0", "-1"), []string{"a", "b", "c"})
}

func TestList_LRange_BoundaryIndexes(t *testing.T) {
	addr, cleanup := startStringTestServer(t)
	defer cleanup()

	rw := connectStringClient(t, addr)

	requireInteger(t, sendCommand(t, rw, "RPUSH", "letters", "a"), 1)
	requireInteger(t, sendCommand(t, rw, "RPUSH", "letters", "b"), 2)
	requireInteger(t, sendCommand(t, rw, "RPUSH", "letters", "c"), 3)

	requireArrayStrings(t, sendCommand(t, rw, "LRANGE", "letters", "-2", "-1"), []string{"b", "c"})
	requireArrayStrings(t, sendCommand(t, rw, "LRANGE", "letters", "0", "100"), []string{"a", "b", "c"})
	requireArrayStrings(t, sendCommand(t, rw, "LRANGE", "letters", "3", "1"), []string{})
}

func TestList_LPop_RPop_NormalPath(t *testing.T) {
	addr, cleanup := startStringTestServer(t)
	defer cleanup()

	rw := connectStringClient(t, addr)

	requireInteger(t, sendCommand(t, rw, "RPUSH", "stack", "a"), 1)
	requireInteger(t, sendCommand(t, rw, "RPUSH", "stack", "b"), 2)
	requireInteger(t, sendCommand(t, rw, "RPUSH", "stack", "c"), 3)

	requireBulkString(t, sendCommand(t, rw, "LPOP", "stack"), "a")
	requireBulkString(t, sendCommand(t, rw, "RPOP", "stack"), "c")
	requireInteger(t, sendCommand(t, rw, "LLEN", "stack"), 1)
	requireArrayStrings(t, sendCommand(t, rw, "LRANGE", "stack", "0", "-1"), []string{"b"})
}

func TestList_NonExistentKey(t *testing.T) {
	addr, cleanup := startStringTestServer(t)
	defer cleanup()

	rw := connectStringClient(t, addr)

	requireInteger(t, sendCommand(t, rw, "LLEN", "missing"), 0)
	requireArrayStrings(t, sendCommand(t, rw, "LRANGE", "missing", "0", "-1"), []string{})
	requireNilBulk(t, sendCommand(t, rw, "LPOP", "missing"))
	requireNilBulk(t, sendCommand(t, rw, "RPOP", "missing"))
}

func TestList_WrongType(t *testing.T) {
	addr, cleanup := startStringTestServer(t)
	defer cleanup()

	rw := connectStringClient(t, addr)

	requireSimpleOK(t, sendCommand(t, rw, "SET", "k", "v"))

	for _, args := range [][]string{
		{"LPUSH", "k", "x"},
		{"RPUSH", "k", "x"},
		{"LRANGE", "k", "0", "-1"},
		{"LLEN", "k"},
		{"LPOP", "k"},
		{"RPOP", "k"},
	} {
		t.Run(args[0], func(t *testing.T) {
			requireErrorContains(t, sendCommand(t, rw, args...), "WRONGTYPE")
		})
	}
}

func TestList_CommandErrors(t *testing.T) {
	addr, cleanup := startStringTestServer(t)
	defer cleanup()

	rw := connectStringClient(t, addr)

	errCases := []struct {
		args []string
		want string
	}{
		{args: []string{"LPUSH", "k"}, want: "wrong number of arguments for 'LPUSH' command"},
		{args: []string{"RPUSH", "k"}, want: "wrong number of arguments for 'RPUSH' command"},
		{args: []string{"LRANGE", "k", "0"}, want: "wrong number of arguments for 'LRANGE' command"},
		{args: []string{"LLEN"}, want: "wrong number of arguments for 'LLEN' command"},
		{args: []string{"LPOP"}, want: "wrong number of arguments for 'LPOP' command"},
		{args: []string{"RPOP"}, want: "wrong number of arguments for 'RPOP' command"},
		{args: []string{"LRANGE", "k", "NaN", "1"}, want: "value is not an integer or out of range"},
		{args: []string{"LRANGE", "k", "0", "NaN"}, want: "value is not an integer or out of range"},
	}

	for _, tc := range errCases {
		t.Run(fmt.Sprintf("%s_%s", tc.args[0], strings.Join(tc.args[1:], "_")), func(t *testing.T) {
			requireErrorContains(t, sendCommand(t, rw, tc.args...), tc.want)
		})
	}
}
