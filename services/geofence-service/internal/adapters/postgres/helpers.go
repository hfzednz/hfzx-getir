package postgres

import (
	"database/sql"
	"database/sql/driver"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/nexora/geofence-service/internal/domain"
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

// VerticesJSON encodes []domain.Point as JSONB.
type VerticesJSON []domain.Point

func (v VerticesJSON) Value() (driver.Value, error) {
	if v == nil {
		return []byte("[]"), nil
	}
	b, err := json.Marshal([]domain.Point(v))
	if err != nil {
		return nil, err
	}
	return b, nil
}

func (v *VerticesJSON) Scan(src any) error {
	if v == nil {
		return fmt.Errorf("nil VerticesJSON destination")
	}
	if src == nil {
		*v = VerticesJSON{}
		return nil
	}
	var b []byte
	switch t := src.(type) {
	case []byte:
		b = t
	case string:
		b = []byte(t)
	default:
		return fmt.Errorf("unsupported VerticesJSON type %T", src)
	}
	if len(b) == 0 {
		*v = VerticesJSON{}
		return nil
	}
	var pts []domain.Point
	if err := json.Unmarshal(b, &pts); err != nil {
		return err
	}
	if pts == nil {
		pts = []domain.Point{}
	}
	*v = VerticesJSON(pts)
	return nil
}

func nullTime(t *time.Time) any {
	if t == nil || t.IsZero() {
		return nil
	}
	return t.UTC()
}

func scanNullTime(nt sql.NullTime) *time.Time {
	if !nt.Valid {
		return nil
	}
	t := nt.Time.UTC()
	return &t
}

func nullFloat(n *float64) any {
	if n == nil {
		return nil
	}
	return *n
}

func scanNullFloat(n sql.NullFloat64) *float64 {
	if !n.Valid {
		return nil
	}
	v := n.Float64
	return &v
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
