package domain

import (
	"fmt"
	"time"

	"github.com/google/uuid"
)

// Streak tracks consecutive engagement days.
type Streak struct {
	ID              uuid.UUID
	TenantID        uuid.UUID
	AccountID       uuid.UUID
	CurrentCount    int
	LongestCount    int
	LastActiveDate  string // YYYY-MM-DD UTC
	Broken          bool
	RecoveryUsed    bool
	UpdatedAt       time.Time
}

// Validate checks streak invariants.
func (s Streak) Validate() error {
	if s.ID == uuid.Nil || s.TenantID == uuid.Nil || s.AccountID == uuid.Nil {
		return fmt.Errorf("%w: streak ids required", ErrInvalidArgument)
	}
	if s.CurrentCount < 0 || s.LongestCount < 0 {
		return fmt.Errorf("%w: streak counts must be non-negative", ErrInvariant)
	}
	return nil
}
