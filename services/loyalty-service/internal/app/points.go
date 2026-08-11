package app

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/nexora/loyalty-service/internal/domain"
)

// EarnPointsInput awards points from an order (idempotent by order_id).
type EarnPointsInput struct {
	TenantID    uuid.UUID
	PrincipalID uuid.UUID
	OrderID     uuid.UUID
	Points      int64
	SpendMinor  int64
	Reference   string
}

// EarnPoints credits points for a completed order; same order_id is idempotent.
func (d *Deps) EarnPoints(ctx context.Context, in EarnPointsInput) (domain.Account, domain.PointLedgerEntry, error) {
	if in.TenantID == uuid.Nil || in.PrincipalID == uuid.Nil || in.OrderID == uuid.Nil {
		return domain.Account{}, domain.PointLedgerEntry{}, fmt.Errorf("%w: tenant, principal, order required", domain.ErrInvalidArgument)
	}
	if in.Points <= 0 {
		return domain.Account{}, domain.PointLedgerEntry{}, fmt.Errorf("%w: points must be > 0", domain.ErrInvalidArgument)
	}

	acct, err := d.EnsureAccount(ctx, EnsureAccountInput{TenantID: in.TenantID, PrincipalID: in.PrincipalID})
	if err != nil {
		return domain.Account{}, domain.PointLedgerEntry{}, err
	}

	idem := "earn:" + in.OrderID.String()
	if existing, err := d.Accounts.GetLedgerByIdempotency(ctx, in.TenantID, idem); err == nil {
		return acct, existing, nil
	}
	if existing, err := d.Accounts.GetLedgerByOrder(ctx, in.TenantID, acct.ID, in.OrderID); err == nil {
		return acct, existing, nil
	}

	now := d.now()
	acct.Points += in.Points
	acct.TierPoints += in.Points
	acct.XP += in.Points
	acct.UpdatedAt = now
	acct.Version++
	if err := acct.Validate(); err != nil {
		return domain.Account{}, domain.PointLedgerEntry{}, err
	}
	if err := d.Accounts.UpdateAccount(ctx, acct); err != nil {
		return domain.Account{}, domain.PointLedgerEntry{}, err
	}

	oid := in.OrderID
	entry := domain.PointLedgerEntry{
		ID: d.newID(), TenantID: in.TenantID, AccountID: acct.ID,
		Kind: domain.PointEarn, Points: in.Points, BalanceAfter: acct.Points,
		OrderID: &oid, Reference: in.Reference, IdempotencyKey: idem,
		Metadata: map[string]any{"spendMinor": in.SpendMinor}, CreatedAt: now,
	}
	if err := d.Accounts.CreateLedgerEntry(ctx, entry); err != nil {
		return domain.Account{}, domain.PointLedgerEntry{}, err
	}

	_, _ = d.Accounts.IncrStat(ctx, in.TenantID, acct.ID, "purchase_count", 1)
	if in.SpendMinor > 0 {
		_, _ = d.Accounts.IncrStat(ctx, in.TenantID, acct.ID, "spend_minor", in.SpendMinor)
	}

	d.emit(ctx, acct, domain.EventPointsEarned, map[string]any{
		"points": in.Points, "orderId": in.OrderID.String(), "entryId": entry.ID.String(),
	})

	_, _, _ = d.EvaluateMembership(ctx, EvaluateMembershipInput{TenantID: in.TenantID, AccountID: acct.ID})
	_ = d.evaluateAchievements(ctx, acct)

	fresh, _ := d.Accounts.GetAccount(ctx, in.TenantID, acct.ID)
	return fresh, entry, nil
}

// RedeemPointsInput spends points (never negative balance).
type RedeemPointsInput struct {
	TenantID       uuid.UUID
	AccountID      uuid.UUID
	Points         int64
	IdempotencyKey string
	Reference      string
}

// RedeemPoints deducts points; fails if balance insufficient.
func (d *Deps) RedeemPoints(ctx context.Context, in RedeemPointsInput) (domain.Account, domain.PointLedgerEntry, error) {
	if in.TenantID == uuid.Nil || in.AccountID == uuid.Nil {
		return domain.Account{}, domain.PointLedgerEntry{}, fmt.Errorf("%w: tenant and account required", domain.ErrInvalidArgument)
	}
	if in.Points <= 0 {
		return domain.Account{}, domain.PointLedgerEntry{}, fmt.Errorf("%w: points must be > 0", domain.ErrInvalidArgument)
	}
	if in.IdempotencyKey == "" {
		return domain.Account{}, domain.PointLedgerEntry{}, fmt.Errorf("%w: idempotency_key required", domain.ErrInvalidArgument)
	}
	if existing, err := d.Accounts.GetLedgerByIdempotency(ctx, in.TenantID, in.IdempotencyKey); err == nil {
		acct, _ := d.Accounts.GetAccount(ctx, in.TenantID, in.AccountID)
		return acct, existing, nil
	}

	acct, err := d.Accounts.GetAccount(ctx, in.TenantID, in.AccountID)
	if err != nil {
		return domain.Account{}, domain.PointLedgerEntry{}, err
	}
	if acct.Points < in.Points {
		return domain.Account{}, domain.PointLedgerEntry{}, fmt.Errorf("%w: balance=%d need=%d", domain.ErrInsufficientPoints, acct.Points, in.Points)
	}

	now := d.now()
	acct.Points -= in.Points
	acct.UpdatedAt = now
	acct.Version++
	if err := acct.Validate(); err != nil {
		return domain.Account{}, domain.PointLedgerEntry{}, err
	}
	if err := d.Accounts.UpdateAccount(ctx, acct); err != nil {
		return domain.Account{}, domain.PointLedgerEntry{}, err
	}

	entry := domain.PointLedgerEntry{
		ID: d.newID(), TenantID: in.TenantID, AccountID: acct.ID,
		Kind: domain.PointRedeem, Points: in.Points, BalanceAfter: acct.Points,
		Reference: in.Reference, IdempotencyKey: in.IdempotencyKey, CreatedAt: now,
	}
	if err := d.Accounts.CreateLedgerEntry(ctx, entry); err != nil {
		return domain.Account{}, domain.PointLedgerEntry{}, err
	}
	d.emit(ctx, acct, domain.EventPointsRedeemed, map[string]any{
		"points": in.Points, "entryId": entry.ID.String(),
	})
	return acct, entry, nil
}

func (d *Deps) grantPoints(ctx context.Context, acct domain.Account, points int64, kind domain.PointEntryKind, ref, idem string, meta map[string]any) (domain.Account, domain.PointLedgerEntry, error) {
	if points <= 0 {
		return acct, domain.PointLedgerEntry{}, fmt.Errorf("%w: points must be > 0", domain.ErrInvalidArgument)
	}
	if existing, err := d.Accounts.GetLedgerByIdempotency(ctx, acct.TenantID, idem); err == nil {
		return acct, existing, nil
	}
	now := d.now()
	acct.Points += points
	if kind == domain.PointEarn || kind == domain.PointGrant {
		acct.TierPoints += points
		acct.XP += points
	}
	acct.UpdatedAt = now
	acct.Version++
	if err := d.Accounts.UpdateAccount(ctx, acct); err != nil {
		return acct, domain.PointLedgerEntry{}, err
	}
	entry := domain.PointLedgerEntry{
		ID: d.newID(), TenantID: acct.TenantID, AccountID: acct.ID,
		Kind: kind, Points: points, BalanceAfter: acct.Points,
		Reference: ref, IdempotencyKey: idem, Metadata: meta, CreatedAt: now,
	}
	if err := d.Accounts.CreateLedgerEntry(ctx, entry); err != nil {
		return acct, domain.PointLedgerEntry{}, err
	}
	d.emit(ctx, acct, domain.EventPointsEarned, map[string]any{"points": points, "entryId": entry.ID.String(), "kind": string(kind)})
	return acct, entry, nil
}
