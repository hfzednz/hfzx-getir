// Package redis provides a Redis cache client.
package redis

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	goredis "github.com/redis/go-redis/v9"
)

// Client is a Redis handle for hot-path cache / rate-limit backends.
type Client struct {
	url    string
	log    *slog.Logger
	client *goredis.Client
}

// NewClient dials Redis when url is set; empty URL leaves a disconnected noop client.
func NewClient(url string, log *slog.Logger) *Client {
	if log == nil {
		log = slog.Default()
	}
	c := &Client{url: url, log: log}
	if url == "" {
		log.Info("redis.noop", "note", "REDIS_URL empty")
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
	log.Info("redis.connected", "addr", opt.Addr)
	return c
}

// Ping checks connectivity.
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
