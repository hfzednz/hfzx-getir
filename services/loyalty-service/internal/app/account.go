package app

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/nexora/loyalty-service/internal/domain"
)

// EnsureAccountInput creates or returns a loyalty account.
type EnsureAccountInput struct {
	TenantID    uuid.UUID
	PrincipalID uuid.UUID
}

// EnsureAccount gets or creates a loyalty account and default membership.
func (d *Deps) EnsureAccount(ctx context.Context, in EnsureAccountInput) (domain.Account, error) {
	if in.TenantID == uuid.Nil || in.PrincipalID == uuid.Nil {
		return domain.Account{}, fmt.Errorf("%w: tenant_id and principal_id required", domain.ErrInvalidArgument)
	}
	if existing, err := d.Accounts.GetAccountByPrincipal(ctx, in.TenantID, in.PrincipalID); err == nil {
		return existing, nil
	}

	now := d.now()
	acct := domain.Account{
		ID: d.newID(), TenantID: in.TenantID, PrincipalID: in.PrincipalID,
		Points: 0, TierPoints: 0, XP: 0, Level: 1, Active: true, Version: 1,
		CreatedAt: now, UpdatedAt: now,
	}
	if err := acct.Validate(); err != nil {
		return domain.Account{}, err
	}
	if err := d.Accounts.CreateAccount(ctx, acct); err != nil {
		if existing, e2 := d.Accounts.GetAccountByPrincipal(ctx, in.TenantID, in.PrincipalID); e2 == nil {
			return existing, nil
		}
		return domain.Account{}, err
	}
	_ = d.Memberships.UpsertMembership(ctx, domain.Membership{
		ID: d.newID(), TenantID: in.TenantID, AccountID: acct.ID,
		Tier: domain.TierStandard, Since: now, UpdatedAt: now,
	})
	return acct, nil
}

// GetAccount returns an account by id.
func (d *Deps) GetAccount(ctx context.Context, tenantID, accountID uuid.UUID) (domain.Account, error) {
	return d.Accounts.GetAccount(ctx, tenantID, accountID)
}

// GetAccountByPrincipal returns an account by principal.
func (d *Deps) GetAccountByPrincipal(ctx context.Context, tenantID, principalID uuid.UUID) (domain.Account, error) {
	return d.Accounts.GetAccountByPrincipal(ctx, tenantID, principalID)
}

// PointsHistory returns ledger history.
func (d *Deps) PointsHistory(ctx context.Context, tenantID, accountID uuid.UUID, limit, offset int) ([]domain.PointLedgerEntry, int, error) {
	if limit <= 0 {
		limit = 50
	}
	return d.Accounts.ListLedger(ctx, tenantID, accountID, limit, offset)
}
