package expiry

import (
	"testing"
	"time"

	"github.com/xiesunsun/mini-redis/internal/store"
	"github.com/xiesunsun/mini-redis/internal/types"
)

// setKey is a helper that inserts a string key with the given expiry into s.
func setKey(s *store.Store, key string, expiry time.Time) {
	s.Set(key, &types.Value{
		DataType: types.StringType,
		Data:     "val",
		Expiry:   expiry,
	})
}

// TestGetOrExpire_NotExpired verifies that a key with a far-future expiry is returned.
func TestGetOrExpire_NotExpired(t *testing.T) {
	s := store.New()
	setKey(s, "k", time.Now().Add(10*time.Second))
	v := GetOrExpire(s, "k")
	if v == nil {
		t.Fatal("expected value, got nil")
	}
}

// TestGetOrExpire_Expired verifies that an expired key is deleted and nil is returned.
func TestGetOrExpire_Expired(t *testing.T) {
	s := store.New()
	setKey(s, "k", time.Now().Add(1*time.Millisecond))
	time.Sleep(5 * time.Millisecond)
	v := GetOrExpire(s, "k")
	if v != nil {
		t.Fatalf("expected nil for expired key, got %v", v)
	}
	if s.Get("k") != nil {
		t.Fatal("expected key to be deleted from store after lazy expiry")
	}
}

// TestGetOrExpire_NoExpiry verifies that a key with zero expiry (never expires) is always returned.
func TestGetOrExpire_NoExpiry(t *testing.T) {
	s := store.New()
	s.Set("k", &types.Value{DataType: types.StringType, Data: "val"})
	v := GetOrExpire(s, "k")
	if v == nil {
		t.Fatal("expected value for key with no expiry, got nil")
	}
}

// TestGetOrExpire_MissingKey verifies that a missing key returns nil.
func TestGetOrExpire_MissingKey(t *testing.T) {
	s := store.New()
	v := GetOrExpire(s, "missing")
	if v != nil {
		t.Fatalf("expected nil for missing key, got %v", v)
	}
}

// TestStartCleaner_DeletesExpiredKeys verifies periodic sweep removes expired keys.
func TestStartCleaner_DeletesExpiredKeys(t *testing.T) {
	s := store.New()
	setKey(s, "a", time.Now().Add(1*time.Millisecond))
	setKey(s, "b", time.Now().Add(1*time.Millisecond))
	setKey(s, "c", time.Now().Add(10*time.Second)) // not expired

	cancel := StartCleaner(s, 5*time.Millisecond)
	defer cancel()

	time.Sleep(30 * time.Millisecond)

	if s.Get("a") != nil {
		t.Error("expected 'a' to be deleted by cleaner")
	}
	if s.Get("b") != nil {
		t.Error("expected 'b' to be deleted by cleaner")
	}
	if s.Get("c") == nil {
		t.Error("expected 'c' to still exist (not expired)")
	}
}

// TestStartCleaner_Stop verifies that cancel stops the background goroutine without panic.
func TestStartCleaner_Stop(t *testing.T) {
	s := store.New()
	cancel := StartCleaner(s, 1*time.Millisecond)
	cancel()
	// Give goroutine time to exit.
	time.Sleep(5 * time.Millisecond)
}
