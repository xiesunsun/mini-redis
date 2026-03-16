package store

import (
	"github.com/xiesunsun/mini-redis/internal/types"
)

// HSet 设置 hash 中的 field。key 不存在时自动创建；类型不匹配返回 ErrWrongType。
// 返回新增 field 的数量（0 表示更新已有 field，1 表示新增 field）。
func (s *Store) HSet(key, field, value string) (int, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	v, ok := s.data[key]
	if !ok {
		s.data[key] = &types.Value{DataType: types.HashType, Data: map[string]string{field: value}}
		return 1, nil
	}
	if v.DataType != types.HashType {
		return 0, ErrWrongType
	}
	h := v.Data.(map[string]string)
	_, existed := h[field]
	h[field] = value
	if existed {
		return 0, nil
	}
	return 1, nil
}

// HGet 返回 hash 中指定 field 的值。
// key 不存在或 field 不存在返回 ErrKeyNotFound；类型不匹配返回 ErrWrongType。
func (s *Store) HGet(key, field string) (string, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	v, ok := s.data[key]
	if !ok {
		return "", ErrKeyNotFound
	}
	if v.DataType != types.HashType {
		return "", ErrWrongType
	}
	h := v.Data.(map[string]string)
	val, ok := h[field]
	if !ok {
		return "", ErrKeyNotFound
	}
	return val, nil
}

// HDel 删除 hash 中指定的 field。
// key 不存在或 field 不存在返回 0；类型不匹配返回 ErrWrongType。
func (s *Store) HDel(key, field string) (int, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	v, ok := s.data[key]
	if !ok {
		return 0, nil
	}
	if v.DataType != types.HashType {
		return 0, ErrWrongType
	}
	h := v.Data.(map[string]string)
	if _, ok := h[field]; !ok {
		return 0, nil
	}
	delete(h, field)
	return 1, nil
}

// HGetAll 返回 hash 中所有 field 和 value，以 [field, value, ...] 交替顺序返回。
// key 不存在返回空切片；类型不匹配返回 ErrWrongType。
func (s *Store) HGetAll(key string) ([]string, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	v, ok := s.data[key]
	if !ok {
		return []string{}, nil
	}
	if v.DataType != types.HashType {
		return nil, ErrWrongType
	}
	h := v.Data.(map[string]string)
	result := make([]string, 0, len(h)*2)
	for f, val := range h {
		result = append(result, f, val)
	}
	return result, nil
}

// HExists 检查 hash 中是否存在指定 field。
// field 存在返回 1，不存在或 key 不存在返回 0；类型不匹配返回 ErrWrongType。
func (s *Store) HExists(key, field string) (int, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	v, ok := s.data[key]
	if !ok {
		return 0, nil
	}
	if v.DataType != types.HashType {
		return 0, ErrWrongType
	}
	h := v.Data.(map[string]string)
	if _, ok := h[field]; ok {
		return 1, nil
	}
	return 0, nil
}
