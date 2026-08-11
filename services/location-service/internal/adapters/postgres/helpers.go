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
	"github.com/nexora/location-service/internal/domain"
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

// JSONComponents marshals AddressComponents to/from JSONB.
type JSONComponents domain.AddressComponents

func (c JSONComponents) Value() (driver.Value, error) {
	b, err := json.Marshal(domain.AddressComponents(c))
	if err != nil {
		return nil, err
	}
	return b, nil
}

func (c *JSONComponents) Scan(src any) error {
	if c == nil {
		return fmt.Errorf("nil JSONComponents destination")
	}
	if src == nil {
		*c = JSONComponents{}
		return nil
	}
	var b []byte
	switch v := src.(type) {
	case []byte:
		b = v
	case string:
		b = []byte(v)
	default:
		return fmt.Errorf("unsupported JSONComponents type %T", src)
	}
	if len(b) == 0 {
		*c = JSONComponents{}
		return nil
	}
	var out domain.AddressComponents
	if err := json.Unmarshal(b, &out); err != nil {
		return err
	}
	*c = JSONComponents(out)
	return nil
}

// JSONGeocode marshals GeocodeResult to/from JSONB.
type JSONGeocode domain.GeocodeResult

func (g JSONGeocode) Value() (driver.Value, error) {
	b, err := json.Marshal(domain.GeocodeResult(g))
	if err != nil {
		return nil, err
	}
	return b, nil
}

func (g *JSONGeocode) Scan(src any) error {
	if g == nil {
		return fmt.Errorf("nil JSONGeocode destination")
	}
	if src == nil {
		*g = JSONGeocode{}
		return nil
	}
	var b []byte
	switch v := src.(type) {
	case []byte:
		b = v
	case string:
		b = []byte(v)
	default:
		return fmt.Errorf("unsupported JSONGeocode type %T", src)
	}
	if len(b) == 0 {
		*g = JSONGeocode{}
		return nil
	}
	var out domain.GeocodeResult
	if err := json.Unmarshal(b, &out); err != nil {
		return err
	}
	*g = JSONGeocode(out)
	return nil
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
