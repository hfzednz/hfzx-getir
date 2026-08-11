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
	"github.com/lib/pq"
	"github.com/nexora/catalog-service/internal/domain"
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
	out := map[string]any{}
	if err := json.Unmarshal(b, &out); err != nil {
		return err
	}
	*m = JSONMap(out)
	return nil
}

// JSONArray marshals []map[string]any to/from JSONB.
type JSONArray []map[string]any

func (a JSONArray) Value() (driver.Value, error) {
	if a == nil {
		return []byte("[]"), nil
	}
	b, err := json.Marshal([]map[string]any(a))
	if err != nil {
		return nil, err
	}
	return b, nil
}

func (a *JSONArray) Scan(src any) error {
	if a == nil {
		return fmt.Errorf("nil JSONArray destination")
	}
	if src == nil {
		*a = JSONArray{}
		return nil
	}
	var b []byte
	switch v := src.(type) {
	case []byte:
		b = v
	case string:
		b = []byte(v)
	default:
		return fmt.Errorf("unsupported JSONArray type %T", src)
	}
	if len(b) == 0 {
		*a = JSONArray{}
		return nil
	}
	var out []map[string]any
	if err := json.Unmarshal(b, &out); err != nil {
		return err
	}
	if out == nil {
		out = []map[string]any{}
	}
	*a = JSONArray(out)
	return nil
}

func nullUUID(id *uuid.UUID) any {
	if id == nil || *id == uuid.Nil {
		return nil
	}
	return *id
}

func scanNullUUID(n uuid.NullUUID) *uuid.UUID {
	if !n.Valid {
		return nil
	}
	id := n.UUID
	return &id
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

func nullFloat64(f *float64) any {
	if f == nil {
		return nil
	}
	return *f
}

func scanNullFloat64(n sql.NullFloat64) *float64 {
	if !n.Valid {
		return nil
	}
	v := n.Float64
	return &v
}

func nullString(s *string) any {
	if s == nil {
		return nil
	}
	return *s
}

func mapNotFound(err error) error {
	if errors.Is(err, sql.ErrNoRows) {
		return domain.ErrNotFound
	}
	return err
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

func textArray(ss []string) any {
	if ss == nil {
		return pq.Array([]string{})
	}
	return pq.Array(ss)
}
