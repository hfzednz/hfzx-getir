package postgres

import (
	"database/sql"
	"database/sql/driver"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/lib/pq"
	"github.com/nexora/data-platform-service/internal/domain"
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
	if err := json.Unmarshal(b, &out); err != nil {
		return err
	}
	if out == nil {
		out = map[string]float64{}
	}
	*m = FloatMap(out)
	return nil
}

type StringMap map[string]string

func (m StringMap) Value() (driver.Value, error) {
	if m == nil {
		return []byte("{}"), nil
	}
	return json.Marshal(map[string]string(m))
}

func (m *StringMap) Scan(src any) error {
	if m == nil {
		return fmt.Errorf("nil StringMap")
	}
	if src == nil {
		*m = StringMap{}
		return nil
	}
	b, err := asBytes(src)
	if err != nil {
		return err
	}
	var out map[string]string
	if err := json.Unmarshal(b, &out); err != nil {
		return err
	}
	if out == nil {
		out = map[string]string{}
	}
	*m = StringMap(out)
	return nil
}

type VariantsJSON []domain.ExperimentVariant

func (v VariantsJSON) Value() (driver.Value, error) {
	if v == nil {
		return []byte("[]"), nil
	}
	return json.Marshal([]domain.ExperimentVariant(v))
}

func (v *VariantsJSON) Scan(src any) error {
	if v == nil {
		return fmt.Errorf("nil VariantsJSON")
	}
	if src == nil {
		*v = VariantsJSON{}
		return nil
	}
	b, err := asBytes(src)
	if err != nil {
		return err
	}
	var out []domain.ExperimentVariant
	if err := json.Unmarshal(b, &out); err != nil {
		return err
	}
	if out == nil {
		out = []domain.ExperimentVariant{}
	}
	*v = VariantsJSON(out)
	return nil
}

func asBytes(src any) ([]byte, error) {
	switch v := src.(type) {
	case []byte:
		return v, nil
	case string:
		return []byte(v), nil
	default:
		return nil, fmt.Errorf("unsupported type %T", src)
	}
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

func isNoRows(err error) bool {
	return errors.Is(err, sql.ErrNoRows)
}

func textArray(ss []string) any {
	if ss == nil {
		return pq.Array([]string{})
	}
	return pq.Array(ss)
}
