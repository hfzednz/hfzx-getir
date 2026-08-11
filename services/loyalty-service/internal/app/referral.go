package app

import (
	"context"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/nexora/loyalty-service/internal/domain"
)

const defaultReferralPoints int64 = 500

// CreateReferralCodeInput creates a referral code for a principal.
type CreateReferralCodeInput struct {
	TenantID    uuid.UUID
	PrincipalID uuid.UUID
	Code        string
}

// CreateReferralCode ensures account and creates/returns referral code.
func (d *Deps) CreateReferralCode(ctx context.Context, in CreateReferralCodeInput) (domain.ReferralCode, error) {
	acct, err := d.EnsureAccount(ctx, EnsureAccountInput{TenantID: in.TenantID, PrincipalID: in.PrincipalID})
	if err != nil {
		return domain.ReferralCode{}, err
	}
	if existing, err := d.Referrals.GetCodeByAccount(ctx, in.TenantID, acct.ID); err == nil {
		return existing, nil
	}
	code := strings.ToUpper(strings.TrimSpace(in.Code))
	if code == "" {
		code = strings.ToUpper(strings.ReplaceAll(d.newID().String()[:8], "-", ""))
	}
	now := d.now()
	c := domain.ReferralCode{
		ID: d.newID(), TenantID: in.TenantID, AccountID: acct.ID, PrincipalID: in.PrincipalID,
		Code: code, Active: true, CreatedAt: now,
	}
	if err := d.Referrals.CreateCode(ctx, c); err != nil {
		return domain.ReferralCode{}, err
	}
	return c, nil
}

// ApplyReferralInput applies a code for a referee (blocks self).
type ApplyReferralInput struct {
	TenantID         uuid.UUID
	RefereePrincipal uuid.UUID
	Code             string
}

// ApplyReferral links referee to referrer; rejects self-referral.
func (d *Deps) ApplyReferral(ctx context.Context, in ApplyReferralInput) (domain.ReferralEvent, error) {
	code, err := d.Referrals.GetCode(ctx, in.TenantID, strings.ToUpper(strings.TrimSpace(in.Code)))
	if err != nil {
		return domain.ReferralEvent{}, fmt.Errorf("%w: %v", domain.ErrReferralInvalid, err)
	}
	if !code.Active {
		return domain.ReferralEvent{}, domain.ErrReferralInvalid
	}
	if code.PrincipalID == in.RefereePrincipal {
		return domain.ReferralEvent{}, domain.ErrSelfReferral
	}

	referee, err := d.EnsureAccount(ctx, EnsureAccountInput{TenantID: in.TenantID, PrincipalID: in.RefereePrincipal})
	if err != nil {
		return domain.ReferralEvent{}, err
	}
	if code.AccountID == referee.ID {
		return domain.ReferralEvent{}, domain.ErrSelfReferral
	}
	if existing, err := d.Referrals.GetEventByReferee(ctx, in.TenantID, referee.ID); err == nil {
		return existing, nil
	}

	now := d.now()
	ev := domain.ReferralEvent{
		ID: d.newID(), TenantID: in.TenantID, CodeID: code.ID,
		ReferrerAccount: code.AccountID, RefereeAccount: referee.ID, RefereePrincipal: in.RefereePrincipal,
		Status: domain.ReferralApplied, CreatedAt: now, UpdatedAt: now,
	}
	if err := ev.Validate(); err != nil {
		return domain.ReferralEvent{}, err
	}
	if err := d.Referrals.CreateEvent(ctx, ev); err != nil {
		return domain.ReferralEvent{}, err
	}
	return ev, nil
}

// CompleteReferralInput completes referral on first order and grants points.
type CompleteReferralInput struct {
	TenantID         uuid.UUID
	RefereePrincipal uuid.UUID
	OrderID          uuid.UUID
	Points           int64
}

// CompleteReferral grants points to referrer on first order hook.
func (d *Deps) CompleteReferral(ctx context.Context, in CompleteReferralInput) (domain.ReferralEvent, error) {
	referee, err := d.Accounts.GetAccountByPrincipal(ctx, in.TenantID, in.RefereePrincipal)
	if err != nil {
		return domain.ReferralEvent{}, err
	}
	ev, err := d.Referrals.GetEventByReferee(ctx, in.TenantID, referee.ID)
	if err != nil {
		return domain.ReferralEvent{}, err
	}
	if ev.Status == domain.ReferralCompleted {
		return ev, nil
	}
	if ev.Status != domain.ReferralApplied {
		return domain.ReferralEvent{}, fmt.Errorf("%w: referral status %s", domain.ErrConflict, ev.Status)
	}

	points := in.Points
	if points <= 0 {
		points = defaultReferralPoints
	}
	referrer, err := d.Accounts.GetAccount(ctx, in.TenantID, ev.ReferrerAccount)
	if err != nil {
		return domain.ReferralEvent{}, err
	}
	oid := in.OrderID
	_, _, err = d.grantPoints(ctx, referrer, points, domain.PointGrant,
		"referral:"+ev.ID.String(), "referral-complete:"+ev.ID.String(),
		map[string]any{"orderId": oid.String()})
	if err != nil {
		return domain.ReferralEvent{}, err
	}

	now := d.now()
	ev.Status = domain.ReferralCompleted
	ev.OrderID = &oid
	ev.PointsGranted = points
	ev.UpdatedAt = now
	if err := d.Referrals.UpdateEvent(ctx, ev); err != nil {
		return domain.ReferralEvent{}, err
	}
	_, _ = d.Accounts.IncrStat(ctx, in.TenantID, referrer.ID, "referral_count", 1)
	_ = d.evaluateAchievements(ctx, referrer)
	d.emit(ctx, referrer, domain.EventReferralCompleted, map[string]any{
		"eventId": ev.ID.String(), "points": points, "refereeId": referee.ID.String(),
	})
	return ev, nil
}
