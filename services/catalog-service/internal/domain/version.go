package domain

import (
	"fmt"
	"time"
	"unicode/utf8"

	"github.com/google/uuid"
)

const maxChangeSummaryLen = 1000

// ProductVersion is an immutable snapshot of a product aggregate.
type ProductVersion struct {
	ID            uuid.UUID
	ProductID     uuid.UUID
	TenantID      uuid.UUID
	VersionNumber int
	Snapshot      map[string]any
	Status        ProductStatus
	ChangeSummary string
	CreatedBy     *uuid.UUID
	CreatedAt     time.Time
}

// Validate checks structural invariants.
func (v ProductVersion) Validate() error {
	if v.ID == uuid.Nil {
		return fmt.Errorf("%w: product_version id required", ErrInvalidArgument)
	}
	if v.ProductID == uuid.Nil {
		return fmt.Errorf("%w: product_id required", ErrInvalidArgument)
	}
	if v.TenantID == uuid.Nil {
		return fmt.Errorf("%w: tenant_id required", ErrInvalidArgument)
	}
	if v.VersionNumber <= 0 {
		return fmt.Errorf("%w: version_number must be > 0", ErrInvalidArgument)
	}
	if v.Snapshot == nil {
		return fmt.Errorf("%w: snapshot required", ErrInvalidArgument)
	}
	if !v.Status.Valid() {
		return fmt.Errorf("%w: invalid version status %q", ErrInvalidArgument, v.Status)
	}
	if utf8.RuneCountInString(v.ChangeSummary) > maxChangeSummaryLen {
		return fmt.Errorf("%w: change_summary too long", ErrInvalidArgument)
	}
	if v.CreatedBy != nil && *v.CreatedBy == uuid.Nil {
		return fmt.Errorf("%w: created_by must not be nil UUID", ErrInvalidArgument)
	}
	return nil
}
