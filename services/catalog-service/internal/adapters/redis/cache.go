// Package redis provides a Redis query cache for catalog-service.
package redis

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	goredis "github.com/redis/go-redis/v9"
)

// Client is a Redis-backed string cache.
type Client struct {
	URL    string
	log    *slog.Logger
	client *goredis.Client
}

// NewClient dials Redis when url is set; empty url keeps a disconnected no-op client.
func NewClient(url string, log *slog.Logger) *Client {
	if log == nil {
		log = slog.Default()
	}
	c := &Client{URL: url, log: log}
	if url == "" {
		return c
	}
	opt, err := goredis.ParseURL(url)
	if err != nil {
		log.Warn("redis.parse", "err", err)
		return c
	}
	rdb := goredis.NewClient(opt)
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	if err := rdb.Ping(ctx).Err(); err != nil {
		log.Warn("redis.ping", "err", err)
		_ = rdb.Close()
		return c
	}
	c.client = rdb
	log.Info("redis.connected", "addr", opt.Addr, "adapter", "catalog-cache")
	return c
}

// Get returns a cached value.
func (c *Client) Get(ctx context.Context, key string) (string, bool) {
	if c == nil || c.client == nil {
		return "", false
	}
	val, err := c.client.Get(ctx, key).Result()
	if err == goredis.Nil {
		return "", false
	}
	if err != nil {
		c.log.Debug("redis.get", "err", err, "key", key)
		return "", false
	}
	return val, true
}

// Set stores a value with a short TTL.
func (c *Client) Set(ctx context.Context, key, value string) error {
	if c == nil || c.client == nil {
		return nil
	}
	return c.client.Set(ctx, key, value, 5*time.Minute).Err()
}

// Ping verifies connectivity when connected.
func (c *Client) Ping(ctx context.Context) error {
	if c == nil || c.client == nil {
		return fmt.Errorf("redis: not connected")
	}
	return c.client.Ping(ctx).Err()
}

// Close closes the underlying connection.
func (c *Client) Close() error {
	if c == nil || c.client == nil {
		return nil
	}
	return c.client.Close()
}
