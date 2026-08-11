package postgres

import (
	"database/sql"
	"database/sql/driver"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/nexora/order-service/internal/domain"
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

// NullJSONMap is a nullable JSONB map (e.g. gift).
type NullJSONMap struct {
	Map   map[string]any
	Valid bool
}

func (n NullJSONMap) Value() (driver.Value, error) {
	if !n.Valid || n.Map == nil {
		return nil, nil
	}
	return json.Marshal(n.Map)
}

func (n *NullJSONMap) Scan(src any) error {
	if n == nil {
		return fmt.Errorf("nil NullJSONMap destination")
	}
	if src == nil {
		n.Map = nil
		n.Valid = false
		return nil
	}
	var jm JSONMap
	if err := jm.Scan(src); err != nil {
		return err
	}
	n.Map = map[string]any(jm)
	n.Valid = true
	return nil
}

// UUIDArray encodes []uuid.UUID as a PostgreSQL UUID[].
type UUIDArray []uuid.UUID

func (a UUIDArray) Value() (driver.Value, error) {
	if a == nil {
		return "{}", nil
	}
	if len(a) == 0 {
		return "{}", nil
	}
	parts := make([]string, len(a))
	for i, id := range a {
		parts[i] = id.String()
	}
	return "{" + strings.Join(parts, ",") + "}", nil
}

func (a *UUIDArray) Scan(src any) error {
	if a == nil {
		return fmt.Errorf("nil UUIDArray destination")
	}
	if src == nil {
		*a = UUIDArray{}
		return nil
	}
	var s string
	switch v := src.(type) {
	case []byte:
		s = string(v)
	case string:
		s = v
	default:
		return fmt.Errorf("unsupported UUIDArray type %T", src)
	}
	s = strings.TrimSpace(s)
	if s == "" || s == "{}" {
		*a = UUIDArray{}
		return nil
	}
	if s[0] != '{' || s[len(s)-1] != '}' {
		return fmt.Errorf("invalid uuid array: %q", s)
	}
	inner := s[1 : len(s)-1]
	if inner == "" {
		*a = UUIDArray{}
		return nil
	}
	parts := strings.Split(inner, ",")
	out := make([]uuid.UUID, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		p = strings.Trim(p, `"`)
		id, err := uuid.Parse(p)
		if err != nil {
			return fmt.Errorf("parse uuid array element %q: %w", p, err)
		}
		out = append(out, id)
	}
	*a = UUIDArray(out)
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

func mapUniqueViolation(err error) error {
	if err == nil {
		return nil
	}
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) && pgErr.Code == "23505" {
		return fmt.Errorf("%w: %s", domain.ErrAlreadyExists, pgErr.ConstraintName)
	}
	return err
}

func isNoRows(err error) bool {
	return errors.Is(err, sql.ErrNoRows)
}
