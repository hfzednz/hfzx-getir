package app

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/nexora/identity-service/internal/app/ports"
	"github.com/nexora/identity-service/internal/domain"
)

// AssignRoleInput binds a role to a principal with optional scope and expiry.
type AssignRoleInput struct {
	PrincipalID uuid.UUID
	RoleID      uuid.UUID
	Scope       domain.Scope
	GrantedBy   uuid.UUID
	ExpiresAt   *time.Time
	AuthCtx     ports.AuthContext
}

// AssignRole creates a principal↔role binding.
func (d *Deps) AssignRole(ctx context.Context, in AssignRoleInput) (domain.PrincipalRole, error) {
	if in.PrincipalID == uuid.Nil || in.RoleID == uuid.Nil {
		return domain.PrincipalRole{}, fmt.Errorf("%w: principal_id and role_id required", domain.ErrInvalidArgument)
	}
	if _, err := d.Principals.GetByID(ctx, in.PrincipalID); err != nil {
		return domain.PrincipalRole{}, err
	}
	if _, err := d.Roles.GetRole(ctx, in.RoleID); err != nil {
		return domain.PrincipalRole{}, err
	}
	now := d.now()
	pr := domain.PrincipalRole{
		ID: d.newID(), PrincipalID: in.PrincipalID, RoleID: in.RoleID,
		Scope: in.Scope, CreatedAt: now, ExpiresAt: in.ExpiresAt,
	}
	if in.GrantedBy != uuid.Nil {
		pr.GrantedBy = &in.GrantedBy
	}
	if err := pr.Validate(); err != nil {
		return domain.PrincipalRole{}, err
	}
	if err := d.Roles.AssignRole(ctx, pr); err != nil {
		return domain.PrincipalRole{}, err
	}
	d.appendAudit(ctx, domain.AuditEvent{
		ActorID: &in.GrantedBy, ActorKind: "admin",
		Action: "rbac.assign_role", ResourceType: "principal_role", ResourceID: pr.ID.String(),
		Outcome: domain.AuditOutcomeSuccess, IP: in.AuthCtx.IP,
		Details: map[string]any{"principal_id": in.PrincipalID.String(), "role_id": in.RoleID.String()},
	})
	return pr, nil
}

// TemporaryGrantInput creates a time-boxed direct permission grant.
type TemporaryGrantInput struct {
	PrincipalID  uuid.UUID
	PermissionID uuid.UUID
	Scope        domain.Scope
	Reason       string
	GrantedBy    uuid.UUID
	ExpiresAt    time.Time
	AuthCtx      ports.AuthContext
}

// TemporaryGrant creates a temporary permission assignment.
func (d *Deps) TemporaryGrant(ctx context.Context, in TemporaryGrantInput) (domain.TemporaryGrant, error) {
	if in.PrincipalID == uuid.Nil || in.PermissionID == uuid.Nil {
		return domain.TemporaryGrant{}, fmt.Errorf("%w: principal_id and permission_id required", domain.ErrInvalidArgument)
	}
	now := d.now()
	if !in.ExpiresAt.After(now) {
		return domain.TemporaryGrant{}, fmt.Errorf("%w: expires_at must be in the future", domain.ErrInvalidArgument)
	}
	if _, err := d.Principals.GetByID(ctx, in.PrincipalID); err != nil {
		return domain.TemporaryGrant{}, err
	}
	if _, err := d.Roles.GetPermission(ctx, in.PermissionID); err != nil {
		return domain.TemporaryGrant{}, err
	}
	g := domain.TemporaryGrant{
		ID: d.newID(), PrincipalID: in.PrincipalID, PermissionID: in.PermissionID,
		Scope: in.Scope, Reason: in.Reason, CreatedAt: now, ExpiresAt: in.ExpiresAt,
	}
	if in.GrantedBy != uuid.Nil {
		g.GrantedBy = &in.GrantedBy
	}
	if err := g.Validate(); err != nil {
		return domain.TemporaryGrant{}, err
	}
	if err := d.Roles.CreateTemporaryGrant(ctx, g); err != nil {
		return domain.TemporaryGrant{}, err
	}
	d.appendAudit(ctx, domain.AuditEvent{
		ActorID: &in.GrantedBy, ActorKind: "admin",
		Action: "rbac.temporary_grant", ResourceType: "temporary_grant", ResourceID: g.ID.String(),
		Outcome: domain.AuditOutcomeSuccess, IP: in.AuthCtx.IP,
		Details: map[string]any{"reason": in.Reason},
	})
	return g, nil
}
