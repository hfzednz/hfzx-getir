package postgres

import (
	"database/sql"
	"database/sql/driver"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/nexora/autonomy-service/internal/domain"
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

// JSONStrings marshals []string to/from JSONB.
type JSONStrings []string

func (s JSONStrings) Value() (driver.Value, error) {
	if s == nil {
		return []byte("[]"), nil
	}
	return json.Marshal([]string(s))
}

func (s *JSONStrings) Scan(src any) error {
	if s == nil {
		return fmt.Errorf("nil JSONStrings destination")
	}
	if src == nil {
		*s = JSONStrings{}
		return nil
	}
	var b []byte
	switch v := src.(type) {
	case []byte:
		b = v
	case string:
		b = []byte(v)
	default:
		return fmt.Errorf("unsupported JSONStrings type %T", src)
	}
	if len(b) == 0 {
		*s = JSONStrings{}
		return nil
	}
	var out []string
	if err := json.Unmarshal(b, &out); err != nil {
		return err
	}
	if out == nil {
		out = []string{}
	}
	*s = JSONStrings(out)
	return nil
}

// JSONBoolMap marshals map[string]bool to/from JSONB.
type JSONBoolMap map[string]bool

func (m JSONBoolMap) Value() (driver.Value, error) {
	if m == nil {
		return []byte("{}"), nil
	}
	return json.Marshal(map[string]bool(m))
}

func (m *JSONBoolMap) Scan(src any) error {
	if m == nil {
		return fmt.Errorf("nil JSONBoolMap destination")
	}
	if src == nil {
		*m = JSONBoolMap{}
		return nil
	}
	var b []byte
	switch v := src.(type) {
	case []byte:
		b = v
	case string:
		b = []byte(v)
	default:
		return fmt.Errorf("unsupported JSONBoolMap type %T", src)
	}
	if len(b) == 0 {
		*m = JSONBoolMap{}
		return nil
	}
	var out map[string]bool
	if err := json.Unmarshal(b, &out); err != nil {
		return err
	}
	if out == nil {
		out = map[string]bool{}
	}
	*m = JSONBoolMap(out)
	return nil
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
