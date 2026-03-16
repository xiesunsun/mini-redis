package store

import (
	"errors"
	"testing"

	"github.com/xiesunsun/mini-redis/internal/types"
)

// ---------- LPush ----------

func TestLPush_NewKey(t *testing.T) {
	s := New()
	n, err := s.LPush("k", "a")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if n != 1 {
		t.Fatalf("expected length 1, got %d", n)
	}
}

func TestLPush_ExistingList(t *testing.T) {
	s := New()
	s.LPush("k", "b")
	n, err := s.LPush("k", "a")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if n != 2 {
		t.Fatalf("expected length 2, got %d", n)
	}
	got, _ := s.LRange("k", 0, -1)
	if len(got) != 2 || got[0] != "a" || got[1] != "b" {
		t.Fatalf("expected [a b], got %v", got)
	}
}

func TestLPush_WrongType(t *testing.T) {
	s := New()
	s.SetString("k", "hello")
	_, err := s.LPush("k", "a")
	if !errors.Is(err, ErrWrongType) {
		t.Fatalf("expected ErrWrongType, got %v", err)
	}
}

// ---------- RPush ----------

func TestRPush_NewKey(t *testing.T) {
	s := New()
	n, err := s.RPush("k", "a")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if n != 1 {
		t.Fatalf("expected length 1, got %d", n)
	}
}

func TestRPush_ExistingList(t *testing.T) {
	s := New()
	s.RPush("k", "a")
	n, err := s.RPush("k", "b")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if n != 2 {
		t.Fatalf("expected length 2, got %d", n)
	}
	got, _ := s.LRange("k", 0, -1)
	if len(got) != 2 || got[0] != "a" || got[1] != "b" {
		t.Fatalf("expected [a b], got %v", got)
	}
}

func TestRPush_WrongType(t *testing.T) {
	s := New()
	s.SetString("k", "hello")
	_, err := s.RPush("k", "a")
	if !errors.Is(err, ErrWrongType) {
		t.Fatalf("expected ErrWrongType, got %v", err)
	}
}

// ---------- LRange ----------

func TestLRange_NormalRange(t *testing.T) {
	s := New()
	s.RPush("k", "a")
	s.RPush("k", "b")
	s.RPush("k", "c")

	got, err := s.LRange("k", 0, 1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(got) != 2 || got[0] != "a" || got[1] != "b" {
		t.Fatalf("expected [a b], got %v", got)
	}
}

func TestLRange_NonExistentKey(t *testing.T) {
	s := New()
	got, err := s.LRange("missing", 0, -1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(got) != 0 {
		t.Fatalf("expected empty list, got %v", got)
	}
}

func TestLRange_WrongType(t *testing.T) {
	s := New()
	s.SetString("k", "hello")
	_, err := s.LRange("k", 0, -1)
	if !errors.Is(err, ErrWrongType) {
		t.Fatalf("expected ErrWrongType, got %v", err)
	}
}

func TestLRange_NegativeIndexes(t *testing.T) {
	s := New()
	s.RPush("k", "a")
	s.RPush("k", "b")
	s.RPush("k", "c")

	got, err := s.LRange("k", 0, -1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(got) != 3 || got[0] != "a" || got[2] != "c" {
		t.Fatalf("expected [a b c], got %v", got)
	}

	got, err = s.LRange("k", -2, -1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(got) != 2 || got[0] != "b" || got[1] != "c" {
		t.Fatalf("expected [b c], got %v", got)
	}
}

func TestLRange_OutOfBounds(t *testing.T) {
	s := New()
	s.RPush("k", "a")
	s.RPush("k", "b")

	got, err := s.LRange("k", 0, 100)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(got) != 2 || got[0] != "a" || got[1] != "b" {
		t.Fatalf("expected [a b], got %v", got)
	}
}

func TestLRange_StartGreaterThanStop(t *testing.T) {
	s := New()
	s.RPush("k", "a")
	s.RPush("k", "b")

	got, err := s.LRange("k", 2, 1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(got) != 0 {
		t.Fatalf("expected empty list, got %v", got)
	}
}

// ---------- LLen ----------

func TestLLen_ExistingList(t *testing.T) {
	s := New()
	s.RPush("k", "a")
	s.RPush("k", "b")

	n, err := s.LLen("k")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if n != 2 {
		t.Fatalf("expected 2, got %d", n)
	}
}

func TestLLen_NonExistentKey(t *testing.T) {
	s := New()
	n, err := s.LLen("missing")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if n != 0 {
		t.Fatalf("expected 0, got %d", n)
	}
}

func TestLLen_WrongType(t *testing.T) {
	s := New()
	s.SetString("k", "hello")
	_, err := s.LLen("k")
	if !errors.Is(err, ErrWrongType) {
		t.Fatalf("expected ErrWrongType, got %v", err)
	}
}

// ---------- LPop ----------

func TestLPop_ExistingList(t *testing.T) {
	s := New()
	s.RPush("k", "a")
	s.RPush("k", "b")

	val, err := s.LPop("k")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if val != "a" {
		t.Fatalf("expected %q, got %q", "a", val)
	}
	n, _ := s.LLen("k")
	if n != 1 {
		t.Fatalf("expected length 1 after pop, got %d", n)
	}
}

func TestLPop_NonExistentKey(t *testing.T) {
	s := New()
	_, err := s.LPop("missing")
	if !errors.Is(err, ErrKeyNotFound) {
		t.Fatalf("expected ErrKeyNotFound, got %v", err)
	}
}

func TestLPop_WrongType(t *testing.T) {
	s := New()
	s.SetString("k", "hello")
	_, err := s.LPop("k")
	if !errors.Is(err, ErrWrongType) {
		t.Fatalf("expected ErrWrongType, got %v", err)
	}
}

// ---------- RPop ----------

func TestRPop_ExistingList(t *testing.T) {
	s := New()
	s.RPush("k", "a")
	s.RPush("k", "b")

	val, err := s.RPop("k")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if val != "b" {
		t.Fatalf("expected %q, got %q", "b", val)
	}
	n, _ := s.LLen("k")
	if n != 1 {
		t.Fatalf("expected length 1 after pop, got %d", n)
	}
}

func TestRPop_NonExistentKey(t *testing.T) {
	s := New()
	_, err := s.RPop("missing")
	if !errors.Is(err, ErrKeyNotFound) {
		t.Fatalf("expected ErrKeyNotFound, got %v", err)
	}
}

func TestRPop_WrongType(t *testing.T) {
	s := New()
	s.SetString("k", "hello")
	_, err := s.RPop("k")
	if !errors.Is(err, ErrWrongType) {
		t.Fatalf("expected ErrWrongType, got %v", err)
	}
}

// ---------- 类型不匹配辅助测试（直接引用 types 包）----------

func TestLPush_WrongTypeViaValue(t *testing.T) {
	s := New()
	s.Set("k", &types.Value{DataType: types.HashType, Data: map[string]string{"f": "v"}})
	_, err := s.LPush("k", "a")
	if !errors.Is(err, ErrWrongType) {
		t.Fatalf("expected ErrWrongType, got %v", err)
	}
}
