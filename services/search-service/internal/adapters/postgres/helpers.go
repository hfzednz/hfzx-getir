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
	"github.com/nexora/search-service/internal/domain"
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

// JSONStringMap marshals map[string]string.
type JSONStringMap map[string]string

func (m JSONStringMap) Value() (driver.Value, error) {
	if m == nil {
		return []byte("{}"), nil
	}
	return json.Marshal(map[string]string(m))
}

func (m *JSONStringMap) Scan(src any) error {
	if m == nil {
		return fmt.Errorf("nil JSONStringMap")
	}
	if src == nil {
		*m = JSONStringMap{}
		return nil
	}
	b, err := asBytes(src)
	if err != nil {
		return err
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

// TextArray encodes []string as Postgres text[].
type TextArray []string

func (a TextArray) Value() (driver.Value, error) {
	if a == nil || len(a) == 0 {
		return "{}", nil
	}
	parts := make([]string, len(a))
	for i, s := range a {
		escaped := strings.ReplaceAll(s, `\`, `\\`)
		escaped = strings.ReplaceAll(escaped, `"`, `\"`)
		parts[i] = `"` + escaped + `"`
	}
	return "{" + strings.Join(parts, ",") + "}", nil
}

func (a *TextArray) Scan(src any) error {
	if a == nil {
		return fmt.Errorf("nil TextArray")
	}
	parts, err := scanPGArray(src)
	if err != nil {
		return err
	}
	*a = TextArray(parts)
	return nil
}

// UUIDArray encodes []uuid.UUID as Postgres uuid[].
type UUIDArray []uuid.UUID

func (a UUIDArray) Value() (driver.Value, error) {
	if a == nil || len(a) == 0 {
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
		return fmt.Errorf("nil UUIDArray")
	}
	parts, err := scanPGArray(src)
	if err != nil {
		return err
	}
	out := make([]uuid.UUID, 0, len(parts))
	for _, p := range parts {
		if p == "" {
			continue
		}
		id, err := uuid.Parse(p)
		if err != nil {
			return err
		}
		out = append(out, id)
	}
	*a = UUIDArray(out)
	return nil
}

func scanPGArray(src any) ([]string, error) {
	if src == nil {
		return []string{}, nil
	}
	var s string
	switch v := src.(type) {
	case []byte:
		s = string(v)
	case string:
		s = v
	default:
		return nil, fmt.Errorf("unsupported array type %T", src)
	}
	s = strings.TrimSpace(s)
	if s == "" || s == "{}" {
		return []string{}, nil
	}
	if s[0] != '{' || s[len(s)-1] != '}' {
		return nil, fmt.Errorf("invalid array: %q", s)
	}
	inner := s[1 : len(s)-1]
	if inner == "" {
		return []string{}, nil
	}
	out := make([]string, 0)
	var cur strings.Builder
	inQuotes := false
	escape := false
	for i := 0; i < len(inner); i++ {
		c := inner[i]
		if escape {
			cur.WriteByte(c)
			escape = false
			continue
		}
		if c == '\\' && inQuotes {
			escape = true
			continue
		}
		if c == '"' {
			inQuotes = !inQuotes
			continue
		}
		if c == ',' && !inQuotes {
			out = append(out, cur.String())
			cur.Reset()
			continue
		}
		cur.WriteByte(c)
	}
	out = append(out, cur.String())
	return out, nil
}

func asBytes(src any) ([]byte, error) {
	switch v := src.(type) {
	case []byte:
		return v, nil
	case string:
		return []byte(v), nil
	default:
		return nil, fmt.Errorf("unsupported JSON type %T", src)
	}
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
