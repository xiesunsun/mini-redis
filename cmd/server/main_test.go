package main

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/xiesunsun/mini-redis/internal/network"
	"github.com/xiesunsun/mini-redis/internal/persistence"
	"github.com/xiesunsun/mini-redis/internal/types"
)

func encodeCommand(args ...string) string {
	var b strings.Builder
	fmt.Fprintf(&b, "*%d\r\n", len(args))
	for _, arg := range args {
		fmt.Fprintf(&b, "$%d\r\n%s\r\n", len(arg), arg)
	}
	return b.String()
}

func sendCommand(t *testing.T, conn net.Conn, args ...string) network.Value {
	t.Helper()

	rw := bufio.NewReadWriter(bufio.NewReader(conn), bufio.NewWriter(conn))
	if _, err := rw.WriteString(encodeCommand(args...)); err != nil {
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

func createTempAOFPath(t *testing.T) string {
	t.Helper()
	f, err := os.CreateTemp("", "main_test_*.aof")
	if err != nil {
		t.Fatalf("create temp file failed: %v", err)
	}
	path := f.Name()
	if err := f.Close(); err != nil {
		t.Fatalf("close temp file failed: %v", err)
	}
	t.Cleanup(func() {
		_ = os.Remove(path)
	})
	return path
}

func TestBuildServer_ReplayAOF(t *testing.T) {
	aofPath := createTempAOFPath(t)
	aof, err := persistence.New(aofPath)
	if err != nil {
		t.Fatalf("open temp AOF failed: %v", err)
	}
	if err := aof.WriteCommand(types.Command{Name: "SET", Args: []string{"replay-key", "replay-value"}}); err != nil {
		t.Fatalf("write AOF command failed: %v", err)
	}
	if err := aof.Close(); err != nil {
		t.Fatalf("close temp AOF failed: %v", err)
	}

	srv, cleanup, err := buildServer("127.0.0.1:0", aofPath, time.Second)
	if err != nil {
		t.Fatalf("build server failed: %v", err)
	}
	defer func() {
		_ = cleanup()
	}()

	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen failed: %v", err)
	}
	defer ln.Close()

	errCh := make(chan error, 1)
	go func() {
		errCh <- srv.Serve(ln)
	}()

	conn, err := net.Dial("tcp", ln.Addr().String())
	if err != nil {
		t.Fatalf("dial failed: %v", err)
	}
	defer conn.Close()

	resp := sendCommand(t, conn, "GET", "replay-key")
	if resp.Type != network.RespBulkString || resp.String != "replay-value" {
		t.Fatalf("unexpected GET response after replay: %+v", resp)
	}

	if err := cleanup(); err != nil {
		t.Fatalf("cleanup failed: %v", err)
	}
	select {
	case serveErr := <-errCh:
		if serveErr != nil {
			t.Fatalf("serve returned error: %v", serveErr)
		}
	case <-time.After(time.Second):
		t.Fatal("server did not stop in time")
	}
}

func TestBuildServer_InvalidAOF(t *testing.T) {
	aofPath := createTempAOFPath(t)
	if err := os.WriteFile(aofPath, []byte("invalid-aof"), 0644); err != nil {
		t.Fatalf("write invalid AOF failed: %v", err)
	}

	_, _, err := buildServer("127.0.0.1:0", aofPath, time.Second)
	if err == nil {
		t.Fatal("expected buildServer to fail with invalid AOF")
	}
}
