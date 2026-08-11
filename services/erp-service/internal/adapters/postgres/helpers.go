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
)

type JSONMap map[string]any

func (m JSONMap) Value() (driver.Value, error) {
	if m == nil {
		return []byte("{}"), nil
	}
	return json.Marshal(map[string]any(m))
}
func (m *JSONMap) Scan(src any) error {
	if m == nil {
		return fmt.Errorf("nil JSONMap")
	}
	if src == nil {
		*m = JSONMap{}
		return nil
	}
	b, err := asBytes(src)
	if err != nil {
		return err
	}
	var out map[string]any
	if len(b) == 0 {
		*m = JSONMap{}
		return nil
	}
	if err := json.Unmarshal(b, &out); err != nil {
		return err
	}
	if out == nil {
		out = map[string]any{}
	}
	*m = JSONMap(out)
	return nil
}

type JSONRaw struct{ V any }

func (j JSONRaw) Value() (driver.Value, error) {
	if j.V == nil {
		return []byte("[]"), nil
	}
	return json.Marshal(j.V)
}

func asBytes(src any) ([]byte, error) {
	switch v := src.(type) {
	case []byte:
		return v, nil
	case string:
		return []byte(v), nil
	default:
		return nil, fmt.Errorf("unsupported %T", src)
	}
}
func nullTime(t *time.Time) any {
	if t == nil || t.IsZero() {
		return nil
	}
	return t.UTC()
}
func nullTimeValue(t time.Time) any {
	if t.IsZero() {
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
func nullString(s string) any {
	if strings.TrimSpace(s) == "" {
		return nil
	}
	return s
}
func isNoRows(err error) bool { return errors.Is(err, sql.ErrNoRows) }
func dateOnly(t time.Time) string { return t.UTC().Format("2006-01-02") }

func nullDate(t time.Time) any {
	if t.IsZero() {
		return nil
	}
	return dateOnly(t)
}

func scanDate(nt sql.NullTime) time.Time {
	if !nt.Valid {
		return time.Time{}
	}
	return nt.Time.UTC()
}

func marshalLines(v any) (driver.Value, error) {
	if v == nil {
		return []byte("[]"), nil
	}
	return json.Marshal(v)
}

func unmarshalLines(src any, dest any) error {
	if src == nil {
		return nil
	}
	b, err := asBytes(src)
	if err != nil {
		return err
	}
	if len(b) == 0 {
		return nil
	}
	return json.Unmarshal(b, dest)
}
