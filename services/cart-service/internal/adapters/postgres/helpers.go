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
	"github.com/nexora/cart-service/internal/domain"
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

// JSONAddons encodes []domain.LineAddon as JSONB.
type JSONAddons []domain.LineAddon

func (a JSONAddons) Value() (driver.Value, error) {
	if a == nil {
		return []byte("[]"), nil
	}
	return json.Marshal([]domain.LineAddon(a))
}

func (a *JSONAddons) Scan(src any) error {
	if a == nil {
		return fmt.Errorf("nil JSONAddons destination")
	}
	if src == nil {
		*a = JSONAddons{}
		return nil
	}
	var b []byte
	switch v := src.(type) {
	case []byte:
		b = v
	case string:
		b = []byte(v)
	default:
		return fmt.Errorf("unsupported JSONAddons type %T", src)
	}
	if len(b) == 0 {
		*a = JSONAddons{}
		return nil
	}
	var out []domain.LineAddon
	if err := json.Unmarshal(b, &out); err != nil {
		return err
	}
	if out == nil {
		out = []domain.LineAddon{}
	}
	*a = JSONAddons(out)
	return nil
}

// JSONLineQuotes encodes []domain.LineQuote as JSONB.
type JSONLineQuotes []domain.LineQuote

func (q JSONLineQuotes) Value() (driver.Value, error) {
	if q == nil {
		return []byte("[]"), nil
	}
	return json.Marshal([]domain.LineQuote(q))
}

func (q *JSONLineQuotes) Scan(src any) error {
	if q == nil {
		return fmt.Errorf("nil JSONLineQuotes destination")
	}
	if src == nil {
		*q = JSONLineQuotes{}
		return nil
	}
	var b []byte
	switch v := src.(type) {
	case []byte:
		b = v
	case string:
		b = []byte(v)
	default:
		return fmt.Errorf("unsupported JSONLineQuotes type %T", src)
	}
	if len(b) == 0 {
		*q = JSONLineQuotes{}
		return nil
	}
	var out []domain.LineQuote
	if err := json.Unmarshal(b, &out); err != nil {
		return err
	}
	if out == nil {
		out = []domain.LineQuote{}
	}
	*q = JSONLineQuotes(out)
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
		return fmt.Errorf("%w: %s", domain.ErrAlreadyExists, pgErr.ConstraintName)
	}
	return err
}

func isNoRows(err error) bool {
	return errors.Is(err, sql.ErrNoRows)
}
