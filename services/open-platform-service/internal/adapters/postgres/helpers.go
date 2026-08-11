package postgres

import (
	"database/sql"
	"database/sql/driver"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"
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

type TextArray []string

func (a TextArray) Value() (driver.Value, error) {
	if a == nil {
		return "{}", nil
	}
	parts := make([]string, len(a))
	for i, s := range a {
		parts[i] = `"` + strings.ReplaceAll(s, `"`, `\"`) + `"`
	}
	return "{" + strings.Join(parts, ",") + "}", nil
}
func (a *TextArray) Scan(src any) error {
	if a == nil {
		return fmt.Errorf("nil TextArray")
	}
	if src == nil {
		*a = TextArray{}
		return nil
	}
	var s string
	switch v := src.(type) {
	case []byte:
		s = string(v)
	case string:
		s = v
	default:
		return fmt.Errorf("unsupported TextArray %T", src)
	}
	s = strings.TrimSpace(s)
	if s == "" || s == "{}" {
		*a = TextArray{}
		return nil
	}
	s = strings.TrimPrefix(strings.TrimSuffix(s, "}"), "{")
	out := []string{}
	cur := strings.Builder{}
	inQ := false
	for i := 0; i < len(s); i++ {
		c := s[i]
		switch {
		case c == '"' && !inQ:
			inQ = true
		case c == '"' && inQ:
			inQ = false
		case c == ',' && !inQ:
			out = append(out, cur.String())
			cur.Reset()
		case c == '\\' && inQ && i+1 < len(s):
			i++
			cur.WriteByte(s[i])
		default:
			cur.WriteByte(c)
		}
	}
	out = append(out, cur.String())
	*a = TextArray(out)
	return nil
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
func scanNullTime(nt sql.NullTime) *time.Time {
	if !nt.Valid {
		return nil
	}
	t := nt.Time.UTC()
	return &t
}
func isNoRows(err error) bool { return errors.Is(err, sql.ErrNoRows) }
