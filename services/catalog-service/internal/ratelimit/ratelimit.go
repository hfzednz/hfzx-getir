// Package ratelimit provides a simple in-memory rate limiter.
package ratelimit

import (
	"context"
	"sync"
	"time"
)

// Limiter checks request quotas.
type Limiter interface {
	Allow(ctx context.Context, key string, limit int, window time.Duration) (bool, error)
}

// MemoryLimiter is a fixed-window counter per key.
type MemoryLimiter struct {
	mu    sync.Mutex
	windows map[string]*window
}

type window struct {
	start time.Time
	count int
}

// NewMemoryLimiter returns an in-memory limiter.
func NewMemoryLimiter() *MemoryLimiter {
	return &MemoryLimiter{windows: make(map[string]*window)}
}

// Allow reports whether the key is under the limit for the window.
func (l *MemoryLimiter) Allow(_ context.Context, key string, limit int, dur time.Duration) (bool, error) {
	if limit <= 0 {
		return true, nil
	}
	now := time.Now()
	l.mu.Lock()
	defer l.mu.Unlock()
	w, ok := l.windows[key]
	if !ok || now.Sub(w.start) >= dur {
		l.windows[key] = &window{start: now, count: 1}
		return true, nil
	}
	if w.count >= limit {
		return false, nil
	}
	w.count++
	return true, nil
}
