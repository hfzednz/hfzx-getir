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
	"github.com/nexora/ai-platform-service/internal/domain"
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

// JSONFloatMap marshals map[string]float64.
type JSONFloatMap map[string]float64

func (m JSONFloatMap) Value() (driver.Value, error) {
	if m == nil {
		return []byte("{}"), nil
	}
	return json.Marshal(map[string]float64(m))
}

func (m *JSONFloatMap) Scan(src any) error {
	if m == nil {
		return fmt.Errorf("nil JSONFloatMap")
	}
	if src == nil {
		*m = JSONFloatMap{}
		return nil
	}
	b, err := asBytes(src)
	if err != nil {
		return err
	}
	if len(b) == 0 {
		*m = JSONFloatMap{}
		return nil
	}
	var out map[string]float64
	if err := json.Unmarshal(b, &out); err != nil {
		return err
	}
	if out == nil {
		out = map[string]float64{}
	}
	*m = JSONFloatMap(out)
	return nil
}

// JSONStringMap marshals map[string]string.
type JSONStringMap map[string]string

func (m JSONStringMap) Value() (driver.Value, error) {
	if m == nil {
		return []byte("{}"), nil
	}
	return json.Marshal(map[string]string(m))
}

func (m *JSONStringMap) Scan(src any) error {
	if m == nil {
		return fmt.Errorf("nil JSONStringMap")
	}
	if src == nil {
		*m = JSONStringMap{}
		return nil
	}
	b, err := asBytes(src)
	if err != nil {
		return err
	}
	if len(b) == 0 {
		*m = JSONStringMap{}
		return nil
	}
	var out map[string]string
	if err := json.Unmarshal(b, &out); err != nil {
		return err
	}
	if out == nil {
		out = map[string]string{}
	}
	*m = JSONStringMap(out)
	return nil
}

// JSONSteps encodes []domain.AgentStep.
type JSONSteps []domain.AgentStep

func (s JSONSteps) Value() (driver.Value, error) {
	if s == nil {
		return []byte("[]"), nil
	}
	return json.Marshal([]domain.AgentStep(s))
}

func (s *JSONSteps) Scan(src any) error {
	if s == nil {
		return fmt.Errorf("nil JSONSteps")
	}
	if src == nil {
		*s = JSONSteps{}
		return nil
	}
	b, err := asBytes(src)
	if err != nil {
		return err
	}
	if len(b) == 0 {
		*s = JSONSteps{}
		return nil
	}
	var out []domain.AgentStep
	if err := json.Unmarshal(b, &out); err != nil {
		return err
	}
	if out == nil {
		out = []domain.AgentStep{}
	}
	*s = JSONSteps(out)
	return nil
}

// JSONConditions encodes []domain.RuleCondition.
type JSONConditions []domain.RuleCondition

func (c JSONConditions) Value() (driver.Value, error) {
	if c == nil {
		return []byte("[]"), nil
	}
	return json.Marshal([]domain.RuleCondition(c))
}

func (c *JSONConditions) Scan(src any) error {
	if c == nil {
		return fmt.Errorf("nil JSONConditions")
	}
	if src == nil {
		*c = JSONConditions{}
		return nil
	}
	b, err := asBytes(src)
	if err != nil {
		return err
	}
	if len(b) == 0 {
		*c = JSONConditions{}
		return nil
	}
	var out []domain.RuleCondition
	if err := json.Unmarshal(b, &out); err != nil {
		return err
	}
	if out == nil {
		out = []domain.RuleCondition{}
	}
	*c = JSONConditions(out)
	return nil
}

// JSONActions encodes []domain.RuleAction.
type JSONActions []domain.RuleAction

func (a JSONActions) Value() (driver.Value, error) {
	if a == nil {
		return []byte("[]"), nil
	}
	return json.Marshal([]domain.RuleAction(a))
}

func (a *JSONActions) Scan(src any) error {
	if a == nil {
		return fmt.Errorf("nil JSONActions")
	}
	if src == nil {
		*a = JSONActions{}
		return nil
	}
	b, err := asBytes(src)
	if err != nil {
		return err
	}
	if len(b) == 0 {
		*a = JSONActions{}
		return nil
	}
	var out []domain.RuleAction
	if err := json.Unmarshal(b, &out); err != nil {
		return err
	}
	if out == nil {
		out = []domain.RuleAction{}
	}
	*a = JSONActions(out)
	return nil
}

func asBytes(src any) ([]byte, error) {
	switch v := src.(type) {
	case []byte:
		return v, nil
	case string:
		return []byte(v), nil
	default:
		return nil, fmt.Errorf("unsupported JSON type %T", src)
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
