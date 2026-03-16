package tests

import (
	"bufio"
	"fmt"
	"net"
	"strings"
	"testing"
	"time"

	"github.com/xiesunsun/mini-redis/internal/command"
	"github.com/xiesunsun/mini-redis/internal/network"
	"github.com/xiesunsun/mini-redis/internal/store"
)

func startStringTestServer(t *testing.T) (string, func()) {
	t.Helper()

	srv := network.NewServer("127.0.0.1:0", &command.Context{Store: store.New()})
	ln, err := net.Listen("tcp", srv.Addr())
	if err != nil {
		t.Fatalf("listen failed: %v", err)
	}

	errCh := make(chan error, 1)
	go func() {
		errCh <- srv.Serve(ln)
	}()

	cleanup := func() {
		t.Helper()
		if err := srv.Close(); err != nil {
			t.Fatalf("close server failed: %v", err)
		}

		select {
		case err := <-errCh:
			if err != nil {
				t.Fatalf("serve returned error: %v", err)
			}
		case <-time.After(time.Second):
			t.Fatalf("server did not stop in time")
		}
	}

	return ln.Addr().String(), cleanup
}

func connectStringClient(t *testing.T, addr string) *bufio.ReadWriter {
	t.Helper()

	conn, err := net.Dial("tcp", addr)
	if err != nil {
		t.Fatalf("dial failed: %v", err)
	}

	t.Cleanup(func() {
		_ = conn.Close()
	})

	return bufio.NewReadWriter(bufio.NewReader(conn), bufio.NewWriter(conn))
}

func encodeRESPCommand(args ...string) (string, error) {
	arr := make([]network.Value, len(args))
	for i, arg := range args {
		arr[i] = network.Value{Type: network.RespBulkString, String: arg}
	}
	return network.Serialize(network.Value{Type: network.RespArray, Array: arr})
}

func sendCommand(t *testing.T, rw *bufio.ReadWriter, args ...string) network.Value {
	t.Helper()

	raw, err := encodeRESPCommand(args...)
	if err != nil {
		t.Fatalf("encode command failed: %v", err)
	}

	if _, err := rw.WriteString(raw); err != nil {
		t.Fatalf("write command failed: %v", err)
	}
	if err := rw.Flush(); err != nil {
		t.Fatalf("flush command failed: %v", err)
	}

	resp, err := network.Parse(rw.Reader)
	if err != nil {
		t.Fatalf("parse response failed: %v", err)
	}
	return resp
}

func requireSimpleOK(t *testing.T, resp network.Value) {
	t.Helper()
	if resp.Type != network.RespSimpleString || resp.String != "OK" {
		t.Fatalf("unexpected simple string response: %+v", resp)
	}
}

func requireInteger(t *testing.T, resp network.Value, want int64) {
	t.Helper()
	if resp.Type != network.RespInteger || resp.Integer != want {
		t.Fatalf("unexpected integer response: got %+v, want %d", resp, want)
	}
}

func requireBulkString(t *testing.T, resp network.Value, want string) {
	t.Helper()
	if resp.Type != network.RespBulkString || resp.IsNull || resp.String != want {
		t.Fatalf("unexpected bulk string response: got %+v, want %q", resp, want)
	}
}

func requireNilBulk(t *testing.T, resp network.Value) {
	t.Helper()
	if resp.Type != network.RespBulkString || !resp.IsNull {
		t.Fatalf("unexpected nil bulk response: %+v", resp)
	}
}

func requireErrorContains(t *testing.T, resp network.Value, want string) {
	t.Helper()
	if resp.Type != network.RespError || !strings.Contains(resp.String, want) {
		t.Fatalf("unexpected error response: got %+v, want contains %q", resp, want)
	}
}

func TestString_SET_GET_NormalPath(t *testing.T) {
	addr, cleanup := startStringTestServer(t)
	defer cleanup()

	rw := connectStringClient(t, addr)

	requireSimpleOK(t, sendCommand(t, rw, "SET", "name", "alice"))
	requireBulkString(t, sendCommand(t, rw, "GET", "name"), "alice")
}

func TestString_SET_EmptyValue(t *testing.T) {
	addr, cleanup := startStringTestServer(t)
	defer cleanup()

	rw := connectStringClient(t, addr)

	requireSimpleOK(t, sendCommand(t, rw, "SET", "empty", ""))
	requireBulkString(t, sendCommand(t, rw, "GET", "empty"), "")
}

func TestString_GET_NonExistentKey(t *testing.T) {
	addr, cleanup := startStringTestServer(t)
	defer cleanup()

	rw := connectStringClient(t, addr)

	requireNilBulk(t, sendCommand(t, rw, "GET", "missing"))
}

func TestString_DEL_ExistingAndNonExistent(t *testing.T) {
	addr, cleanup := startStringTestServer(t)
	defer cleanup()

	rw := connectStringClient(t, addr)

	requireSimpleOK(t, sendCommand(t, rw, "SET", "k", "v"))
	requireInteger(t, sendCommand(t, rw, "DEL", "k"), 1)
	requireInteger(t, sendCommand(t, rw, "DEL", "k"), 0)
	requireNilBulk(t, sendCommand(t, rw, "GET", "k"))
}

func TestString_EXPIRE_And_TTL(t *testing.T) {
	addr, cleanup := startStringTestServer(t)
	defer cleanup()

	rw := connectStringClient(t, addr)

	requireSimpleOK(t, sendCommand(t, rw, "SET", "session", "abc"))
	requireInteger(t, sendCommand(t, rw, "TTL", "session"), -1)
	requireInteger(t, sendCommand(t, rw, "EXPIRE", "session", "1"), 1)

	ttlResp := sendCommand(t, rw, "TTL", "session")
	if ttlResp.Type != network.RespInteger || ttlResp.Integer < 0 || ttlResp.Integer > 1 {
		t.Fatalf("unexpected ttl response: %+v", ttlResp)
	}

	time.Sleep(1100 * time.Millisecond)
	requireNilBulk(t, sendCommand(t, rw, "GET", "session"))
	requireInteger(t, sendCommand(t, rw, "TTL", "session"), -2)
}

func TestString_EXPIRE_BoundaryCases(t *testing.T) {
	addr, cleanup := startStringTestServer(t)
	defer cleanup()

	rw := connectStringClient(t, addr)

	requireInteger(t, sendCommand(t, rw, "EXPIRE", "missing", "10"), 0)

	requireSimpleOK(t, sendCommand(t, rw, "SET", "temp", "v"))
	requireErrorContains(t, sendCommand(t, rw, "EXPIRE", "temp", "NaN"), "value is not an integer")
	requireInteger(t, sendCommand(t, rw, "EXPIRE", "temp", "-1"), 1)
	requireInteger(t, sendCommand(t, rw, "TTL", "temp"), -2)
}

func TestString_CommandArityErrors(t *testing.T) {
	addr, cleanup := startStringTestServer(t)
	defer cleanup()

	rw := connectStringClient(t, addr)

	errCases := []struct {
		args []string
		want string
	}{
		{args: []string{"SET", "k"}, want: "wrong number of arguments for 'SET' command"},
		{args: []string{"GET"}, want: "wrong number of arguments for 'GET' command"},
		{args: []string{"DEL"}, want: "wrong number of arguments for 'DEL' command"},
		{args: []string{"EXPIRE", "k"}, want: "wrong number of arguments for 'EXPIRE' command"},
		{args: []string{"TTL"}, want: "wrong number of arguments for 'TTL' command"},
	}

	for _, tc := range errCases {
		t.Run(fmt.Sprintf("%s_arity", tc.args[0]), func(t *testing.T) {
			requireErrorContains(t, sendCommand(t, rw, tc.args...), tc.want)
		})
	}
}

func TestRedisCompat_ExtraArgs_ReturnWrongNumberOfArguments(t *testing.T) {
	addr, cleanup := startStringTestServer(t)
	defer cleanup()

	rw := connectStringClient(t, addr)

	errCases := []struct {
		args []string
		want string
	}{
		{args: []string{"SET", "k", "v", "extra"}, want: "wrong number of arguments for 'SET' command"},
		{args: []string{"GET", "k", "extra"}, want: "wrong number of arguments for 'GET' command"},
		{args: []string{"EXPIRE", "k", "10", "extra"}, want: "wrong number of arguments for 'EXPIRE' command"},
		{args: []string{"LPUSH", "list", "v", "extra"}, want: "wrong number of arguments for 'LPUSH' command"},
		{args: []string{"RPUSH", "list", "v", "extra"}, want: "wrong number of arguments for 'RPUSH' command"},
		{args: []string{"HSET", "h", "f", "v", "extra"}, want: "wrong number of arguments for 'HSET' command"},
	}

	for _, tc := range errCases {
		t.Run(fmt.Sprintf("%s_extra_args", tc.args[0]), func(t *testing.T) {
			requireErrorContains(t, sendCommand(t, rw, tc.args...), tc.want)
		})
	}
}
