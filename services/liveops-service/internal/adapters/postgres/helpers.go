package postgres

import (
	"database/sql"
	"database/sql/driver"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/nexora/liveops-service/internal/domain"
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

// JSONRules encodes []domain.TargetRule.
type JSONRules []domain.TargetRule

func (r JSONRules) Value() (driver.Value, error) {
	if r == nil {
		return []byte("[]"), nil
	}
	return json.Marshal([]domain.TargetRule(r))
}

func (r *JSONRules) Scan(src any) error {
	if r == nil {
		return fmt.Errorf("nil JSONRules")
	}
	if src == nil {
		*r = JSONRules{}
		return nil
	}
	b, err := asBytes(src)
	if err != nil {
		return err
	}
	if len(b) == 0 {
		*r = JSONRules{}
		return nil
	}
	var out []domain.TargetRule
	if err := json.Unmarshal(b, &out); err != nil {
		return err
	}
	if out == nil {
		out = []domain.TargetRule{}
	}
	*r = JSONRules(out)
	return nil
}

// JSONVariants encodes []domain.Variant.
type JSONVariants []domain.Variant

func (v JSONVariants) Value() (driver.Value, error) {
	if v == nil {
		return []byte("[]"), nil
	}
	return json.Marshal([]domain.Variant(v))
}

func (v *JSONVariants) Scan(src any) error {
	if v == nil {
		return fmt.Errorf("nil JSONVariants")
	}
	if src == nil {
		*v = JSONVariants{}
		return nil
	}
	b, err := asBytes(src)
	if err != nil {
		return err
	}
	if len(b) == 0 {
		*v = JSONVariants{}
		return nil
	}
	var out []domain.Variant
	if err := json.Unmarshal(b, &out); err != nil {
		return err
	}
	if out == nil {
		out = []domain.Variant{}
	}
	*v = JSONVariants(out)
	return nil
}

// TextArray encodes []string as Postgres text[].
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
		return fmt.Errorf("unsupported TextArray type %T", src)
	}
	s = strings.TrimSpace(s)
	if s == "" || s == "{}" {
		*a = TextArray{}
		return nil
	}
	s = strings.TrimPrefix(s, "{")
	s = strings.TrimSuffix(s, "}")
	if s == "" {
		*a = TextArray{}
		return nil
	}
	// naive split respecting quoted tokens
	out := []string{}
	cur := strings.Builder{}
	inQuote := false
	for i := 0; i < len(s); i++ {
		c := s[i]
		switch {
		case c == '"' && !inQuote:
			inQuote = true
		case c == '"' && inQuote:
			inQuote = false
		case c == ',' && !inQuote:
			out = append(out, cur.String())
			cur.Reset()
		case c == '\\' && inQuote && i+1 < len(s):
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
		return nil, fmt.Errorf("unsupported JSON type %T", src)
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
