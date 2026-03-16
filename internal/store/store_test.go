package store

import (
	"testing"

	"github.com/xiesunsun/mini-redis/internal/types"
)

func TestGet_ExistingKey(t *testing.T) {
	s := New()
	v := &types.Value{DataType: types.StringType, Data: "hello"}
	s.Set("k", v)

	got := s.Get("k")
	if got != v {
		t.Fatalf("expected %v, got %v", v, got)
	}
}

func TestGet_NonExistentKey(t *testing.T) {
	s := New()

	got := s.Get("missing")
	if got != nil {
		t.Fatalf("expected nil, got %v", got)
	}
}

func TestSet_NewKey(t *testing.T) {
	s := New()
	v := &types.Value{DataType: types.StringType, Data: "world"}
	s.Set("key", v)

	got := s.Get("key")
	if got != v {
		t.Fatalf("expected %v, got %v", v, got)
	}
}

func TestSet_OverwriteExistingKey(t *testing.T) {
	s := New()
	v1 := &types.Value{DataType: types.StringType, Data: "first"}
	v2 := &types.Value{DataType: types.StringType, Data: "second"}
	s.Set("key", v1)
	s.Set("key", v2)

	got := s.Get("key")
	if got != v2 {
		t.Fatalf("expected %v, got %v", v2, got)
	}
}

func TestDelete_ExistingKey(t *testing.T) {
	s := New()
	v := &types.Value{DataType: types.StringType, Data: "val"}
	s.Set("key", v)
	s.Delete("key")

	got := s.Get("key")
	if got != nil {
		t.Fatalf("expected nil after delete, got %v", got)
	}
}

func TestDelete_NonExistentKey(t *testing.T) {
	s := New()
	// deleting a non-existent key should not panic
	s.Delete("missing")

	got := s.Get("missing")
	if got != nil {
		t.Fatalf("expected nil, got %v", got)
	}
}

func TestKeys_EmptyStore(t *testing.T) {
	s := New()

	keys := s.Keys()
	if len(keys) != 0 {
		t.Fatalf("expected empty keys, got %v", keys)
	}
}

func TestKeys_MultipleKeys(t *testing.T) {
	s := New()
	s.Set("a", &types.Value{DataType: types.StringType, Data: "1"})
	s.Set("b", &types.Value{DataType: types.StringType, Data: "2"})
	s.Set("c", &types.Value{DataType: types.StringType, Data: "3"})

	keys := s.Keys()
	if len(keys) != 3 {
		t.Fatalf("expected 3 keys, got %d: %v", len(keys), keys)
	}

	keySet := make(map[string]bool)
	for _, k := range keys {
		keySet[k] = true
	}
	for _, expected := range []string{"a", "b", "c"} {
		if !keySet[expected] {
			t.Fatalf("expected key %q in keys %v", expected, keys)
		}
	}
}
