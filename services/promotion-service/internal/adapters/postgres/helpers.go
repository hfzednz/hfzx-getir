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
	"github.com/nexora/promotion-service/internal/domain"
)

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
	var b []byte
	switch v := src.(type) {
	case []byte:
		b = v
	case string:
		b = []byte(v)
	default:
		return fmt.Errorf("unsupported JSONMap %T", src)
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
	inner := strings.TrimPrefix(strings.TrimSuffix(s, "}"), "{")
	if inner == "" {
		*a = TextArray{}
		return nil
	}
	parts := strings.Split(inner, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		p = strings.Trim(p, `"`)
		out = append(out, p)
	}
	*a = TextArray(out)
	return nil
}

// UUIDArray encodes []uuid.UUID as PostgreSQL uuid[].
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
		return fmt.Errorf("unsupported UUIDArray %T", src)
	}
	s = strings.TrimSpace(s)
	if s == "" || s == "{}" {
		*a = UUIDArray{}
		return nil
	}
	inner := strings.TrimPrefix(strings.TrimSuffix(s, "}"), "{")
	parts := strings.Split(inner, ",")
	out := make([]uuid.UUID, 0, len(parts))
	for _, p := range parts {
		p = strings.Trim(strings.TrimSpace(p), `"`)
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
		return domain.ErrAlreadyExists
	}
	return err
}

func isNoRows(err error) bool { return errors.Is(err, sql.ErrNoRows) }
