// Package ratelimit provides rate-limit interfaces and an in-memory limiter for tests.
package ratelimit

import (
	"context"
	"errors"
	"sync"
	"time"
)

var (
	ErrDenied = errors.New("ratelimit: request denied")
)

// Limiter is the rate-limit port used by identity-service auth endpoints.
type Limiter interface {
	// Allow reports whether key may proceed under limit requests per window.
	// limit is max events; window is the sliding/fixed window duration.
	Allow(ctx context.Context, key string, limit int, window time.Duration) (bool, error)
}

// Result is an optional richer Allow response for adapters that need remaining quota.
type Result struct {
	Allowed   bool
	Remaining int
	ResetAt   time.Time
}

// MemoryLimiter is a fixed-window in-memory limiter suitable for unit tests
// and single-process development. Not for multi-instance production.
type MemoryLimiter struct {
	mu      sync.Mutex
	buckets map[string]*bucket
	now     func() time.Time
}

type bucket struct {
	count   int
	resetAt time.Time
}

// NewMemoryLimiter creates an empty MemoryLimiter.
func NewMemoryLimiter() *MemoryLimiter {
	return &MemoryLimiter{
		buckets: make(map[string]*bucket),
		now:     time.Now,
	}
}

// Allow implements Limiter with a fixed window per key.
func (m *MemoryLimiter) Allow(_ context.Context, key string, limit int, window time.Duration) (bool, error) {
	if key == "" {
		return false, errors.New("ratelimit: empty key")
	}
	if limit <= 0 {
		return false, nil
	}
	if window <= 0 {
		window = time.Minute
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	now := m.now()
	b, ok := m.buckets[key]
	if !ok || !now.Before(b.resetAt) {
		m.buckets[key] = &bucket{count: 1, resetAt: now.Add(window)}
		return true, nil
	}
	if b.count >= limit {
		return false, nil
	}
	b.count++
	return true, nil
}

// Reset clears all buckets (tests).
func (m *MemoryLimiter) Reset() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.buckets = make(map[string]*bucket)
}

// SetNowFunc overrides the clock (tests).
func (m *MemoryLimiter) SetNowFunc(fn func() time.Time) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if fn == nil {
		m.now = time.Now
		return
	}
	m.now = fn
}

// Ensure MemoryLimiter satisfies Limiter.
var _ Limiter = (*MemoryLimiter)(nil)

// RedisClient is the minimal Redis surface a RedisLimiter needs.
// Production adapters wrap go-redis (or similar) behind this interface.
type RedisClient interface {
	// Incr increments key and returns the new value.
	Incr(ctx context.Context, key string) (int64, error)
	// Expire sets TTL on key if not already set (or always — implementation choice).
	Expire(ctx context.Context, key string, window time.Duration) error
	// TTL returns remaining TTL; negative if no expiry / missing.
	TTL(ctx context.Context, key string) (time.Duration, error)
	// Del deletes keys.
	Del(ctx context.Context, keys ...string) error
}

// RedisLimiter is a fixed-window limiter backed by Redis.
// Construct with a RedisClient; methods are ready for production wiring.
type RedisLimiter struct {
	client RedisClient
	prefix string
}

// NewRedisLimiter returns a RedisLimiter. client may be nil until wired;
// Allow will return an error if client is nil.
func NewRedisLimiter(client RedisClient, keyPrefix string) *RedisLimiter {
	if keyPrefix == "" {
		keyPrefix = "rl:"
	}
	return &RedisLimiter{client: client, prefix: keyPrefix}
}

// Allow implements Limiter using INCR + EXPIRE fixed windows.
func (r *RedisLimiter) Allow(ctx context.Context, key string, limit int, window time.Duration) (bool, error) {
	if r == nil || r.client == nil {
		return false, errors.New("ratelimit: redis client not configured")
	}
	if key == "" {
		return false, errors.New("ratelimit: empty key")
	}
	if limit <= 0 {
		return false, nil
	}
	if window <= 0 {
		window = time.Minute
	}
	rk := r.prefix + key
	n, err := r.client.Incr(ctx, rk)
	if err != nil {
		return false, err
	}
	if n == 1 {
		if err := r.client.Expire(ctx, rk, window); err != nil {
			return false, err
		}
	}
	return n <= int64(limit), nil
}

// Ensure RedisLimiter satisfies Limiter.
var _ Limiter = (*RedisLimiter)(nil)
