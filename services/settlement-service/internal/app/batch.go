package app

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/nexora/settlement-service/internal/app/ports"
	"github.com/nexora/settlement-service/internal/domain"
)

// CreateBatchInput creates a draft settlement batch.
type CreateBatchInput struct {
	TenantID       uuid.UUID
	Currency       string
	PeriodStart    time.Time
	PeriodEnd      time.Time
	Description    string
	IdempotencyKey string
	ActorID        uuid.UUID
}

// CreateBatch creates a draft batch.
func (d *Deps) CreateBatch(ctx context.Context, in CreateBatchInput) (domain.SettlementBatch, error) {
	if in.TenantID == uuid.Nil {
		return domain.SettlementBatch{}, fmt.Errorf("%w: tenant_id", domain.ErrInvalidArgument)
	}
	currency := strings.ToUpper(strings.TrimSpace(in.Currency))
	if _, err := domain.NewMoney(0, currency); err != nil {
		return domain.SettlementBatch{}, err
	}
	if in.IdempotencyKey != "" {
		if existing, err := d.Batches.GetByIdempotencyKey(ctx, in.TenantID, in.IdempotencyKey); err == nil {
			return existing, nil
		} else if !errors.Is(err, domain.ErrNotFound) {
			return domain.SettlementBatch{}, err
		}
	}
	now := d.now()
	b := domain.SettlementBatch{
		ID:             d.newID(),
		TenantID:       in.TenantID,
		Status:         domain.BatchStatusDraft,
		Currency:       currency,
		PeriodStart:    in.PeriodStart.UTC(),
		PeriodEnd:      in.PeriodEnd.UTC(),
		Description:    in.Description,
		IdempotencyKey: in.IdempotencyKey,
		Lines:          []domain.SettlementLine{},
		CreatedAt:      now,
		UpdatedAt:      now,
		Version:        1,
	}
	if err := b.Validate(); err != nil {
		return domain.SettlementBatch{}, err
	}
	if err := d.Batches.Create(ctx, b); err != nil {
		return domain.SettlementBatch{}, err
	}
	actor := in.ActorID
	_ = d.appendEvent(ctx, b.ID, b.TenantID, domain.EventSettlementCreated, map[string]any{
		"currency": b.Currency,
	}, &actor)
	return b, nil
}

// AddLineInput adds a line to a draft batch.
type AddLineInput struct {
	TenantID    uuid.UUID
	BatchID     uuid.UUID
	PayeeType   domain.PayeeType
	PayeeRef    string
	AmountMinor int64
	ExternalRef string
	Memo        string
}

// AddLine appends a payable line (draft only).
func (d *Deps) AddLine(ctx context.Context, in AddLineInput) (domain.SettlementBatch, error) {
	b, err := d.Batches.GetByID(ctx, in.TenantID, in.BatchID)
	if err != nil {
		return domain.SettlementBatch{}, err
	}
	if b.Status != domain.BatchStatusDraft {
		return domain.SettlementBatch{}, fmt.Errorf("%w: can only add lines in draft", domain.ErrInvalidTransition)
	}
	line := domain.SettlementLine{
		ID:          d.newID(),
		PayeeType:   in.PayeeType,
		PayeeRef:    strings.TrimSpace(in.PayeeRef),
		AmountMinor: in.AmountMinor,
		Currency:    b.Currency,
		ExternalRef: in.ExternalRef,
		Memo:        in.Memo,
	}
	if err := line.Validate(); err != nil {
		return domain.SettlementBatch{}, err
	}
	b.Lines = append(b.Lines, line)
	b.RecalcTotal()
	b.UpdatedAt = d.now()
	b.Version++
	if err := d.Batches.Update(ctx, b); err != nil {
		return domain.SettlementBatch{}, err
	}
	return b, nil
}

// SubmitInput submits a batch for dual-control approval.
type SubmitInput struct {
	TenantID uuid.UUID
	BatchID  uuid.UUID
	ActorID  uuid.UUID
}

// Submit moves draft → pending_approval.
func (d *Deps) Submit(ctx context.Context, in SubmitInput) (domain.SettlementBatch, error) {
	if in.ActorID == uuid.Nil {
		return domain.SettlementBatch{}, fmt.Errorf("%w: actor_id required", domain.ErrInvalidArgument)
	}
	b, err := d.Batches.GetByID(ctx, in.TenantID, in.BatchID)
	if err != nil {
		return domain.SettlementBatch{}, err
	}
	if len(b.Lines) == 0 {
		return domain.SettlementBatch{}, domain.ErrBatchNotEmpty
	}
	if err := d.transition(&b, domain.BatchStatusPendingApproval); err != nil {
		return domain.SettlementBatch{}, err
	}
	now := d.now()
	actor := in.ActorID
	b.SubmittedBy = &actor
	b.SubmittedAt = &now
	if err := d.Batches.Update(ctx, b); err != nil {
		return domain.SettlementBatch{}, err
	}
	_ = d.appendEvent(ctx, b.ID, b.TenantID, domain.EventSettlementSubmitted, map[string]any{
		"totalMinor": b.TotalMinor,
	}, &actor)
	return b, nil
}

// ApproveInput dual-control approves a pending batch.
type ApproveInput struct {
	TenantID uuid.UUID
	BatchID  uuid.UUID
	ActorID  uuid.UUID
}

// Approve moves pending_approval → approved. Approver must differ from submitter.
func (d *Deps) Approve(ctx context.Context, in ApproveInput) (domain.SettlementBatch, error) {
	if in.ActorID == uuid.Nil {
		return domain.SettlementBatch{}, fmt.Errorf("%w: actor_id required", domain.ErrInvalidArgument)
	}
	b, err := d.Batches.GetByID(ctx, in.TenantID, in.BatchID)
	if err != nil {
		return domain.SettlementBatch{}, err
	}
	if b.SubmittedBy != nil && *b.SubmittedBy == in.ActorID {
		return domain.SettlementBatch{}, domain.ErrDualControl
	}
	if err := d.transition(&b, domain.BatchStatusApproved); err != nil {
		return domain.SettlementBatch{}, err
	}
	now := d.now()
	actor := in.ActorID
	b.ApprovedBy = &actor
	b.ApprovedAt = &now
	if err := d.Batches.Update(ctx, b); err != nil {
		return domain.SettlementBatch{}, err
	}
	_ = d.appendEvent(ctx, b.ID, b.TenantID, domain.EventSettlementApproved, map[string]any{
		"totalMinor": b.TotalMinor,
	}, &actor)
	return b, nil
}

// ExecutePayoutsInput executes payouts for an approved batch.
type ExecutePayoutsInput struct {
	TenantID uuid.UUID
	BatchID  uuid.UUID
	ActorID  uuid.UUID
}

// ExecutePayouts approved → paying → completed|failed; posts ledger journals and payouts.
func (d *Deps) ExecutePayouts(ctx context.Context, in ExecutePayoutsInput) (domain.SettlementBatch, error) {
	b, err := d.Batches.GetByID(ctx, in.TenantID, in.BatchID)
	if err != nil {
		return domain.SettlementBatch{}, err
	}
	if err := d.transition(&b, domain.BatchStatusPaying); err != nil {
		return domain.SettlementBatch{}, err
	}
	if err := d.Batches.Update(ctx, b); err != nil {
		return domain.SettlementBatch{}, err
	}

	actor := in.ActorID
	for _, line := range b.Lines {
		payout := domain.PayoutInstruction{
			ID: d.newID(), BatchID: b.ID, LineID: line.ID, TenantID: b.TenantID,
			PayeeType: line.PayeeType, PayeeRef: line.PayeeRef,
			AmountMinor: line.AmountMinor, Currency: line.Currency,
			Status: "pending", CreatedAt: d.now(), UpdatedAt: d.now(),
		}
		if d.Ledger != nil {
			_, err := d.Ledger.PostSettlementJournal(ctx, ports.LedgerPostRequest{
				TenantID: b.TenantID, Currency: b.Currency,
				Reference: "settle:" + b.ID.String() + ":" + line.ID.String(),
				IdempotencyKey: "settle-ledger-" + line.ID.String(),
				DebitAccount: "SETTLEMENT_EXPENSE", CreditAccount: "SETTLEMENT_CLEARING",
				AmountMinor: line.AmountMinor,
			})
			if err != nil {
				b.FailureReason = err.Error()
				_ = d.transition(&b, domain.BatchStatusFailed)
				_ = d.Batches.Update(ctx, b)
				_ = d.appendEvent(ctx, b.ID, b.TenantID, domain.EventSettlementFailed, map[string]any{
					"reason": b.FailureReason,
				}, &actor)
				return b, err
			}
		}
		if d.Payout != nil {
			res, err := d.Payout.Execute(ctx, ports.PayoutRequest{
				InstructionID: payout.ID, TenantID: b.TenantID,
				PayeeType: line.PayeeType, PayeeRef: line.PayeeRef,
				AmountMinor: line.AmountMinor, Currency: line.Currency,
			})
			if err != nil || !res.Succeeded {
				msg := "payout failed"
				if err != nil {
					msg = err.Error()
				} else if res.Error != "" {
					msg = res.Error
				}
				payout.Status = "failed"
				payout.UpdatedAt = d.now()
				_ = d.Batches.SavePayout(ctx, payout)
				b.FailureReason = msg
				_ = d.transition(&b, domain.BatchStatusFailed)
				_ = d.Batches.Update(ctx, b)
				_ = d.appendEvent(ctx, b.ID, b.TenantID, domain.EventSettlementFailed, map[string]any{
					"reason": msg,
				}, &actor)
				return b, fmt.Errorf("%w: %s", domain.ErrInvariant, msg)
			}
			payout.Status = "succeeded"
			payout.ProviderRef = res.ProviderRef
		} else {
			payout.Status = "succeeded"
			payout.ProviderRef = "stub-" + payout.ID.String()
		}
		payout.UpdatedAt = d.now()
		_ = d.Batches.SavePayout(ctx, payout)
	}

	if err := d.transition(&b, domain.BatchStatusCompleted); err != nil {
		return domain.SettlementBatch{}, err
	}
	if err := d.Batches.Update(ctx, b); err != nil {
		return domain.SettlementBatch{}, err
	}
	_ = d.appendEvent(ctx, b.ID, b.TenantID, domain.EventSettlementCompleted, map[string]any{
		"totalMinor": b.TotalMinor,
	}, &actor)
	return b, nil
}

// ListBatchesInput filters batches.
type ListBatchesInput struct {
	TenantID uuid.UUID
	Status   *domain.BatchStatus
	Limit    int
	Offset   int
}

// List returns settlement batches.
func (d *Deps) List(ctx context.Context, in ListBatchesInput) ([]domain.SettlementBatch, int, error) {
	if in.TenantID == uuid.Nil {
		return nil, 0, fmt.Errorf("%w: tenant_id", domain.ErrInvalidArgument)
	}
	limit := in.Limit
	if limit <= 0 {
		limit = 50
	}
	return d.Batches.List(ctx, ports.BatchFilter{
		TenantID: in.TenantID, Status: in.Status, Limit: limit, Offset: in.Offset,
	})
}

// GetBatch returns a batch by id.
func (d *Deps) GetBatch(ctx context.Context, tenantID, id uuid.UUID) (domain.SettlementBatch, error) {
	return d.Batches.GetByID(ctx, tenantID, id)
}
