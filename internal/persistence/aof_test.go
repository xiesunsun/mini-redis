package persistence

import (
	"os"
	"strings"
	"testing"

	"github.com/xiesunsun/mini-redis/internal/types"
)

func tempAOF(t *testing.T) (*AOF, string) {
	t.Helper()
	f, err := os.CreateTemp("", "aof_test_*.aof")
	if err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}
	path := f.Name()
	f.Close()

	aof, err := New(path)
	if err != nil {
		os.Remove(path)
		t.Fatalf("failed to open AOF: %v", err)
	}
	t.Cleanup(func() {
		aof.Close()
		os.Remove(path)
	})
	return aof, path
}

func tempAOFWithRawContent(t *testing.T, content string) *AOF {
	t.Helper()

	f, err := os.CreateTemp("", "aof_test_*.aof")
	if err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}
	if _, err := f.WriteString(content); err != nil {
		f.Close()
		os.Remove(f.Name())
		t.Fatalf("failed to write temp file: %v", err)
	}
	path := f.Name()
	if err := f.Close(); err != nil {
		os.Remove(path)
		t.Fatalf("failed to close temp file: %v", err)
	}

	aof, err := New(path)
	if err != nil {
		os.Remove(path)
		t.Fatalf("failed to open AOF: %v", err)
	}
	t.Cleanup(func() {
		aof.Close()
		os.Remove(path)
	})
	return aof
}

func TestReplay_EmptyFile(t *testing.T) {
	aof, _ := tempAOF(t)
	cmds, err := aof.Replay()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(cmds) != 0 {
		t.Fatalf("expected 0 commands, got %d", len(cmds))
	}
}

func TestWriteCommand_SingleCommand(t *testing.T) {
	aof, _ := tempAOF(t)
	cmd := types.Command{Name: "SET", Args: []string{"foo", "bar"}}
	if err := aof.WriteCommand(cmd); err != nil {
		t.Fatalf("WriteCommand error: %v", err)
	}

	cmds, err := aof.Replay()
	if err != nil {
		t.Fatalf("Replay error: %v", err)
	}
	if len(cmds) != 1 {
		t.Fatalf("expected 1 command, got %d", len(cmds))
	}
	if cmds[0].Name != "SET" || len(cmds[0].Args) != 2 || cmds[0].Args[0] != "foo" || cmds[0].Args[1] != "bar" {
		t.Errorf("unexpected command: %+v", cmds[0])
	}
}

func TestWriteAndReplay_MultipleCommands(t *testing.T) {
	aof, _ := tempAOF(t)

	inputs := []types.Command{
		{Name: "SET", Args: []string{"key1", "value1"}},
		{Name: "LPUSH", Args: []string{"mylist", "elem"}},
		{Name: "HSET", Args: []string{"myhash", "field", "val"}},
		{Name: "DEL", Args: []string{"key1"}},
	}
	for _, cmd := range inputs {
		if err := aof.WriteCommand(cmd); err != nil {
			t.Fatalf("WriteCommand(%v) error: %v", cmd.Name, err)
		}
	}

	cmds, err := aof.Replay()
	if err != nil {
		t.Fatalf("Replay error: %v", err)
	}
	if len(cmds) != len(inputs) {
		t.Fatalf("expected %d commands, got %d", len(inputs), len(cmds))
	}
	for i, want := range inputs {
		got := cmds[i]
		if got.Name != want.Name {
			t.Errorf("cmd[%d].Name: want %q, got %q", i, want.Name, got.Name)
		}
		if len(got.Args) != len(want.Args) {
			t.Errorf("cmd[%d].Args length: want %d, got %d", i, len(want.Args), len(got.Args))
			continue
		}
		for j := range want.Args {
			if got.Args[j] != want.Args[j] {
				t.Errorf("cmd[%d].Args[%d]: want %q, got %q", i, j, want.Args[j], got.Args[j])
			}
		}
	}
}

func TestReplay_CommandNameUppercased(t *testing.T) {
	aof, _ := tempAOF(t)
	// 写入小写命令名，Replay 应返回大写
	cmd := types.Command{Name: "set", Args: []string{"k", "v"}}
	if err := aof.WriteCommand(cmd); err != nil {
		t.Fatalf("WriteCommand error: %v", err)
	}
	cmds, err := aof.Replay()
	if err != nil {
		t.Fatalf("Replay error: %v", err)
	}
	if len(cmds) != 1 || cmds[0].Name != "SET" {
		t.Errorf("expected command name SET, got %q", cmds[0].Name)
	}
}

func TestReplay_MultipleReplays(t *testing.T) {
	aof, _ := tempAOF(t)
	cmd := types.Command{Name: "GET", Args: []string{"foo"}}
	if err := aof.WriteCommand(cmd); err != nil {
		t.Fatalf("WriteCommand error: %v", err)
	}

	// 多次 Replay 应返回相同结果
	for i := 0; i < 3; i++ {
		cmds, err := aof.Replay()
		if err != nil {
			t.Fatalf("Replay[%d] error: %v", i, err)
		}
		if len(cmds) != 1 || cmds[0].Name != "GET" {
			t.Errorf("Replay[%d]: unexpected result %+v", i, cmds)
		}
	}
}

func TestReplay_LargePayload(t *testing.T) {
	aof, _ := tempAOF(t)
	largeValue := strings.Repeat("x", 128*1024)

	if err := aof.WriteCommand(types.Command{Name: "SET", Args: []string{"k", largeValue}}); err != nil {
		t.Fatalf("WriteCommand error: %v", err)
	}

	cmds, err := aof.Replay()
	if err != nil {
		t.Fatalf("Replay error: %v", err)
	}
	if len(cmds) != 1 {
		t.Fatalf("expected 1 command, got %d", len(cmds))
	}
	if cmds[0].Name != "SET" || len(cmds[0].Args) != 2 {
		t.Fatalf("unexpected command shape: %+v", cmds[0])
	}
	if cmds[0].Args[1] != largeValue {
		t.Fatalf("payload mismatch: want length %d, got %d", len(largeValue), len(cmds[0].Args[1]))
	}
}

func TestReplay_InvalidRESP(t *testing.T) {
	cases := []struct {
		name string
		raw  string
	}{
		{
			name: "invalid array prefix",
			raw:  "$1\r\na\r\n",
		},
		{
			name: "invalid array length",
			raw:  "*x\r\n",
		},
		{
			name: "zero argc",
			raw:  "*0\r\n",
		},
		{
			name: "invalid bulk header prefix",
			raw:  "*1\r\n!3\r\nSET\r\n",
		},
		{
			name: "invalid bulk length",
			raw:  "*1\r\n$abc\r\nSET\r\n",
		},
		{
			name: "negative bulk length",
			raw:  "*1\r\n$-2\r\n",
		},
		{
			name: "truncated bulk body",
			raw:  "*2\r\n$3\r\nSET\r\n$5\r\nabc\r\n",
		},
		{
			name: "bulk missing CRLF terminator at boundary",
			raw:  "*2\r\n$3\r\nSET\r\n$5\r\nabcdex\r\n",
		},
		{
			name: "line without CRLF",
			raw:  "*1\r\n$3\r\nSET\n",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			aof := tempAOFWithRawContent(t, tc.raw)
			if _, err := aof.Replay(); err == nil {
				t.Fatal("expected Replay to fail, got nil error")
			}
		})
	}
}
