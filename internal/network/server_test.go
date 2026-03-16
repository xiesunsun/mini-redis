package network

import (
	"bufio"
	"fmt"
	"net"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/xiesunsun/mini-redis/internal/command"
	"github.com/xiesunsun/mini-redis/internal/store"
)

func startTestServer(t *testing.T) (*Server, string, func()) {
	t.Helper()

	srv := NewServer("127.0.0.1:0", &command.Context{Store: store.New()})
	ln, err := net.Listen("tcp", srv.addr)
	if err != nil {
		t.Fatalf("listen failed: %v", err)
	}

	errCh := make(chan error, 1)
	go func() {
		errCh <- srv.Serve(ln)
	}()

	cleanup := func() {
		t.Helper()
		_ = srv.Close()

		select {
		case err := <-errCh:
			if err != nil {
				t.Fatalf("serve returned error: %v", err)
			}
		case <-time.After(time.Second):
			t.Fatalf("server did not stop in time")
		}
	}

	return srv, ln.Addr().String(), cleanup
}

func encodeCommand(args ...string) string {
	var b strings.Builder
	fmt.Fprintf(&b, "*%d\r\n", len(args))
	for _, arg := range args {
		fmt.Fprintf(&b, "$%d\r\n%s\r\n", len(arg), arg)
	}
	return b.String()
}

func sendAndRead(rw *bufio.ReadWriter, args ...string) (Value, error) {
	if _, err := rw.WriteString(encodeCommand(args...)); err != nil {
		return Value{}, err
	}
	if err := rw.Flush(); err != nil {
		return Value{}, err
	}
	return Parse(rw.Reader)
}

func TestServer_HandleCommandsInSingleConnection(t *testing.T) {
	_, addr, cleanup := startTestServer(t)
	defer cleanup()

	conn, err := net.Dial("tcp", addr)
	if err != nil {
		t.Fatalf("dial failed: %v", err)
	}
	defer conn.Close()

	rw := bufio.NewReadWriter(bufio.NewReader(conn), bufio.NewWriter(conn))

	setResp, err := sendAndRead(rw, "SET", "name", "alice")
	if err != nil {
		t.Fatalf("SET failed: %v", err)
	}
	if setResp.Type != RespSimpleString || setResp.String != "OK" {
		t.Fatalf("unexpected SET response: %+v", setResp)
	}

	getResp, err := sendAndRead(rw, "GET", "name")
	if err != nil {
		t.Fatalf("GET failed: %v", err)
	}
	if getResp.Type != RespBulkString || getResp.String != "alice" {
		t.Fatalf("unexpected GET response: %+v", getResp)
	}
}

func TestServer_InvalidArgumentType_ConnectionKeepsAlive(t *testing.T) {
	_, addr, cleanup := startTestServer(t)
	defer cleanup()

	conn, err := net.Dial("tcp", addr)
	if err != nil {
		t.Fatalf("dial failed: %v", err)
	}
	defer conn.Close()

	rw := bufio.NewReadWriter(bufio.NewReader(conn), bufio.NewWriter(conn))

	if _, err := rw.WriteString("*1\r\n+PING\r\n"); err != nil {
		t.Fatalf("write invalid command failed: %v", err)
	}
	if err := rw.Flush(); err != nil {
		t.Fatalf("flush invalid command failed: %v", err)
	}

	errResp, err := Parse(rw.Reader)
	if err != nil {
		t.Fatalf("read invalid command response failed: %v", err)
	}
	if errResp.Type != RespError || !strings.Contains(errResp.String, "protocol error") {
		t.Fatalf("unexpected protocol error response: %+v", errResp)
	}

	setResp, err := sendAndRead(rw, "SET", "k", "v")
	if err != nil {
		t.Fatalf("SET after invalid request failed: %v", err)
	}
	if setResp.Type != RespSimpleString || setResp.String != "OK" {
		t.Fatalf("unexpected SET response: %+v", setResp)
	}
}

func TestServer_ConcurrentClients(t *testing.T) {
	_, addr, cleanup := startTestServer(t)
	defer cleanup()

	const clients = 8

	var wg sync.WaitGroup
	errCh := make(chan error, clients)

	for i := 0; i < clients; i++ {
		i := i
		wg.Add(1)
		go func() {
			defer wg.Done()

			conn, err := net.Dial("tcp", addr)
			if err != nil {
				errCh <- fmt.Errorf("dial client %d failed: %w", i, err)
				return
			}
			defer conn.Close()

			rw := bufio.NewReadWriter(bufio.NewReader(conn), bufio.NewWriter(conn))
			key := fmt.Sprintf("key-%d", i)
			value := fmt.Sprintf("value-%d", i)

			resp, err := sendAndRead(rw, "SET", key, value)
			if err != nil {
				errCh <- fmt.Errorf("SET client %d failed: %w", i, err)
				return
			}
			if resp.Type != RespSimpleString || resp.String != "OK" {
				errCh <- fmt.Errorf("SET client %d unexpected response: %+v", i, resp)
				return
			}

			resp, err = sendAndRead(rw, "GET", key)
			if err != nil {
				errCh <- fmt.Errorf("GET client %d failed: %w", i, err)
				return
			}
			if resp.Type != RespBulkString || resp.String != value {
				errCh <- fmt.Errorf("GET client %d unexpected response: %+v", i, resp)
				return
			}
		}()
	}

	wg.Wait()
	close(errCh)

	for err := range errCh {
		t.Fatal(err)
	}
}
