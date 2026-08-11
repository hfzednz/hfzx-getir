package redis

import (
	"context"
	"fmt"
	"time"

	goredis "github.com/redis/go-redis/v9"
)

// Client is a Redis handle for rate-limit / cache backends.
type Client struct {
	URL    string
	client *goredis.Client
}

// Open dials Redis and verifies connectivity with Ping.
func Open(redisURL string) (*Client, error) {
	if redisURL == "" {
		return nil, fmt.Errorf("REDIS_URL required")
	}
	opt, err := goredis.ParseURL(redisURL)
	if err != nil {
		return nil, fmt.Errorf("redis parse: %w", err)
	}
	rdb := goredis.NewClient(opt)
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	if err := rdb.Ping(ctx).Err(); err != nil {
		_ = rdb.Close()
		return nil, fmt.Errorf("redis ping: %w", err)
	}
	return &Client{URL: redisURL, client: rdb}, nil
}

// Ping checks connectivity.
func (c *Client) Ping() error {
	if c == nil || c.client == nil {
		return fmt.Errorf("redis: not connected")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	return c.client.Ping(ctx).Err()
}

// Close closes the underlying connection.
func (c *Client) Close() error {
	if c == nil || c.client == nil {
		return nil
	}
	return c.client.Close()
}
