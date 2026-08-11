// Package redis provides a Redis soft-hold cache using go-redis.
package redis

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	goredis "github.com/redis/go-redis/v9"
)

// Client is a Redis-backed soft reservation hold cache.
// When URL is empty, methods are no-ops (dev). When URL is set, a real go-redis client is used.
type Client struct {
	URL    string
	TTL    time.Duration
	client *goredis.Client
	log    *slog.Logger
}

// NewClient returns a Redis client. Empty URL keeps a no-op client (no connection).
// Non-empty URL dials Redis and verifies connectivity.
func NewClient(url string, log *slog.Logger) (*Client, error) {
	if log == nil {
		log = slog.Default()
	}
	c := &Client{URL: url, TTL: 15 * time.Minute, log: log}
	if url == "" {
		return c, nil
	}
	opt, err := goredis.ParseURL(url)
	if err != nil {
		return nil, fmt.Errorf("redis: parse url: %w", err)
	}
	client := goredis.NewClient(opt)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := client.Ping(ctx).Err(); err != nil {
		_ = client.Close()
		return nil, fmt.Errorf("redis: ping: %w", err)
	}
	c.client = client
	log.Info("redis.connected", "addr", opt.Addr)
	return c, nil
}

// GetHold returns the hold value when present.
func (c *Client) GetHold(ctx context.Context, key string) (string, bool) {
	if c.client == nil {
		return "", false
	}
	v, err := c.client.Get(ctx, holdKey(key)).Result()
	if err == goredis.Nil || err != nil {
		return "", false
	}
	return v, true
}

// SetHold stores a soft hold with the client TTL.
func (c *Client) SetHold(ctx context.Context, key, value string) error {
	if c.client == nil {
		return nil
	}
	ttl := c.TTL
	if ttl <= 0 {
		ttl = 15 * time.Minute
	}
	return c.client.Set(ctx, holdKey(key), value, ttl).Err()
}

// DeleteHold removes a hold.
func (c *Client) DeleteHold(ctx context.Context, key string) error {
	if c.client == nil {
		return nil
	}
	return c.client.Del(ctx, holdKey(key)).Err()
}

// Close closes the underlying Redis client.
func (c *Client) Close() error {
	if c.client == nil {
		return nil
	}
	return c.client.Close()
}

// Ping verifies Redis connectivity when configured.
func (c *Client) Ping(ctx context.Context) error {
	if c.client == nil {
		return nil
	}
	return c.client.Ping(ctx).Err()
}

func holdKey(key string) string {
	return "inv:hold:" + key
}
