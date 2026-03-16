package network

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"net"
	"sync"
	"sync/atomic"

	"github.com/xiesunsun/mini-redis/internal/command"
)

const defaultListenAddr = ":6379"

// Server is a TCP server that handles RESP requests and dispatches commands.
type Server struct {
	addr   string
	router *command.Router

	mu       sync.Mutex
	listener net.Listener

	conns    sync.Map
	wg       sync.WaitGroup
	once     sync.Once
	closeErr error
	closed   atomic.Bool
}

// NewServer creates a network server with command context.
func NewServer(addr string, ctx *command.Context) *Server {
	if addr == "" {
		addr = defaultListenAddr
	}

	return &Server{
		addr:   addr,
		router: command.NewRouter(ctx),
	}
}

// Addr returns configured listening address. After Serve starts, returns actual listener address.
func (s *Server) Addr() string {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.listener != nil {
		return s.listener.Addr().String()
	}
	return s.addr
}

// ListenAndServe starts listening and serves incoming connections.
func (s *Server) ListenAndServe() error {
	ln, err := net.Listen("tcp", s.addr)
	if err != nil {
		return err
	}
	return s.Serve(ln)
}

// Serve accepts connections from an existing listener.
func (s *Server) Serve(listener net.Listener) error {
	if listener == nil {
		return errors.New("listener is nil")
	}

	s.mu.Lock()
	s.listener = listener
	s.mu.Unlock()

	for {
		conn, err := listener.Accept()
		if err != nil {
			if s.closed.Load() || errors.Is(err, net.ErrClosed) {
				return nil
			}
			var netErr net.Error
			if errors.As(err, &netErr) && netErr.Temporary() {
				continue
			}
			return err
		}

		s.wg.Add(1)
		go func() {
			defer s.wg.Done()
			s.handleConnection(conn)
		}()
	}
}

// Close stops accepting new connections and waits running handlers to finish.
func (s *Server) Close() error {
	s.once.Do(func() {
		s.closed.Store(true)

		s.mu.Lock()
		ln := s.listener
		s.mu.Unlock()

		if ln != nil {
			s.closeErr = ln.Close()
		}
		s.conns.Range(func(key, _ any) bool {
			if conn, ok := key.(net.Conn); ok {
				_ = conn.Close()
			}
			return true
		})
		s.wg.Wait()
	})

	return s.closeErr
}

func (s *Server) handleConnection(conn net.Conn) {
	defer conn.Close()
	s.conns.Store(conn, struct{}{})
	defer s.conns.Delete(conn)

	reader := bufio.NewReader(conn)
	for {
		value, err := Parse(reader)
		if err != nil {
			if errors.Is(err, io.EOF) || errors.Is(err, net.ErrClosed) {
				return
			}
			_, _ = io.WriteString(conn, respErr("protocol error: "+err.Error()))
			return
		}

		name, args, err := parseCommand(value)
		if err != nil {
			_, _ = io.WriteString(conn, respErr(err.Error()))
			continue
		}

		response := s.router.DispatchParts(name, args)
		if _, err := io.WriteString(conn, response); err != nil {
			return
		}
	}
}

func parseCommand(value Value) (string, []string, error) {
	if value.Type != RespArray || value.IsNull {
		return "", nil, errors.New("protocol error: expected array request")
	}
	if len(value.Array) == 0 {
		return "", nil, errors.New("protocol error: empty command")
	}

	args := make([]string, len(value.Array))
	for i, item := range value.Array {
		if item.Type != RespBulkString || item.IsNull {
			return "", nil, fmt.Errorf("protocol error: argument %d must be non-null bulk string", i)
		}
		args[i] = item.String
	}

	return args[0], args[1:], nil
}

func respErr(msg string) string {
	return fmt.Sprintf("-ERR %s\r\n", msg)
}
