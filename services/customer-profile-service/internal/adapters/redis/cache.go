// Package redisadapter provides a profile cache with Redis when configured, else in-process map.
package redisadapter

import (
	"context"
	"sync"
	"time"

	goredis "github.com/redis/go-redis/v9"

	"github.com/nexora/customer-profile-service/internal/domain"
	"github.com/nexora/customer-profile-service/internal/observability"
)

// Cache stores profile payloads. Empty RedisURL uses an in-process map.
type Cache struct {
	URL    string
	client *goredis.Client

	mu   sync.RWMutex
	data map[string]cacheEntry
}

type cacheEntry struct {
	payload []byte
	expires time.Time
}

// NewCache dials Redis when url is set; otherwise uses in-process storage.
func NewCache(redisURL string) *Cache {
	c := &Cache{
		URL:  redisURL,
		data: make(map[string]cacheEntry),
	}
	if redisURL == "" {
		return c
	}
	opt, err := goredis.ParseURL(redisURL)
	if err != nil {
		return c
	}
	rdb := goredis.NewClient(opt)
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	if err := rdb.Ping(ctx).Err(); err != nil {
		_ = rdb.Close()
		return c
	}
	c.client = rdb
	return c
}

// Put stores a value with optional TTL.
func (c *Cache) Put(ctx context.Context, key string, payload []byte, ttl time.Duration) error {
	if c.client != nil {
		if ttl <= 0 {
			ttl = 5 * time.Minute
		}
		return c.client.Set(ctx, key, payload, ttl).Err()
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	exp := time.Time{}
	if ttl > 0 {
		exp = time.Now().UTC().Add(ttl)
	}
	c.data[key] = cacheEntry{payload: append([]byte(nil), payload...), expires: exp}
	return nil
}

// Get returns a cached value or domain.ErrNotFound.
func (c *Cache) Get(ctx context.Context, key string) ([]byte, error) {
	if c.client != nil {
		val, err := c.client.Get(ctx, key).Bytes()
		if err == goredis.Nil {
			observability.Default.CacheMisses.Add(1)
			return nil, domain.ErrNotFound
		}
		if err != nil {
			observability.Default.CacheMisses.Add(1)
			return nil, err
		}
		observability.Default.CacheHits.Add(1)
		return val, nil
	}
	c.mu.RLock()
	defer c.mu.RUnlock()
	e, ok := c.data[key]
	if !ok {
		observability.Default.CacheMisses.Add(1)
		return nil, domain.ErrNotFound
	}
	if !e.expires.IsZero() && time.Now().UTC().After(e.expires) {
		observability.Default.CacheMisses.Add(1)
		return nil, domain.ErrNotFound
	}
	observability.Default.CacheHits.Add(1)
	return append([]byte(nil), e.payload...), nil
}

// Delete removes a key.
func (c *Cache) Delete(ctx context.Context, key string) error {
	if c.client != nil {
		return c.client.Del(ctx, key).Err()
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	delete(c.data, key)
	return nil
}

// Ping reports readiness.
func (c *Cache) Ping(ctx context.Context) error {
	if c.client != nil {
		return c.client.Ping(ctx).Err()
	}
	return nil
}

// Close closes the Redis connection when present.
func (c *Cache) Close() error {
	if c == nil || c.client == nil {
		return nil
	}
	return c.client.Close()
}

// Connected reports whether Redis is active.
func (c *Cache) Connected() bool {
	return c != nil && c.client != nil
}
