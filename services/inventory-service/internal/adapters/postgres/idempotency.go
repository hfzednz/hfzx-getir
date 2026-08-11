package postgres

import (
	"context"
	"database/sql"
	"encoding/json"
	"time"
)

// IdempotencyStore persists command results by key (multi-instance safe).
type IdempotencyStore struct {
	DB  *sql.DB
	TTL time.Duration
}

func (s *IdempotencyStore) Get(ctx context.Context, key string) (any, bool, error) {
	if key == "" {
		return nil, false, nil
	}
	var raw []byte
	err := s.DB.QueryRowContext(ctx, `
		SELECT payload FROM inventory_idempotency
		WHERE key=$1 AND (expires_at IS NULL OR expires_at > now())`, key).Scan(&raw)
	if err == sql.ErrNoRows {
		return nil, false, nil
	}
	if err != nil {
		return nil, false, err
	}
	var v any
	if err := json.Unmarshal(raw, &v); err != nil {
		return nil, false, err
	}
	return v, true, nil
}

func (s *IdempotencyStore) Put(ctx context.Context, key string, value any) error {
	if key == "" {
		return nil
	}
	raw, err := json.Marshal(value)
	if err != nil {
		return err
	}
	ttl := s.TTL
	if ttl <= 0 {
		ttl = 24 * time.Hour
	}
	expires := time.Now().UTC().Add(ttl)
	_, err = s.DB.ExecContext(ctx, `
		INSERT INTO inventory_idempotency (key, payload, created_at, expires_at)
		VALUES ($1, $2::jsonb, now(), $3)
		ON CONFLICT (key) DO UPDATE SET payload=EXCLUDED.payload, expires_at=EXCLUDED.expires_at`,
		key, raw, expires)
	return err
}
