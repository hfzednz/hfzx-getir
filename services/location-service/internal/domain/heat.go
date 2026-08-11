package domain

import (
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
)

// HeatCell is a demand/density aggregate for a grid cell.
type HeatCell struct {
	ID          uuid.UUID
	TenantID    uuid.UUID
	GridCell    string
	DemandScore float64
	Density     float64
	UpdatedAt   time.Time
}

// Validate checks heat cell invariants.
func (h HeatCell) Validate() error {
	if h.ID == uuid.Nil {
		return fmt.Errorf("%w: heat cell id required", ErrInvalidArgument)
	}
	if h.TenantID == uuid.Nil {
		return fmt.Errorf("%w: tenant_id required", ErrInvalidArgument)
	}
	if strings.TrimSpace(h.GridCell) == "" {
		return fmt.Errorf("%w: grid_cell required", ErrInvalidArgument)
	}
	if h.DemandScore < 0 {
		return fmt.Errorf("%w: demand_score must be >= 0", ErrInvalidArgument)
	}
	if h.Density < 0 {
		return fmt.Errorf("%w: density must be >= 0", ErrInvalidArgument)
	}
	return nil
}
