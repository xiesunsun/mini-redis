package persistence

import (
	"bufio"
	"fmt"
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

	var cmds []types.Command
	scanner := bufio.NewScanner(a.file)
	for scanner.Scan() {
		line := scanner.Text()
		if !strings.HasPrefix(line, "*") {
			return nil, fmt.Errorf("invalid AOF format: expected '*', got %q", line)
		}
		argc, err := strconv.Atoi(line[1:])
		if err != nil {
			return nil, fmt.Errorf("invalid argument count: %v", err)
		}

		args := make([]string, argc)
		for i := 0; i < argc; i++ {
			if !scanner.Scan() {
				return nil, fmt.Errorf("unexpected EOF reading bulk length")
			}
			bulkLine := scanner.Text()
			if !strings.HasPrefix(bulkLine, "$") {
				return nil, fmt.Errorf("invalid bulk format: expected '$', got %q", bulkLine)
			}
			if !scanner.Scan() {
				return nil, fmt.Errorf("unexpected EOF reading bulk data")
			}
			args[i] = scanner.Text()
		}

		cmds = append(cmds, types.Command{
			Name: strings.ToUpper(args[0]),
			Args: args[1:],
		})
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}
	return cmds, nil
}

// Close 关闭 AOF 文件。
func (a *AOF) Close() error {
	a.mu.Lock()
	defer a.mu.Unlock()
	return a.file.Close()
}
