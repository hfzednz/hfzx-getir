package app

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/nexora/identity-service/internal/app/ports"
	"github.com/nexora/identity-service/internal/domain"
)

// CreateGuestSessionInput creates a limited anonymous principal + session.
type CreateGuestSessionInput struct {
	TenantID uuid.UUID
	AuthCtx  ports.AuthContext
}

// CreateGuestSession provisions a guest principal and issues a limited session.
func (d *Deps) CreateGuestSession(ctx context.Context, in CreateGuestSessionInput) (AuthResult, error) {
	if in.TenantID == uuid.Nil {
		return AuthResult{}, fmt.Errorf("%w: tenant_id required", domain.ErrInvalidArgument)
	}
	now := d.now()
	p := domain.Principal{
		ID: d.newID(), TenantID: in.TenantID, Kind: domain.PrincipalKindGuest,
		Status: domain.PrincipalStatusActive, DisplayName: "Guest",
		CreatedAt: now, UpdatedAt: now,
	}
	if err := d.Principals.Create(ctx, p); err != nil {
		return AuthResult{}, err
	}
	d.publishPrincipal(ctx, "created", p)
	score := d.scoreSignals(in.AuthCtx.Signals)
	res, err := d.createSessionWithTokens(ctx, p, []string{"guest"}, "aal0", nil, in.AuthCtx, score)
	if err != nil {
		return AuthResult{}, err
	}
	d.appendAudit(ctx, domain.AuditEvent{
		TenantID: &in.TenantID, ActorID: &p.ID, ActorKind: string(p.Kind),
		Action: "auth.guest.create", ResourceType: "principal", ResourceID: p.ID.String(),
		Outcome: domain.AuditOutcomeSuccess, IP: in.AuthCtx.IP, SessionID: &res.Session.ID,
	})
	return res, nil
}
