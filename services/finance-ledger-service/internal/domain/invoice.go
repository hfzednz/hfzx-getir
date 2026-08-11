package domain

import (
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
)

// InvoiceStatus is the invoice lifecycle.
type InvoiceStatus string

const (
	InvoiceStatusDraft    InvoiceStatus = "draft"
	InvoiceStatusIssued   InvoiceStatus = "issued"
	InvoiceStatusPaid     InvoiceStatus = "paid"
	InvoiceStatusVoid     InvoiceStatus = "void"
	InvoiceStatusCredited InvoiceStatus = "credited"
)

// Valid reports whether the invoice status is recognized.
func (s InvoiceStatus) Valid() bool {
	switch s {
	case InvoiceStatusDraft, InvoiceStatusIssued, InvoiceStatusPaid, InvoiceStatusVoid, InvoiceStatusCredited:
		return true
	default:
		return false
	}
}

// InvoiceLine is a single invoice line in minor units.
type InvoiceLine struct {
	ID          uuid.UUID
	Description string
	Qty         int64
	UnitMinor   int64
	TaxMinor    int64
	TotalMinor  int64
	TaxCode     string
}

// Validate checks invoice line invariants.
func (l InvoiceLine) Validate() error {
	if strings.TrimSpace(l.Description) == "" {
		return fmt.Errorf("%w: line description required", ErrInvalidArgument)
	}
	if l.Qty <= 0 {
		return fmt.Errorf("%w: qty must be positive", ErrInvalidArgument)
	}
	if l.UnitMinor < 0 || l.TaxMinor < 0 || l.TotalMinor < 0 {
		return fmt.Errorf("%w: line amounts", ErrNegativeMoney)
	}
	return nil
}

// Invoice is billing metadata (no cart/order aggregate).
type Invoice struct {
	ID             uuid.UUID
	TenantID       uuid.UUID
	Status         InvoiceStatus
	Currency       string
	CounterpartyRef string // opaque party ref
	ExternalRef    string // opaque business ref
	IdempotencyKey string
	SubtotalMinor  int64
	TaxMinor       int64
	TotalMinor     int64
	Lines          []InvoiceLine
	IssuedAt       *time.Time
	CreatedAt      time.Time
	UpdatedAt      time.Time
	Version        int64
}

// Validate checks invoice invariants.
func (inv Invoice) Validate() error {
	if inv.ID == uuid.Nil {
		return fmt.Errorf("%w: invoice id required", ErrInvalidArgument)
	}
	if inv.TenantID == uuid.Nil {
		return fmt.Errorf("%w: tenant_id required", ErrInvalidArgument)
	}
	if !inv.Status.Valid() {
		return fmt.Errorf("%w: invalid invoice status %q", ErrInvalidArgument, inv.Status)
	}
	if _, err := NewMoney(0, inv.Currency); err != nil {
		return err
	}
	if len(inv.Lines) == 0 {
		return fmt.Errorf("%w: invoice requires lines", ErrInvalidArgument)
	}
	for i, line := range inv.Lines {
		if err := line.Validate(); err != nil {
			return fmt.Errorf("%w: line[%d]", err, i)
		}
	}
	if inv.SubtotalMinor < 0 || inv.TaxMinor < 0 || inv.TotalMinor < 0 {
		return fmt.Errorf("%w: totals", ErrNegativeMoney)
	}
	return nil
}

// CreditNote reverses (part of) an issued invoice.
type CreditNote struct {
	ID             uuid.UUID
	TenantID       uuid.UUID
	InvoiceID      uuid.UUID
	Currency       string
	AmountMinor    int64
	Reason         string
	IdempotencyKey string
	CreatedAt      time.Time
}

// Validate checks credit note invariants.
func (c CreditNote) Validate() error {
	if c.ID == uuid.Nil {
		return fmt.Errorf("%w: credit note id required", ErrInvalidArgument)
	}
	if c.TenantID == uuid.Nil {
		return fmt.Errorf("%w: tenant_id required", ErrInvalidArgument)
	}
	if c.InvoiceID == uuid.Nil {
		return fmt.Errorf("%w: invoice_id required", ErrInvalidArgument)
	}
	if c.AmountMinor <= 0 {
		return fmt.Errorf("%w: amount must be positive", ErrInvalidArgument)
	}
	if _, err := NewMoney(c.AmountMinor, c.Currency); err != nil {
		return err
	}
	return nil
}
