package domain

import (
	"fmt"
	"time"

	"github.com/google/uuid"
)

// RelationType classifies product-to-product links.
type RelationType string

const (
	RelationTypeRelated        RelationType = "related"
	RelationTypeAlternative    RelationType = "alternative"
	RelationTypeAccessory      RelationType = "accessory"
	RelationTypeReplacement    RelationType = "replacement"
	RelationTypeComplementary  RelationType = "complementary"
	RelationTypeAI             RelationType = "ai"
)

func (t RelationType) Valid() bool {
	switch t {
	case RelationTypeRelated, RelationTypeAlternative, RelationTypeAccessory,
		RelationTypeReplacement, RelationTypeComplementary, RelationTypeAI:
		return true
	default:
		return false
	}
}

// ProductRelation is a directed merchandising / AI relation.
type ProductRelation struct {
	ID              uuid.UUID
	TenantID        uuid.UUID
	SourceProductID uuid.UUID
	TargetProductID uuid.UUID
	Type            RelationType
	SortOrder       int
	Score           *float64
	Metadata        map[string]any
	CreatedAt       time.Time
	UpdatedAt       time.Time
}

// Validate checks structural invariants.
func (r ProductRelation) Validate() error {
	if r.ID == uuid.Nil {
		return fmt.Errorf("%w: relation id required", ErrInvalidArgument)
	}
	if r.TenantID == uuid.Nil {
		return fmt.Errorf("%w: tenant_id required", ErrInvalidArgument)
	}
	if r.SourceProductID == uuid.Nil {
		return fmt.Errorf("%w: source_product_id required", ErrInvalidArgument)
	}
	if r.TargetProductID == uuid.Nil {
		return fmt.Errorf("%w: target_product_id required", ErrInvalidArgument)
	}
	if r.SourceProductID == r.TargetProductID {
		return fmt.Errorf("%w: product cannot relate to itself", ErrInvariant)
	}
	if !r.Type.Valid() {
		return fmt.Errorf("%w: invalid relation type %q", ErrInvalidArgument, r.Type)
	}
	if r.Score != nil && (*r.Score < 0 || *r.Score > 1) {
		return fmt.Errorf("%w: score must be in [0,1]", ErrInvalidArgument)
	}
	return nil
}
