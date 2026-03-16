package types

import "time"

// DataType 表示 Redis 支持的数据类型。
type DataType int

const (
	StringType DataType = iota
	ListType
	HashType
)

// Value 表示存储在内存中的值。
type Value struct {
	DataType DataType
	// Data 实际存储的数据
	// StringType → string
	// ListType   → []string
	// HashType   → map[string]string
	Data   interface{}
	Expiry time.Time // 零值表示永不过期
}
