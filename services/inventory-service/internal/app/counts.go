package app

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/nexora/inventory-service/internal/domain"
)

// StartCountCmd starts an inventory count session.
type StartCountCmd struct {
	TenantID    uuid.UUID
	WarehouseID uuid.UUID
	LocationID  *uuid.UUID
	Type        domain.CountSessionType
	ActorID     *uuid.UUID
	Notes       string
	Lines       []StartCountLine
}

// StartCountLine seeds a count line with system qty.
type StartCountLine struct {
	VariantID  uuid.UUID
	SKUCode    string
	LocationID *uuid.UUID
	LotID      *uuid.UUID
	SystemQty  int64
}

// StartCount creates a session and transitions to in_progress.
func (d *Deps) StartCount(ctx context.Context, in StartCountCmd) (domain.CountSession, error) {
	if in.TenantID == uuid.Nil || in.WarehouseID == uuid.Nil {
		return domain.CountSession{}, domain.ErrInvalidArgument
	}
	typ := in.Type
	if typ == "" {
		typ = domain.CountSessionTypeCycle
	}
	now := d.now()
	s := domain.CountSession{
		ID: d.newID(), TenantID: in.TenantID, WarehouseID: in.WarehouseID,
		LocationID: in.LocationID, Type: typ, Status: domain.CountSessionStatusDraft,
		Notes: in.Notes, Metadata: map[string]any{}, CreatedAt: now, UpdatedAt: now,
	}
	for _, line := range in.Lines {
		sys := line.SystemQty
		if sys < 0 {
			return domain.CountSession{}, domain.ErrInvalidArgument
		}
		s.Lines = append(s.Lines, domain.CountLine{
			ID: d.newID(), SessionID: s.ID, VariantID: line.VariantID,
			SKUCode: line.SKUCode, LocationID: line.LocationID, LotID: line.LotID,
			SystemQty: sys, Metadata: map[string]any{}, CreatedAt: now, UpdatedAt: now,
		})
	}
	if err := s.Validate(); err != nil {
		return domain.CountSession{}, err
	}
	if err := d.Counts.Create(ctx, s); err != nil {
		return domain.CountSession{}, err
	}
	if err := s.TransitionTo(domain.CountSessionStatusInProgress, in.ActorID); err != nil {
		return domain.CountSession{}, err
	}
	if err := d.Counts.Update(ctx, s); err != nil {
		return domain.CountSession{}, err
	}
	return s, nil
}

// SubmitCountLine is a counted qty submission.
type SubmitCountLine struct {
	LineID     uuid.UUID
	CountedQty int64
}

// SubmitCountCmd submits counted lines.
type SubmitCountCmd struct {
	TenantID  uuid.UUID
	SessionID uuid.UUID
	ActorID   *uuid.UUID
	Lines     []SubmitCountLine
}

// SubmitCount records counted quantities and moves to submitted.
func (d *Deps) SubmitCount(ctx context.Context, in SubmitCountCmd) (domain.CountSession, error) {
	s, err := d.Counts.GetByID(ctx, in.TenantID, in.SessionID)
	if err != nil {
		return domain.CountSession{}, err
	}
	byID := map[uuid.UUID]*domain.CountLine{}
	for i := range s.Lines {
		byID[s.Lines[i].ID] = &s.Lines[i]
	}
	for _, line := range in.Lines {
		cl, ok := byID[line.LineID]
		if !ok {
			return domain.CountSession{}, fmt.Errorf("%w: count line %s", domain.ErrNotFound, line.LineID)
		}
		if err := cl.SetCounted(line.CountedQty); err != nil {
			return domain.CountSession{}, err
		}
	}
	if err := s.TransitionTo(domain.CountSessionStatusSubmitted, in.ActorID); err != nil {
		return domain.CountSession{}, err
	}
	if err := d.Counts.Update(ctx, s); err != nil {
		return domain.CountSession{}, err
	}
	return s, nil
}

// ApproveCountCmd approves variance and posts adjustments.
type ApproveCountCmd struct {
	TenantID       uuid.UUID
	SessionID      uuid.UUID
	ActorID        *uuid.UUID
	IdempotencyKey string
}

// ApproveCount applies variance adjustments to balances.
func (d *Deps) ApproveCount(ctx context.Context, in ApproveCountCmd) (domain.CountSession, error) {
	if in.IdempotencyKey == "" {
		return domain.CountSession{}, fmt.Errorf("%w: idempotency_key required", domain.ErrInvalidArgument)
	}
	if v, ok := d.idemGet(ctx, "count:"+in.IdempotencyKey); ok {
		if s, ok := v.(domain.CountSession); ok {
			return s, nil
		}
	}
	s, err := d.Counts.GetByID(ctx, in.TenantID, in.SessionID)
	if err != nil {
		return domain.CountSession{}, err
	}
	if s.Status == domain.CountSessionStatusApproved {
		d.idemPut(ctx, "count:"+in.IdempotencyKey, s)
		return s, nil
	}
	if s.Status == domain.CountSessionStatusSubmitted {
		_ = s.TransitionTo(domain.CountSessionStatusPendingApproval, in.ActorID)
	}
	for i := range s.Lines {
		line := &s.Lines[i]
		if line.CountedQty == nil {
			return domain.CountSession{}, fmt.Errorf("%w: line %s not counted", domain.ErrInvariant, line.ID)
		}
		delta := *line.CountedQty - line.SystemQty
		if delta != 0 {
			loc := line.LocationID
			if loc == nil {
				loc = s.LocationID
			}
			_, _, err := d.Adjust(ctx, AdjustStockInput{
				TenantID: in.TenantID, WarehouseID: s.WarehouseID,
				VariantID: line.VariantID, SKUCode: line.SKUCode, LocationID: loc,
				Delta: delta, IdempotencyKey: in.IdempotencyKey + ":line:" + line.ID.String(),
				ActorID: in.ActorID, Reason: "count variance",
			})
			if err != nil {
				return domain.CountSession{}, err
			}
		}
		approved := true
		line.Approved = &approved
	}
	if err := s.TransitionTo(domain.CountSessionStatusApproved, in.ActorID); err != nil {
		return domain.CountSession{}, err
	}
	if err := d.Counts.Update(ctx, s); err != nil {
		return domain.CountSession{}, err
	}
	d.publishEvent(ctx, domain.EventStockCountCompleted, in.TenantID, s.WarehouseID, uuid.Nil, map[string]any{
		"sessionId": s.ID,
	})
	d.idemPut(ctx, "count:"+in.IdempotencyKey, s)
	return s, nil
}

// GetCount returns a count session.
func (d *Deps) GetCount(ctx context.Context, tenantID, id uuid.UUID) (domain.CountSession, error) {
	return d.Counts.GetByID(ctx, tenantID, id)
}
