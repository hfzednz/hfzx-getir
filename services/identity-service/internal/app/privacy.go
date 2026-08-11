package app

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/nexora/identity-service/internal/app/ports"
	"github.com/nexora/identity-service/internal/domain"
)

// PrincipalDataExport is a privacy export payload.
type PrincipalDataExport struct {
	Principal   domain.Principal
	Identifiers []domain.Identifier
	Devices     []domain.Device
	Sessions    []domain.Session
	Consents    []domain.Consent
	Roles       []domain.PrincipalRole
	ExportedAt  time.Time
}

// ExportPrincipalData assembles a subject-access / data-portability export.
func (d *Deps) ExportPrincipalData(ctx context.Context, principalID uuid.UUID) (PrincipalDataExport, error) {
	if principalID == uuid.Nil {
		return PrincipalDataExport{}, fmt.Errorf("%w: principal_id required", domain.ErrInvalidArgument)
	}
	p, err := d.Principals.GetByID(ctx, principalID)
	if err != nil {
		return PrincipalDataExport{}, err
	}
	idents, _ := d.Principals.ListIdentifiers(ctx, principalID)
	devices, _ := d.Devices.ListByPrincipal(ctx, principalID)
	sessions, _ := d.Sessions.ListByPrincipal(ctx, principalID)
	consents, _ := d.Principals.ListConsents(ctx, principalID)
	roles, _ := d.Roles.ListPrincipalRoles(ctx, principalID)
	exp := PrincipalDataExport{
		Principal: p, Identifiers: idents, Devices: devices, Sessions: sessions,
		Consents: consents, Roles: roles, ExportedAt: d.now(),
	}
	d.appendAudit(ctx, domain.AuditEvent{
		TenantID: &p.TenantID, ActorID: &p.ID, ActorKind: string(p.Kind),
		Action: "privacy.export", ResourceType: "principal", ResourceID: p.ID.String(),
		Outcome: domain.AuditOutcomeSuccess,
	})
	return exp, nil
}

// RequestDeletion marks a principal for deletion (soft-delete / status deleted).
func (d *Deps) RequestDeletion(ctx context.Context, principalID uuid.UUID, authCtx ports.AuthContext) error {
	if principalID == uuid.Nil {
		return fmt.Errorf("%w: principal_id required", domain.ErrInvalidArgument)
	}
	p, err := d.Principals.GetByID(ctx, principalID)
	if err != nil {
		return err
	}
	now := d.now()
	p.Status = domain.PrincipalStatusDeleted
	p.DeletedAt = &now
	p.UpdatedAt = now
	if err := d.Principals.Update(ctx, p); err != nil {
		return err
	}
	// Revoke all sessions.
	sessions, _ := d.Sessions.ListByPrincipal(ctx, principalID)
	for _, s := range sessions {
		_ = d.Sessions.Revoke(ctx, s.ID, now, "principal_deleted")
	}
	d.publishPrincipal(ctx, "deleted", p)
	d.appendAudit(ctx, domain.AuditEvent{
		TenantID: &p.TenantID, ActorID: &p.ID, ActorKind: string(p.Kind),
		Action: "privacy.delete_request", ResourceType: "principal", ResourceID: p.ID.String(),
		Outcome: domain.AuditOutcomeSuccess, IP: authCtx.IP, RequestID: authCtx.RequestID,
	})
	return nil
}
