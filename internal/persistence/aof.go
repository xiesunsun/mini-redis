package persistence

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
	"sync"

	"github.com/xiesunsun/mini-redis/internal/types"
)

// AOF 实现 Append-Only File 持久化。
type AOF struct {
	file *os.File
	mu   sync.Mutex
}

// New 打开（或创建）指定路径的 AOF 文件，返回 AOF 实例。
func New(path string) (*AOF, error) {
	f, err := os.OpenFile(path, os.O_CREATE|os.O_RDWR|os.O_APPEND, 0644)
	if err != nil {
		return nil, err
	}
	return &AOF{file: f}, nil
}

// WriteCommand 将一条命令以 RESP 格式追加到 AOF 文件。
func (a *AOF) WriteCommand(cmd types.Command) error {
	a.mu.Lock()
	defer a.mu.Unlock()

	args := append([]string{cmd.Name}, cmd.Args...)
	if _, err := fmt.Fprintf(a.file, "*%d\r\n", len(args)); err != nil {
		return err
	}
	for _, arg := range args {
		if _, err := fmt.Fprintf(a.file, "$%d\r\n%s\r\n", len(arg), arg); err != nil {
			return err
		}
	}
	return nil
}

// Replay 读取 AOF 文件中的所有命令并返回。
func (a *AOF) Replay() ([]types.Command, error) {
	a.mu.Lock()
	defer a.mu.Unlock()

	if _, err := a.file.Seek(0, 0); err != nil {
		return nil, err
	}

	reader := bufio.NewReader(a.file)
	var cmds []types.Command
	for {
		cmd, err := readAOFCommand(reader)
		if err != nil {
			if errors.Is(err, io.EOF) {
				break
			}
			return nil, err
		}
		cmds = append(cmds, cmd)
	}
	return cmds, nil
}

func readAOFCommand(r *bufio.Reader) (types.Command, error) {
	header, err := readRESPLine(r)
	if err != nil {
		return types.Command{}, err
	}
	if !strings.HasPrefix(header, "*") {
		return types.Command{}, fmt.Errorf("invalid AOF format: expected array header, got %q", header)
	}

	argc, err := strconv.Atoi(header[1:])
	if err != nil {
		return types.Command{}, fmt.Errorf("invalid argument count %q", header[1:])
	}
	if argc <= 0 {
		return types.Command{}, fmt.Errorf("invalid argument count %d", argc)
	}

	args := make([]string, 0, argc)
	for i := 0; i < argc; i++ {
		bulkHeader, err := readRESPLine(r)
		if err != nil {
			return types.Command{}, fmt.Errorf("read bulk[%d] header: %w", i, err)
		}
		if !strings.HasPrefix(bulkHeader, "$") {
			return types.Command{}, fmt.Errorf("invalid bulk header %q", bulkHeader)
		}

		n, err := strconv.Atoi(bulkHeader[1:])
		if err != nil {
			return types.Command{}, fmt.Errorf("invalid bulk length %q", bulkHeader[1:])
		}
		if n < 0 {
			return types.Command{}, fmt.Errorf("invalid bulk length %d", n)
		}

		buf := make([]byte, n+2)
		if _, err := io.ReadFull(r, buf); err != nil {
			return types.Command{}, fmt.Errorf("read bulk[%d] body: %w", i, err)
		}
		if buf[n] != '\r' || buf[n+1] != '\n' {
			return types.Command{}, fmt.Errorf("bulk[%d] missing CRLF terminator", i)
		}
		args = append(args, string(buf[:n]))
	}

	return types.Command{
		Name: strings.ToUpper(args[0]),
		Args: args[1:],
	}, nil
}

func readRESPLine(r *bufio.Reader) (string, error) {
	line, err := r.ReadString('\n')
	if err != nil {
		if errors.Is(err, io.EOF) {
			if len(line) == 0 {
				return "", io.EOF
			}
			return "", io.ErrUnexpectedEOF
		}
		return "", err
	}
	if len(line) < 2 || !strings.HasSuffix(line, "\r\n") {
		return "", errors.New("line missing CRLF terminator")
	}
	return strings.TrimSuffix(line, "\r\n"), nil
}

// Close 关闭 AOF 文件。
func (a *AOF) Close() error {
	a.mu.Lock()
	defer a.mu.Unlock()
	return a.file.Close()
}
