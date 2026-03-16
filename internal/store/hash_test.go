package store

import (
	"errors"
	"testing"

	"github.com/xiesunsun/mini-redis/internal/types"
)

// ---------- HSet ----------

func TestHSet_NewKey(t *testing.T) {
	s := New()
	n, err := s.HSet("k", "f", "v")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if n != 1 {
		t.Fatalf("expected 1 (new field), got %d", n)
	}
}

func TestHSet_UpdateExistingField(t *testing.T) {
	s := New()
	s.HSet("k", "f", "v1")
	n, err := s.HSet("k", "f", "v2")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if n != 0 {
		t.Fatalf("expected 0 (update), got %d", n)
	}
	got, _ := s.HGet("k", "f")
	if got != "v2" {
		t.Fatalf("expected v2, got %q", got)
	}
}

func TestHSet_WrongType(t *testing.T) {
	s := New()
	s.SetString("k", "hello")
	_, err := s.HSet("k", "f", "v")
	if !errors.Is(err, ErrWrongType) {
		t.Fatalf("expected ErrWrongType, got %v", err)
	}
}

// ---------- HGet ----------

func TestHGet_ExistingField(t *testing.T) {
	s := New()
	s.HSet("k", "f", "v")
	got, err := s.HGet("k", "f")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != "v" {
		t.Fatalf("expected %q, got %q", "v", got)
	}
}

func TestHGet_NonExistentKey(t *testing.T) {
	s := New()
	_, err := s.HGet("missing", "f")
	if !errors.Is(err, ErrKeyNotFound) {
		t.Fatalf("expected ErrKeyNotFound, got %v", err)
	}
}

func TestHGet_NonExistentField(t *testing.T) {
	s := New()
	s.HSet("k", "f", "v")
	_, err := s.HGet("k", "no_such_field")
	if !errors.Is(err, ErrKeyNotFound) {
		t.Fatalf("expected ErrKeyNotFound, got %v", err)
	}
}

func TestHGet_WrongType(t *testing.T) {
	s := New()
	s.SetString("k", "hello")
	_, err := s.HGet("k", "f")
	if !errors.Is(err, ErrWrongType) {
		t.Fatalf("expected ErrWrongType, got %v", err)
	}
}

// ---------- HDel ----------

func TestHDel_ExistingField(t *testing.T) {
	s := New()
	s.HSet("k", "f", "v")
	n, err := s.HDel("k", "f")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if n != 1 {
		t.Fatalf("expected 1, got %d", n)
	}
	_, err = s.HGet("k", "f")
	if !errors.Is(err, ErrKeyNotFound) {
		t.Fatalf("expected ErrKeyNotFound after del, got %v", err)
	}
}

func TestHDel_NonExistentKey(t *testing.T) {
	s := New()
	n, err := s.HDel("missing", "f")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if n != 0 {
		t.Fatalf("expected 0, got %d", n)
	}
}

func TestHDel_NonExistentField(t *testing.T) {
	s := New()
	s.HSet("k", "f", "v")
	n, err := s.HDel("k", "no_such_field")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if n != 0 {
		t.Fatalf("expected 0, got %d", n)
	}
}

func TestHDel_WrongType(t *testing.T) {
	s := New()
	s.SetString("k", "hello")
	_, err := s.HDel("k", "f")
	if !errors.Is(err, ErrWrongType) {
		t.Fatalf("expected ErrWrongType, got %v", err)
	}
}

// ---------- HGetAll ----------

func TestHGetAll_ExistingHash(t *testing.T) {
	s := New()
	s.HSet("k", "f1", "v1")
	s.HSet("k", "f2", "v2")

	got, err := s.HGetAll("k")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(got) != 4 {
		t.Fatalf("expected 4 elements (2 field-value pairs), got %d", len(got))
	}
	// Build map from result to verify field-value pairs regardless of order
	m := make(map[string]string)
	for i := 0; i < len(got); i += 2 {
		m[got[i]] = got[i+1]
	}
	if m["f1"] != "v1" || m["f2"] != "v2" {
		t.Fatalf("unexpected field-value pairs: %v", m)
	}
}

func TestHGetAll_NonExistentKey(t *testing.T) {
	s := New()
	got, err := s.HGetAll("missing")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(got) != 0 {
		t.Fatalf("expected empty slice, got %v", got)
	}
}

func TestHGetAll_WrongType(t *testing.T) {
	s := New()
	s.SetString("k", "hello")
	_, err := s.HGetAll("k")
	if !errors.Is(err, ErrWrongType) {
		t.Fatalf("expected ErrWrongType, got %v", err)
	}
}

// ---------- HExists ----------

func TestHExists_FieldExists(t *testing.T) {
	s := New()
	s.HSet("k", "f", "v")
	n, err := s.HExists("k", "f")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if n != 1 {
		t.Fatalf("expected 1, got %d", n)
	}
}

func TestHExists_NonExistentKey(t *testing.T) {
	s := New()
	n, err := s.HExists("missing", "f")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if n != 0 {
		t.Fatalf("expected 0, got %d", n)
	}
}

func TestHExists_NonExistentField(t *testing.T) {
	s := New()
	s.HSet("k", "f", "v")
	n, err := s.HExists("k", "no_such_field")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if n != 0 {
		t.Fatalf("expected 0, got %d", n)
	}
}

func TestHExists_WrongType(t *testing.T) {
	s := New()
	s.SetString("k", "hello")
	_, err := s.HExists("k", "f")
	if !errors.Is(err, ErrWrongType) {
		t.Fatalf("expected ErrWrongType, got %v", err)
	}
}

// ---------- 类型不匹配辅助测试（直接引用 types 包）----------

func TestHSet_WrongTypeViaValue(t *testing.T) {
	s := New()
	s.Set("k", &types.Value{DataType: types.ListType, Data: []string{"a"}})
	_, err := s.HSet("k", "f", "v")
	if !errors.Is(err, ErrWrongType) {
		t.Fatalf("expected ErrWrongType, got %v", err)
	}
}
