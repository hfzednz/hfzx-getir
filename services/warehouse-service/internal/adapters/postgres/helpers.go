package postgres

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/nexora/warehouse-service/internal/domain"
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

// JSONRaw marshals arbitrary JSON values to/from JSONB.
type JSONRaw struct {
	V any
}

func (j JSONRaw) Value() (driver.Value, error) {
	if j.V == nil {
		return []byte("null"), nil
	}
	b, err := json.Marshal(j.V)
	if err != nil {
		return nil, err
	}
	return b, nil
}

// TextArray encodes []string as PostgreSQL text[] or JSONB arrays interchangeably.
type TextArray []string

func (a TextArray) Value() (driver.Value, error) {
	if a == nil {
		return []byte("[]"), nil
	}
	b, err := json.Marshal([]string(a))
	if err != nil {
		return nil, err
	}
	return b, nil
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
	return *t
}

func nullTimeValue(t time.Time) any {
	if t.IsZero() {
		return nil
	}
	return t
}

func scanNullTime(nt sql.NullTime) *time.Time {
	if !nt.Valid {
		return nil
	}
	t := nt.Time.UTC()
	return &t
}

func scanTimeOrZero(nt sql.NullTime) time.Time {
	if !nt.Valid {
		return time.Time{}
	}
	return nt.Time.UTC()
}

func nullInt64(v *int64) any {
	if v == nil {
		return nil
	}
	return *v
}

func nullInt(v *int) any {
	if v == nil {
		return nil
	}
	return *v
}

func scanNullInt64(n sql.NullInt64) *int64 {
	if !n.Valid {
		return nil
	}
	v := n.Int64
	return &v
}

func scanNullInt(n sql.NullInt64) *int {
	if !n.Valid {
		return nil
	}
	v := int(n.Int64)
	return &v
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

// ensureWarehouse upserts a minimal warehouses row so FK children can be inserted.
func ensureWarehouse(ctx context.Context, db execer, tenantID, warehouseID uuid.UUID) error {
	if warehouseID == uuid.Nil {
		return fmt.Errorf("%w: warehouse_id required", domain.ErrInvalidArgument)
	}
	if tenantID == uuid.Nil {
		tenantID = uuid.Nil
	}
	code := "WH" + strings.ReplaceAll(warehouseID.String(), "-", "")[:12]
	_, err := db.ExecContext(ctx, `
		INSERT INTO warehouses (id, tenant_id, code, name, type, status, timezone, metadata, created_at, updated_at)
		VALUES ($1, $2, $3, '', 'dark_store', 'active', 'UTC', '{}'::jsonb, now(), now())
		ON CONFLICT (id) DO NOTHING`,
		warehouseID, tenantID, code,
	)
	return err
}

type execer interface {
	ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error)
}

func mergeMeta(base map[string]any, extra map[string]any) JSONMap {
	out := map[string]any{}
	for k, v := range base {
		out[k] = v
	}
	for k, v := range extra {
		out[k] = v
	}
	return JSONMap(out)
}

func metaGetMap(m map[string]any) map[string]any {
	if m == nil {
		return map[string]any{}
	}
	return m
}

func zoneCode(s domain.Station) string {
	if s.ZoneCode != "" {
		return s.ZoneCode
	}
	return s.Zone
}

func lineQty(l domain.FulfillmentLine) int {
	if l.Qty > 0 {
		return l.Qty
	}
	return int(l.OrderedQty())
}
