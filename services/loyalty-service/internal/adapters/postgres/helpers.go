package postgres

import (
	"database/sql"
	"database/sql/driver"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/nexora/loyalty-service/internal/domain"
)

// JSONMap marshals map[string]any to/from JSONB.
type JSONMap map[string]any

func (m JSONMap) Value() (driver.Value, error) {
	if m == nil {
		return []byte("{}"), nil
	}
	b, err := json.Marshal(map[string]any(m))
	if err != nil {
		return nil, err
	}
	return b, nil
}

func (m *JSONMap) Scan(src any) error {
	if m == nil {
		return fmt.Errorf("nil JSONMap destination")
	}
	if src == nil {
		*m = JSONMap{}
		return nil
	}
	var b []byte
	switch v := src.(type) {
	case []byte:
		b = v
	case string:
		b = []byte(v)
	default:
		return fmt.Errorf("unsupported JSONMap type %T", src)
	}
	if len(b) == 0 {
		*m = JSONMap{}
		return nil
	}
	var out map[string]any
	if err := json.Unmarshal(b, &out); err != nil {
		return err
	}
	if out == nil {
		out = map[string]any{}
	}
	*m = JSONMap(out)
	return nil
}

// JSONPrizes marshals []domain.SpinPrize to/from JSONB.
type JSONPrizes []domain.SpinPrize

func (p JSONPrizes) Value() (driver.Value, error) {
	if p == nil {
		return []byte("[]"), nil
	}
	b, err := json.Marshal([]domain.SpinPrize(p))
	if err != nil {
		return nil, err
	}
	return b, nil
}

func (p *JSONPrizes) Scan(src any) error {
	if p == nil {
		return fmt.Errorf("nil JSONPrizes destination")
	}
	if src == nil {
		*p = JSONPrizes{}
		return nil
	}
	var b []byte
	switch v := src.(type) {
	case []byte:
		b = v
	case string:
		b = []byte(v)
	default:
		return fmt.Errorf("unsupported JSONPrizes type %T", src)
	}
	if len(b) == 0 {
		*p = JSONPrizes{}
		return nil
	}
	var out []domain.SpinPrize
	if err := json.Unmarshal(b, &out); err != nil {
		return err
	}
	if out == nil {
		out = []domain.SpinPrize{}
	}
	*p = JSONPrizes(out)
	return nil
}

func nullUUID(id *uuid.UUID) any {
	if id == nil || *id == uuid.Nil {
		return nil
	}
	return *id
}

func nullUUIDValue(id uuid.UUID) any {
	if id == uuid.Nil {
		return nil
	}
	return id
}

func scanNullUUID(n uuid.NullUUID) *uuid.UUID {
	if !n.Valid {
		return nil
	}
	id := n.UUID
	return &id
}

func scanUUIDOrNil(n uuid.NullUUID) uuid.UUID {
	if !n.Valid {
		return uuid.Nil
	}
	return n.UUID
}

func nullTime(t *time.Time) any {
	if t == nil || t.IsZero() {
		return nil
	}
	return *t
}

func scanNullTime(nt sql.NullTime) *time.Time {
	if !nt.Valid {
		return nil
	}
	t := nt.Time.UTC()
	return &t
}

func mapUniqueViolation(err error) error {
	if err == nil {
		return nil
	}
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) && pgErr.Code == "23505" {
		return domain.ErrAlreadyExists
	}
	return err
}

func isNoRows(err error) bool {
	return errors.Is(err, sql.ErrNoRows)
}
