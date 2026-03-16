package expiry

import (
	"context"
	"time"

	"github.com/xiesunsun/mini-redis/internal/store"
	"github.com/xiesunsun/mini-redis/internal/types"
)

// isExpired reports whether v has a non-zero expiry that is in the past.
func isExpired(v *types.Value) bool {
	return !v.Expiry.IsZero() && time.Now().After(v.Expiry)
}

// GetOrExpire implements lazy expiry: it returns the value for key from s,
// deleting and returning nil if the value has expired.
func GetOrExpire(s *store.Store, key string) *types.Value {
	v := s.Get(key)
	if v == nil {
		return nil
	}
	if isExpired(v) {
		s.Delete(key)
		return nil
	}
	return v
}

// StartCleaner starts a background goroutine that periodically deletes all
// expired keys from s. The returned function stops the cleaner.
func StartCleaner(s *store.Store, interval time.Duration) context.CancelFunc {
	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				sweep(s)
			case <-ctx.Done():
				return
			}
		}
	}()
	return cancel
}

// sweep scans all keys and deletes those that have expired.
func sweep(s *store.Store) {
	for _, key := range s.Keys() {
		v := s.Get(key)
		if v != nil && isExpired(v) {
			s.Delete(key)
		}
	}
}
