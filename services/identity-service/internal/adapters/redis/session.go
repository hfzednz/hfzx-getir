package redisadapter

import (
	"context"
	"fmt"
	"time"

	"github.com/nexora/identity-service/internal/domain"
	"github.com/redis/go-redis/v9"
)

// SessionCache is a Redis-backed session/token cache.
type SessionCache struct {
	client *redis.Client
}

func NewSessionCache(redisURL string) (*SessionCache, error) {
	if redisURL == "" {
		return nil, fmt.Errorf("redis: empty REDIS_URL")
	}
	opt, err := redis.ParseURL(redisURL)
	if err != nil {
		return nil, fmt.Errorf("redis: parse url: %w", err)
	}
	client := redis.NewClient(opt)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := client.Ping(ctx).Err(); err != nil {
		_ = client.Close()
		return nil, fmt.Errorf("redis: ping: %w", err)
	}
	return &SessionCache{client: client}, nil
}

func (c *SessionCache) Put(ctx context.Context, key string, payload []byte, ttl time.Duration) error {
	return c.client.Set(ctx, key, payload, ttl).Err()
}

func (c *SessionCache) Get(ctx context.Context, key string) ([]byte, error) {
	b, err := c.client.Get(ctx, key).Bytes()
	if err == redis.Nil {
		return nil, domain.ErrNotFound
	}
	return b, err
}

func (c *SessionCache) Delete(ctx context.Context, key string) error {
	return c.client.Del(ctx, key).Err()
}

func (c *SessionCache) Ping(ctx context.Context) error {
	return c.client.Ping(ctx).Err()
}

func (c *SessionCache) Close() error {
	if c.client == nil {
		return nil
	}
	return c.client.Close()
}
