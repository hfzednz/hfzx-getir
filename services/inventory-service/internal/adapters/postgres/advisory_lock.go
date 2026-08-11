package postgres

import (
	"context"
	"database/sql"
	"hash/fnv"
)

// AdvisoryStockLocker uses Postgres session advisory locks for stock mutations.
type AdvisoryStockLocker struct{ DB *sql.DB }

func (l *AdvisoryStockLocker) WithLock(ctx context.Context, key string, fn func() error) error {
	if l == nil || l.DB == nil {
		return fn()
	}
	conn, err := l.DB.Conn(ctx)
	if err != nil {
		return err
	}
	defer func() { _ = conn.Close() }()

	lockKey := advisoryKey(key)
	if _, err := conn.ExecContext(ctx, `SELECT pg_advisory_lock($1)`, lockKey); err != nil {
		return err
	}
	defer func() {
		_, _ = conn.ExecContext(context.Background(), `SELECT pg_advisory_unlock($1)`, lockKey)
	}()
	return fn()
}

func advisoryKey(s string) int64 {
	h := fnv.New64a()
	_, _ = h.Write([]byte(s))
	v := h.Sum64()
	return int64(v & 0x7fffffffffffffff)
}
