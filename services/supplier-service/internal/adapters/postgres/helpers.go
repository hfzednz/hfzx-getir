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
	"github.com/nexora/supplier-service/internal/domain"
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

func marshalJSON(v any) (driver.Value, error) {
	if v == nil {
		return []byte("[]"), nil
	}
	return json.Marshal(v)
}

func unmarshalJSON(src any, dest any) error {
	if src == nil {
		return nil
	}
	b, err := asBytes(src)
	if err != nil {
		return err
	}
	if len(b) == 0 {
		return nil
	}
	return json.Unmarshal(b, dest)
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

func nullBool(b *bool) any {
	if b == nil {
		return nil
	}
	return *b
}

func scanNullBool(nb sql.NullBool) *bool {
	if !nb.Valid {
		return nil
	}
	v := nb.Bool
	return &v
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

func partnerKindsToStrings(kinds []domain.PartnerKind) []string {
	out := make([]string, len(kinds))
	for i, k := range kinds {
		out[i] = string(k)
	}
	return out
}

func stringsToPartnerKinds(ss []string) []domain.PartnerKind {
	out := make([]domain.PartnerKind, len(ss))
	for i, s := range ss {
		out[i] = domain.PartnerKind(s)
	}
	return out
}
