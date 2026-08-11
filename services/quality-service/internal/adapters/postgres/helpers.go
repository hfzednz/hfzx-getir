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
	"github.com/nexora/quality-service/internal/domain"
)

type JSONMap map[string]any

func (m JSONMap) Value() (driver.Value, error) {
	if m == nil {
		return []byte("{}"), nil
	}
	return json.Marshal(map[string]any(m))
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
	var out map[string]any
	if len(b) == 0 {
		*m = JSONMap{}
		return nil
	}
	if err := json.Unmarshal(b, &out); err != nil {
		return err
	}
	if out == nil {
		out = map[string]any{}
	}
	*m = JSONMap(out)
	return nil
}

type FloatMap map[string]float64

func (m FloatMap) Value() (driver.Value, error) {
	if m == nil {
		return []byte("{}"), nil
	}
	return json.Marshal(map[string]float64(m))
}
func (m *FloatMap) Scan(src any) error {
	if m == nil {
		return fmt.Errorf("nil FloatMap")
	}
	if src == nil {
		*m = FloatMap{}
		return nil
	}
	b, err := asBytes(src)
	if err != nil {
		return err
	}
	var out map[string]float64
	if len(b) == 0 {
		*m = FloatMap{}
		return nil
	}
	if err := json.Unmarshal(b, &out); err != nil {
		return err
	}
	if out == nil {
		out = map[string]float64{}
	}
	*m = FloatMap(out)
	return nil
}

type SummaryJSON domain.RunSummary

func (s SummaryJSON) Value() (driver.Value, error) {
	return json.Marshal(domain.RunSummary(s))
}
func (s *SummaryJSON) Scan(src any) error {
	if s == nil {
		return fmt.Errorf("nil SummaryJSON")
	}
	if src == nil {
		*s = SummaryJSON{}
		return nil
	}
	b, err := asBytes(src)
	if err != nil {
		return err
	}
	var out domain.RunSummary
	if len(b) == 0 {
		*s = SummaryJSON{}
		return nil
	}
	if err := json.Unmarshal(b, &out); err != nil {
		return err
	}
	*s = SummaryJSON(out)
	return nil
}

type UUIDArray []uuid.UUID

func (a UUIDArray) Value() (driver.Value, error) {
	if a == nil {
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
	s = strings.TrimPrefix(strings.TrimSuffix(s, "}"), "{")
	parts := strings.Split(s, ",")
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

func asBytes(src any) ([]byte, error) {
	switch v := src.(type) {
	case []byte:
		return v, nil
	case string:
		return []byte(v), nil
	default:
		return nil, fmt.Errorf("unsupported %T", src)
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
func isNoRows(err error) bool { return errors.Is(err, sql.ErrNoRows) }
