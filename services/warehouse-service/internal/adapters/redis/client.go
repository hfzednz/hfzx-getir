// Package redis provides a Redis task-queue for warehouse station work.
package redis

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	goredis "github.com/redis/go-redis/v9"
)

// Client is a Redis client for task queues / station locks.
type Client struct {
	URL    string
	log    *slog.Logger
	client *goredis.Client
}

// NewClient dials Redis when url is set; empty URL leaves a disconnected noop client.
func NewClient(url string, log *slog.Logger) *Client {
	if log == nil {
		log = slog.Default()
	}
	c := &Client{URL: url, log: log}
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
	log.Info("redis.connected", "addr", opt.Addr, "adapter", "warehouse-queue")
	return c
}

func (c *Client) queueKey(queue string) string {
	return "nexora:warehouse:queue:" + queue
}

// EnqueueTask pushes a task id onto a warehouse queue.
func (c *Client) EnqueueTask(ctx context.Context, queue, taskID string) error {
	if c == nil || c.client == nil {
		return nil
	}
	return c.client.LPush(ctx, c.queueKey(queue), taskID).Err()
}

// DequeueTask pops a task id (FIFO via RPOP).
func (c *Client) DequeueTask(ctx context.Context, queue string) (string, bool) {
	if c == nil || c.client == nil {
		return "", false
	}
	val, err := c.client.RPop(ctx, c.queueKey(queue)).Result()
	if err == goredis.Nil || err != nil {
		return "", false
	}
	return val, true
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
