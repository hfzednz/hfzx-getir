package app

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/nexora/settlement-service/internal/domain"
)

// ReconcileProviderReportInput compares provider reported total vs batch total.
type ReconcileProviderReportInput struct {
	TenantID      uuid.UUID
	BatchID       uuid.UUID
	ProviderRef   string
	ReportedMinor int64
	ActorID       uuid.UUID
}

// ReconcileResult is the outcome of reconciliation.
type ReconcileResult struct {
	Reconciliation domain.Reconciliation
	Mismatch       *domain.Mismatch
	Matched        bool
}

// ReconcileProviderReport detects mismatches between expected and reported amounts.
func (d *Deps) ReconcileProviderReport(ctx context.Context, in ReconcileProviderReportInput) (ReconcileResult, error) {
	if in.TenantID == uuid.Nil || in.BatchID == uuid.Nil {
		return ReconcileResult{}, fmt.Errorf("%w: tenant_id and batch_id required", domain.ErrInvalidArgument)
	}
	if in.ReportedMinor < 0 {
		return ReconcileResult{}, fmt.Errorf("%w: reported", domain.ErrNegativeMoney)
	}
	b, err := d.Batches.GetByID(ctx, in.TenantID, in.BatchID)
	if err != nil {
		return ReconcileResult{}, err
	}
	now := d.now()
	matched := b.TotalMinor == in.ReportedMinor
	rec := domain.Reconciliation{
		ID:            d.newID(),
		TenantID:      in.TenantID,
		BatchID:       b.ID,
		ProviderRef:   in.ProviderRef,
		ExpectedMinor: b.TotalMinor,
		ReportedMinor: in.ReportedMinor,
		Matched:       matched,
		CreatedAt:     now,
	}
	if err := d.Batches.SaveReconciliation(ctx, rec); err != nil {
		return ReconcileResult{}, err
	}
	result := ReconcileResult{Reconciliation: rec, Matched: matched}
	actor := in.ActorID
	if !matched {
		mm := domain.Mismatch{
			ID:            d.newID(),
			TenantID:      in.TenantID,
			BatchID:       b.ID,
			ReconcileID:   rec.ID,
			ExpectedMinor: b.TotalMinor,
			ReportedMinor: in.ReportedMinor,
			DeltaMinor:    in.ReportedMinor - b.TotalMinor,
			Detail:        "provider report amount mismatch",
			CreatedAt:     now,
		}
		if err := d.Batches.SaveMismatch(ctx, mm); err != nil {
			return ReconcileResult{}, err
		}
		result.Mismatch = &mm
	}
	_ = d.appendEvent(ctx, b.ID, b.TenantID, domain.EventSettlementReconciled, map[string]any{
		"matched": matched, "expectedMinor": b.TotalMinor, "reportedMinor": in.ReportedMinor,
	}, &actor)
	return result, nil
}
