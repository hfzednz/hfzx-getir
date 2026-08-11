package app

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/nexora/loyalty-service/internal/domain"
)

// AdminManualGrantInput is an audited points grant.
type AdminManualGrantInput struct {
	TenantID    uuid.UUID
	PrincipalID uuid.UUID
	Points      int64
	Reason      string
	ActorID     *uuid.UUID
}

// AdminManualGrant grants points with audit trail.
func (d *Deps) AdminManualGrant(ctx context.Context, in AdminManualGrantInput) (domain.Account, domain.PointLedgerEntry, error) {
	if in.Points <= 0 {
		return domain.Account{}, domain.PointLedgerEntry{}, fmt.Errorf("%w: points must be > 0", domain.ErrInvalidArgument)
	}
	if in.Reason == "" {
		return domain.Account{}, domain.PointLedgerEntry{}, fmt.Errorf("%w: reason required", domain.ErrInvalidArgument)
	}
	acct, err := d.EnsureAccount(ctx, EnsureAccountInput{TenantID: in.TenantID, PrincipalID: in.PrincipalID})
	if err != nil {
		return domain.Account{}, domain.PointLedgerEntry{}, err
	}
	idem := "admin-grant:" + d.newID().String()
	acct, entry, err := d.grantPoints(ctx, acct, in.Points, domain.PointAdjust, in.Reason, idem, map[string]any{
		"reason": in.Reason,
	})
	if err != nil {
		return domain.Account{}, domain.PointLedgerEntry{}, err
	}
	_ = d.Accounts.CreateAudit(ctx, domain.AuditEntry{
		ID: d.newID(), TenantID: in.TenantID, AccountID: acct.ID,
		Action: "manual_grant", ActorID: in.ActorID,
		Detail: map[string]any{"points": in.Points, "reason": in.Reason, "entryId": entry.ID.String()},
		CreatedAt: d.now(),
	})
	_, _, _ = d.EvaluateMembership(ctx, EvaluateMembershipInput{TenantID: in.TenantID, AccountID: acct.ID})
	return acct, entry, nil
}
