package network

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"io"
	"strconv"
	"strings"
)

const (
	// RESP type prefixes.
	RespSimpleString byte = '+'
	RespError        byte = '-'
	RespInteger      byte = ':'
	RespBulkString   byte = '$'
	RespArray        byte = '*'
)

// Value is a parsed RESP value.
// String is used by simple string / error / bulk string.
// Integer is used by integer.
// Array is used by array.
// IsNull is used by bulk string and array for RESP null values.
type Value struct {
	Type    byte
	String  string
	Integer int64
	Array   []Value
	IsNull  bool
}

// Parse decodes a single RESP value from input.
// Supported input types: string, []byte, io.Reader, *bufio.Reader.
func Parse(input any) (Value, error) {
	switch v := input.(type) {
	case string:
		return parseFromReader(bufio.NewReader(strings.NewReader(v)))
	case []byte:
		return parseFromReader(bufio.NewReader(bytes.NewReader(v)))
	case *bufio.Reader:
		return parseFromReader(v)
	case io.Reader:
		return parseFromReader(bufio.NewReader(v))
	default:
		return Value{}, fmt.Errorf("unsupported parse input type %T", input)
	}
}

// Serialize encodes a single RESP value.
func Serialize(v Value) (string, error) {
	switch v.Type {
	case RespSimpleString:
		return "+" + v.String + "\r\n", nil
	case RespError:
		return "-" + v.String + "\r\n", nil
	case RespInteger:
		return ":" + strconv.FormatInt(v.Integer, 10) + "\r\n", nil
	case RespBulkString:
		if v.IsNull {
			return "$-1\r\n", nil
		}
		return fmt.Sprintf("$%d\r\n%s\r\n", len(v.String), v.String), nil
	case RespArray:
		if v.IsNull {
			return "*-1\r\n", nil
		}
		var b strings.Builder
		b.WriteString(fmt.Sprintf("*%d\r\n", len(v.Array)))
		for _, item := range v.Array {
			encoded, err := Serialize(item)
			if err != nil {
				return "", err
			}
			b.WriteString(encoded)
		}
		return b.String(), nil
	default:
		return "", fmt.Errorf("unsupported resp type %q", string(v.Type))
	}
}

func parseFromReader(r *bufio.Reader) (Value, error) {
	prefix, err := r.ReadByte()
	if err != nil {
		return Value{}, err
	}

	switch prefix {
	case RespSimpleString:
		line, err := readLine(r)
		if err != nil {
			return Value{}, err
		}
		return Value{Type: RespSimpleString, String: line}, nil
	case RespError:
		line, err := readLine(r)
		if err != nil {
			return Value{}, err
		}
		return Value{Type: RespError, String: line}, nil
	case RespInteger:
		line, err := readLine(r)
		if err != nil {
			return Value{}, err
		}
		n, err := strconv.ParseInt(line, 10, 64)
		if err != nil {
			return Value{}, fmt.Errorf("invalid integer %q", line)
		}
		return Value{Type: RespInteger, Integer: n}, nil
	case RespBulkString:
		line, err := readLine(r)
		if err != nil {
			return Value{}, err
		}
		n, err := strconv.Atoi(line)
		if err != nil {
			return Value{}, fmt.Errorf("invalid bulk length %q", line)
		}
		if n == -1 {
			return Value{Type: RespBulkString, IsNull: true}, nil
		}
		if n < 0 {
			return Value{}, fmt.Errorf("invalid bulk length %d", n)
		}
		buf := make([]byte, n+2)
		if _, err := io.ReadFull(r, buf); err != nil {
			return Value{}, err
		}
		if !bytes.HasSuffix(buf, []byte("\r\n")) {
			return Value{}, errors.New("bulk string missing CRLF terminator")
		}
		return Value{
			Type:   RespBulkString,
			String: string(buf[:n]),
		}, nil
	case RespArray:
		line, err := readLine(r)
		if err != nil {
			return Value{}, err
		}
		n, err := strconv.Atoi(line)
		if err != nil {
			return Value{}, fmt.Errorf("invalid array length %q", line)
		}
		if n == -1 {
			return Value{Type: RespArray, IsNull: true}, nil
		}
		if n < 0 {
			return Value{}, fmt.Errorf("invalid array length %d", n)
		}
		items := make([]Value, 0, n)
		for i := 0; i < n; i++ {
			item, err := parseFromReader(r)
			if err != nil {
				return Value{}, err
			}
			items = append(items, item)
		}
		return Value{Type: RespArray, Array: items}, nil
	default:
		return Value{}, fmt.Errorf("unknown RESP type prefix %q", string(prefix))
	}
}

func readLine(r *bufio.Reader) (string, error) {
	line, err := r.ReadString('\n')
	if err != nil {
		return "", err
	}
	if len(line) < 2 || !strings.HasSuffix(line, "\r\n") {
		return "", errors.New("line must end with CRLF")
	}
	return strings.TrimSuffix(line, "\r\n"), nil
}
