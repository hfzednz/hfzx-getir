package domain

import (
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
)

// BatchStatus is the settlement batch lifecycle.
type BatchStatus string

const (
	BatchStatusDraft           BatchStatus = "draft"
	BatchStatusPendingApproval BatchStatus = "pending_approval"
	BatchStatusApproved        BatchStatus = "approved"
	BatchStatusPaying          BatchStatus = "paying"
	BatchStatusCompleted       BatchStatus = "completed"
	BatchStatusFailed          BatchStatus = "failed"
)

// Valid reports whether the batch status is recognized.
func (s BatchStatus) Valid() bool {
	switch s {
	case BatchStatusDraft, BatchStatusPendingApproval, BatchStatusApproved,
		BatchStatusPaying, BatchStatusCompleted, BatchStatusFailed:
		return true
	default:
		return false
	}
}

// PayeeType classifies settlement line recipients.
type PayeeType string

const (
	PayeeCourier  PayeeType = "courier"
	PayeeSupplier PayeeType = "supplier"
	PayeeMerchant PayeeType = "merchant"
	PayeePartner  PayeeType = "partner"
)

// Valid reports whether the payee type is recognized.
func (t PayeeType) Valid() bool {
	switch t {
	case PayeeCourier, PayeeSupplier, PayeeMerchant, PayeePartner:
		return true
	default:
		return false
	}
}

// SettlementLine is a payable line in a batch.
type SettlementLine struct {
	ID           uuid.UUID
	PayeeType    PayeeType
	PayeeRef     string // opaque payee id
	AmountMinor  int64
	Currency     string
	ExternalRef  string // opaque business ref
	Memo         string
}

// Validate checks line invariants.
func (l SettlementLine) Validate() error {
	if !l.PayeeType.Valid() {
		return fmt.Errorf("%w: invalid payee_type %q", ErrInvalidArgument, l.PayeeType)
	}
	if strings.TrimSpace(l.PayeeRef) == "" {
		return fmt.Errorf("%w: payee_ref required", ErrInvalidArgument)
	}
	if l.AmountMinor <= 0 {
		return fmt.Errorf("%w: amount must be positive", ErrInvalidArgument)
	}
	if _, err := NewMoney(l.AmountMinor, l.Currency); err != nil {
		return err
	}
	return nil
}

// SettlementBatch aggregates payable lines for dual-control approval and payout.
type SettlementBatch struct {
	ID             uuid.UUID
	TenantID       uuid.UUID
	Status         BatchStatus
	Currency       string
	PeriodStart    time.Time
	PeriodEnd      time.Time
	Description    string
	IdempotencyKey string
	Lines          []SettlementLine
	TotalMinor     int64
	SubmittedBy    *uuid.UUID
	SubmittedAt    *time.Time
	ApprovedBy     *uuid.UUID
	ApprovedAt     *time.Time
	CompletedAt    *time.Time
	FailedAt       *time.Time
	FailureReason  string
	CreatedAt      time.Time
	UpdatedAt      time.Time
	Version        int64
}

// Validate checks batch structural invariants.
func (b SettlementBatch) Validate() error {
	if b.ID == uuid.Nil {
		return fmt.Errorf("%w: batch id required", ErrInvalidArgument)
	}
	if b.TenantID == uuid.Nil {
		return fmt.Errorf("%w: tenant_id required", ErrInvalidArgument)
	}
	if !b.Status.Valid() {
		return fmt.Errorf("%w: invalid status %q", ErrInvalidArgument, b.Status)
	}
	if _, err := NewMoney(0, b.Currency); err != nil {
		return err
	}
	for i, line := range b.Lines {
		if err := line.Validate(); err != nil {
			return fmt.Errorf("%w: line[%d]", err, i)
		}
		if !strings.EqualFold(line.Currency, b.Currency) {
			return fmt.Errorf("%w: line[%d]", ErrCurrencyMismatch, i)
		}
	}
	return nil
}

// RecalcTotal sets TotalMinor from lines.
func (b *SettlementBatch) RecalcTotal() {
	var sum int64
	for _, l := range b.Lines {
		sum += l.AmountMinor
	}
	b.TotalMinor = sum
}

// ValidateTransition checks allowed status transitions.
func ValidateTransition(from, to BatchStatus) error {
	if from == to {
		return nil
	}
	ok := map[BatchStatus][]BatchStatus{
		BatchStatusDraft:           {BatchStatusPendingApproval},
		BatchStatusPendingApproval: {BatchStatusApproved, BatchStatusDraft},
		BatchStatusApproved:        {BatchStatusPaying},
		BatchStatusPaying:          {BatchStatusCompleted, BatchStatusFailed},
	}
	allowed, exists := ok[from]
	if !exists {
		return fmt.Errorf("%w: %s -> %s", ErrInvalidTransition, from, to)
	}
	for _, a := range allowed {
		if a == to {
			return nil
		}
	}
	return fmt.Errorf("%w: %s -> %s", ErrInvalidTransition, from, to)
}

// PayoutInstruction is a bank/PSP payout request generated from a line.
type PayoutInstruction struct {
	ID          uuid.UUID
	BatchID     uuid.UUID
	LineID      uuid.UUID
	TenantID    uuid.UUID
	PayeeType   PayeeType
	PayeeRef    string
	AmountMinor int64
	Currency    string
	Status      string // pending|sent|succeeded|failed
	ProviderRef string
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

// Reconciliation compares batch totals vs provider report.
type Reconciliation struct {
	ID              uuid.UUID
	TenantID        uuid.UUID
	BatchID         uuid.UUID
	ProviderRef     string
	ExpectedMinor   int64
	ReportedMinor   int64
	Matched         bool
	CreatedAt       time.Time
}

// Mismatch records a reconciliation discrepancy.
type Mismatch struct {
	ID            uuid.UUID
	TenantID      uuid.UUID
	BatchID       uuid.UUID
	ReconcileID   uuid.UUID
	ExpectedMinor int64
	ReportedMinor int64
	DeltaMinor    int64
	Detail        string
	CreatedAt     time.Time
}
