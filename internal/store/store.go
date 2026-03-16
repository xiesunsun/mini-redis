package store

import (
	"sync"

	"github.com/xiesunsun/mini-redis/internal/types"
)

// Store 是线程安全的内存键值存储。
type Store struct {
	mu   sync.RWMutex
	data map[string]*types.Value
}

// New 创建并返回一个空的 Store 实例。
func New() *Store {
	return &Store{
		data: make(map[string]*types.Value),
	}
}

// Get 返回 key 对应的 Value，不存在时返回 nil。
func (s *Store) Get(key string) *types.Value {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.data[key]
}

// Set 将 key 设置为 value。
func (s *Store) Set(key string, value *types.Value) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.data[key] = value
}

// Delete 删除指定的 key。
func (s *Store) Delete(key string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.data, key)
}

// Keys 返回当前所有 key 的列表（顺序不定）。
func (s *Store) Keys() []string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	keys := make([]string, 0, len(s.data))
	for k := range s.data {
		keys = append(keys, k)
	}
	return keys
}
