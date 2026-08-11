package ratelimit_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/nexora/identity-service/internal/ratelimit"
)

func TestMemoryLimiterAllow(t *testing.T) {
	lim := ratelimit.NewMemoryLimiter()
	ctx := context.Background()
	key := "login:user@example.com"
	window := time.Minute

	tests := []struct {
		name  string
		limit int
		calls int
		want  []bool
	}{
		{name: "under limit", limit: 3, calls: 3, want: []bool{true, true, true}},
		{name: "exceeds", limit: 2, calls: 3, want: []bool{true, true, false}},
		{name: "zero limit", limit: 0, calls: 1, want: []bool{false}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			lim.Reset()
			for i := 0; i < tt.calls; i++ {
				ok, err := lim.Allow(ctx, key, tt.limit, window)
				if err != nil {
					t.Fatalf("call %d: %v", i, err)
				}
				if ok != tt.want[i] {
					t.Fatalf("call %d: got %v want %v", i, ok, tt.want[i])
				}
			}
		})
	}
}

func TestMemoryLimiterWindowReset(t *testing.T) {
	lim := ratelimit.NewMemoryLimiter()
	ctx := context.Background()
	now := time.Date(2026, 8, 6, 12, 0, 0, 0, time.UTC)
	lim.SetNowFunc(func() time.Time { return now })

	ok, err := lim.Allow(ctx, "k", 1, time.Minute)
	if err != nil || !ok {
		t.Fatalf("first: ok=%v err=%v", ok, err)
	}
	ok, err = lim.Allow(ctx, "k", 1, time.Minute)
	if err != nil || ok {
		t.Fatalf("second should deny: ok=%v err=%v", ok, err)
	}

	now = now.Add(time.Minute)
	ok, err = lim.Allow(ctx, "k", 1, time.Minute)
	if err != nil || !ok {
		t.Fatalf("after window: ok=%v err=%v", ok, err)
	}
}

func TestMemoryLimiterEmptyKey(t *testing.T) {
	lim := ratelimit.NewMemoryLimiter()
	_, err := lim.Allow(context.Background(), "", 5, time.Minute)
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestRedisLimiterNilClient(t *testing.T) {
	lim := ratelimit.NewRedisLimiter(nil, "rl:")
	_, err := lim.Allow(context.Background(), "k", 5, time.Minute)
	if err == nil {
		t.Fatal("expected error")
	}
}

type mockRedis struct {
	vals map[string]int64
	ttl  map[string]time.Duration
}

func newMockRedis() *mockRedis {
	return &mockRedis{vals: map[string]int64{}, ttl: map[string]time.Duration{}}
}

func (m *mockRedis) Incr(_ context.Context, key string) (int64, error) {
	m.vals[key]++
	return m.vals[key], nil
}

func (m *mockRedis) Expire(_ context.Context, key string, window time.Duration) error {
	m.ttl[key] = window
	return nil
}

func (m *mockRedis) TTL(_ context.Context, key string) (time.Duration, error) {
	d, ok := m.ttl[key]
	if !ok {
		return -1, nil
	}
	return d, nil
}

func (m *mockRedis) Del(_ context.Context, keys ...string) error {
	for _, k := range keys {
		delete(m.vals, k)
		delete(m.ttl, k)
	}
	return nil
}

func TestRedisLimiterAllow(t *testing.T) {
	client := newMockRedis()
	lim := ratelimit.NewRedisLimiter(client, "rl:")
	ctx := context.Background()

	tests := []struct {
		name string
		n    int
		want bool
	}{
		{name: "1", n: 1, want: true},
		{name: "2", n: 2, want: true},
		{name: "3 denied", n: 3, want: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ok, err := lim.Allow(ctx, "user:1", 2, time.Minute)
			if err != nil {
				t.Fatal(err)
			}
			if ok != tt.want {
				t.Fatalf("got %v want %v", ok, tt.want)
			}
		})
	}

	if client.ttl["rl:user:1"] != time.Minute {
		t.Fatalf("ttl=%v", client.ttl["rl:user:1"])
	}
}

func TestRedisLimiterIncrError(t *testing.T) {
	lim := ratelimit.NewRedisLimiter(&errRedis{}, "rl:")
	_, err := lim.Allow(context.Background(), "k", 5, time.Minute)
	if err == nil {
		t.Fatal("expected error")
	}
}

type errRedis struct{}

func (e *errRedis) Incr(context.Context, string) (int64, error) {
	return 0, errors.New("redis down")
}
func (e *errRedis) Expire(context.Context, string, time.Duration) error { return nil }
func (e *errRedis) TTL(context.Context, string) (time.Duration, error)  { return 0, nil }
func (e *errRedis) Del(context.Context, ...string) error                { return nil }
