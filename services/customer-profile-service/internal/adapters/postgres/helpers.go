package postgres

import (
	"database/sql"
	"database/sql/driver"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/nexora/customer-profile-service/internal/domain"
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

// JSONRaw marshals arbitrary JSON values.
type JSONRaw struct{ V any }

func (j JSONRaw) Value() (driver.Value, error) {
	if j.V == nil {
		return []byte("null"), nil
	}
	return json.Marshal(j.V)
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
		p = strings.TrimSpace(strings.Trim(p, `"`))
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

// IntArray encodes []int as PostgreSQL int[].
type IntArray []int

func (a IntArray) Value() (driver.Value, error) {
	if a == nil || len(a) == 0 {
		return "{}", nil
	}
	parts := make([]string, len(a))
	for i, n := range a {
		parts[i] = fmt.Sprintf("%d", n)
	}
	return "{" + strings.Join(parts, ",") + "}", nil
}

func (a *IntArray) Scan(src any) error {
	if a == nil {
		return fmt.Errorf("nil IntArray destination")
	}
	if src == nil {
		*a = IntArray{}
		return nil
	}
	var s string
	switch v := src.(type) {
	case []byte:
		s = string(v)
	case string:
		s = v
	default:
		return fmt.Errorf("unsupported IntArray type %T", src)
	}
	s = strings.TrimSpace(s)
	if s == "" || s == "{}" {
		*a = IntArray{}
		return nil
	}
	if s[0] != '{' || s[len(s)-1] != '}' {
		return fmt.Errorf("invalid int array: %q", s)
	}
	inner := s[1 : len(s)-1]
	if inner == "" {
		*a = IntArray{}
		return nil
	}
	parts := strings.Split(inner, ",")
	out := make([]int, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p == "" {
			continue
		}
		var n int
		if _, err := fmt.Sscanf(p, "%d", &n); err != nil {
			return err
		}
		out = append(out, n)
	}
	*a = IntArray(out)
	return nil
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

func nullIP(ip net.IP) any {
	if len(ip) == 0 {
		return nil
	}
	return ip.String()
}

func scanIP(ns sql.NullString) net.IP {
	if !ns.Valid || ns.String == "" {
		return nil
	}
	return net.ParseIP(ns.String)
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

func rowsAffectedOrNotFound(res sql.Result, err error) error {
	if err != nil {
		return mapUniqueViolation(err)
	}
	n, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if n == 0 {
		return domain.ErrNotFound
	}
	return nil
}

func metaGetMap(m map[string]any) map[string]any {
	if m == nil {
		return map[string]any{}
	}
	return m
}

type scannable interface {
	Scan(dest ...any) error
}
