package app

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/nexora/identity-service/internal/app/ports"
	"github.com/nexora/identity-service/internal/domain"
)

// SearchPrincipals searches principals within a tenant.
func (d *Deps) SearchPrincipals(ctx context.Context, tenantID uuid.UUID, query string, limit int) ([]domain.Principal, error) {
	if tenantID == uuid.Nil {
		return nil, fmt.Errorf("%w: tenant_id required", domain.ErrInvalidArgument)
	}
	if limit <= 0 {
		limit = 50
	}
	return d.Principals.Search(ctx, tenantID, query, limit)
}

// LockPrincipal locks a principal (admin).
func (d *Deps) LockPrincipal(ctx context.Context, actorID, principalID uuid.UUID, reason string, authCtx ports.AuthContext) error {
	return d.setPrincipalStatus(ctx, actorID, principalID, domain.PrincipalStatusLocked, "principal.lock", reason, authCtx)
}

// SuspendPrincipal suspends a principal (admin).
func (d *Deps) SuspendPrincipal(ctx context.Context, actorID, principalID uuid.UUID, reason string, authCtx ports.AuthContext) error {
	return d.setPrincipalStatus(ctx, actorID, principalID, domain.PrincipalStatusSuspended, "principal.suspend", reason, authCtx)
}

// UnlockPrincipal restores an active status (admin).
func (d *Deps) UnlockPrincipal(ctx context.Context, actorID, principalID uuid.UUID, authCtx ports.AuthContext) error {
	return d.setPrincipalStatus(ctx, actorID, principalID, domain.PrincipalStatusActive, "principal.unlock", "", authCtx)
}

func (d *Deps) setPrincipalStatus(
	ctx context.Context,
	actorID, principalID uuid.UUID,
	status domain.PrincipalStatus,
	action, reason string,
	authCtx ports.AuthContext,
) error {
	if principalID == uuid.Nil {
		return fmt.Errorf("%w: principal_id required", domain.ErrInvalidArgument)
	}
	p, err := d.Principals.GetByID(ctx, principalID)
	if err != nil {
		return err
	}
	if p.Status == domain.PrincipalStatusDeleted {
		return domain.ErrPrincipalDeleted
	}
	now := d.now()
	p.Status = status
	p.UpdatedAt = now
	if err := d.Principals.Update(ctx, p); err != nil {
		return err
	}
	d.publishPrincipal(ctx, string(status), p)
	details := map[string]any{}
	if reason != "" {
		details["reason"] = reason
	}
	d.appendAudit(ctx, domain.AuditEvent{
		TenantID: &p.TenantID, ActorID: &actorID, ActorKind: "admin",
		Action: action, ResourceType: "principal", ResourceID: p.ID.String(),
		Outcome: domain.AuditOutcomeSuccess, IP: authCtx.IP, RequestID: authCtx.RequestID,
		Details: details,
	})
	return nil
}
