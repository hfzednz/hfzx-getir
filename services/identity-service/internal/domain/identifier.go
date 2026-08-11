package domain

import (
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
)

// IdentifierType is a login handle kind.
type IdentifierType string

const (
	IdentifierTypeEmail    IdentifierType = "email"
	IdentifierTypePhone    IdentifierType = "phone"
	IdentifierTypeUsername IdentifierType = "username"
	IdentifierTypeExternal IdentifierType = "external"
)

func (t IdentifierType) Valid() bool {
	switch t {
	case IdentifierTypeEmail, IdentifierTypePhone, IdentifierTypeUsername, IdentifierTypeExternal:
		return true
	default:
		return false
	}
}

// Identifier binds a unique login handle to a principal within a tenant.
type Identifier struct {
	ID          uuid.UUID
	PrincipalID uuid.UUID
	TenantID    uuid.UUID
	Type        IdentifierType
	Value       string
	VerifiedAt  *time.Time
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

// IsVerified reports whether ownership has been proven.
func (i Identifier) IsVerified() bool {
	return i.VerifiedAt != nil
}

// Normalize returns a canonical identifier value for uniqueness checks.
func NormalizeIdentifier(t IdentifierType, value string) string {
	v := strings.TrimSpace(value)
	switch t {
	case IdentifierTypeEmail, IdentifierTypeUsername:
		return strings.ToLower(v)
	default:
		return v
	}
}

// Validate checks structural invariants.
func (i Identifier) Validate() error {
	if i.ID == uuid.Nil {
		return fmt.Errorf("%w: identifier id required", ErrInvalidArgument)
	}
	if i.PrincipalID == uuid.Nil {
		return fmt.Errorf("%w: principal_id required", ErrInvalidArgument)
	}
	if i.TenantID == uuid.Nil {
		return fmt.Errorf("%w: tenant_id required", ErrInvalidArgument)
	}
	if !i.Type.Valid() {
		return fmt.Errorf("%w: invalid identifier type %q", ErrInvalidArgument, i.Type)
	}
	if strings.TrimSpace(i.Value) == "" {
		return fmt.Errorf("%w: identifier value required", ErrInvalidArgument)
	}
	return nil
}
