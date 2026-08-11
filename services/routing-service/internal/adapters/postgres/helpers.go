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
	"github.com/nexora/routing-service/internal/domain"
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

type waypointDTO struct {
	ID       uuid.UUID  `json:"id"`
	Sequence int        `json:"sequence"`
	Kind     string     `json:"kind"`
	Lat      float64    `json:"lat"`
	Lon      float64    `json:"lon"`
	OrderID  *uuid.UUID `json:"orderId,omitempty"`
	Label    string     `json:"label,omitempty"`
	ETAAt    *time.Time `json:"etaAt,omitempty"`
}

// WaypointsJSON encodes []domain.Waypoint as JSONB.
type WaypointsJSON []domain.Waypoint

func (w WaypointsJSON) Value() (driver.Value, error) {
	if w == nil {
		return []byte("[]"), nil
	}
	dtos := make([]waypointDTO, len(w))
	for i, wp := range w {
		dtos[i] = waypointDTO{
			ID: wp.ID, Sequence: wp.Sequence, Kind: string(wp.Kind),
			Lat: wp.Lat, Lon: wp.Lon, OrderID: wp.OrderID, Label: wp.Label, ETAAt: wp.ETAAt,
		}
	}
	b, err := json.Marshal(dtos)
	if err != nil {
		return nil, err
	}
	return b, nil
}

func (w *WaypointsJSON) Scan(src any) error {
	if w == nil {
		return fmt.Errorf("nil WaypointsJSON destination")
	}
	if src == nil {
		*w = WaypointsJSON{}
		return nil
	}
	var b []byte
	switch v := src.(type) {
	case []byte:
		b = v
	case string:
		b = []byte(v)
	default:
		return fmt.Errorf("unsupported WaypointsJSON type %T", src)
	}
	if len(b) == 0 {
		*w = WaypointsJSON{}
		return nil
	}
	var dtos []waypointDTO
	if err := json.Unmarshal(b, &dtos); err != nil {
		return err
	}
	out := make([]domain.Waypoint, 0, len(dtos))
	for _, d := range dtos {
		out = append(out, domain.Waypoint{
			ID: d.ID, Sequence: d.Sequence, Kind: domain.WaypointKind(d.Kind),
			Lat: d.Lat, Lon: d.Lon, OrderID: d.OrderID, Label: d.Label, ETAAt: d.ETAAt,
		})
	}
	*w = WaypointsJSON(out)
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

func mapNotFound(err error) error {
	if isNoRows(err) {
		return fmt.Errorf("%w", domain.ErrNotFound)
	}
	return err
}
