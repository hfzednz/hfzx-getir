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
	"github.com/nexora/notification-service/internal/domain"
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

// JSONStringMap marshals map[string]string to/from JSONB.
type JSONStringMap map[string]string

func (m JSONStringMap) Value() (driver.Value, error) {
	if m == nil {
		return []byte("{}"), nil
	}
	b, err := json.Marshal(map[string]string(m))
	if err != nil {
		return nil, err
	}
	return b, nil
}

func (m *JSONStringMap) Scan(src any) error {
	if m == nil {
		return fmt.Errorf("nil JSONStringMap destination")
	}
	if src == nil {
		*m = JSONStringMap{}
		return nil
	}
	var b []byte
	switch v := src.(type) {
	case []byte:
		b = v
	case string:
		b = []byte(v)
	default:
		return fmt.Errorf("unsupported JSONStringMap type %T", src)
	}
	if len(b) == 0 {
		*m = JSONStringMap{}
		return nil
	}
	var out map[string]string
	if err := json.Unmarshal(b, &out); err != nil {
		return err
	}
	if out == nil {
		out = map[string]string{}
	}
	*m = JSONStringMap(out)
	return nil
}

// JSONBoolMap marshals map[string]bool (channel opt-out) to/from JSONB.
type JSONBoolMap map[string]bool

func (m JSONBoolMap) Value() (driver.Value, error) {
	if m == nil {
		return []byte("{}"), nil
	}
	b, err := json.Marshal(map[string]bool(m))
	if err != nil {
		return nil, err
	}
	return b, nil
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
	return t.UTC()
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

func channelOptOutToJSON(m map[domain.Channel]bool) JSONBoolMap {
	out := JSONBoolMap{}
	for k, v := range m {
		out[string(k)] = v
	}
	return out
}

func channelOptOutFromJSON(m JSONBoolMap) map[domain.Channel]bool {
	out := map[domain.Channel]bool{}
	for k, v := range m {
		out[domain.Channel(k)] = v
	}
	return out
}
