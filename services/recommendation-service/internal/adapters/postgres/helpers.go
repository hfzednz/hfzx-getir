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
	"github.com/nexora/recommendation-service/internal/domain"
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

// TextArray encodes []string as PostgreSQL text[].
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
		return fmt.Errorf("nil TextArray destination")
	}
	if src == nil {
		*a = TextArray{}
		return nil
	}
	var b []byte
	switch v := src.(type) {
	case []byte:
		b = v
	case string:
		b = []byte(v)
	default:
		return fmt.Errorf("unsupported TextArray type %T", src)
	}
	s := strings.TrimSpace(string(b))
	if s == "" || s == "{}" || s == "[]" || s == "null" {
		*a = TextArray{}
		return nil
	}
	if strings.HasPrefix(s, "[") {
		var out []string
		if err := json.Unmarshal(b, &out); err != nil {
			return err
		}
		if out == nil {
			out = []string{}
		}
		*a = TextArray(out)
		return nil
	}
	if s[0] != '{' || s[len(s)-1] != '}' {
		return fmt.Errorf("invalid text array: %q", s)
	}
	inner := s[1 : len(s)-1]
	if inner == "" {
		*a = TextArray{}
		return nil
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
	*a = TextArray(out)
	return nil
}

func nullUUIDValue(id uuid.UUID) any {
	if id == uuid.Nil {
		return nil
	}
	return id
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

func mapNotFound(err error) error {
	if isNoRows(err) {
		return domain.ErrNotFound
	}
	return err
}
