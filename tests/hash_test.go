package tests

import (
	"fmt"
	"strings"
	"testing"

	"github.com/xiesunsun/mini-redis/internal/network"
)

func requireHashPairs(t *testing.T, resp network.Value, want map[string]string) {
	t.Helper()
	if resp.Type != network.RespArray || resp.IsNull {
		t.Fatalf("unexpected array response: %+v", resp)
	}
	if len(resp.Array)%2 != 0 {
		t.Fatalf("unexpected hash array length: got %d, want even number", len(resp.Array))
	}

	got := make(map[string]string, len(resp.Array)/2)
	for i := 0; i < len(resp.Array); i += 2 {
		field := resp.Array[i]
		value := resp.Array[i+1]
		if field.Type != network.RespBulkString || field.IsNull {
			t.Fatalf("unexpected hash field at index %d: %+v", i, field)
		}
		if value.Type != network.RespBulkString || value.IsNull {
			t.Fatalf("unexpected hash value at index %d: %+v", i+1, value)
		}
		got[field.String] = value.String
	}

	if len(got) != len(want) {
		t.Fatalf("unexpected hash size: got %d, want %d", len(got), len(want))
	}

	for field, wantValue := range want {
		gotValue, ok := got[field]
		if !ok {
			t.Fatalf("missing field %q in response: %+v", field, got)
		}
		if gotValue != wantValue {
			t.Fatalf("unexpected value for field %q: got %q, want %q", field, gotValue, wantValue)
		}
	}
}

func TestHash_HSet_HGet_HGetAll_HExists_NormalPath(t *testing.T) {
	addr, cleanup := startStringTestServer(t)
	defer cleanup()

	rw := connectStringClient(t, addr)

	requireInteger(t, sendCommand(t, rw, "HSET", "user:1", "name", "alice"), 1)
	requireInteger(t, sendCommand(t, rw, "HSET", "user:1", "age", "18"), 1)
	requireInteger(t, sendCommand(t, rw, "HSET", "user:1", "age", "20"), 0)

	requireBulkString(t, sendCommand(t, rw, "HGET", "user:1", "age"), "20")
	requireInteger(t, sendCommand(t, rw, "HEXISTS", "user:1", "name"), 1)
	requireInteger(t, sendCommand(t, rw, "HEXISTS", "user:1", "missing"), 0)
	requireHashPairs(t, sendCommand(t, rw, "HGETALL", "user:1"), map[string]string{
		"name": "alice",
		"age":  "20",
	})
}

func TestHash_HDel_ExistingAndNonExistent(t *testing.T) {
	addr, cleanup := startStringTestServer(t)
	defer cleanup()

	rw := connectStringClient(t, addr)

	requireInteger(t, sendCommand(t, rw, "HSET", "user:2", "name", "bob"), 1)
	requireInteger(t, sendCommand(t, rw, "HDEL", "user:2", "name"), 1)
	requireInteger(t, sendCommand(t, rw, "HDEL", "user:2", "name"), 0)
	requireInteger(t, sendCommand(t, rw, "HDEL", "missing", "name"), 0)
	requireNilBulk(t, sendCommand(t, rw, "HGET", "user:2", "name"))
}

func TestHash_NonExistentKey(t *testing.T) {
	addr, cleanup := startStringTestServer(t)
	defer cleanup()

	rw := connectStringClient(t, addr)

	requireNilBulk(t, sendCommand(t, rw, "HGET", "missing", "f"))
	requireInteger(t, sendCommand(t, rw, "HEXISTS", "missing", "f"), 0)
	requireHashPairs(t, sendCommand(t, rw, "HGETALL", "missing"), map[string]string{})
}

func TestHash_WrongType(t *testing.T) {
	addr, cleanup := startStringTestServer(t)
	defer cleanup()

	rw := connectStringClient(t, addr)

	requireSimpleOK(t, sendCommand(t, rw, "SET", "k", "v"))

	for _, args := range [][]string{
		{"HSET", "k", "f", "x"},
		{"HGET", "k", "f"},
		{"HDEL", "k", "f"},
		{"HGETALL", "k"},
		{"HEXISTS", "k", "f"},
	} {
		t.Run(args[0], func(t *testing.T) {
			requireErrorContains(t, sendCommand(t, rw, args...), "WRONGTYPE")
		})
	}
}

func TestHash_CommandErrors(t *testing.T) {
	addr, cleanup := startStringTestServer(t)
	defer cleanup()

	rw := connectStringClient(t, addr)

	errCases := []struct {
		args []string
		want string
	}{
		{args: []string{"HSET", "k", "f"}, want: "wrong number of arguments for 'HSET' command"},
		{args: []string{"HGET", "k"}, want: "wrong number of arguments for 'HGET' command"},
		{args: []string{"HDEL", "k"}, want: "wrong number of arguments for 'HDEL' command"},
		{args: []string{"HGETALL"}, want: "wrong number of arguments for 'HGETALL' command"},
		{args: []string{"HEXISTS", "k"}, want: "wrong number of arguments for 'HEXISTS' command"},
	}

	for _, tc := range errCases {
		t.Run(fmt.Sprintf("%s_%s", tc.args[0], strings.Join(tc.args[1:], "_")), func(t *testing.T) {
			requireErrorContains(t, sendCommand(t, rw, tc.args...), tc.want)
		})
	}
}
