package store

import (
	"github.com/xiesunsun/mini-redis/internal/types"
)

// LPush 从左侧插入 value，返回插入后列表长度。
// key 不存在时自动创建新列表；类型不匹配返回 ErrWrongType。
func (s *Store) LPush(key, value string) (int, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	v, ok := s.data[key]
	if !ok {
		s.data[key] = &types.Value{DataType: types.ListType, Data: []string{value}}
		return 1, nil
	}
	if v.DataType != types.ListType {
		return 0, ErrWrongType
	}
	list := v.Data.([]string)
	list = append([]string{value}, list...)
	v.Data = list
	return len(list), nil
}

// RPush 从右侧插入 value，返回插入后列表长度。
// key 不存在时自动创建新列表；类型不匹配返回 ErrWrongType。
func (s *Store) RPush(key, value string) (int, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	v, ok := s.data[key]
	if !ok {
		s.data[key] = &types.Value{DataType: types.ListType, Data: []string{value}}
		return 1, nil
	}
	if v.DataType != types.ListType {
		return 0, ErrWrongType
	}
	list := v.Data.([]string)
	list = append(list, value)
	v.Data = list
	return len(list), nil
}

// LRange 返回列表 [start, stop] 范围内的元素（含两端）。
// 支持负数索引（-1 为最后一个元素）；超出范围的索引自动截断。
// key 不存在返回空列表；类型不匹配返回 ErrWrongType。
func (s *Store) LRange(key string, start, stop int) ([]string, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	v, ok := s.data[key]
	if !ok {
		return []string{}, nil
	}
	if v.DataType != types.ListType {
		return nil, ErrWrongType
	}
	list := v.Data.([]string)
	n := len(list)

	if start < 0 {
		start = n + start
	}
	if stop < 0 {
		stop = n + stop
	}
	if start < 0 {
		start = 0
	}
	if stop >= n {
		stop = n - 1
	}
	if start > stop {
		return []string{}, nil
	}
	result := make([]string, stop-start+1)
	copy(result, list[start:stop+1])
	return result, nil
}

// LLen 返回列表的长度。
// key 不存在返回 0；类型不匹配返回 ErrWrongType。
func (s *Store) LLen(key string) (int, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	v, ok := s.data[key]
	if !ok {
		return 0, nil
	}
	if v.DataType != types.ListType {
		return 0, ErrWrongType
	}
	return len(v.Data.([]string)), nil
}

// LPop 弹出并返回列表最左侧元素。
// key 不存在返回 ErrKeyNotFound；类型不匹配返回 ErrWrongType。
func (s *Store) LPop(key string) (string, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	v, ok := s.data[key]
	if !ok {
		return "", ErrKeyNotFound
	}
	if v.DataType != types.ListType {
		return "", ErrWrongType
	}
	list := v.Data.([]string)
	if len(list) == 0 {
		return "", ErrKeyNotFound
	}
	val := list[0]
	v.Data = list[1:]
	return val, nil
}

// RPop 弹出并返回列表最右侧元素。
// key 不存在返回 ErrKeyNotFound；类型不匹配返回 ErrWrongType。
func (s *Store) RPop(key string) (string, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	v, ok := s.data[key]
	if !ok {
		return "", ErrKeyNotFound
	}
	if v.DataType != types.ListType {
		return "", ErrWrongType
	}
	list := v.Data.([]string)
	if len(list) == 0 {
		return "", ErrKeyNotFound
	}
	val := list[len(list)-1]
	v.Data = list[:len(list)-1]
	return val, nil
}
