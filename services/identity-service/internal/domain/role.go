package domain

import (
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
)

// RoleKind classifies role definitions.
type RoleKind string

const (
	RoleKindPlatform   RoleKind = "platform"
	RoleKindTenant     RoleKind = "tenant"
	RoleKindDepartment RoleKind = "department"
	RoleKindCustom     RoleKind = "custom"
)

func (k RoleKind) Valid() bool {
	switch k {
	case RoleKindPlatform, RoleKindTenant, RoleKindDepartment, RoleKindCustom:
		return true
	default:
		return false
	}
}

// Role is a named permission bundle, optionally tenant-scoped.
type Role struct {
	ID          uuid.UUID
	TenantID    *uuid.UUID
	Name        string
	Kind        RoleKind
	Description string
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

func (r Role) Validate() error {
	if r.ID == uuid.Nil {
		return fmt.Errorf("%w: role id required", ErrInvalidArgument)
	}
	if strings.TrimSpace(r.Name) == "" {
		return fmt.Errorf("%w: role name required", ErrInvalidArgument)
	}
	if !r.Kind.Valid() {
		return fmt.Errorf("%w: invalid role kind %q", ErrInvalidArgument, r.Kind)
	}
	if r.Kind == RoleKindPlatform && r.TenantID != nil {
		return fmt.Errorf("%w: platform role must not have tenant_id", ErrInvariant)
	}
	if r.Kind == RoleKindTenant && r.TenantID == nil {
		return fmt.Errorf("%w: tenant role requires tenant_id", ErrInvariant)
	}
	return nil
}

// Scope binds an authorization grant to optional org dimensions.
type Scope struct {
	TenantID    *uuid.UUID
	CityID      *uuid.UUID
	WarehouseID *uuid.UUID
}

// Matches reports whether the request attributes are covered by this scope.
// A nil scope dimension is unrestricted for that axis.
func (s Scope) Matches(tenantID, cityID, warehouseID *uuid.UUID) bool {
	if s.TenantID != nil {
		if tenantID == nil || *s.TenantID != *tenantID {
			return false
		}
	}
	if s.CityID != nil {
		if cityID == nil || *s.CityID != *cityID {
			return false
		}
	}
	if s.WarehouseID != nil {
		if warehouseID == nil || *s.WarehouseID != *warehouseID {
			return false
		}
	}
	return true
}

// PrincipalRole is a scoped role binding on a principal.
type PrincipalRole struct {
	ID          uuid.UUID
	PrincipalID uuid.UUID
	RoleID      uuid.UUID
	Scope       Scope
	GrantedBy   *uuid.UUID
	CreatedAt   time.Time
	ExpiresAt   *time.Time
}

func (pr PrincipalRole) IsActive(now time.Time) bool {
	return pr.ExpiresAt == nil || now.Before(*pr.ExpiresAt)
}

func (pr PrincipalRole) Validate() error {
	if pr.ID == uuid.Nil {
		return fmt.Errorf("%w: principal_role id required", ErrInvalidArgument)
	}
	if pr.PrincipalID == uuid.Nil {
		return fmt.Errorf("%w: principal_id required", ErrInvalidArgument)
	}
	if pr.RoleID == uuid.Nil {
		return fmt.Errorf("%w: role_id required", ErrInvalidArgument)
	}
	return nil
}

// TemporaryGrant is a time-boxed direct permission assignment.
type TemporaryGrant struct {
	ID           uuid.UUID
	PrincipalID  uuid.UUID
	PermissionID uuid.UUID
	Scope        Scope
	Reason       string
	GrantedBy    *uuid.UUID
	CreatedAt    time.Time
	ExpiresAt    time.Time
	RevokedAt    *time.Time
}

func (g TemporaryGrant) IsActive(now time.Time) bool {
	return g.RevokedAt == nil && now.Before(g.ExpiresAt)
}

func (g TemporaryGrant) Validate() error {
	if g.ID == uuid.Nil {
		return fmt.Errorf("%w: temporary_grant id required", ErrInvalidArgument)
	}
	if g.PrincipalID == uuid.Nil {
		return fmt.Errorf("%w: principal_id required", ErrInvalidArgument)
	}
	if g.PermissionID == uuid.Nil {
		return fmt.Errorf("%w: permission_id required", ErrInvalidArgument)
	}
	if !g.ExpiresAt.After(g.CreatedAt) {
		return fmt.Errorf("%w: expires_at must be after created_at", ErrInvariant)
	}
	return nil
}
