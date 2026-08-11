package app

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/nexora/promotion-service/internal/domain"
)

// IssueVoucherInput issues a personal voucher.
type IssueVoucherInput struct {
	TenantID     uuid.UUID
	PrincipalID  uuid.UUID
	PromotionID  *uuid.UUID
	Code         string
	ValueMinor   int64
	Currency     string
	StartsAt     *time.Time
	EndsAt       *time.Time
}

// IssueVoucher creates an issued voucher for a principal.
func (d *Deps) IssueVoucher(ctx context.Context, in IssueVoucherInput) (domain.Voucher, error) {
	if in.TenantID == uuid.Nil || in.PrincipalID == uuid.Nil {
		return domain.Voucher{}, fmt.Errorf("%w: tenant_id and principal_id required", domain.ErrInvalidArgument)
	}
	if in.ValueMinor <= 0 {
		return domain.Voucher{}, fmt.Errorf("%w: value_minor must be > 0", domain.ErrInvalidArgument)
	}
	currency := strings.ToUpper(strings.TrimSpace(in.Currency))
	if len(currency) != 3 {
		return domain.Voucher{}, fmt.Errorf("%w: currency required", domain.ErrInvalidArgument)
	}
	now := d.now()
	code := strings.ToUpper(strings.TrimSpace(in.Code))
	if code == "" {
		code = "V" + strings.ToUpper(strings.ReplaceAll(d.newID().String(), "-", "")[:11])
	}
	v := domain.Voucher{
		ID:             d.newID(),
		TenantID:       in.TenantID,
		PromotionID:    in.PromotionID,
		Code:           code,
		PrincipalID:    in.PrincipalID,
		Status:         domain.VoucherIssued,
		ValueMinor:     in.ValueMinor,
		Currency:       currency,
		RemainingMinor: in.ValueMinor,
		StartsAt:       in.StartsAt,
		EndsAt:         in.EndsAt,
		CreatedAt:      now,
		UpdatedAt:      now,
	}
	if err := v.Validate(); err != nil {
		return domain.Voucher{}, err
	}
	if err := d.Vouchers.Create(ctx, v); err != nil {
		return domain.Voucher{}, err
	}
	d.emit(ctx, v.TenantID, v.ID, domain.EventVoucherIssued, map[string]any{
		"code": v.Code, "valueMinor": v.ValueMinor, "principalId": v.PrincipalID.String(),
	})
	return v, nil
}

// RedeemVoucherInput redeems voucher value idempotently.
type RedeemVoucherInput struct {
	TenantID       uuid.UUID
	Code           string
	PrincipalID    uuid.UUID
	IdempotencyKey string
	OrderRef       string
	AmountMinor    int64
	Currency       string
}

// RedeemVoucher spends voucher remaining balance.
func (d *Deps) RedeemVoucher(ctx context.Context, in RedeemVoucherInput) (domain.VoucherRedemption, error) {
	if in.TenantID == uuid.Nil || in.Code == "" || in.IdempotencyKey == "" {
		return domain.VoucherRedemption{}, fmt.Errorf("%w: tenant_id, code, and idempotency_key required", domain.ErrInvalidArgument)
	}
	if existing, err := d.Vouchers.GetRedemptionByIdempotency(ctx, in.TenantID, in.IdempotencyKey); err == nil {
		return existing, nil
	}
	v, err := d.Vouchers.GetByCode(ctx, in.TenantID, in.Code)
	if err != nil {
		return domain.VoucherRedemption{}, fmt.Errorf("%w: %v", domain.ErrVoucherInvalid, err)
	}
	now := d.now()
	if !v.IsRedeemableAt(now) {
		return domain.VoucherRedemption{}, domain.ErrVoucherInvalid
	}
	if v.PrincipalID != in.PrincipalID {
		return domain.VoucherRedemption{}, domain.ErrForbidden
	}
	amount := in.AmountMinor
	if amount <= 0 {
		amount = v.RemainingMinor
	}
	if amount > v.RemainingMinor {
		return domain.VoucherRedemption{}, domain.ErrVoucherExhausted
	}
	currency := strings.ToUpper(strings.TrimSpace(in.Currency))
	if currency == "" {
		currency = v.Currency
	}
	if currency != v.Currency {
		return domain.VoucherRedemption{}, domain.ErrCurrencyMismatch
	}
	red := domain.VoucherRedemption{
		ID:             d.newID(),
		TenantID:       in.TenantID,
		VoucherID:      v.ID,
		PrincipalID:    in.PrincipalID,
		IdempotencyKey: in.IdempotencyKey,
		OrderRef:       in.OrderRef,
		AmountMinor:    amount,
		Currency:       currency,
		RedeemedAt:     now,
		CreatedAt:      now,
	}
	if err := d.Vouchers.CreateRedemption(ctx, red); err != nil {
		if existing, e2 := d.Vouchers.GetRedemptionByIdempotency(ctx, in.TenantID, in.IdempotencyKey); e2 == nil {
			return existing, nil
		}
		return domain.VoucherRedemption{}, err
	}
	v.RemainingMinor -= amount
	if v.RemainingMinor == 0 {
		v.Status = domain.VoucherRedeemed
	}
	v.UpdatedAt = now
	if err := d.Vouchers.Update(ctx, v); err != nil {
		return domain.VoucherRedemption{}, err
	}
	d.emit(ctx, v.TenantID, v.ID, domain.EventVoucherRedeemed, map[string]any{
		"code": v.Code, "amountMinor": amount, "orderRef": in.OrderRef,
	})
	return red, nil
}
