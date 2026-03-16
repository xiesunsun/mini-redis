package network

import (
	"bufio"
	"strings"
	"testing"
)

func TestParse_SimpleString(t *testing.T) {
	v, err := Parse("+OK\r\n")
	if err != nil {
		t.Fatalf("parse failed: %v", err)
	}
	if v.Type != RespSimpleString || v.String != "OK" {
		t.Fatalf("unexpected value: %+v", v)
	}
}

func TestParse_Error(t *testing.T) {
	v, err := Parse("-ERR wrong type\r\n")
	if err != nil {
		t.Fatalf("parse failed: %v", err)
	}
	if v.Type != RespError || v.String != "ERR wrong type" {
		t.Fatalf("unexpected value: %+v", v)
	}
}

func TestParse_Integer(t *testing.T) {
	v, err := Parse(":123\r\n")
	if err != nil {
		t.Fatalf("parse failed: %v", err)
	}
	if v.Type != RespInteger || v.Integer != 123 {
		t.Fatalf("unexpected value: %+v", v)
	}
}

func TestParse_BulkString(t *testing.T) {
	v, err := Parse("$5\r\nhello\r\n")
	if err != nil {
		t.Fatalf("parse failed: %v", err)
	}
	if v.Type != RespBulkString || v.String != "hello" || v.IsNull {
		t.Fatalf("unexpected value: %+v", v)
	}
}

func TestParse_NilBulkString(t *testing.T) {
	v, err := Parse("$-1\r\n")
	if err != nil {
		t.Fatalf("parse failed: %v", err)
	}
	if v.Type != RespBulkString || !v.IsNull {
		t.Fatalf("unexpected value: %+v", v)
	}
}

func TestParse_Array(t *testing.T) {
	raw := "*3\r\n+OK\r\n:7\r\n$3\r\nhey\r\n"
	v, err := Parse(raw)
	if err != nil {
		t.Fatalf("parse failed: %v", err)
	}
	if v.Type != RespArray {
		t.Fatalf("expected array, got %+v", v)
	}
	if len(v.Array) != 3 {
		t.Fatalf("expected 3 items, got %d", len(v.Array))
	}
	if v.Array[0].Type != RespSimpleString || v.Array[0].String != "OK" {
		t.Fatalf("unexpected first item: %+v", v.Array[0])
	}
	if v.Array[1].Type != RespInteger || v.Array[1].Integer != 7 {
		t.Fatalf("unexpected second item: %+v", v.Array[1])
	}
	if v.Array[2].Type != RespBulkString || v.Array[2].String != "hey" {
		t.Fatalf("unexpected third item: %+v", v.Array[2])
	}
}

func TestParse_NilArray(t *testing.T) {
	v, err := Parse("*-1\r\n")
	if err != nil {
		t.Fatalf("parse failed: %v", err)
	}
	if v.Type != RespArray || !v.IsNull {
		t.Fatalf("unexpected value: %+v", v)
	}
}

func TestParse_ReaderInput(t *testing.T) {
	reader := bufio.NewReader(strings.NewReader("+PONG\r\n"))
	v, err := Parse(reader)
	if err != nil {
		t.Fatalf("parse failed: %v", err)
	}
	if v.Type != RespSimpleString || v.String != "PONG" {
		t.Fatalf("unexpected value: %+v", v)
	}
}

func TestSerialize_AllTypes(t *testing.T) {
	tests := []struct {
		name string
		in   Value
		want string
	}{
		{
			name: "simple string",
			in:   Value{Type: RespSimpleString, String: "OK"},
			want: "+OK\r\n",
		},
		{
			name: "error",
			in:   Value{Type: RespError, String: "ERR bad"},
			want: "-ERR bad\r\n",
		},
		{
			name: "integer",
			in:   Value{Type: RespInteger, Integer: 42},
			want: ":42\r\n",
		},
		{
			name: "bulk string",
			in:   Value{Type: RespBulkString, String: "redis"},
			want: "$5\r\nredis\r\n",
		},
		{
			name: "nil bulk string",
			in:   Value{Type: RespBulkString, IsNull: true},
			want: "$-1\r\n",
		},
		{
			name: "array",
			in: Value{
				Type: RespArray,
				Array: []Value{
					{Type: RespBulkString, String: "SET"},
					{Type: RespBulkString, String: "k"},
					{Type: RespBulkString, String: "v"},
				},
			},
			want: "*3\r\n$3\r\nSET\r\n$1\r\nk\r\n$1\r\nv\r\n",
		},
		{
			name: "nil array",
			in:   Value{Type: RespArray, IsNull: true},
			want: "*-1\r\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Serialize(tt.in)
			if err != nil {
				t.Fatalf("serialize failed: %v", err)
			}
			if got != tt.want {
				t.Fatalf("serialize got %q, want %q", got, tt.want)
			}
		})
	}
}

func TestSerialize_UnsupportedType(t *testing.T) {
	_, err := Serialize(Value{Type: '?'})
	if err == nil {
		t.Fatalf("expected error for unsupported type")
	}
}
