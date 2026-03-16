package store

import (
	"errors"
	"testing"

	"github.com/xiesunsun/mini-redis/internal/types"
)

func TestGetString_ExistingKey(t *testing.T) {
	s := New()
	s.SetString("k", "hello")

	got, err := s.GetString("k")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != "hello" {
		t.Fatalf("expected %q, got %q", "hello", got)
	}
}

func TestGetString_NonExistentKey(t *testing.T) {
	s := New()

	_, err := s.GetString("missing")
	if !errors.Is(err, ErrKeyNotFound) {
		t.Fatalf("expected ErrKeyNotFound, got %v", err)
	}
}

func TestGetString_WrongType(t *testing.T) {
	s := New()
	s.Set("k", &types.Value{DataType: types.ListType, Data: []string{"a", "b"}})

	_, err := s.GetString("k")
	if !errors.Is(err, ErrWrongType) {
		t.Fatalf("expected ErrWrongType, got %v", err)
	}
}

func TestDeleteKey_ExistingKey(t *testing.T) {
	s := New()
	s.SetString("k", "value")
	s.DeleteKey("k")

	_, err := s.GetString("k")
	if !errors.Is(err, ErrKeyNotFound) {
		t.Fatalf("expected ErrKeyNotFound after delete, got %v", err)
	}
}

func TestDeleteKey_NonExistentKey(t *testing.T) {
	s := New()
	// deleting a non-existent key should not panic
	s.DeleteKey("missing")

	_, err := s.GetString("missing")
	if !errors.Is(err, ErrKeyNotFound) {
		t.Fatalf("expected ErrKeyNotFound, got %v", err)
	}
}
