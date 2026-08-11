package domain

import (
	"fmt"
	"regexp"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/google/uuid"
)

// AttributeType enumerates supported attribute data types.
type AttributeType string

const (
	AttributeTypeText      AttributeType = "text"
	AttributeTypeNumber    AttributeType = "number"
	AttributeTypeBoolean   AttributeType = "boolean"
	AttributeTypeDate      AttributeType = "date"
	AttributeTypeColor     AttributeType = "color"
	AttributeTypeSize      AttributeType = "size"
	AttributeTypeWeight    AttributeType = "weight"
	AttributeTypeDimension AttributeType = "dimension"
	AttributeTypeVolume    AttributeType = "volume"
	AttributeTypeEnergy    AttributeType = "energy"
	AttributeTypeNutrition AttributeType = "nutrition"
	AttributeTypeCustom    AttributeType = "custom"
)

func (t AttributeType) Valid() bool {
	switch t {
	case AttributeTypeText, AttributeTypeNumber, AttributeTypeBoolean, AttributeTypeDate,
		AttributeTypeColor, AttributeTypeSize, AttributeTypeWeight, AttributeTypeDimension,
		AttributeTypeVolume, AttributeTypeEnergy, AttributeTypeNutrition, AttributeTypeCustom:
		return true
	default:
		return false
	}
}

var attributeCodeRegexp = regexp.MustCompile(`^[a-z][a-z0-9_]{1,63}$`)

const (
	maxAttributeNameLen = 120
	maxAttributeCodeLen = 64
)

// AttributeDef is a tenant-scoped attribute definition (schema-driven).
type AttributeDef struct {
	ID            uuid.UUID
	TenantID      uuid.UUID
	Code          string
	Name          string
	Description   string
	Type          AttributeType
	Schema        map[string]any
	IsRequired    bool
	IsFilterable  bool
	IsVariantAxis bool
	SortOrder     int
	CreatedAt     time.Time
	UpdatedAt     time.Time
	DeletedAt     *time.Time
}

// Validate checks structural invariants.
func (a AttributeDef) Validate() error {
	if a.ID == uuid.Nil {
		return fmt.Errorf("%w: attribute_def id required", ErrInvalidArgument)
	}
	if a.TenantID == uuid.Nil {
		return fmt.Errorf("%w: tenant_id required", ErrInvalidArgument)
	}
	code := strings.TrimSpace(a.Code)
	if !attributeCodeRegexp.MatchString(code) {
		return fmt.Errorf("%w: invalid attribute code %q", ErrInvalidArgument, a.Code)
	}
	if strings.TrimSpace(a.Name) == "" {
		return fmt.Errorf("%w: attribute name required", ErrInvalidArgument)
	}
	if utf8.RuneCountInString(a.Name) > maxAttributeNameLen {
		return fmt.Errorf("%w: attribute name too long", ErrInvalidArgument)
	}
	if !a.Type.Valid() {
		return fmt.Errorf("%w: invalid attribute type %q", ErrInvalidArgument, a.Type)
	}
	if a.Schema == nil {
		return fmt.Errorf("%w: schema required", ErrInvalidArgument)
	}
	return nil
}

// ProductAttribute is a value of an attribute on a product.
type ProductAttribute struct {
	ID             uuid.UUID
	ProductID      uuid.UUID
	AttributeDefID uuid.UUID
	TenantID       uuid.UUID
	Value          map[string]any
	Locale         string
	CreatedAt      time.Time
	UpdatedAt      time.Time
}

// Validate checks structural invariants.
func (p ProductAttribute) Validate() error {
	if p.ID == uuid.Nil {
		return fmt.Errorf("%w: product_attribute id required", ErrInvalidArgument)
	}
	if p.ProductID == uuid.Nil {
		return fmt.Errorf("%w: product_id required", ErrInvalidArgument)
	}
	if p.AttributeDefID == uuid.Nil {
		return fmt.Errorf("%w: attribute_def_id required", ErrInvalidArgument)
	}
	if p.TenantID == uuid.Nil {
		return fmt.Errorf("%w: tenant_id required", ErrInvalidArgument)
	}
	if p.Value == nil {
		return fmt.Errorf("%w: value required", ErrInvalidArgument)
	}
	if p.Locale != "" {
		if err := ValidateLang(p.Locale); err != nil {
			return err
		}
	}
	return nil
}
