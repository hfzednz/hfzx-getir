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
	"github.com/nexora/checkout-service/internal/domain"
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

// JSONAddress encodes domain.AddressSnapshot as JSONB.
type JSONAddress domain.AddressSnapshot

func (a JSONAddress) Value() (driver.Value, error) {
	return json.Marshal(domain.AddressSnapshot(a))
}

func (a *JSONAddress) Scan(src any) error {
	if a == nil {
		return fmt.Errorf("nil JSONAddress destination")
	}
	var out domain.AddressSnapshot
	if err := scanJSON(src, &out); err != nil {
		return err
	}
	*a = JSONAddress(out)
	return nil
}

// JSONSlot encodes domain.SlotSnapshot as JSONB.
type JSONSlot domain.SlotSnapshot

func (s JSONSlot) Value() (driver.Value, error) {
	return json.Marshal(domain.SlotSnapshot(s))
}

func (s *JSONSlot) Scan(src any) error {
	if s == nil {
		return fmt.Errorf("nil JSONSlot destination")
	}
	var out domain.SlotSnapshot
	if err := scanJSON(src, &out); err != nil {
		return err
	}
	*s = JSONSlot(out)
	return nil
}

// JSONGift encodes domain.GiftPrefs as JSONB.
type JSONGift domain.GiftPrefs

func (g JSONGift) Value() (driver.Value, error) {
	return json.Marshal(domain.GiftPrefs(g))
}

func (g *JSONGift) Scan(src any) error {
	if g == nil {
		return fmt.Errorf("nil JSONGift destination")
	}
	var out domain.GiftPrefs
	if err := scanJSON(src, &out); err != nil {
		return err
	}
	*g = JSONGift(out)
	return nil
}

// JSONInvoice encodes domain.InvoicePrefs as JSONB.
type JSONInvoice domain.InvoicePrefs

func (i JSONInvoice) Value() (driver.Value, error) {
	return json.Marshal(domain.InvoicePrefs(i))
}

func (i *JSONInvoice) Scan(src any) error {
	if i == nil {
		return fmt.Errorf("nil JSONInvoice destination")
	}
	var out domain.InvoicePrefs
	if err := scanJSON(src, &out); err != nil {
		return err
	}
	*i = JSONInvoice(out)
	return nil
}

// JSONValidation encodes domain.ValidationResults as JSONB.
type JSONValidation domain.ValidationResults

func (v JSONValidation) Value() (driver.Value, error) {
	return json.Marshal(domain.ValidationResults(v))
}

func (v *JSONValidation) Scan(src any) error {
	if v == nil {
		return fmt.Errorf("nil JSONValidation destination")
	}
	var out domain.ValidationResults
	if err := scanJSON(src, &out); err != nil {
		return err
	}
	*v = JSONValidation(out)
	return nil
}

// JSONQuote encodes domain.QuoteSnapshot as JSONB.
type JSONQuote domain.QuoteSnapshot

func (q JSONQuote) Value() (driver.Value, error) {
	return json.Marshal(domain.QuoteSnapshot(q))
}

func (q *JSONQuote) Scan(src any) error {
	if q == nil {
		return fmt.Errorf("nil JSONQuote destination")
	}
	var out domain.QuoteSnapshot
	if err := scanJSON(src, &out); err != nil {
		return err
	}
	*q = JSONQuote(out)
	return nil
}

// StringArray encodes []string as a PostgreSQL TEXT[].
type StringArray []string

func (a StringArray) Value() (driver.Value, error) {
	if a == nil {
		return "{}", nil
	}
	if len(a) == 0 {
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

func (a *StringArray) Scan(src any) error {
	if a == nil {
		return fmt.Errorf("nil StringArray destination")
	}
	if src == nil {
		*a = StringArray{}
		return nil
	}
	var s string
	switch v := src.(type) {
	case []byte:
		s = string(v)
	case string:
		s = v
	default:
		return fmt.Errorf("unsupported StringArray type %T", src)
	}
	s = strings.TrimSpace(s)
	if s == "" || s == "{}" {
		*a = StringArray{}
		return nil
	}
	if s[0] != '{' || s[len(s)-1] != '}' {
		return fmt.Errorf("invalid text array: %q", s)
	}
	inner := s[1 : len(s)-1]
	if inner == "" {
		*a = StringArray{}
		return nil
	}
	// Prefer JSON-style parse when possible; fall back to pq-like split.
	parts, err := splitPGTextArray(inner)
	if err != nil {
		return err
	}
	*a = StringArray(parts)
	return nil
}

func splitPGTextArray(inner string) ([]string, error) {
	out := make([]string, 0)
	var cur strings.Builder
	inQuote := false
	escape := false
	for i := 0; i < len(inner); i++ {
		c := inner[i]
		if escape {
			cur.WriteByte(c)
			escape = false
			continue
		}
		if c == '\\' && inQuote {
			escape = true
			continue
		}
		if c == '"' {
			inQuote = !inQuote
			continue
		}
		if c == ',' && !inQuote {
			out = append(out, cur.String())
			cur.Reset()
			continue
		}
		cur.WriteByte(c)
	}
	out = append(out, cur.String())
	return out, nil
}

func scanJSON(src any, dest any) error {
	if src == nil {
		return nil
	}
	var b []byte
	switch v := src.(type) {
	case []byte:
		b = v
	case string:
		b = []byte(v)
	default:
		return fmt.Errorf("unsupported JSON type %T", src)
	}
	if len(b) == 0 {
		return nil
	}
	return json.Unmarshal(b, dest)
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

func nullString(s string) any {
	if s == "" {
		return nil
	}
	return s
}

func scanNullString(ns sql.NullString) string {
	if !ns.Valid {
		return ""
	}
	return ns.String
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
