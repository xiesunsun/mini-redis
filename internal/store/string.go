package store

import (
	"errors"

	"github.com/xiesunsun/mini-redis/internal/types"
)

var (
	ErrKeyNotFound = errors.New("key not found")
	ErrWrongType   = errors.New("WRONGTYPE Operation against a key holding the wrong kind of value")
)

// SetString 将 key 设置为 string 类型的值。
func (s *Store) SetString(key, value string) {
	s.Set(key, &types.Value{DataType: types.StringType, Data: value})
}

// GetString 读取 key 对应的 string 值。
// key 不存在返回 ErrKeyNotFound，类型不匹配返回 ErrWrongType。
func (s *Store) GetString(key string) (string, error) {
	v := s.Get(key)
	if v == nil {
		return "", ErrKeyNotFound
	}
	if v.DataType != types.StringType {
		return "", ErrWrongType
	}
	return v.Data.(string), nil
}

// DeleteKey 删除指定的 key。
func (s *Store) DeleteKey(key string) {
	s.Delete(key)
}
